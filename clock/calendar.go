package clock

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"time"
)

//go:embed images/xmastree.png
var xmasImageSource []byte
var xmasImage = mustLoadImage(xmasImageSource)

//go:embed images/f1.png
var f1ImageSource []byte
var f1Image = mustLoadImage(f1ImageSource)

func mustLoadImage(src []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		panic(fmt.Errorf("error loading image: %w", err))
	}

	return img
}

type Event struct {
	Name      string
	Timestamp string
	Image     image.Image
}

func (e Event) Until() time.Duration {
	startsAt, _ := time.Parse(time.RFC3339, e.Timestamp)
	return time.Until(startsAt)
}

var events = []Event{
	{
		Name:      "BEL",
		Timestamp: "2022-08-28T14:00:00.000+01:00",
		Image:     f1Image,
	},
	{
		Name:      "NED",
		Timestamp: "2022-09-04T14:00:00.000+01:00",
		Image:     f1Image,
	},
	{
		Name:      "ITA",
		Timestamp: "2022-09-11T14:00:00.000+01:00",
		Image:     f1Image,
	},
	{
		Name:      "ELLIE",
		Timestamp: "2022-09-18T00:00:00.000+01:00",
		Image:     nil,
	},
	{
		Name:      "SIN",
		Timestamp: "2022-10-02T13:00:00.000+01:00",
		Image:     f1Image,
	},
	{
		Name:      "JAP",
		Timestamp: "2022-10-09T06:00:00.000+01:00",
		Image:     f1Image,
	},
	{
		Name:      "USA",
		Timestamp: "2022-10-23T20:00:00.000+01:00",
		Image:     f1Image,
	},
	{
		Name:      "MEX",
		Timestamp: "2022-10-30T20:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "BRA",
		Timestamp: "2022-11-13T18:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "ABU",
		Timestamp: "2022-11-20T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "XMAS",
		Timestamp: "2022-12-25T00:00:00.000Z",
		Image:     xmasImage,
	},
	{
		Name:      "2023",
		Timestamp: "2023-01-0100:00:00.000Z",
		Image:     xmasImage,
	},
	{
		Name:      "GEORGE",
		Timestamp: "2023-07-1100:00:00.000+01:00",
		Image:     nil,
	},
}

func getNextEvent() *Event {
	for _, r := range events {
		until := r.Until()

		if until.Seconds() < 0 {
			continue
		}

		return &r
	}

	return nil
}

func formatDuration(u time.Duration) string {
	u = u.Round(time.Minute)

	// not actually days - 24h periods because that's much easier and honestly who needs daylight savings
	d := u / (time.Hour * 24)
	u -= d * (time.Hour * 24)

	h := u / time.Hour
	u -= h * time.Hour

	m := u / time.Minute
	u -= m * time.Minute

	s := u / time.Second

	// less than one day to go, render more precise countdown
	if d <= 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", h, m, s)
	}

	return fmt.Sprintf("%02dd %02dh %02dm", d, h, m)
}
