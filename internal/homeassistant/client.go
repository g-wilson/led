package homeassistant

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

func New(baseURL, token string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	return &Client{
		baseURL: baseURL,
		token:   token,
		client:  client,
	}
}

func (c *Client) GetState(entityID string) (StateResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/states/%s", c.baseURL, entityID), nil)
	if err != nil {
		return StateResponse{}, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return StateResponse{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StateResponse{}, err
	}

	switch {
	case resp.StatusCode > 499:
		err = errors.New(resp.Status)
	case resp.StatusCode > 399:
		err = fmt.Errorf("home assistant API error: %s", resp.Status)
	case resp.StatusCode < 200:
		err = fmt.Errorf("unhandled_status %d", resp.StatusCode)
	}
	if err != nil {
		return StateResponse{}, err
	}

	var state StateResponse
	err = json.Unmarshal(body, &state)
	if err != nil {
		return StateResponse{}, err
	}

	return state, nil
}
