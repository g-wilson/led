package weather

import (
	"log"
	"time"

	"github.com/g-wilson/led/pkg/darksky"
)

type DayWeather struct {
	ApparentTemperatureHigh float32
	ApparentTemperatureLow  float32
	SunriseTime             time.Time
	SunsetTime              time.Time
	Rainy                   bool
	Windy                   bool
	Cloudy                  bool
	Humidity                float32
}

type Agent struct {
	client  *darksky.Client
	options AgentOptions

	todayData    DayWeather
	tomorrowData DayWeather
}

type AgentOptions struct {
	Latitude  string
	Longitude string
	Refresh   int
}

func New(client *darksky.Client, options AgentOptions) *Agent {
	a := &Agent{
		client:  client,
		options: options,
	}

	a.populateCache()

	go func() {
		ticker := time.NewTicker(time.Duration(options.Refresh) * time.Second)
		for range ticker.C {
			_ = a.populateCache()
		}
	}()

	return a
}

func (a *Agent) populateCache() (err error) {
	log.Println("fetching weather")
	dw, err := a.client.GetDailyWeather(a.options.Latitude, a.options.Longitude)
	if err != nil {
		return
	}

	a.todayData = dayWeatherFromDarksky(dw.Days[0])
	a.tomorrowData = dayWeatherFromDarksky(dw.Days[1])

	return
}

func (a *Agent) GetToday() DayWeather {
	return a.todayData
}

func (a *Agent) GetTomorrow() DayWeather {
	return a.tomorrowData
}

func dayWeatherFromDarksky(in darksky.DayWeather) DayWeather {
	return DayWeather{
		ApparentTemperatureHigh: in.ApparentTemperatureHigh,
		ApparentTemperatureLow:  in.ApparentTemperatureLow,
		SunriseTime:             time.Unix(in.SunriseTime, 0),
		SunsetTime:              time.Unix(in.SunsetTime, 0),
		Rainy:                   in.PrecipType == "rain" && in.PrecipProbability > 0.6,
		Windy:                   in.WindSpeed > 10 || in.WindGust > 20,
		Cloudy:                  in.CloudCover > 0.4,
		Humidity:                in.Humidity,
	}
}
