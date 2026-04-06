package clock

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"sync/atomic"
	"time"

	"github.com/g-wilson/led/config"
	"github.com/g-wilson/led/internal/calendar"
	"github.com/g-wilson/led/internal/diagnostics"
	"github.com/g-wilson/led/internal/hasensors"
	"github.com/g-wilson/led/internal/weather"

	"github.com/toelsiba/fopix"
	"golang.org/x/image/draw"
)

//go:embed fonts/tom-thumb-new.json
var fontSource []byte

type page struct {
	name string
	fn   func(c *image.RGBA) error
}

type ClockRenderer struct {
	font        *fopix.Drawer
	weather     *weather.Agent
	diagnostics *diagnostics.Agent
	sensors     *hasensors.Agent
	calendar    *calendar.Agent
	location    *time.Location
	pages       []page
	currentPage atomic.Int32
	pageInterval time.Duration
	debug       bool
}

// New creates a ClockRenderer. All agent dependencies must be fully initialised
// before being passed in; New does no network I/O itself.
//
// sensors may be nil, in which case area pages are omitted.
func New(
	ctx context.Context,
	cfg *config.Settings,
	weatherAgent *weather.Agent,
	diagAgent *diagnostics.Agent,
	sensorsAgent *hasensors.Agent,
	calendarAgent *calendar.Agent,
) (*ClockRenderer, error) {
	fontInfo := fopix.FontInfo{}
	if err := json.Unmarshal(fontSource, &fontInfo); err != nil {
		return nil, fmt.Errorf("error loading font info file: %w", err)
	}
	font, err := fopix.NewDrawer(fontInfo)
	if err != nil {
		return nil, fmt.Errorf("error creating font drawer: %w", err)
	}
	font.SetScale(1)

	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("cannot determine timezone: %w", err)
	}

	r := &ClockRenderer{
		font:         font,
		weather:      weatherAgent,
		diagnostics:  diagAgent,
		sensors:      sensorsAgent,
		calendar:     calendarAgent,
		location:     location,
		pageInterval: 5 * time.Second,
		debug:        cfg.Debug,
	}

	// Phase 1: static pages
	r.pages = []page{
		{name: "today", fn: r.renderToday},
		{name: "tomorrow", fn: r.renderTomorrow},
		{name: "daylight", fn: r.renderDaylight},
		{name: "moon", fn: r.renderMoon},
		{name: "countdown", fn: r.renderCountdown},
		{name: "diag", fn: r.renderDiag},
	}

	// Phase 2: dynamic area pages (only if sensors agent was provided)
	if sensorsAgent != nil {
		for _, areaName := range sensorsAgent.GetAreas() {
			r.pages = append(r.pages, page{
				name: "area-" + areaName,
				fn: func(c *image.RGBA) error {
					return r.renderArea(c, areaName)
				},
			})
		}
	}

	// start the page iterator to continuously tick through pages in the background.
	r.startPageIterator(ctx)

	return r, nil
}

// DrawFrame renders the current clock display into the provided target buffer.
func (r *ClockRenderer) DrawFrame(c *image.RGBA) error {
	// clear the image to black as a background for the page
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// if it's overnight, don't render anything
	if r.isCurrentlyOvernight() && !r.debug {
		return nil
	}

	// all pages - clock
	r.addText(c, image.Point{X: 0, Y: -1}, r.getTimeString(), color.RGBA{200, 200, 200, 255})

	// page content from the current page
	return r.pages[r.currentPage.Load()].fn(c)
}

// startPageIterator kicks off a goroutine ticking continuously through
// the length of the pages array, updating the current page each time
func (r *ClockRenderer) startPageIterator(ctx context.Context) {
	go func() {
		i := int32(0)
		ticker := time.NewTicker(r.pageInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				i++
				if int(i) > len(r.pages)-1 {
					i = 0
				}
				r.currentPage.Store(i)
			}
		}
	}()
}

// PageNames returns the names of all registered pages in display order.
func (r *ClockRenderer) PageNames() []string {
	names := make([]string, len(r.pages))
	for i, p := range r.pages {
		names[i] = p.name
	}
	return names
}

// CaptureFrame renders the named page into the provided buffer and returns the
// result. Unlike DrawFrame, it bypasses overnight blackout and page rotation
// state, making it suitable for generating static snapshots during development.
func (r *ClockRenderer) CaptureFrame(pageName string, c *image.RGBA) error {
	for _, p := range r.pages {
		if p.name == pageName {
			draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			r.addText(c, image.Point{X: 0, Y: -1}, r.getTimeString(), color.RGBA{200, 200, 200, 255})
			return p.fn(c)
		}
	}
	return fmt.Errorf("unknown page: %q", pageName)
}

func (r *ClockRenderer) addText(c *image.RGBA, pos image.Point, text string, col color.RGBA) {
	r.font.SetColor(col)
	r.font.DrawText(c, pos, text)
}

func (r *ClockRenderer) getTimeString() string {
	return time.Now().UTC().In(r.location).Format("15:04 Mon Jan 2")
}

func (r *ClockRenderer) isCurrentlyOvernight() bool {
	now := time.Now().In(r.location)
	year, month, day := now.Date()
	today8pm := time.Date(year, month, day, 20, 0, 0, 0, r.location)
	today6am := time.Date(year, month, day, 6, 0, 0, 0, r.location)

	return today8pm.Before(now) || today6am.After(now)
}

func formatShortDuration(d time.Duration) string {
	if d < 0 {
		return "0m"
	}
	d = d.Round(time.Minute)
	if d < time.Minute {
		return "0m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}

	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func formatDuration(u time.Duration) string {
	u = u.Round(time.Minute)

	// not actually days - 24h periods because that's much easier and honestly who needs daylight savings
	d := u / (time.Hour * 24)
	u -= d * (time.Hour * 24)

	h := u / time.Hour
	u -= h * time.Hour

	m := u / time.Minute
	u -= m * time.Minute

	s := u / time.Second

	// less than one day to go, render more precise countdown
	if d <= 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", h, m, s)
	}

	return fmt.Sprintf("%02dd %02dh %02dm", d, h, m)
}
