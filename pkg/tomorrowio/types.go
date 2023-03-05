package tomorrowio

import (
	"time"

	"github.com/g-wilson/led/pkg/weather"
)

type ErrorResponse struct {
	Code    int64  `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ForecastResponse struct {
	Timelines Timelines `json:"timelines"`
	Location  Location  `json:"location"`
}

type Values struct {
	CloudBaseAvg                float64   `json:"cloudBaseAvg"`
	CloudBaseMax                float64   `json:"cloudBaseMax"`
	CloudBaseMin                int       `json:"cloudBaseMin"`
	CloudCeilingAvg             float64   `json:"cloudCeilingAvg"`
	CloudCeilingMax             float64   `json:"cloudCeilingMax"`
	CloudCeilingMin             int       `json:"cloudCeilingMin"`
	CloudCoverAvg               float64   `json:"cloudCoverAvg"`
	CloudCoverMax               int       `json:"cloudCoverMax"`
	CloudCoverMin               float64   `json:"cloudCoverMin"`
	DewPointAvg                 float64   `json:"dewPointAvg"`
	DewPointMax                 float64   `json:"dewPointMax"`
	DewPointMin                 float64   `json:"dewPointMin"`
	EvapotranspirationAvg       float64   `json:"evapotranspirationAvg"`
	EvapotranspirationMax       float64   `json:"evapotranspirationMax"`
	EvapotranspirationMin       float64   `json:"evapotranspirationMin"`
	EvapotranspirationSum       float64   `json:"evapotranspirationSum"`
	FreezingRainIntensityAvg    int       `json:"freezingRainIntensityAvg"`
	FreezingRainIntensityMax    int       `json:"freezingRainIntensityMax"`
	FreezingRainIntensityMin    int       `json:"freezingRainIntensityMin"`
	HumidityAvg                 float64   `json:"humidityAvg"`
	HumidityMax                 float64   `json:"humidityMax"`
	HumidityMin                 float64   `json:"humidityMin"`
	IceAccumulationAvg          int       `json:"iceAccumulationAvg"`
	IceAccumulationLweAvg       int       `json:"iceAccumulationLweAvg"`
	IceAccumulationLweMax       int       `json:"iceAccumulationLweMax"`
	IceAccumulationLweMin       int       `json:"iceAccumulationLweMin"`
	IceAccumulationMax          int       `json:"iceAccumulationMax"`
	IceAccumulationMin          int       `json:"iceAccumulationMin"`
	IceAccumulationSum          int       `json:"iceAccumulationSum"`
	MoonriseTime                time.Time `json:"moonriseTime"`
	MoonsetTime                 time.Time `json:"moonsetTime"`
	PrecipitationProbabilityAvg int       `json:"precipitationProbabilityAvg"`
	PrecipitationProbabilityMax int       `json:"precipitationProbabilityMax"`
	PrecipitationProbabilityMin int       `json:"precipitationProbabilityMin"`
	PressureSurfaceLevelAvg     float64   `json:"pressureSurfaceLevelAvg"`
	PressureSurfaceLevelMax     float64   `json:"pressureSurfaceLevelMax"`
	PressureSurfaceLevelMin     float64   `json:"pressureSurfaceLevelMin"`
	RainAccumulationAvg         float64   `json:"rainAccumulationAvg"`
	RainAccumulationLweAvg      float64   `json:"rainAccumulationLweAvg"`
	RainAccumulationLweMax      float64   `json:"rainAccumulationLweMax"`
	RainAccumulationLweMin      int       `json:"rainAccumulationLweMin"`
	RainAccumulationMax         float64   `json:"rainAccumulationMax"`
	RainAccumulationMin         int       `json:"rainAccumulationMin"`
	RainAccumulationSum         float64   `json:"rainAccumulationSum"`
	RainIntensityAvg            float64   `json:"rainIntensityAvg"`
	RainIntensityMax            float64   `json:"rainIntensityMax"`
	RainIntensityMin            int       `json:"rainIntensityMin"`
	SleetAccumulationAvg        int       `json:"sleetAccumulationAvg"`
	SleetAccumulationLweAvg     int       `json:"sleetAccumulationLweAvg"`
	SleetAccumulationLweMax     int       `json:"sleetAccumulationLweMax"`
	SleetAccumulationLweMin     int       `json:"sleetAccumulationLweMin"`
	SleetAccumulationMax        int       `json:"sleetAccumulationMax"`
	SleetAccumulationMin        int       `json:"sleetAccumulationMin"`
	SleetIntensityAvg           int       `json:"sleetIntensityAvg"`
	SleetIntensityMax           int       `json:"sleetIntensityMax"`
	SleetIntensityMin           int       `json:"sleetIntensityMin"`
	SnowAccumulationAvg         float64   `json:"snowAccumulationAvg"`
	SnowAccumulationLweAvg      int       `json:"snowAccumulationLweAvg"`
	SnowAccumulationLweMax      float64   `json:"snowAccumulationLweMax"`
	SnowAccumulationLweMin      int       `json:"snowAccumulationLweMin"`
	SnowAccumulationMax         float64   `json:"snowAccumulationMax"`
	SnowAccumulationMin         int       `json:"snowAccumulationMin"`
	SnowAccumulationSum         float64   `json:"snowAccumulationSum"`
	SnowIntensityAvg            int       `json:"snowIntensityAvg"`
	SnowIntensityMax            int       `json:"snowIntensityMax"`
	SnowIntensityMin            int       `json:"snowIntensityMin"`
	SunriseTime                 time.Time `json:"sunriseTime"`
	SunsetTime                  time.Time `json:"sunsetTime"`
	TemperatureApparentAvg      float64   `json:"temperatureApparentAvg"`
	TemperatureApparentMax      float64   `json:"temperatureApparentMax"`
	TemperatureApparentMin      float64   `json:"temperatureApparentMin"`
	TemperatureAvg              float64   `json:"temperatureAvg"`
	TemperatureMax              float64   `json:"temperatureMax"`
	TemperatureMin              float64   `json:"temperatureMin"`
	UvHealthConcernAvg          int       `json:"uvHealthConcernAvg"`
	UvHealthConcernMax          int       `json:"uvHealthConcernMax"`
	UvHealthConcernMin          int       `json:"uvHealthConcernMin"`
	UvIndexAvg                  int       `json:"uvIndexAvg"`
	UvIndexMax                  int       `json:"uvIndexMax"`
	UvIndexMin                  int       `json:"uvIndexMin"`
	VisibilityAvg               int       `json:"visibilityAvg"`
	VisibilityMax               int       `json:"visibilityMax"`
	VisibilityMin               int       `json:"visibilityMin"`
	WeatherCodeMax              int       `json:"weatherCodeMax"`
	WeatherCodeMin              int       `json:"weatherCodeMin"`
	WindDirectionAvg            float64   `json:"windDirectionAvg"`
	WindGustAvg                 float64   `json:"windGustAvg"`
	WindGustMax                 float64   `json:"windGustMax"`
	WindGustMin                 float64   `json:"windGustMin"`
	WindSpeedAvg                float64   `json:"windSpeedAvg"`
	WindSpeedMax                float64   `json:"windSpeedMax"`
	WindSpeedMin                float64   `json:"windSpeedMin"`
}

type Daily struct {
	Time   time.Time `json:"time"`
	Values Values    `json:"values"`
}

type Timelines struct {
	Daily []Daily `json:"daily"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (d Values) ToDomain() weather.DayWeather {
	return weather.DayWeather{
		ApparentTemperatureHigh: float32(d.TemperatureApparentMax),
		ApparentTemperatureLow:  float32(d.TemperatureApparentMin),
		SunriseTime:             d.SunriseTime,
		SunsetTime:              d.SunsetTime,
		Rainy:                   d.PrecipitationProbabilityAvg > 25,
		Windy:                   d.WindSpeedAvg > 10 || d.WindGustAvg > 20,
		Cloudy:                  d.CloudCoverAvg > 0.6,
		Humidity:                float32(d.HumidityAvg),
	}
}
