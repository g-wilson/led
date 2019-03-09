package led

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/g-wilson/led/f1"
	"github.com/g-wilson/led/weather"

	"github.com/joho/godotenv"
	"github.com/toelsiba/fopix"
)

var fontFace *fopix.Font

var weatherCache *weather.Cache

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fontFace = getFontFace()

	weatherRefresh, _ := strconv.ParseInt(os.Getenv("WEATHER_REFRESH"), 10, 32)
	weatherCache = weather.NewAgent(getWeatherClient(), weather.AgentOptions{
		Refresh:   int(weatherRefresh),
		Latitude:  os.Getenv("WEATHER_LATITUDE"),
		Longitude: os.Getenv("WEATHER_LONGITUDE"),
	})
}

func getFontFace() *fopix.Font {
	f, err := fopix.NewFromFile("./fonts/tom-thumb-new.json")
	if err != nil {
		log.Fatalln(err)
	}

	return f
}

func getWeatherClient() *weather.DarkskyClient {
	apiKey := os.Getenv("DARKSKY_API_KEY")

	if len(apiKey) == 0 {
		log.Fatalln(errors.New("environment variable DARKSKY_API_KEY is required"))
	}

	return weather.New(apiKey, nil)
}

// NewFrameChannel returns a channel which can recieve image frames on each "tick"
func NewFrameChannel(bounds image.Rectangle, frametime int) <-chan image.Image {
	frames := make(chan image.Image)

	canvas := image.NewRGBA(bounds)

	draw.Draw(canvas, bounds, &image.Uniform{color.Black}, image.ZP, draw.Src)

	go func() {
		ticker := time.Tick(time.Duration(frametime) * time.Millisecond)
		for range ticker {
			frame, err := drawFrame(canvas)

			if err != nil {
				log.Println(err.Error())
				close(frames)
				return
			}

			frames <- frame
		}
	}()

	return frames
}

func drawFrame(c *image.RGBA) (*image.RGBA, error) {
	// clear canvas before drawing?
	draw.Draw(c, c.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	addText(c, 0, -1, time.Now().UTC().Format("15:04"), &color.RGBA{255, 255, 255, 255})

	weatherX := 0
	weatherY := 12
	addText(c, weatherX, weatherY, fmt.Sprintf("%02.f", weatherCache.Today.ApparentTemperatureLow)+"oC", &color.RGBA{80, 80, 255, 255})
	addText(c, weatherX+17, weatherY, fmt.Sprintf("%02.f", weatherCache.Today.ApparentTemperatureHigh)+"oC", &color.RGBA{255, 150, 0, 255})

	race := f1.GetNextRace()
	if race != nil {
		addText(c, 0, 26, race.Location+": "+formatDuration(race.Until()), &color.RGBA{255, 20, 20, 255})
	}

	return c, nil
}

func addText(c *image.RGBA, x, y int, text string, col *color.RGBA) {
	if col == nil {
		col = &color.RGBA{255, 255, 255, 255}
	}

	fontFace.Scale(1)
	fontFace.Color(col)
	fontFace.DrawText(c, image.Point{X: x, Y: y}, text)
}

// Not actually days - 24h periods because that's much easier and honestly who needs daylight savings
func formatDuration(u time.Duration) string {
	u = u.Round(time.Minute)

	d := u / (time.Hour * 24)
	u -= d * (time.Hour * 24)

	h := u / time.Hour
	u -= h * time.Hour

	m := u / time.Minute

	return fmt.Sprintf("%02dd %02dh %02dm", d, h, m)
}
