package clock

import (
	"fmt"
	"image"
	"image/color"
)

var _ page = (*ClockRenderer)(nil).renderAirQuality

func (r *ClockRenderer) renderAirQuality(c *image.RGBA) error {
	air := r.airQuality.Get()

	r.addText(c, image.Point{X: 0, Y: 8}, "Air Quality", color.RGBA{180, 180, 180, 255})

	aqiText := fmt.Sprintf("AQI %s %s", air.AQI.Value, air.AQI.Level)
	r.addText(c, image.Point{X: 0, Y: 15}, aqiText, air.AQI.Color)

	pm25Text := fmt.Sprintf("PM2.5 %s", air.PM25.Value)
	r.addText(c, image.Point{X: 0, Y: 22}, pm25Text, air.PM25.Color)

	return nil
}
