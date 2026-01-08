package main

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/g-wilson/led/clock"
	"github.com/g-wilson/led/internal/framestreamer"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	rows, _ := strconv.ParseInt(os.Getenv("LED_ROWS"), 10, 32)
	cols, _ := strconv.ParseInt(os.Getenv("LED_COLS"), 10, 32)

	bounds := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: int(cols), Y: int(rows)},
	}

	clockApp, err := clock.New()
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
			case err := <-fs.E:
				log.Fatalln(err)
			case frame := <-fs.C:
				saveImage("output.png", frame)
			}
		}
	}()
	go fs.Start()

	buf := bufio.NewReader(os.Stdin)
	fmt.Println("Press return to exit")
	_, _ = buf.ReadBytes('\n') // block for user input

	fs.Stop()
	os.Exit(0)
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
