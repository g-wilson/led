package main

import (
	"image"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/g-wilson/led/clock"
	"github.com/g-wilson/led/internal/framestreamer"
	"github.com/g-wilson/led/internal/windowrenderer"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// Lock the main goroutine to the OS thread (required for GLFW on macOS)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rows, _ := strconv.ParseInt(os.Getenv("LED_ROWS"), 10, 32)
	cols, _ := strconv.ParseInt(os.Getenv("LED_COLS"), 10, 32)
	ledRows := int(rows)
	ledCols := int(cols)

	// Initialize GLFW (must be on main thread on macOS)
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	// Create window renderer
	renderer, err := windowrenderer.New(ledRows, ledCols)
	if err != nil {
		log.Fatalln("failed to create window renderer:", err)
	}
	defer renderer.Cleanup()

	// Create clock renderer
	clockApp, err := clock.New()
	if err != nil {
		log.Fatalln(err)
	}

	// Create framestreamer
	fs := framestreamer.New(framestreamer.Params{
		Bounds: image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: ledCols, Y: ledRows},
		},
		FrametimeMs: framestreamer.OneFPS,
		Renderer:    clockApp,
	})

	// Start framestreamer
	go fs.Start()
	defer fs.Stop()

	// Frame receiver goroutine - sends frames to renderer for main thread processing
	go func() {
		for {
			select {
			case err := <-fs.E:
				if err != nil {
					log.Fatalln("framestreamer error:", err)
				}
			case frame := <-fs.C:
				if frame != nil {
					renderer.SendFrame(frame)
				}
			}
		}
	}()

	// Main render loop - processes frames on main thread
	renderer.Run()
}
