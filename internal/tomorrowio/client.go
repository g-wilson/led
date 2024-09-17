package tomorrowio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/g-wilson/led/internal/weather"
)

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

func (c *Client) GetTwoDayWeatherAtLocation(lat, lon string) (weather.TwoDayWeather, error) {
	resp, err := c.getForecast(&url.Values{
		"location":  []string{fmt.Sprintf("%s,%s", lat, lon)},
		"fields":    []string{"core"},
		"units":     []string{"metric"},
		"timesteps": []string{"1d"},
		"apikey":    []string{c.apiKey},
	})
	if err != nil {
		return weather.TwoDayWeather{}, err
	}

	return weather.TwoDayWeather{
		Today:    resp.Timelines.Daily[0].Values.ToDomain(),
		Tomorrow: resp.Timelines.Daily[1].Values.ToDomain(),
	}, nil
}

func (c *Client) getForecast(params *url.Values) (f *ForecastResponse, err error) {
	req, err := http.NewRequest("GET", "https://api.tomorrow.io/v4/weather/forecast", nil)
	if err != nil {
		return nil, err
	}
	if params != nil {
		req.URL.RawQuery = params.Encode()
	}
	req.Header.Add("accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch {
	case resp.StatusCode > 499:
		err = errors.New(resp.Status)
	case resp.StatusCode > 400:
		e := &ErrorResponse{}
		err2 := json.Unmarshal(body, e)
		if err2 != nil {
			err = errors.New(resp.Status)
		} else {
			err = errors.New(e.Message)
		}
	case resp.StatusCode < 199:
		err = fmt.Errorf("unhandled_status %d", resp.StatusCode)
	}
	if err != nil {
		return
	}

	f = &ForecastResponse{}
	err = json.Unmarshal(body, f)
	if err != nil {
		return nil, err
	}

	return
}
