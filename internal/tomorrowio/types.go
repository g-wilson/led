package tomorrowio

import (
	"time"

	"github.com/g-wilson/led/internal/weather"
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
	CloudBaseMin                float64   `json:"cloudBaseMin"`
	CloudCeilingAvg             float64   `json:"cloudCeilingAvg"`
	CloudCeilingMax             float64   `json:"cloudCeilingMax"`
	CloudCeilingMin             float64   `json:"cloudCeilingMin"`
	CloudCoverAvg               float64   `json:"cloudCoverAvg"`
	CloudCoverMax               float64   `json:"cloudCoverMax"`
	CloudCoverMin               float64   `json:"cloudCoverMin"`
	DewPointAvg                 float64   `json:"dewPointAvg"`
	DewPointMax                 float64   `json:"dewPointMax"`
	DewPointMin                 float64   `json:"dewPointMin"`
	EvapotranspirationAvg       float64   `json:"evapotranspirationAvg"`
	EvapotranspirationMax       float64   `json:"evapotranspirationMax"`
	EvapotranspirationMin       float64   `json:"evapotranspirationMin"`
	EvapotranspirationSum       float64   `json:"evapotranspirationSum"`
	FreezingRainIntensityAvg    float64   `json:"freezingRainIntensityAvg"`
	FreezingRainIntensityMax    float64   `json:"freezingRainIntensityMax"`
	FreezingRainIntensityMin    float64   `json:"freezingRainIntensityMin"`
	HumidityAvg                 float64   `json:"humidityAvg"`
	HumidityMax                 float64   `json:"humidityMax"`
	HumidityMin                 float64   `json:"humidityMin"`
	IceAccumulationAvg          float64   `json:"iceAccumulationAvg"`
	IceAccumulationLweAvg       float64   `json:"iceAccumulationLweAvg"`
	IceAccumulationLweMax       float64   `json:"iceAccumulationLweMax"`
	IceAccumulationLweMin       float64   `json:"iceAccumulationLweMin"`
	IceAccumulationMax          float64   `json:"iceAccumulationMax"`
	IceAccumulationMin          float64   `json:"iceAccumulationMin"`
	IceAccumulationSum          float64   `json:"iceAccumulationSum"`
	MoonriseTime                time.Time `json:"moonriseTime"`
	MoonsetTime                 time.Time `json:"moonsetTime"`
	PrecipitationProbabilityAvg float64   `json:"precipitationProbabilityAvg"`
	PrecipitationProbabilityMax float64   `json:"precipitationProbabilityMax"`
	PrecipitationProbabilityMin float64   `json:"precipitationProbabilityMin"`
	PressureSurfaceLevelAvg     float64   `json:"pressureSurfaceLevelAvg"`
	PressureSurfaceLevelMax     float64   `json:"pressureSurfaceLevelMax"`
	PressureSurfaceLevelMin     float64   `json:"pressureSurfaceLevelMin"`
	RainAccumulationAvg         float64   `json:"rainAccumulationAvg"`
	RainAccumulationLweAvg      float64   `json:"rainAccumulationLweAvg"`
	RainAccumulationLweMax      float64   `json:"rainAccumulationLweMax"`
	RainAccumulationLweMin      float64   `json:"rainAccumulationLweMin"`
	RainAccumulationMax         float64   `json:"rainAccumulationMax"`
	RainAccumulationMin         float64   `json:"rainAccumulationMin"`
	RainAccumulationSum         float64   `json:"rainAccumulationSum"`
	RainIntensityAvg            float64   `json:"rainIntensityAvg"`
	RainIntensityMax            float64   `json:"rainIntensityMax"`
	RainIntensityMin            float64   `json:"rainIntensityMin"`
	SleetAccumulationAvg        float64   `json:"sleetAccumulationAvg"`
	SleetAccumulationLweAvg     float64   `json:"sleetAccumulationLweAvg"`
	SleetAccumulationLweMax     float64   `json:"sleetAccumulationLweMax"`
	SleetAccumulationLweMin     float64   `json:"sleetAccumulationLweMin"`
	SleetAccumulationMax        float64   `json:"sleetAccumulationMax"`
	SleetAccumulationMin        float64   `json:"sleetAccumulationMin"`
	SleetIntensityAvg           float64   `json:"sleetIntensityAvg"`
	SleetIntensityMax           float64   `json:"sleetIntensityMax"`
	SleetIntensityMin           float64   `json:"sleetIntensityMin"`
	SnowAccumulationAvg         float64   `json:"snowAccumulationAvg"`
	SnowAccumulationLweAvg      float64   `json:"snowAccumulationLweAvg"`
	SnowAccumulationLweMax      float64   `json:"snowAccumulationLweMax"`
	SnowAccumulationLweMin      float64   `json:"snowAccumulationLweMin"`
	SnowAccumulationMax         float64   `json:"snowAccumulationMax"`
	SnowAccumulationMin         float64   `json:"snowAccumulationMin"`
	SnowAccumulationSum         float64   `json:"snowAccumulationSum"`
	SnowIntensityAvg            float64   `json:"snowIntensityAvg"`
	SnowIntensityMax            float64   `json:"snowIntensityMax"`
	SnowIntensityMin            float64   `json:"snowIntensityMin"`
	SunriseTime                 time.Time `json:"sunriseTime"`
	SunsetTime                  time.Time `json:"sunsetTime"`
	TemperatureApparentAvg      float64   `json:"temperatureApparentAvg"`
	TemperatureApparentMax      float64   `json:"temperatureApparentMax"`
	TemperatureApparentMin      float64   `json:"temperatureApparentMin"`
	TemperatureAvg              float64   `json:"temperatureAvg"`
	TemperatureMax              float64   `json:"temperatureMax"`
	TemperatureMin              float64   `json:"temperatureMin"`
	UvHealthConcernAvg          float64   `json:"uvHealthConcernAvg"`
	UvHealthConcernMax          float64   `json:"uvHealthConcernMax"`
	UvHealthConcernMin          float64   `json:"uvHealthConcernMin"`
	UvIndexAvg                  float64   `json:"uvIndexAvg"`
	UvIndexMax                  float64   `json:"uvIndexMax"`
	UvIndexMin                  float64   `json:"uvIndexMin"`
	VisibilityAvg               float64   `json:"visibilityAvg"`
	VisibilityMax               float64   `json:"visibilityMax"`
	VisibilityMin               float64   `json:"visibilityMin"`
	WeatherCodeMax              float64   `json:"weatherCodeMax"`
	WeatherCodeMin              float64   `json:"weatherCodeMin"`
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
		TemperatureHigh: float32(d.TemperatureMax),
		TemperatureLow:  float32(d.TemperatureMin),
		SunriseTime:     d.SunriseTime,
		SunsetTime:      d.SunsetTime,
		Rainy:           d.PrecipitationProbabilityAvg > 25,
		Windy:           d.WindSpeedAvg > 6 || d.WindGustAvg > 12,
		Cloudy:          d.CloudCoverAvg > 0.6,
		Humidity:        float32(d.HumidityAvg),
	}
}
