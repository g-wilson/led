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

	"github.com/g-wilson/led"

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

	refresh, _ := strconv.ParseInt(os.Getenv("LED_REFRESH"), 10, 32)
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
		panic(err)
	}

	c := rgbmatrix.NewCanvas(m)

	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	frames := led.NewFrameChannel(c.Bounds(), int(refresh))

	go func() {
		for frame := range frames {
			draw.Draw(c, c.Bounds(), frame, image.ZP, draw.Src)
			c.Render()
		}
		fmt.Println("Frame channel closed, exiting")
		os.Exit(1)
	}()

	buf := bufio.NewReader(os.Stdin)
	fmt.Println("Press return to exit")
	_, _ = buf.ReadBytes('\n')
	os.Exit(0)
}
