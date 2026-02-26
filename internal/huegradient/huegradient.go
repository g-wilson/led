package huegradient

import (
	"image/color"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

// Gradient steps through the Oklch hue wheel at a fixed interval, producing
// perceptually uniform colours with consistent lightness and chroma.
type Gradient struct {
	BaseHue float64
	Step    float64
}

// Color returns the Oklch-derived RGBA colour for iteration i.
func (g Gradient) Color(i int) color.RGBA {
	hue := math.Mod(g.BaseHue+float64(i)*g.Step, 360)
	c := colorful.OkLch(0.75, 0.12, hue).Clamped()
	return color.RGBA{uint8(c.R * 255), uint8(c.G * 255), uint8(c.B * 255), 255}
}
