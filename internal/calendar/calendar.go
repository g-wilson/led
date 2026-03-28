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

var builtinImages map[string]image.Image

var sortedEvents eventList

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

func (s eventList) Len() int {
	return len(s)
}

func (s eventList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s eventList) Less(i, j int) bool {
	return s[i].StartsAt.Before(s[j].StartsAt)
}

// Load initialises the calendar. It parses the embedded default event files
// and any additional YAML files provided in the files slice.
func Load(files []string) error {
	f1Img, _, err := image.Decode(bytes.NewReader(f1ImageSource))
	if err != nil {
		return fmt.Errorf("calendar: failed to decode builtin f1 image: %w", err)
	}

	xmasImg, _, err := image.Decode(bytes.NewReader(xmasImageSource))
	if err != nil {
		return fmt.Errorf("calendar: failed to decode builtin xmastree image: %w", err)
	}

	builtinImages = map[string]image.Image{
		"f1":       f1Img,
		"xmastree": xmasImg,
	}

	defaults, err := loadYAMLBytes(calendars.EventsYAML, "")
	if err != nil {
		return fmt.Errorf("calendar: failed to parse embedded events.yaml: %w", err)
	}

	all := defaults

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("calendar: skipping file %q: %v", path, err)
			continue
		}

		events, err := loadYAMLBytes(data, filepath.Dir(path))
		if err != nil {
			log.Printf("calendar: skipping file %q: invalid YAML: %v", path, err)
			continue
		}

		all = append(all, events...)
	}

	sort.Sort(eventList(all))
	sortedEvents = all

	return nil
}

func GetNextEvent() *Event {
	for i := range sortedEvents {
		if sortedEvents[i].Until().Seconds() >= 0 {
			return &sortedEvents[i]
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

func loadYAMLBytes(data []byte, dir string) (eventList, error) {
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
			Image:     resolveImage(ye.Image, dir),
		})
	}

	return result, nil
}

func resolveImage(ref string, dir string) image.Image {
	if ref == "" {
		return nil
	}

	if strings.HasPrefix(ref, "builtin:") {
		alias := strings.TrimPrefix(ref, "builtin:")
		img, ok := builtinImages[alias]
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

