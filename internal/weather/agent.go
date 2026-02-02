package weather

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type DayWeather struct {
	TemperatureHigh float32
	TemperatureLow  float32
	SunriseTime     time.Time
	SunsetTime      time.Time
	Rainy           bool
	Windy           bool
	Cloudy          bool
	Snowy           bool
	Humidity        float32
}

type TwoDayWeather struct {
	Today    DayWeather
	Tomorrow DayWeather
}

type DayWeatherProvider interface {
	GetTwoDayWeatherAtLocation(lat, lon string) (TwoDayWeather, error)
}

type Agent struct {
	client  DayWeatherProvider
	options AgentOptions

	mu           sync.RWMutex
	todayData    DayWeather
	tomorrowData DayWeather
}

type AgentOptions struct {
	Latitude  string
	Longitude string
	Refresh   int
}

func New(client DayWeatherProvider, options AgentOptions) (*Agent, error) {
	a := &Agent{
		client:  client,
		options: options,
	}

	err := a.populateCache()
	if err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(time.Duration(options.Refresh) * time.Second)
		for range ticker.C {
			err := a.populateCache()
			if err != nil {
				log.Println(fmt.Errorf("error fetching weather: %w", err))
			}
		}
	}()

	return a, nil
}

func (a *Agent) populateCache() (err error) {
	log.Println("fetching weather")
	dw, err := a.client.GetTwoDayWeatherAtLocation(a.options.Latitude, a.options.Longitude)
	if err != nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.todayData = dw.Today
	a.tomorrowData = dw.Tomorrow

	return
}

func (a *Agent) GetToday() DayWeather {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.todayData
}

func (a *Agent) GetTomorrow() DayWeather {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.tomorrowData
}
