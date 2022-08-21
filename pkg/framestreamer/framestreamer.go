package framestreamer

import (
	"image"
	"time"
)

const (
	OneFPS        = 1000
	TenFPS        = 1000 / 10
	FifteenFPS    = 1000 / 15
	TwentyFourFPS = 1000 / 24
	ThirtyFPS     = 1000 / 30
)

// Renderer delivers single image frames to our stream
type Renderer interface {
	DrawFrame(bounds image.Rectangle) (*image.RGBA, error)
}

// FrameStreamer will stream image frames, from a provided renderer, at provided intervals, over a channel
type FrameStreamer struct {
	C chan image.Image
	E chan error

	renderer Renderer
	bounds   image.Rectangle
	ticker   *time.Ticker
	started  bool
}

type Params struct {
	Bounds      image.Rectangle
	Renderer    Renderer
	FrametimeMs int64
}

// New creates a FrameStreamer but does not start rendering or sending until Start is called
func New(params Params) FrameStreamer {
	return FrameStreamer{
		C: make(chan image.Image),
		E: make(chan error),

		renderer: params.Renderer,
		bounds:   params.Bounds,
		ticker:   time.NewTicker(time.Duration(params.FrametimeMs) * time.Millisecond),
	}
}

// Start loops on the ticker and on each tick renders a frame, sending it to the frame channel
func (fs FrameStreamer) Start() {
	if fs.started {
		return
	}
	fs.started = true

	// ticker used instead of sleep:
	// tickers drop ticks for slow recievers i.e. if recieving on the
	// FS channel is blocked, ticks slow and the renderer is not called un-necessarily
	for range fs.ticker.C {
		frame, err := fs.renderer.DrawFrame(fs.bounds)
		if err != nil {
			fs.E <- err
			return
		}

		fs.C <- frame
	}
}

// Stop ends the ticker and closes the channels
func (fs FrameStreamer) Stop() {
	fs.ticker.Stop()

	close(fs.C)
	close(fs.E)
}
