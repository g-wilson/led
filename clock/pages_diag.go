package clock

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/g-wilson/led/internal/diagnostics"
	"github.com/g-wilson/led/internal/huegradient"
)

var (
	diagGreen  = huegradient.Gradient{BaseHue: 150}.Color(0)
	diagYellow = huegradient.Gradient{BaseHue: 100}.Color(0)
	diagOrange = huegradient.Gradient{BaseHue: 50}.Color(0)
	diagRed    = huegradient.Gradient{BaseHue: 26}.Color(0)
)

var _ page = (*ClockRenderer)(nil).renderDiag

func (r *ClockRenderer) renderDiag(c *image.RGBA) error {
	status := r.diagnostics.GetStatus()
	sinceText, sinceColor := diagSinceText(status)
	pingText, pingColor := diagPingText(status)

	r.addText(c, image.Point{X: 1, Y: 10}, sinceText, sinceColor)
	r.addText(c, image.Point{X: 1, Y: 18}, pingText, pingColor)

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
