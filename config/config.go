package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Settings struct {
	// Weather
	TomorrowIOAPIKey string `env:"TOMORROWIO_API_KEY,required"`
	WeatherLatitude  string `env:"WEATHER_LATITUDE,required"`
	WeatherLongitude string `env:"WEATHER_LONGITUDE,required"`
	WeatherRefresh   int    `env:"WEATHER_REFRESH"    envDefault:"1800"`

	// Clock
	Timezone string `env:"TIMEZONE" envDefault:"Europe/London"`
	Debug    bool   `env:"DEBUG"    envDefault:"false"`

	// Calendar
	CalendarFiles []string `env:"CALENDAR_FILES" envSeparator:","`

	// LED hardware
	LEDRows       int    `env:"LED_ROWS"       envDefault:"32"`
	LEDCols       int    `env:"LED_COLS"       envDefault:"64"`
	LEDPWMBits    int    `env:"LED_PWM_BITS"   envDefault:"11"`
	LEDPWMLSBNano int    `env:"LED_PWM_LSB"    envDefault:"130"`
	LEDBrightness int    `env:"LED_BRIGHTNESS" envDefault:"30"`
	LEDHardware   string `env:"LED_HARDWARE"   envDefault:"adafruit-hat"`

	// Home Assistant (all optional — only used when all three are set)
	HAURL          string   `env:"HA_URL"`
	HAToken        string   `env:"HA_TOKEN"`
	HASensors      []string `env:"HA_SENSORS"       envSeparator:","`
	HAMediaPlayers []string `env:"HA_MEDIA_PLAYERS" envSeparator:","`
}

func Load() (*Settings, error) {
	// Best-effort: real environment variables take precedence over .env file.
	_ = godotenv.Load()

	s := &Settings{}
	if err := env.Parse(s); err != nil {
		return nil, err
	}

	return s, nil
}
