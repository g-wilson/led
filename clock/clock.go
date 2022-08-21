package clock

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"strconv"
	"time"

	"github.com/g-wilson/led/pkg/darksky"
	"github.com/g-wilson/led/pkg/weather"

	"github.com/joho/godotenv"
	"github.com/toelsiba/fopix"
	"golang.org/x/image/draw"
)

//go:embed fonts/tom-thumb-new.json
var fontSource []byte

type ClockRenderer struct {
	font         *fopix.Drawer
	weather      WeatherProvider
	location     *time.Location
	pages        []string
	currentPage  int
	pageInterval time.Duration
}

type WeatherProvider interface {
	GetToday() weather.DayWeather
	GetTomorrow() weather.DayWeather
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

	darkskyAPIKey := os.Getenv("DARKSKY_API_KEY")
	if len(darkskyAPIKey) == 0 {
		return nil, errors.New("environment variable DARKSKY_API_KEY is required")
	}

	darkskyClient := darksky.New(darkskyAPIKey, nil)
	weatherRefresh, _ := strconv.ParseInt(os.Getenv("WEATHER_REFRESH"), 10, 32)
	weatherAgent := weather.New(darkskyClient, weather.AgentOptions{
		Refresh:   int(weatherRefresh),
		Latitude:  os.Getenv("WEATHER_LATITUDE"),
		Longitude: os.Getenv("WEATHER_LONGITUDE"),
	})

	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("cannot determine timezone: %w", err)
	}

	r := &ClockRenderer{
		font:         font,
		weather:      weatherAgent,
		location:     location,
		pages:        []string{"today", "tomorrow", "daylight"},
		currentPage:  0,
		pageInterval: 5 * time.Second,
	}

	r.startPageIterator()

	return r, nil
}

func (r *ClockRenderer) DrawFrame(bounds image.Rectangle) (*image.RGBA, error) {
	c := image.NewRGBA(bounds)
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// all pages - clock
	r.addText(c, image.Point{X: 0, Y: -1}, r.getTimeString(), color.RGBA{200, 200, 200, 255})

	// Countdown timer
	// event := getNextEvent()
	// if event != nil {
	// 	r.addText(c, image.Point{X: 0, Y: 26}, event.Name+":"+formatDuration(event.Until()), color.RGBA{255, 20, 20, 255})
	// }

	switch r.pages[r.currentPage] {
	// Page 1: Today's weather
	case "today":
		w := r.weather.GetToday()
		r.addText(c, image.Point{X: 0, Y: 10}, "Today", color.RGBA{215, 0, 88, 255})
		r.renderWeather(c, w)

	// Page 2: Tomorrow's weather
	case "tomorrow":
		w := r.weather.GetTomorrow()
		r.addText(c, image.Point{X: 0, Y: 10}, "Tomorrow", color.RGBA{215, 0, 88, 255})
		r.renderWeather(c, w)

	// Page 3: Sunset + Sunrise
	case "daylight":
		w := r.weather.GetToday()
		sunrise := w.SunriseTime.UTC().In(r.location).Format("15:04")
		sunset := w.SunsetTime.UTC().In(r.location).Format("15:04")
		r.addText(c, image.Point{X: 4, Y: 10}, fmt.Sprintf("Sunrise %s", sunrise), color.RGBA{152, 168, 27, 255})
		r.addText(c, image.Point{X: 8, Y: 18}, fmt.Sprintf("Sunset %s", sunset), color.RGBA{194, 27, 27, 255})

	}

	return c, nil
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
	r.addText(c, image.Point{X: 0, Y: 17}, fmt.Sprintf("%02.foC", w.ApparentTemperatureLow), color.RGBA{80, 80, 255, 255})
	r.addText(c, image.Point{X: 17, Y: 17}, fmt.Sprintf("%02.foC", w.ApparentTemperatureHigh), color.RGBA{255, 150, 0, 255})

	summaryStart := 38
	if w.Cloudy {
		r.addText(c, image.Point{X: summaryStart, Y: 17}, "C", color.RGBA{179, 161, 136, 255})
	} else {
		r.addText(c, image.Point{X: summaryStart, Y: 17}, "S", color.RGBA{255, 213, 0, 255})
	}
	if w.Rainy {
		r.addText(c, image.Point{X: summaryStart + 4, Y: 17}, "R", color.RGBA{0, 113, 237, 255})
	} else {
		r.addText(c, image.Point{X: summaryStart + 4, Y: 17}, "D", color.RGBA{179, 161, 136, 255})
	}
	if w.Windy {
		r.addText(c, image.Point{X: summaryStart + 8, Y: 17}, "W", color.RGBA{0, 247, 255, 255})
	}

	r.addText(c, image.Point{X: summaryStart + 15, Y: 17}, fmt.Sprintf("H%02.f", (w.Humidity*100)), color.RGBA{230, 77, 0, 255})
}
