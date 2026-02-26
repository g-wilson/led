package clock

import (
	"image"
	"image/color"

	"github.com/g-wilson/led/internal/calendar"

	"golang.org/x/image/draw"
)

func (r *ClockRenderer) renderCountdown(c *image.RGBA) error {
	if event := calendar.GetNextEvent(); event != nil {
		if event.Image != nil {
			draw.Draw(c, c.Bounds(), event.Image, image.Point{X: -44, Y: -9}, draw.Over)
		}
		halfway := 32 - int(float64((4*len(event.Name))/2))
		r.addText(c, image.Point{X: halfway, Y: 15}, event.Name, color.RGBA{60, 60, 215, 255})
		r.addText(c, image.Point{X: 10, Y: 22}, formatDuration(event.Until()), color.RGBA{215, 0, 0, 255})
	}
	return nil
}
