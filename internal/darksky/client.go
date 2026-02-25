package darksky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/g-wilson/led/internal/weather"
)

type Forecast struct {
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Timezone  string         `json:"timezone"`
	Offset    float32        `json:"offset"`
	Current   CurrentWeather `json:"currently"`
	Hourly    HourlyWeather  `json:"hourly"`
	Daily     DailyWeather   `json:"daily"`
}

type HourlyWeather struct {
	Summary string           `json:"summary"`
	Icon    string           `json:"icon"`
	Hours   []CurrentWeather `json:"data"`
}

type DailyWeather struct {
	Summary string       `json:"summary"`
	Icon    string       `json:"icon"`
	Days    []DayWeather `json:"data"`
}

type CurrentWeather struct {
	// Time                 time.Time `json:"time"`
	Summary              string  `json:"summary"`
	Icon                 string  `json:"icon"`
	NearestStormDistance int     `json:"nearestStormDistance"`
	NearestStormBearing  int     `json:"nearestStormBearing"`
	PrecipIntensity      float32 `json:"precipIntensity"`
	PrecipProbability    float32 `json:"precipProbability"`
	Temperature          float32 `json:"temperature"`
	ApparentTemperature  float32 `json:"apparentTemperature"`
	DewPoint             float32 `json:"dewPoint"`
	Humidity             float32 `json:"humidity"`
	Pressure             float32 `json:"pressure"`
	WindSpeed            float32 `json:"windSpeed"`
	WindGust             float32 `json:"windGust"`
	WindBearing          float32 `json:"windBearing"`
	CloudCover           float32 `json:"cloudCover"`
	UVIndex              uint8   `json:"uvIndex"`
	Visibility           float32 `json:"visibility"`
	Ozone                float32 `json:"ozone"`
}

type DayWeather struct {
	// Time                 time.Time `json:"time"`
	Summary     string `json:"summary"`
	Icon        string `json:"icon"`
	SunriseTime int64  `json:"sunriseTime"`
	SunsetTime  int64  `json:"sunsetTime"`
	// "moonPhase": 0.9,
	// "precipIntensity": 0.3226,
	// "precipIntensityMax": 0.7341,
	// "precipIntensityMaxTime": 1551600000,
	PrecipProbability float32 `json:"precipProbability"`
	PrecipType        string  `json:"precipType"`
	TemperatureHigh   float32 `json:"temperatureHigh"`
	// "temperatureHighTime": 1551628800,
	ApparentTemperatureHigh float32 `json:"apparentTemperatureHigh"`
	// "apparentTemperatureHighTime": 1551628800,
	TemperatureLow float32 `json:"temperatureLow"`
	// "temperatureLowTime": 1551679200,
	ApparentTemperatureLow float32 `json:"apparentTemperatureLow"`
	// "apparentTemperatureLowTime": 1551682800,
	DewPoint    float32 `json:"dewPoint"`
	Humidity    float32 `json:"humidity"`
	Pressure    float32 `json:"pressure"`
	WindSpeed   float32 `json:"windSpeed"`
	WindGust    float32 `json:"windGust"`
	WindBearing float32 `json:"windBearing"`
	CloudCover  float32 `json:"cloudCover"`
	UVIndex     uint8   `json:"uvIndex"`
	// "uvIndexTime": 1551610800,
	Visibility float32 `json:"visibility"`
	Ozone      float32 `json:"ozone"`
	// "temperatureMin": 9.58,
	// "temperatureMinTime": 1551654000,
	// "temperatureMax": 13.27,
	// "temperatureMaxTime": 1551628800,
	// "apparentTemperatureMin": 6.13,
	// "apparentTemperatureMinTime": 1551654000,
	// "apparentTemperatureMax": 13.27,
	// "apparentTemperatureMaxTime": 1551628800
}

func (d DayWeather) ToDomain() weather.DayWeather {
	return weather.DayWeather{
		TemperatureHigh: d.TemperatureHigh,
		TemperatureLow:  d.TemperatureLow,
		SunriseTime:     time.Unix(d.SunriseTime, 0),
		SunsetTime:      time.Unix(d.SunsetTime, 0),
		Rainy:           d.PrecipProbability > 0.25,
		Windy:           d.WindSpeed > 10 || d.WindGust > 20,
		Cloudy:          d.CloudCover > 0.6,
		Humidity:        d.Humidity,
	}
}

type Client struct {
	apiKey string
	client *http.Client
}

func New(apiKey string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	return &Client{
		client: client,
		apiKey: apiKey,
	}
}

func (c *Client) GetTwoDayWeatherAtLocation(ctx context.Context, lat, lon string) (weather.TwoDayWeather, error) {
	resp, err := c.getDailyWeather(ctx, lat, lon)
	if err != nil {
		return weather.TwoDayWeather{}, err
	}

	return weather.TwoDayWeather{
		Today:    resp.Days[0].ToDomain(),
		Tomorrow: resp.Days[1].ToDomain(),
	}, nil
}

func (c *Client) getForecast(ctx context.Context, path string, params *url.Values) (f *Forecast, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.darksky.net/forecast/"+c.apiKey+path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 199 {
		err = fmt.Errorf("unhandled_status %d", resp.StatusCode)
	} else if resp.StatusCode > 399 {
		err = errors.New(resp.Status)
	}

	if err != nil {
		return
	}

	f = &Forecast{}
	err = json.Unmarshal(body, f)
	if err != nil {
		return nil, err
	}

	return
}

//nolint:staticcheck
func (c *Client) getCurrentWeather(ctx context.Context, lat, lon string) (cw CurrentWeather, err error) {
	f, err := c.getForecast(ctx, "/"+lat+","+lon, &url.Values{
		"exclude": {"minutely,hourly,daily,alerts,flags"},
		"units":   {"uk2"},
	})
	if err != nil {
		return
	}

	return f.Current, nil
}

//nolint:staticcheck
func (c *Client) getHourlyWeather(ctx context.Context, lat, lon string) (hw HourlyWeather, err error) {
	f, err := c.getForecast(ctx, "/"+lat+","+lon, &url.Values{
		"exclude": {"currently,minutely,daily,alerts,flags"},
		"units":   {"uk2"},
	})
	if err != nil {
		return
	}

	return f.Hourly, nil
}

func (c *Client) getDailyWeather(ctx context.Context, lat, lon string) (dw DailyWeather, err error) {
	f, err := c.getForecast(ctx, "/"+lat+","+lon, &url.Values{
		"exclude": {"currently,minutely,hourly,alerts,flags"},
		"units":   {"uk2"},
	})
	if err != nil {
		return
	}

	return f.Daily, nil
}
