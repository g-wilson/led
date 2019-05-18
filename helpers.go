package led

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"
)

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

func getTimeString() string {
	return time.Now().UTC().In(location).Format("15:04 Mon Jan 2")
}

func loadImageFile(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, imageType, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	if imageType != "png" {
		return nil, errors.New("image must be a png")
	}

	file.Seek(0, 0)

	return png.Decode(file)
}
