package weather

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const populateCacheTimeout = 15 * time.Second

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
	GetTwoDayWeatherAtLocation(ctx context.Context, lat, lon string) (TwoDayWeather, error)
}

type Agent struct {
	ctx     context.Context
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

func New(ctx context.Context, client DayWeatherProvider, options AgentOptions) (*Agent, error) {
	a := &Agent{
		ctx:     ctx,
		client:  client,
		options: options,
	}

	err := a.populateCache()
	if err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(time.Duration(options.Refresh) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.populateCache(); err != nil {
					log.Println(fmt.Errorf("error fetching weather: %w", err))
				}
			}
		}
	}()

	return a, nil
}

func (a *Agent) populateCache() (err error) {
	log.Println("fetching weather")

	ctx, cancel := context.WithTimeout(a.ctx, populateCacheTimeout)
	defer cancel()

	dw, err := a.client.GetTwoDayWeatherAtLocation(ctx, a.options.Latitude, a.options.Longitude)
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
