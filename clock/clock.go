package clock

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"strconv"
	"time"

	"github.com/g-wilson/led/clock/weather"

	"github.com/joho/godotenv"
	"github.com/toelsiba/fopix"
	"golang.org/x/image/draw"
)

//go:embed fonts/tom-thumb-new.json
var fontSource []byte

type ClockRenderer struct {
	font         *fopix.Drawer
	weatherCache *weather.Cache
	location     *time.Location
}

func New() (ClockRenderer, error) {
	err := godotenv.Load()
	if err != nil {
		return ClockRenderer{}, fmt.Errorf("error loading .env file: %w", err)
	}

	fontInfo := fopix.FontInfo{}
	err = json.Unmarshal(fontSource, &fontInfo)
	if err != nil {
		return ClockRenderer{}, fmt.Errorf("error loading font info file: %w", err)
	}
	font, err := fopix.NewDrawer(fontInfo)
	if err != nil {
		return ClockRenderer{}, fmt.Errorf("error creating font drawer: %w", err)
	}
	font.SetScale(1)

	weatherAPIKey := os.Getenv("DARKSKY_API_KEY")
	if len(weatherAPIKey) == 0 {
		return ClockRenderer{}, errors.New("environment variable DARKSKY_API_KEY is required")
	}

	weatherClient := weather.New(weatherAPIKey, nil)
	weatherRefresh, _ := strconv.ParseInt(os.Getenv("WEATHER_REFRESH"), 10, 32)
	weatherCache := weather.NewAgent(weatherClient, weather.AgentOptions{
		Refresh:   int(weatherRefresh),
		Latitude:  os.Getenv("WEATHER_LATITUDE"),
		Longitude: os.Getenv("WEATHER_LONGITUDE"),
	})

	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return ClockRenderer{}, fmt.Errorf("cannot determine timezone: %w", err)
	}

	return ClockRenderer{
		font:         font,
		weatherCache: weatherCache,
		location:     location,
	}, nil
}

func (r ClockRenderer) DrawFrame(bounds image.Rectangle) (*image.RGBA, error) {
	c := image.NewRGBA(bounds)
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	r.addText(c, image.Point{X: 0, Y: -1}, r.getTimeString(), color.RGBA{255, 255, 255, 255})

	r.addText(c, image.Point{X: 0, Y: 12}, fmt.Sprintf("%02.foC", r.weatherCache.Today.ApparentTemperatureLow), color.RGBA{80, 80, 255, 255})
	r.addText(c, image.Point{X: 17, Y: 12}, fmt.Sprintf("%02.foC", r.weatherCache.Today.ApparentTemperatureHigh), color.RGBA{255, 150, 0, 255})

	event := getNextEvent()
	if event != nil {
		r.addText(c, image.Point{X: 0, Y: 26}, event.Name+":"+formatDuration(event.Until()), color.RGBA{255, 20, 20, 255})
	}

	return c, nil
}

func (r ClockRenderer) addText(c *image.RGBA, pos image.Point, text string, col color.RGBA) {
	r.font.SetColor(col)
	r.font.DrawText(c, pos, text)
}

func (r ClockRenderer) getTimeString() string {
	return time.Now().UTC().In(r.location).Format("15:04 Mon Jan 2")
}

func formatDuration(u time.Duration) string {
	u = u.Round(time.Minute)

	// not actually days - 24h periods because that's much easier and honestly who needs daylight savings
	d := u / (time.Hour * 24)
	u -= d * (time.Hour * 24)

	h := u / time.Hour
	u -= h * time.Hour

	m := u / time.Minute

	return fmt.Sprintf("%02dd %02dh %02dm", d, h, m)
}

func formatDurationSeconds(u time.Duration) string {
	u = u.Round(time.Second)

	// d := u / (time.Hour * 24)
	// u -= d * (time.Hour * 24)

	h := u / time.Hour
	u -= h * time.Hour

	m := u / time.Minute
	u -= m * time.Minute

	s := u / time.Second

	return fmt.Sprintf("%02dh %02dm %02ds", h, m, s)
}
