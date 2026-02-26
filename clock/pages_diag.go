package clock

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/g-wilson/led/internal/diagnostics"
)

var (
	diagGreen  = color.RGBA{0, 200, 0, 255}
	diagYellow = color.RGBA{200, 200, 0, 255}
	diagOrange = color.RGBA{255, 140, 0, 255}
	diagRed    = color.RGBA{200, 0, 0, 255}
)

func (r *ClockRenderer) renderDiag(c *image.RGBA) error {
	status := r.diagnostics.GetStatus()
	sinceText, sinceColor := diagSinceText(status)
	pingText, pingColor := diagPingText(status)
	r.addText(c, image.Point{X: 0, Y: 10}, sinceText, sinceColor)
	r.addText(c, image.Point{X: 0, Y: 18}, pingText, pingColor)
	return nil
}

func diagSinceText(status diagnostics.Status) (string, color.RGBA) {
	if status.LastHealthyAt.IsZero() {
		return "Last ok never", diagRed
	}

	since := time.Since(status.LastHealthyAt)
	sinceText := fmt.Sprintf("Last ok %s", formatShortDuration(since))
	if status.IsStale(time.Now()) {
		return sinceText, diagRed
	}

	return sinceText, diagGreen
}

func diagPingText(status diagnostics.Status) (string, color.RGBA) {
	if !status.LastPingOk {
		return "Ping n/a", diagRed
	}

	pingText := fmt.Sprintf("Ping %dms", status.LastPing.Milliseconds())
	level := status.PingLevel()
	return pingText, diagPingColor(level)
}

func diagPingColor(level diagnostics.PingLevel) color.RGBA {
	switch level {
	case diagnostics.PingLevelGreen:
		return diagGreen
	case diagnostics.PingLevelYellow:
		return diagYellow
	case diagnostics.PingLevelOrange:
		return diagOrange
	default:
		return diagRed
	}
}
