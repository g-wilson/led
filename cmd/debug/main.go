package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	bounds := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: cfg.LEDCols, Y: cfg.LEDRows},
	}

	ctx := context.Background()

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

	clockApp, err := clock.New(ctx, cfg, weatherAgent, diagAgent, sensorsAgent, calAgent)
	if err != nil {
		log.Fatalln(err)
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
