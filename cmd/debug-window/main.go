package main

import (
	"context"
	"image"
	"log"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/g-wilson/led/clock"
	"github.com/g-wilson/led/config"
	"github.com/g-wilson/led/internal/calendar"
	"github.com/g-wilson/led/internal/diagnostics"
	"github.com/g-wilson/led/internal/hasensors"
	"github.com/g-wilson/led/internal/homeassistant"
	"github.com/g-wilson/led/internal/tomorrowio"
	"github.com/g-wilson/led/internal/weather"
	"github.com/g-wilson/led/internal/framestreamer"
	"github.com/g-wilson/led/internal/windowrenderer"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	// Lock the main goroutine to the OS thread (required for GLFW on macOS)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize GLFW (must be on main thread on macOS)
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	fs := framestreamer.New(framestreamer.Params{
		Bounds: image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: cfg.LEDCols, Y: cfg.LEDRows},
		},
		FrametimeMs: framestreamer.OneFPS,
		Renderer:    clockApp,
	})

	renderer, err := windowrenderer.New("LED Matrix Debug", cfg.LEDRows, cfg.LEDCols, fs.C, fs.E)
	if err != nil {
		log.Fatalln("failed to create window renderer:", err)
	}
	defer renderer.Cleanup()

	go fs.Start()
	defer fs.Stop()

	if err := renderer.Run(); err != nil {
		log.Fatalln("renderer error:", err)
	}
}
