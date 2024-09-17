package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"strconv"

	"github.com/g-wilson/led/clock"
	"github.com/g-wilson/led/internal/framestreamer"

	"github.com/joho/godotenv"
	rgbmatrix "github.com/mcuadros/go-rpi-rgb-led-matrix"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	config := &rgbmatrix.DefaultConfig

	rows, _ := strconv.ParseInt(os.Getenv("LED_ROWS"), 10, 32)
	cols, _ := strconv.ParseInt(os.Getenv("LED_COLS"), 10, 32)
	pwmb, _ := strconv.ParseInt(os.Getenv("LED_PWM_BITS"), 10, 32)
	pwmlsb, _ := strconv.ParseInt(os.Getenv("LED_PWM_LSB"), 10, 32)
	brt, _ := strconv.ParseInt(os.Getenv("LED_BRIGHTNESS"), 10, 32)

	config.Rows = int(rows)
	config.Cols = int(cols)
	config.PWMBits = int(pwmb)
	config.PWMLSBNanoseconds = int(pwmlsb)
	config.Brightness = int(brt)
	config.HardwareMapping = os.Getenv("LED_HARDWARE")

	m, err := rgbmatrix.NewRGBLedMatrix(config)
	if err != nil {
		log.Fatalln(err)
	}

	c := rgbmatrix.NewCanvas(m)
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	clockApp, err := clock.New()
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
			case err := <-fs.E:
				log.Fatalln(err)
			case frame := <-fs.C:
				draw.Draw(c, c.Bounds(), frame, image.Point{}, draw.Src)
				c.Render()
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
