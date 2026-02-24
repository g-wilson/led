package clock

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/g-wilson/led/internal/calendar"
	"github.com/g-wilson/led/internal/diagnostics"
	"github.com/g-wilson/led/internal/hasensors"
	"github.com/g-wilson/led/internal/homeassistant"
	"github.com/g-wilson/led/internal/tomorrowio"
	"github.com/g-wilson/led/internal/weather"

	"github.com/joho/godotenv"
	"github.com/toelsiba/fopix"
	"golang.org/x/image/draw"
)

//go:embed fonts/tom-thumb-new.json
var fontSource []byte

type page func(c *image.RGBA) error

type ClockRenderer struct {
	font         *fopix.Drawer
	weather      *weather.Agent
	diagnostics  *diagnostics.Agent
	sensors      *hasensors.Agent
	location     *time.Location
	pages        []page
	currentPage  int
	pageInterval time.Duration
	debug        bool
}

func New() (*ClockRenderer, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	fontInfo := fopix.FontInfo{}
	err = json.Unmarshal(fontSource, &fontInfo)
	if err != nil {
		return nil, fmt.Errorf("error loading font info file: %w", err)
	}
	font, err := fopix.NewDrawer(fontInfo)
	if err != nil {
		return nil, fmt.Errorf("error creating font drawer: %w", err)
	}
	font.SetScale(1)

	tomorrowIoAPIKey := os.Getenv("TOMORROWIO_API_KEY")
	if len(tomorrowIoAPIKey) == 0 {
		return nil, errors.New("environment variable TOMORROWIO_API_KEY is required")
	}

	tomorrowIoClient := tomorrowio.New(tomorrowIoAPIKey, nil)
	weatherRefresh, _ := strconv.ParseInt(os.Getenv("WEATHER_REFRESH"), 10, 32)
	weatherAgent, err := weather.New(tomorrowIoClient, weather.AgentOptions{
		Refresh:   int(weatherRefresh),
		Latitude:  os.Getenv("WEATHER_LATITUDE"),
		Longitude: os.Getenv("WEATHER_LONGITUDE"),
	})
	if err != nil {
		return nil, fmt.Errorf("error initiating weather agent: %w", err)
	}

	diagAgent, err := diagnostics.New()
	if err != nil {
		return nil, fmt.Errorf("error initiating diagnostics agent: %w", err)
	}

	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("cannot determine timezone: %w", err)
	}

	r := &ClockRenderer{
		font:         font,
		weather:      weatherAgent,
		diagnostics:  diagAgent,
		location:     location,
		currentPage:  0,
		pageInterval: 5 * time.Second,
		debug:        os.Getenv("DEBUG") == "true",
	}

	// Phase 1: static pages
	r.pages = []page{
		r.renderToday,
		r.renderTomorrow,
		r.renderDaylight,
		r.renderCountdown,
		r.renderDiag,
	}

	// Phase 2: dynamic area pages (skipped entirely if HA env vars not set)
	haURL := os.Getenv("HA_URL")
	haToken := os.Getenv("HA_TOKEN")
	haEntityIDs := strings.Split(os.Getenv("HA_SENSORS"), ",")

	if haURL != "" && haToken != "" && len(haEntityIDs) > 0 {
		haClient := homeassistant.New(haURL, haToken, nil)
		sensorsAgent, err := hasensors.New(haClient, haEntityIDs)
		if err != nil {
			log.Printf("sensors agent unavailable, skipping area pages: %v", err)
		} else {
			r.sensors = sensorsAgent
			for _, areaName := range sensorsAgent.GetAreas() {
				areaName := areaName
				r.pages = append(r.pages, func(c *image.RGBA) error {
					return r.renderArea(c, areaName)
				})
			}
		}
	}

	r.startPageIterator()

	return r, nil
}

// DrawFrame renders the current clock display into the provided target buffer.
// The buffer is expected to be pre-cleared by the caller (FrameStreamer).
func (r *ClockRenderer) DrawFrame(c *image.RGBA) error {
	// clear the image to black as a background for the page
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// if it's overnight, don't render anything
	if r.isCurrentlyOvernight() && !r.debug {
		return nil
	}

	// all pages - clock
	r.addText(c, image.Point{X: 0, Y: -1}, r.getTimeString(), color.RGBA{200, 200, 200, 255})

	return r.pages[r.currentPage](c)
}

func (r *ClockRenderer) renderToday(c *image.RGBA) error {
	w := r.weather.GetToday()
	r.addText(c, image.Point{X: 0, Y: 8}, "Today", color.RGBA{215, 0, 88, 255})
	r.renderWeather(c, w)
	return nil
}

func (r *ClockRenderer) renderTomorrow(c *image.RGBA) error {
	w := r.weather.GetTomorrow()
	r.addText(c, image.Point{X: 0, Y: 8}, "Tomorrow", color.RGBA{215, 0, 88, 255})
	r.renderWeather(c, w)
	return nil
}

func (r *ClockRenderer) renderDaylight(c *image.RGBA) error {
	w := r.weather.GetToday()
	sunrise := w.SunriseTime.UTC().In(r.location).Format("15:04")
	sunset := w.SunsetTime.UTC().In(r.location).Format("15:04")
	r.addText(c, image.Point{X: 4, Y: 10}, fmt.Sprintf("Sunrise %s", sunrise), color.RGBA{152, 168, 27, 255})
	r.addText(c, image.Point{X: 8, Y: 18}, fmt.Sprintf("Sunset %s", sunset), color.RGBA{194, 27, 27, 255})
	return nil
}

