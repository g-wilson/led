package clock

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/g-wilson/led/internal/huegradient"
)

var sensorGradient = huegradient.Gradient{BaseHue: 60, Step: 50}

func (r *ClockRenderer) renderArea(c *image.RGBA, area string) error {
	r.addText(c, image.Point{X: 0, Y: 5}, area, color.RGBA{215, 0, 88, 255})

	if as, ok := r.sensors.GetArea(area); ok {
		for i, s := range as.Sensors {
			y := 12 + (i * 6)
			r.addText(c, image.Point{X: 0, Y: y}, fmt.Sprintf("%s %s%s", shortenSensorName(s.Name), s.State, s.Unit), sensorGradient.Color(i))
		}
	}

	return nil
}

// shortenSensorName turns "main bedroom temperature" into "temp".
func shortenSensorName(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	w := []rune(words[len(words)-1])
	if len(w) > 4 {
		return string(w[:4])
	}
	return string(w)
}
