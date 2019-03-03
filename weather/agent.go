package weather

import (
	"log"
	"time"
)

type Cache struct {
	Current CurrentWeather
	Today   DayWeather
}

type AgentOptions struct {
	Latitude  string
	Longitude string
	Refresh   int
}

func NewAgent(client *DarkskyClient, options AgentOptions) *Cache {
	cache := &Cache{}

	populateCache(client, cache, options)

	go func() {
		ticker := time.Tick(time.Duration(options.Refresh) * time.Second)
		for range ticker {
			_ = populateCache(client, cache, options)
		}
	}()

	return cache
}

func populateCache(client *DarkskyClient, cache *Cache, options AgentOptions) (err error) {
	log.Println("fetching weather")
	// cache.Current, err = client.GetCurrentWeather(options.Latitude, options.Longitude)
	dw, err := client.GetDailyWeather(options.Latitude, options.Longitude)
	if err != nil {
		return
	}

	cache.Today = dw.Days[2]

	return
}
