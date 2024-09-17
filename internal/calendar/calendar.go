package calendar

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"sort"
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

type eventList []Event

func (s eventList) Len() int {
	return len(s)
}

func (s eventList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s eventList) Less(i, j int) bool {
	iStartsAt, _ := time.Parse(time.RFC3339, s[i].Timestamp)
	jStartsAt, _ := time.Parse(time.RFC3339, s[j].Timestamp)

	return iStartsAt.Before(jStartsAt)
}

var sortedEvents = (func() eventList {
	e := append(eventList{}, events...)

	sort.Sort(e)

	return e
})()

func GetNextEvent() *Event {
	for _, r := range sortedEvents {
		until := r.Until()

		if until.Seconds() < 0 {
			continue
		}

		return &r
	}

	return nil
}
