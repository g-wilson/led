package clock

import (
	"fmt"
	"image"
	"image/color"

	"github.com/g-wilson/led/internal/huegradient"
	"github.com/g-wilson/led/internal/weather"
)

var (
	colourDayName = color.RGBA{215, 0, 88, 255}

	colourTempLow  = huegradient.Gradient{BaseHue: 240}.Color(0)
	colourTempHigh = huegradient.Gradient{BaseHue: 80}.Color(0)

	colourSunGradient  = huegradient.Gradient{BaseHue: 40, Step: 40}
	colourMoonGradient = huegradient.Gradient{BaseHue: 280, Step: 40}
)

var _ page = (*ClockRenderer)(nil).renderToday
var _ page = (*ClockRenderer)(nil).renderTomorrow
var _ page = (*ClockRenderer)(nil).renderDaylight

func (r *ClockRenderer) renderToday(c *image.RGBA) error {
	w := r.weather.GetToday()
	r.addText(c, image.Point{X: 0, Y: 8}, "Today", colourDayName)
	r.renderWeather(c, w)
	return nil
}

func (r *ClockRenderer) renderTomorrow(c *image.RGBA) error {
	w := r.weather.GetTomorrow()
	r.addText(c, image.Point{X: 0, Y: 8}, "Tomorrow", colourDayName)
	r.renderWeather(c, w)
	return nil
}

func (r *ClockRenderer) renderDaylight(c *image.RGBA) error {
	w := r.weather.GetToday()
	xOffset := 2

	sunrise := w.SunriseTime.In(r.location).Format("15:04")
	sunset := w.SunsetTime.In(r.location).Format("15:04")
	r.addText(c, image.Point{X: xOffset, Y: 8}, fmt.Sprintf(" Sunrise %s", sunrise), colourSunGradient.Color(0))
	r.addText(c, image.Point{X: xOffset, Y: 14}, fmt.Sprintf("  Sunset %s", sunset), colourSunGradient.Color(1))

	if !w.MoonriseTime.IsZero() {
		moonrise := w.MoonriseTime.In(r.location).Format("15:04")
		r.addText(c, image.Point{X: xOffset, Y: 20}, fmt.Sprintf("Moonrise %s", moonrise), colourMoonGradient.Color(0))
	}
	if !w.MoonsetTime.IsZero() {
		moonset := w.MoonsetTime.In(r.location).Format("15:04")
		r.addText(c, image.Point{X: xOffset, Y: 26}, fmt.Sprintf(" Moonset %s", moonset), colourMoonGradient.Color(1))
	}

	return nil
}

func (r *ClockRenderer) renderWeather(c *image.RGBA, w weather.DayWeather) {
	yOffset := 15
	summaryStart := 36

	r.addText(c, image.Point{X: 0, Y: yOffset}, fmt.Sprintf("%02.foC", w.TemperatureLow), colourTempLow)
	r.addText(c, image.Point{X: 17, Y: yOffset}, fmt.Sprintf("%02.foC", w.TemperatureHigh), colourTempHigh)

	// Underneath temperatures, always shows
	if w.Cloudy {
		r.addText(c, image.Point{X: 0, Y: yOffset + 7}, "Cloudy", color.RGBA{179, 161, 136, 255})
	} else {
		r.addText(c, image.Point{X: 0, Y: yOffset + 7}, "Sunny", color.RGBA{255, 213, 0, 255})
	}

	// To the right, conditionally shows
	if w.Snowy {
		r.addText(c, image.Point{X: summaryStart, Y: yOffset}, "Snow", color.RGBA{255, 255, 255, 255})
	} else if w.Rainy {
		r.addText(c, image.Point{X: summaryStart, Y: yOffset}, "Rain", color.RGBA{0, 113, 237, 255})
	}
	if w.Windy {
		r.addText(c, image.Point{X: summaryStart, Y: yOffset + 7}, "Windy", color.RGBA{0, 247, 255, 255})
	}

	// Humidity, hidden for now
	// r.addText(c, image.Point{X: summaryStart + 15, Y: yOffset}, fmt.Sprintf("H%02.f", (w.Humidity*100)), color.RGBA{230, 77, 0, 255})
}
