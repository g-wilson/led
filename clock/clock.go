package clock

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
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
		return ClockRenderer{}, err
	}
	font, err := fopix.NewDrawer(fontInfo)
	if err != nil {
		return ClockRenderer{}, err
	}

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
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	r.addText(c, 0, -1, r.getTimeString(), color.RGBA{255, 255, 255, 255})

	weatherX := 0
	weatherY := 12
	r.addText(c, weatherX, weatherY, fmt.Sprintf("%02.f", r.weatherCache.Today.ApparentTemperatureLow)+"oC", color.RGBA{80, 80, 255, 255})
	r.addText(c, weatherX+17, weatherY, fmt.Sprintf("%02.f", r.weatherCache.Today.ApparentTemperatureHigh)+"oC", color.RGBA{255, 150, 0, 255})

	event := getNextEvent()
	if event != nil {
		r.addText(c, 0, 26, event.Name+":"+formatDuration(event.Until()), color.RGBA{255, 20, 20, 255})
	}

	return c, nil
}

func (r ClockRenderer) addText(c *image.RGBA, x, y int, text string, col color.RGBA) {
	r.font.SetScale(1)
	r.font.SetColor(col)
	r.font.DrawText(c, image.Point{X: x, Y: y}, text)
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
