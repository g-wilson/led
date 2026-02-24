package homeassistant

type StateResponse struct {
	EntityID    string         `json:"entity_id"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastChanged string         `json:"last_changed"`
	LastUpdated string         `json:"last_updated"`
}

type AreaSensorsResponse struct {
	Area     string   `json:"area"`
	Entities []string `json:"entities"`
}
