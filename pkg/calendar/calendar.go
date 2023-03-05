package calendar

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

func (e *Event) Until() time.Duration {
	startsAt, _ := time.Parse(time.RFC3339, e.Timestamp)
	return time.Until(startsAt)
}

var events = []Event{
	{
		Name:      "BAH",
		Timestamp: "2023-03-05T15:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "SAU",
		Timestamp: "2023-03-19T17:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "AUS",
		Timestamp: "2023-04-02T05:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "AZE",
		Timestamp: "2023-04-30T11:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "MIA",
		Timestamp: "2023-05-07T19:30:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "EMI",
		Timestamp: "2023-05-21T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "MON",
		Timestamp: "2023-05-28T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "ESP",
		Timestamp: "2023-06-04T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "CAN",
		Timestamp: "2023-06-18T18:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "AUT",
		Timestamp: "2023-06-18T18:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "GBR",
		Timestamp: "2023-07-09T14:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "HUN",
		Timestamp: "2023-07-23T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "BEL",
		Timestamp: "2023-07-30T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "NED",
		Timestamp: "2023-08-27T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "ITA",
		Timestamp: "2023-09-01T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "SIN",
		Timestamp: "2023-09-17T12:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "JAP",
		Timestamp: "2023-09-24T05:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "QAT",
		Timestamp: "2023-10-08T14:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "USA",
		Timestamp: "2023-10-22T19:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "MEX",
		Timestamp: "2023-10-29T20:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "BRA",
		Timestamp: "2023-11-05T17:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "LSV",
		Timestamp: "2023-11-19T04:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "ABU",
		Timestamp: "2023-11-26T13:00:00.000Z",
		Image:     f1Image,
	},
	{
		Name:      "XMAS",
		Timestamp: "2023-12-25T00:00:00.000Z",
		Image:     xmasImage,
	},
	{
		Name:      "2024",
		Timestamp: "2024-01-0100:00:00.000Z",
		Image:     xmasImage,
	},
	{
		Name:      "GEORGE",
		Timestamp: "2023-07-1100:00:00.000+01:00",
		Image:     nil,
	},
	{
		Name:      "ELLIE",
		Timestamp: "2023-09-18T00:00:00.000+01:00",
		Image:     nil,
	},
}

func GetNextEvent() *Event {
	for _, r := range events {
		until := r.Until()

		if until.Seconds() < 0 {
			continue
		}

		return &r
	}

	return nil
}
