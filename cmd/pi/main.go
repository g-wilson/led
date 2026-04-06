package main

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os/signal"
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

	rgbmatrix "github.com/mcuadros/go-rpi-rgb-led-matrix"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	matrixConfig := &rgbmatrix.DefaultConfig
	matrixConfig.Rows = cfg.LEDRows
	matrixConfig.Cols = cfg.LEDCols
	matrixConfig.PWMBits = cfg.LEDPWMBits
	matrixConfig.PWMLSBNanoseconds = cfg.LEDPWMLSBNano
	matrixConfig.Brightness = cfg.LEDBrightness
	matrixConfig.HardwareMapping = cfg.LEDHardware

	m, err := rgbmatrix.NewRGBLedMatrix(matrixConfig)
	if err != nil {
		log.Fatalln(err)
	}

	c := rgbmatrix.NewCanvas(m)
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

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
		Bounds:      c.Bounds(),
		FrametimeMs: framestreamer.OneFPS,
		Renderer:    clockApp,
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-fs.E:
				if !ok {
					return
				}
				log.Fatalln(err)
			case frame, ok := <-fs.C:
				if !ok {
					return
				}
				draw.Draw(c, c.Bounds(), frame, image.Point{}, draw.Src)
				c.Render()
			}
		}
	}()
	go fs.Start()

	<-ctx.Done()
	fs.Stop()
}
