package homeassistant

import (
	"bytes"
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

const areaSensorsTemplate = `{%- set ns = namespace(result=[]) -%}{%- for aid in areas() -%}  {%- set sensors = area_entities(aid)        | select('match', '^sensor\\.')        | list -%}  {%- if sensors -%}    {%- set ns.result = ns.result + [        {"area": area_name(aid), "entities": sensors}    ] -%}  {%- endif -%}{%- endfor -%}{{ ns.result | to_json }}`

func (c *Client) RunTemplateAreaSensors() ([]AreaSensorsResponse, error) {
	payload, err := json.Marshal(map[string]string{"template": areaSensorsTemplate})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/template", c.baseURL), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch {
	case resp.StatusCode >= 500:
		err = errors.New(resp.Status)
	case resp.StatusCode >= 400:
		err = fmt.Errorf("home assistant API error: %s", resp.Status)
	case resp.StatusCode < 200:
		err = fmt.Errorf("unhandled_status %d", resp.StatusCode)
	}
	if err != nil {
		return nil, err
	}

	var areas []AreaSensorsResponse
	if err := json.Unmarshal(body, &areas); err != nil {
		return nil, fmt.Errorf("error parsing area sensors response: %w", err)
	}

	return areas, nil
}

func (c *Client) GetState(entityID string) (StateResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/states/%s", c.baseURL, entityID), nil)
	if err != nil {
		return StateResponse{}, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)

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
	case resp.StatusCode >= 500:
		err = errors.New(resp.Status)
	case resp.StatusCode >= 400:
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
