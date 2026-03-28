package main

import (
	"context"
	"image"
	"image/png"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/g-wilson/led/clock"
	"github.com/g-wilson/led/config"
	"github.com/g-wilson/led/internal/framestreamer"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	clockApp, err := clock.New(ctx, cfg)
	if err != nil {
		log.Fatalln(err)
	}

	fs := framestreamer.New(framestreamer.Params{
		Bounds:      bounds,
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
				saveImage("output.png", frame)
			}
		}
	}()
	go fs.Start()

	<-ctx.Done()
	fs.Stop()
}

func saveImage(filename string, img *image.RGBA) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return err
	}

	return nil
}
