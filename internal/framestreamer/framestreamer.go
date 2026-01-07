package framestreamer

import (
	"image"
	"image/color"
	"image/draw"
	"time"
)

const (
	OneFPS        = 1000
	TenFPS        = 1000 / 10
	FifteenFPS    = 1000 / 15
	TwentyFourFPS = 1000 / 24
	ThirtyFPS     = 1000 / 30
	SixtyFPS      = 1000 / 60

	// bufferCount is the number of pre-allocated buffers for triple buffering.
	// Triple buffering ensures safety without explicit "return to pool" calls:
	// - Buffer N: Being drawn into
	// - Buffer N-1: In channel / being processed
	// - Buffer N-2: Consumer might still be using
	bufferCount = 3
)

// Renderer delivers single image frames to our stream by drawing into a provided buffer.
type Renderer interface {
	DrawFrame(target *image.RGBA) error
}

// FrameStreamer will stream image frames, from a provided renderer, at provided intervals, over a channel.
// It manages a triple buffer pool internally to avoid per-frame allocations.
type FrameStreamer struct {
	C chan *image.RGBA
	E chan error

	renderer Renderer
	bounds   image.Rectangle
	ticker   *time.Ticker
	started  bool

	// Buffer pool - triple buffering for zero-allocation frame streaming
	buffers [bufferCount]*image.RGBA
	current int
}

type Params struct {
	Bounds      image.Rectangle
	Renderer    Renderer
	FrametimeMs int64
}

// New creates a FrameStreamer but does not start rendering or sending until Start is called.
// It pre-allocates a triple buffer pool based on the provided bounds.
func New(params Params) *FrameStreamer {
	fs := &FrameStreamer{
		C: make(chan *image.RGBA),
		E: make(chan error),

		renderer: params.Renderer,
		bounds:   params.Bounds,
		ticker:   time.NewTicker(time.Duration(params.FrametimeMs) * time.Millisecond),
		current:  0,
	}

	// Pre-allocate triple buffer pool
	for i := range fs.buffers {
		fs.buffers[i] = image.NewRGBA(params.Bounds)
	}

	return fs
}

// Start loops on the ticker and on each tick renders a frame, sending it to the frame channel.
// Buffers are rotated through the pool to avoid allocations.
func (fs *FrameStreamer) Start() {
	if fs.started {
		return
	}
	fs.started = true

	// ticker used instead of sleep:
	// tickers drop ticks for slow recievers i.e. if recieving on the
	// FS channel is blocked, ticks slow and the renderer is not called un-necessarily
	for range fs.ticker.C {
		// Rotate to next buffer
		fs.current = (fs.current + 1) % bufferCount
		buf := fs.buffers[fs.current]

		// Clear buffer to black before rendering
		draw.Draw(buf, buf.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

		// Renderer draws into the provided buffer
		err := fs.renderer.DrawFrame(buf)
		if err != nil {
			fs.E <- err
			return
		}

		fs.C <- buf
	}
}

// Stop ends the ticker and closes the channels
func (fs *FrameStreamer) Stop() {
	fs.ticker.Stop()

	close(fs.C)
	close(fs.E)
}