func (r *ClockRenderer) renderCountdown(c *image.RGBA) error {
	if event := calendar.GetNextEvent(); event != nil {
		if event.Image != nil {
			draw.Draw(c, c.Bounds(), event.Image, image.Point{X: -44, Y: -9}, draw.Over)
		}
		halfway := 32 - int(float64((4*len(event.Name))/2))
		r.addText(c, image.Point{X: halfway, Y: 15}, event.Name, color.RGBA{60, 60, 215, 255})
		r.addText(c, image.Point{X: 10, Y: 22}, formatDuration(event.Until()), color.RGBA{215, 0, 0, 255})
	}
	return nil
}

func (r *ClockRenderer) renderDiag(c *image.RGBA) error {
	status := r.diagnostics.GetStatus()
	sinceText, sinceColor := diagSinceText(status)
	pingText, pingColor := diagPingText(status)
	r.addText(c, image.Point{X: 0, Y: 10}, sinceText, sinceColor)
	r.addText(c, image.Point{X: 0, Y: 18}, pingText, pingColor)
	return nil
}

func (r *ClockRenderer) renderArea(c *image.RGBA, area string) error {
	r.addText(c, image.Point{X: 0, Y: 8}, area, color.RGBA{215, 0, 88, 255})

	if as, ok := r.sensors.GetArea(area); ok {
		for i, s := range as.Sensors {
			y := 16 + (i * 8)
			r.addText(c, image.Point{X: 0, Y: y}, fmt.Sprintf("%s %s%s", s.Name, s.State, s.Unit), color.RGBA{200, 200, 200, 255})
		}
	}

	return nil
}

func (r *ClockRenderer) addText(c *image.RGBA, pos image.Point, text string, col color.RGBA) {
	r.font.SetColor(col)
	r.font.DrawText(c, pos, text)
}

func (r *ClockRenderer) getTimeString() string {
	return time.Now().UTC().In(r.location).Format("15:04 Mon Jan 2")
}

// startPageIterator kicks off a goroutine ticking continuously through
// the length of the pages array, updating the current page each time
func (r *ClockRenderer) startPageIterator() {
	go func() {
		i := 0
		ticker := time.NewTicker(r.pageInterval)
		for range ticker.C {
			i += 1
			if i > (len(r.pages) - 1) {
				i = 0
			}

			r.currentPage = i
		}
	}()
}

func (r *ClockRenderer) renderWeather(c *image.RGBA, w weather.DayWeather) {
	yOffset := 15
	summaryStart := 36

	r.addText(c, image.Point{X: 0, Y: yOffset}, fmt.Sprintf("%02.foC", w.TemperatureLow), color.RGBA{80, 80, 255, 255})
	r.addText(c, image.Point{X: 17, Y: yOffset}, fmt.Sprintf("%02.foC", w.TemperatureHigh), color.RGBA{255, 150, 0, 255})

	// Underneath temperatures, always shows
	if w.Cloudy {
		r.addText(c, image.Point{X: 0, Y: yOffset + 7}, "Cloudy", color.RGBA{179, 161, 136, 255})
	} else {
		r.addText(c, image.Point{X: 0, Y: yOffset + 7}, "Sunny", color.RGBA{255, 213, 0, 255})
	}

	// To the right, conditionally shows
	if w.Snowy {
		r.addText(c, image.Point{X: summaryStart, Y: yOffset}, "Snow", color.RGBA{255, 255, 255, 255})
	} else if w.Rainy {
		r.addText(c, image.Point{X: summaryStart, Y: yOffset}, "Rain", color.RGBA{0, 113, 237, 255})
	}
	if w.Windy {
		r.addText(c, image.Point{X: summaryStart, Y: yOffset + 7}, "Windy", color.RGBA{0, 247, 255, 255})
	}

	// Humidity, hidden for now
	// r.addText(c, image.Point{X: summaryStart + 15, Y: yOffset}, fmt.Sprintf("H%02.f", (w.Humidity*100)), color.RGBA{230, 77, 0, 255})
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

func (r *ClockRenderer) isCurrentlyOvernight() bool {
	now := time.Now().In(r.location)
	year, month, day := now.Date()
	today8pm := time.Date(year, month, day, 20, 0, 0, 0, r.location)
	today6am := time.Date(year, month, day, 6, 0, 0, 0, r.location)

	return today8pm.Before(now) || today6am.After(now)
}

var (
	diagGreen  = color.RGBA{0, 200, 0, 255}
	diagYellow = color.RGBA{200, 200, 0, 255}
	diagOrange = color.RGBA{255, 140, 0, 255}
	diagRed    = color.RGBA{200, 0, 0, 255}
)

func diagSinceText(status diagnostics.Status) (string, color.RGBA) {
	if status.LastHealthyAt.IsZero() {
		return "Last ok never", diagRed
	}

	since := time.Since(status.LastHealthyAt)
	sinceText := fmt.Sprintf("Last ok %s", formatShortDuration(since))
	if status.IsStale(time.Now()) {
		return sinceText, diagRed
	}

	return sinceText, diagGreen
}

func diagPingText(status diagnostics.Status) (string, color.RGBA) {
	if !status.LastPingOk {
		return "Ping n/a", diagRed
	}

	pingText := fmt.Sprintf("Ping %dms", status.LastPing.Milliseconds())
	level := status.PingLevel()
	return pingText, diagPingColor(level)
}

func diagPingColor(level diagnostics.PingLevel) color.RGBA {
	switch level {
	case diagnostics.PingLevelGreen:
		return diagGreen
	case diagnostics.PingLevelYellow:
		return diagYellow
	case diagnostics.PingLevelOrange:
		return diagOrange
	default:
		return diagRed
	}
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
