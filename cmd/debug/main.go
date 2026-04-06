package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/g-wilson/led/clock"
	"github.com/g-wilson/led/config"
	"github.com/g-wilson/led/internal/calendar"
	"github.com/g-wilson/led/internal/diagnostics"
	"github.com/g-wilson/led/internal/hasensors"
	"github.com/g-wilson/led/internal/homeassistant"
	"github.com/g-wilson/led/internal/tomorrowio"
	"github.com/g-wilson/led/internal/weather"
)

func main() {
	useMock := flag.Bool("mock", false, "use fake data instead of real API clients")
	flag.Parse()

	ctx := context.Background()

	var (
		clockApp *clock.ClockRenderer
		bounds   image.Rectangle
	)

	if *useMock {
		cfg := &config.Settings{
			Timezone: "Europe/London",
			Debug:    true,
			LEDRows:  32,
			LEDCols:  64,
		}
		bounds = image.Rectangle{Max: image.Point{X: cfg.LEDCols, Y: cfg.LEDRows}}

		calAgent, err := calendar.New(fakeCalendarLoader{})
		if err != nil {
			log.Fatalln(err)
		}
		weatherAgent, err := weather.New(ctx, fakeWeatherProvider{}, weather.AgentOptions{
			Refresh:   3600,
			Latitude:  "51.5",
			Longitude: "-0.1",
		})
		if err != nil {
			log.Fatalln(err)
		}
		diagAgent, err := diagnostics.New(ctx, fakePinger{})
		if err != nil {
			log.Fatalln(err)
		}
		clockApp, err = clock.New(ctx, cfg, weatherAgent, diagAgent, nil, calAgent)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalln(err)
		}
		bounds = image.Rectangle{Max: image.Point{X: cfg.LEDCols, Y: cfg.LEDRows}}

		calLoader := calendar.NewFileLoader(cfg.CalendarFiles)
		calAgent, err := calendar.New(calLoader)
		if err != nil {
			log.Fatalln(err)
		}
		tioClient := tomorrowio.New(cfg.TomorrowIOAPIKey, nil)
		weatherAgent, err := weather.New(ctx, tioClient, weather.AgentOptions{
			Refresh:   cfg.WeatherRefresh,
			Latitude:  cfg.WeatherLatitude,
			Longitude: cfg.WeatherLongitude,
		})
		if err != nil {
			log.Fatalln(err)
		}
		diagAgent, err := diagnostics.New(ctx, diagnostics.NetPinger{})
		if err != nil {
			log.Fatalln(err)
		}
		var sensorsAgent *hasensors.Agent
		if cfg.HAURL != "" && cfg.HAToken != "" && len(cfg.HASensors) > 0 {
			haClient := homeassistant.New(cfg.HAURL, cfg.HAToken, nil)
			sensorsAgent, err = hasensors.New(ctx, haClient, cfg.HASensors)
			if err != nil {
				log.Printf("sensors agent unavailable, skipping area pages: %v", err)
			}
		}
		clockApp, err = clock.New(ctx, cfg, weatherAgent, diagAgent, sensorsAgent, calAgent)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err := os.MkdirAll("frames", 0755); err != nil {
		log.Fatalln(err)
	}

	for _, name := range clockApp.PageNames() {
		buf := image.NewRGBA(bounds)
		if err := clockApp.CaptureFrame(name, buf); err != nil {
			log.Fatalf("error capturing page %q: %v", name, err)
		}
		path := fmt.Sprintf("frames/page-%s.png", name)
		if err := saveImage(path, buf); err != nil {
			log.Fatalf("error saving page %q: %v", name, err)
		}
		log.Printf("saved %s", path)
	}
}

func saveImage(filename string, img *image.RGBA) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// --- mock implementations ---

type fakePinger struct{}

func (fakePinger) Ping(_ context.Context, _ string) (time.Duration, error) {
	return 45 * time.Millisecond, nil
}

type fakeWeatherProvider struct{}

func (fakeWeatherProvider) GetTwoDayWeatherAtLocation(_ context.Context, _, _ string) (weather.TwoDayWeather, error) {
	base := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	return weather.TwoDayWeather{
		Today: weather.DayWeather{
			TemperatureHigh: 18,
			TemperatureLow:  9,
			SunriseTime:     base.Add(5*time.Hour + 48*time.Minute),
			SunsetTime:      base.Add(19*time.Hour + 12*time.Minute),
			MoonriseTime:    base.Add(21*time.Hour + 30*time.Minute),
			MoonsetTime:     base.Add(7*time.Hour + 15*time.Minute),
			Windy:           true,
		},
		Tomorrow: weather.DayWeather{
			TemperatureHigh: 13,
			TemperatureLow:  7,
			SunriseTime:     base.Add(24*time.Hour + 5*time.Hour + 46*time.Minute),
			SunsetTime:      base.Add(24*time.Hour + 19*time.Hour + 14*time.Minute),
			Cloudy:          true,
			Rainy:           true,
		},
	}, nil
}

type fakeCalendarLoader struct{}

func (fakeCalendarLoader) Load() ([]calendar.EventSource, error) {
	return []calendar.EventSource{{
		Data: []byte(`events:
  - name: "Monaco GP"
    time: "2026-05-24T14:00:00Z"
    image: "builtin:f1"
`),
	}}, nil
}
