package clock

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/soniakeys/meeus/v3/julian"
	"github.com/soniakeys/meeus/v3/moonillum"
)

var colourMoon = color.RGBA{50, 50, 70, 255}
var litColour = color.RGBA{210, 215, 230, 255}
var darkColour = color.RGBA{25, 25, 45, 255}

func moonPhaseIllumination(t time.Time) (illum float64, waxing bool) {
	jde := julian.TimeToJD(t)
	angle := moonillum.PhaseAngle3(jde)
	illum = (1 + math.Cos(float64(angle))) / 2

	// compare to 12 hours ahead to determine waxing vs waning
	jdeFuture := julian.TimeToJD(t.Add(12 * time.Hour))
	angleFuture := moonillum.PhaseAngle3(jdeFuture)
	illumFuture := (1 + math.Cos(float64(angleFuture))) / 2
	waxing = illumFuture > illum
	return
}

func moonPhaseName(illum float64, waxing bool) string {
	switch {
	case illum < 0.02:
		return "New Moon"
	case illum > 0.98:
		return "Full Moon"
	case illum < 0.48 && waxing:
		return "Wax Crescent"
	case illum >= 0.48 && illum <= 0.52 && waxing:
		return "1st Quarter"
	case illum > 0.52 && waxing:
		return "Wax Gibbous"
	case illum > 0.52:
		return "Wan Gibbous"
	case illum >= 0.48 && illum <= 0.52:
		return "3rd Quarter"
	default:
		return "Wan Crescent"
	}
}

func drawMoonDisc(c *image.RGBA, centre image.Point, radius int, illum float64, waxing bool) {
	r2 := radius * radius
	for dy := -radius; dy <= radius; dy++ {
		hw := math.Sqrt(float64(r2 - dy*dy))
		termX := hw * (1 - 2*illum)
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > r2 {
				continue
			}
			px := centre.X + dx
			py := centre.Y + dy
			if !image.Pt(px, py).In(c.Bounds()) {
				continue
			}
			var lit bool
			if waxing {
				lit = float64(dx) >= termX
			} else {
				lit = float64(dx) <= -termX
			}
			if lit {
				c.SetRGBA(px, py, litColour)
			} else {
				c.SetRGBA(px, py, darkColour)
			}
		}
	}
}

func (r *ClockRenderer) renderMoon(c *image.RGBA) error {
	illum, waxing := moonPhaseIllumination(time.Now())
	name := moonPhaseName(illum, waxing)

	drawMoonDisc(c, image.Point{X: 32, Y: 15}, 9, illum, waxing)

	centreX := 32 - (4*len(name))/2
	r.addText(c, image.Point{X: centreX, Y: 26}, name, colourMoon)

	return nil
}
