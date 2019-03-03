package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/g-wilson/led"
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

	c := image.NewRGBA(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: int(cols), Y: int(rows)},
	})

	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	frames := led.NewFrameChannel(c.Bounds(), 1000)

	go func() {
		for frame := range frames {
			draw.Draw(c, c.Bounds(), frame, image.ZP, draw.Src)
			saveImage("output.png", c)
		}
		fmt.Println("Frame channel closed, exiting")
		os.Exit(1)
	}()

	buf := bufio.NewReader(os.Stdin)
	fmt.Println("Press return to exit")
	_, _ = buf.ReadBytes('\n')
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
