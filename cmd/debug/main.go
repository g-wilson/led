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

	clockApp, err := clock.New(context.Background(), cfg)
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
