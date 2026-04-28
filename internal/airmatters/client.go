package airmatters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
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

func (c *Client) GetNearbyAirCondition(ctx context.Context, lat, lon string) (AirCondition, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.air-matters.app/nearby_air_condition", nil)
	if err != nil {
		return AirCondition{}, err
	}

	req.URL.RawQuery = url.Values{
		"lat":      []string{lat},
		"lon":      []string{lon},
		"lang":     []string{"en"},
		"standard": []string{"aqi_us"},
	}.Encode()

	req.Header.Add("Authorization", c.apiKey)
	req.Header.Add("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return AirCondition{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AirCondition{}, err
	}

	switch {
	case resp.StatusCode > 499:
		return AirCondition{}, errors.New(resp.Status)
	case resp.StatusCode >= 400:
		return AirCondition{}, fmt.Errorf("air matters api error: %s", resp.Status)
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return AirCondition{}, err
	}

	return r.toDomain(), nil
}

func (r Response) toDomain() AirCondition {
	ac := AirCondition{
		PlaceName: r.Place.Name,
	}
	for _, reading := range r.Latest.Readings {
		switch reading.Kind {
		case "aqi":
			ac.AQI = ReadingData{
				Value: reading.Value,
				Level: reading.Level,
				Color: parseHexColor(reading.Color),
			}
		case "pm25":
			ac.PM25 = ReadingData{
				Value: reading.Value,
				Level: reading.Level,
				Color: parseHexColor(reading.Color),
			}
		}
	}
	return ac
}

func parseHexColor(s string) color.RGBA {
	if len(s) != 7 || s[0] != '#' {
		return color.RGBA{200, 200, 200, 255}
	}
	r, err1 := strconv.ParseUint(s[1:3], 16, 8)
	g, err2 := strconv.ParseUint(s[3:5], 16, 8)
	b, err3 := strconv.ParseUint(s[5:7], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return color.RGBA{200, 200, 200, 255}
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}
