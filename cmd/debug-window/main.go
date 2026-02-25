package main

import (
	"context"
	"image"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

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

	// Create clock renderer
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	clockApp, err := clock.New(ctx)
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

	// Create window renderer with direct channel access to framestreamer
	renderer, err := windowrenderer.New("LED Matrix Debug", ledRows, ledCols, fs.C, fs.E)
	if err != nil {
		log.Fatalln("failed to create window renderer:", err)
	}
	defer renderer.Cleanup()

	// Start framestreamer - calls the clock app to render frames at the given framerate
	go fs.Start()
	defer fs.Stop()

	// Main render loop - processes frames on main thread
	// No intermediate goroutine needed - renderer reads directly from framestreamer channels
	if err := renderer.Run(); err != nil {
		log.Fatalln("renderer error:", err)
	}
}
