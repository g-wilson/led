package clock

import (
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
	{
		Name:      "XMAS",
		Timestamp: "2021-12-25T00:00:00.000Z",
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
