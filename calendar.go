package led

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

var Events = []Event{
	Event{
		Name:      "TEST",
		Timestamp: "2021-08-31T13:00:00.000Z",
	},
}

func GetNextEvent() (r *Event) {
	for _, r := range Events {
		until := r.Until()

		if until.Seconds() < 0 {
			continue
		}

		return &r
	}

	return nil
}
