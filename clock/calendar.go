package clock

import (
	"fmt"
	"time"
)

type Event struct {
	Name      string
	Timestamp string
}

func (e Event) Until() time.Duration {
	startsAt, _ := time.Parse(time.RFC3339, e.Timestamp)
	return time.Until(startsAt)
}

var events = []Event{
	// {
	// 	Name:      "XMAS",
	// 	Timestamp: "2021-12-25T00:00:00.000Z",
	// },
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

	return fmt.Sprintf("%02dd %02dh %02dm", d, h, m)
}
