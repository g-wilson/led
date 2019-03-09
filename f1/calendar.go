package f1

import (
	"time"
)

type Race struct {
	Location  string
	Timestamp string
}

func (r Race) Until() time.Duration {
	startsAt, _ := time.Parse(time.RFC3339, r.Timestamp)
	return time.Until(startsAt)
}

var Races = []Race{
	Race{
		Location:  "AUS",
		Timestamp: "2019-03-17T05:10:00.000Z",
	},
	Race{
		Location:  "BAH",
		Timestamp: "2019-03-31T16:10:00.000Z",
	},
	Race{
		Location:  "CHN",
		Timestamp: "2019-04-14T07:10:00.000Z",
	},
	Race{
		Location:  "AZE",
		Timestamp: "2019-04-28T13:10:00.000Z",
	},
	Race{
		Location:  "ESP",
		Timestamp: "2019-05-21T14:10:00.000Z",
	},
	Race{
		Location:  "MCO",
		Timestamp: "2019-05-26T14:10:00.000Z",
	},
	Race{
		Location:  "CAN",
		Timestamp: "2019-06-09T19:10:00.000Z",
	},
	Race{
		Location:  "FRA",
		Timestamp: "2019-06-23T14:10:00.000Z",
	},
	Race{
		Location:  "AUT",
		Timestamp: "2019-06-30T14:10:00.000Z",
	},
	Race{
		Location:  "GBR",
		Timestamp: "2019-07-14T14:10:00.000Z",
	},
	Race{
		Location:  "GER",
		Timestamp: "2019-07-28T14:10:00.000Z",
	},
	Race{
		Location:  "HUN",
		Timestamp: "2019-08-04T14:10:00.000Z",
	},
	Race{
		Location:  "BEL",
		Timestamp: "2019-09-01T14:10:00.000Z",
	},
	Race{
		Location:  "ITA",
		Timestamp: "2019-09-08T14:10:00.000Z",
	},
	Race{
		Location:  "SGP",
		Timestamp: "2019-09-22T13:10:00.000Z",
	},
	Race{
		Location:  "RUS",
		Timestamp: "2019-09-29T12:10:00.000Z",
	},
	Race{
		Location:  "JAP",
		Timestamp: "2019-10-13T06:10:00.000Z",
	},
	Race{
		Location:  "MEX",
		Timestamp: "2019-10-27T19:10:00.000Z",
	},
	Race{
		Location:  "USA",
		Timestamp: "2019-11-03T19:10:00.000Z",
	},
	Race{
		Location:  "BRA",
		Timestamp: "2019-11-17T17:10:00.000Z",
	},
	Race{
		Location:  "ARE",
		Timestamp: "2019-12-01T17:10:00.000Z",
	},
}

func GetNextRace() (r *Race) {
	for _, r := range Races {
		until := r.Until()

		if until.Seconds() < 0 {
			continue
		}

		return &r
	}

	return nil
}
