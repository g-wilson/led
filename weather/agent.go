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

	cache.Today = dw.Days[0]

	// now := time.Now()
	// afternoon, err := time.Parse(time.RFC3339,
	// 	fmt.Sprintf("%d-%02d-%02dT16:00:00-00:00", now.Year(), now.Month(), now.Day()),
	// )

	// if now.Before(afternoon) {
	// 	cache.Today = dw.Days[0]
	// } else {
	// 	cache.Today = dw.Days[1]
	// }

	return
}
