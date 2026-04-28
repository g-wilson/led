package airmatters

import "image/color"

type Response struct {
	Place  Place  `json:"place"`
	Latest Latest `json:"latest"`
}

type Place struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	PlaceID string  `json:"place_id"`
}

type Latest struct {
	Readings   []Reading `json:"readings"`
	UpdateTime string    `json:"update_time"`
}

type Reading struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Color string `json:"color"`
	Level string `json:"level"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type AirCondition struct {
	PlaceName string
	AQI       ReadingData
	PM25      ReadingData
}

type ReadingData struct {
	Value string
	Level string
	Color color.RGBA
}
