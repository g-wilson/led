package calendar

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/g-wilson/led/calendars"
	"gopkg.in/yaml.v3"
)

//go:embed images/xmastree.png
var xmasImageSource []byte

//go:embed images/f1.png
var f1ImageSource []byte

// EventSource holds raw YAML bytes and the base directory used to resolve
// relative image paths within that source.
type EventSource struct {
	Data []byte
	Dir  string
}

// Loader is the data-fetching boundary for the calendar agent. It returns the
// set of YAML sources to load events from.
type Loader interface {
	Load() ([]EventSource, error)
}

// FileLoader is the production Loader. It always includes the embedded default
// events and appends any additional YAML files provided at construction time.
type FileLoader struct {
	files []string
}

func NewFileLoader(files []string) *FileLoader {
	return &FileLoader{files: files}
}

func (l *FileLoader) Load() ([]EventSource, error) {
	sources := []EventSource{
		{Data: calendars.EventsYAML, Dir: ""},
	}
	for _, path := range l.files {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("calendar: skipping file %q: %v", path, err)
			continue
		}
		sources = append(sources, EventSource{Data: data, Dir: filepath.Dir(path)})
	}
	return sources, nil
}

// Event represents a single calendar entry.
type Event struct {
	Name      string
	Timestamp string
	StartsAt  time.Time
	Image     image.Image
}

func (e *Event) Until() time.Duration {
	return time.Until(e.StartsAt)
}

type eventList []Event

func (s eventList) Len() int           { return len(s) }
func (s eventList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s eventList) Less(i, j int) bool { return s[i].StartsAt.Before(s[j].StartsAt) }

// Agent holds a sorted list of calendar events loaded at construction time.
type Agent struct {
	events eventList
}

// New creates a calendar Agent by loading and parsing all sources from the
// provided Loader.
func New(loader Loader) (*Agent, error) {
	f1Img, _, err := image.Decode(bytes.NewReader(f1ImageSource))
	if err != nil {
		return nil, fmt.Errorf("calendar: failed to decode builtin f1 image: %w", err)
	}

	xmasImg, _, err := image.Decode(bytes.NewReader(xmasImageSource))
	if err != nil {
		return nil, fmt.Errorf("calendar: failed to decode builtin xmastree image: %w", err)
	}

	builtins := map[string]image.Image{
		"f1":       f1Img,
		"xmastree": xmasImg,
	}

	sources, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("calendar: loader failed: %w", err)
	}

	var all eventList
	for _, src := range sources {
		events, err := parseYAMLBytes(src.Data, src.Dir, builtins)
		if err != nil {
			log.Printf("calendar: skipping source: invalid YAML: %v", err)
			continue
		}
		all = append(all, events...)
	}

	sort.Sort(all)

	return &Agent{events: all}, nil
}

// GetNextEvent returns the next upcoming event, or nil if none remain.
func (a *Agent) GetNextEvent() *Event {
	for i := range a.events {
		if a.events[i].Until().Seconds() >= 0 {
			return &a.events[i]
		}
	}
	return nil
}

type yamlFile struct {
	Events []yamlEvent `yaml:"events"`
}

type yamlEvent struct {
	Name  string `yaml:"name"`
	Time  string `yaml:"time"`
	Image string `yaml:"image"`
}

func parseYAMLBytes(data []byte, dir string, builtins map[string]image.Image) (eventList, error) {
	var f yamlFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}

	result := make(eventList, 0, len(f.Events))
	for _, ye := range f.Events {
		startsAt, err := time.Parse(time.RFC3339, ye.Time)
		if err != nil {
			log.Printf("calendar: skipping event %q: invalid time %q: %v", ye.Name, ye.Time, err)
			continue
		}

		result = append(result, Event{
			Name:      ye.Name,
			Timestamp: ye.Time,
			StartsAt:  startsAt,
			Image:     resolveImage(ye.Image, dir, builtins),
		})
	}

	return result, nil
}

func resolveImage(ref string, dir string, builtins map[string]image.Image) image.Image {
	if ref == "" {
		return nil
	}

	if strings.HasPrefix(ref, "builtin:") {
		alias := strings.TrimPrefix(ref, "builtin:")
		img, ok := builtins[alias]
		if !ok {
			log.Printf("calendar: unknown builtin image %q", alias)
			return nil
		}
		return img
	}

	if !filepath.IsAbs(ref) && dir != "" {
		ref = filepath.Join(dir, ref)
	}

	f, err := os.Open(ref)
	if err != nil {
		log.Printf("calendar: cannot open image %q: %v", ref, err)
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("calendar: cannot decode image %q: %v", ref, err)
		return nil
	}

	return img
}
