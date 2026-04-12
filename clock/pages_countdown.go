package clock

import (
	"image"
	"image/color"

	"github.com/g-wilson/led/internal/calendar"
	"github.com/g-wilson/led/internal/huegradient"

	"golang.org/x/image/draw"
)

var (
	colourEventName = huegradient.Gradient{BaseHue: 260}.Color(0)
	colourCountdown = color.RGBA{215, 0, 0, 255}
)

var _ page = (*ClockRenderer)(nil).renderCountdown

func (r *ClockRenderer) renderCountdown(c *image.RGBA) error {
	if event := calendar.GetNextEvent(); event != nil {
		if event.Image != nil {
			draw.Draw(c, c.Bounds(), event.Image, image.Point{X: -44, Y: -9}, draw.Over)
		}
		halfway := 32 - int(float64((4*len(event.Name))/2))
		r.addText(c, image.Point{X: halfway, Y: 15}, event.Name, colourEventName)
		r.addText(c, image.Point{X: 10, Y: 22}, formatDuration(event.Until()), colourCountdown)
	}
	return nil
}
