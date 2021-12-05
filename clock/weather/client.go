package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
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
	Summary string `json:"summary"`
	Icon    string `json:"icon"`
	// "sunriseTime": 1551595369,
	// "sunsetTime": 1551635074,
	// "moonPhase": 0.9,
	// "precipIntensity": 0.3226,
	// "precipIntensityMax": 0.7341,
	// "precipIntensityMaxTime": 1551600000,
	// "precipProbability": 0.98,
	// "precipType": "rain",
	// "temperatureHigh": 13.27,
	// "temperatureHighTime": 1551628800,
	// "temperatureLow": 5.01,
	// "temperatureLowTime": 1551679200,
	// "apparentTemperatureHigh": 13.27,
	// "apparentTemperatureHighTime": 1551628800,
	// "apparentTemperatureLow": 1.45,
	// "apparentTemperatureLowTime": 1551682800,
	TemperatureHigh         float32 `json:"temperatureHigh"`
	ApparentTemperatureHigh float32 `json:"apparentTemperatureHigh"`
	TemperatureLow          float32 `json:"temperatureLow"`
	ApparentTemperatureLow  float32 `json:"apparentTemperatureLow"`
	// "dewPoint": 8.73,
	// "humidity": 0.83,
	// "pressure": 998.94,
	// "windSpeed": 16.72,
	// "windGust": 45.56,
	// "windGustTime": 1551639600,
	// "windBearing": 225,
	// "cloudCover": 0.87,
	CloudCover float32 `json:"cloudCover"`
	// "uvIndex": 2,
	// "uvIndexTime": 1551610800,
	// "visibility": 7.28,
	// "ozone": 337.38,
	// "temperatureMin": 9.58,
	// "temperatureMinTime": 1551654000,
	// "temperatureMax": 13.27,
	// "temperatureMaxTime": 1551628800,
	// "apparentTemperatureMin": 6.13,
	// "apparentTemperatureMinTime": 1551654000,
	// "apparentTemperatureMax": 13.27,
	// "apparentTemperatureMaxTime": 1551628800
}

type DarkskyClient struct {
	apiKey string
	client *http.Client
}

func New(apiKey string, client *http.Client) *DarkskyClient {
	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	return &DarkskyClient{
		client: client,
		apiKey: apiKey,
	}
}

func (c *DarkskyClient) GetForecast(path string, params *url.Values) (f *Forecast, err error) {
	req, err := http.NewRequest("GET", "https://api.darksky.net/forecast/"+c.apiKey+path, nil)
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

func (c *DarkskyClient) GetCurrentWeather(lat, long string) (cw CurrentWeather, err error) {
	f, err := c.GetForecast("/"+lat+","+long, &url.Values{
		"exclude": {"minutely,hourly,daily,alerts,flags"},
		"units":   {"uk2"},
	})
	if err != nil {
		return
	}

	return f.Current, nil
}

func (c *DarkskyClient) GetHourlyWeather(lat, long string) (hw HourlyWeather, err error) {
	f, err := c.GetForecast("/"+lat+","+long, &url.Values{
		"exclude": {"currently,minutely,daily,alerts,flags"},
		"units":   {"uk2"},
	})
	if err != nil {
		return
	}

	return f.Hourly, nil
}

func (c *DarkskyClient) GetDailyWeather(lat, long string) (dw DailyWeather, err error) {
	f, err := c.GetForecast("/"+lat+","+long, &url.Values{
		"exclude": {"currently,minutely,hourly,alerts,flags"},
		"units":   {"uk2"},
	})
	if err != nil {
		return
	}

	return f.Daily, nil
}
