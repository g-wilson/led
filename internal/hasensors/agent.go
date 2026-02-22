package hasensors

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/g-wilson/led/internal/homeassistant"
)

// Measurement represents a single key/value/unit tuple from a sensor.
type Measurement struct {
	Key   string
	Value string
	Unit  string
}

// SensorState represents the current state of a single HA sensor.
type SensorState struct {
	EntityID     string
	Name         string
	State        string
	Unit         string
	Attributes   map[string]string
	Measurements []Measurement
	LastUpdated  string
}

// StateProvider abstracts the Home Assistant API client.
type StateProvider interface {
	GetState(entityID string) (homeassistant.StateResponse, error)
}

type Agent struct {
	client    StateProvider
	entityIDs []string

	mu      sync.RWMutex
	sensors map[string]SensorState
}

type AgentOptions struct {
	EntityIDs []string
	Refresh   int
}

func New(client StateProvider, options AgentOptions) (*Agent, error) {
	if len(options.EntityIDs) == 0 {
		return nil, fmt.Errorf("at least one entity ID is required")
	}
	if options.Refresh <= 0 {
		return nil, fmt.Errorf("refresh interval must be positive")
	}

	a := &Agent{
		client:    client,
		entityIDs: options.EntityIDs,
		sensors:   make(map[string]SensorState),
	}

	a.populateCache()

	go func() {
		ticker := time.NewTicker(time.Duration(options.Refresh) * time.Second)
		for range ticker.C {
			a.populateCache()
		}
	}()

	return a, nil
}

func (a *Agent) populateCache() {
	log.Println("fetching HA sensors")

	for _, entityID := range a.entityIDs {
		resp, err := a.client.GetState(entityID)
		if err != nil {
			log.Println(fmt.Errorf("error fetching HA sensor %s: %w", entityID, err))
			continue
		}

		a.mu.Lock()
		a.sensors[entityID] = toDomain(entityID, resp)
		a.mu.Unlock()
	}
}

func (a *Agent) GetSensor(entityID string) (SensorState, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	s, ok := a.sensors[entityID]
	return s, ok
}

func (a *Agent) GetAllSensors() []SensorState {
	a.mu.RLock()
	defer a.mu.RUnlock()

	out := make([]SensorState, 0, len(a.sensors))
	for _, s := range a.sensors {
		out = append(out, s)
	}

	return out
}

// metaAttributes are attribute keys that are excluded from the Measurements
// and Attributes fields since they are represented by dedicated SensorState fields.
var metaAttributes = map[string]bool{
	"friendly_name":       true,
	"unit_of_measurement": true,
	"icon":                true,
	"device_class":        true,
	"state_class":         true,
	"entity_picture":      true,
}

func toDomain(entityID string, resp homeassistant.StateResponse) SensorState {
	name := entityID
	if fn, ok := resp.Attributes["friendly_name"].(string); ok {
		name = fn
	}

	unit := ""
	if u, ok := resp.Attributes["unit_of_measurement"].(string); ok {
		unit = u
	}

	// Primary measurement key: the portion after the dot in the entity ID
	primaryKey := entityID
	if parts := strings.SplitN(entityID, ".", 2); len(parts) == 2 {
		primaryKey = parts[1]
	}

	measurements := []Measurement{
		{Key: primaryKey, Value: resp.State, Unit: unit},
	}

	attrs := make(map[string]string)

	for k, v := range resp.Attributes {
		if metaAttributes[k] {
			continue
		}

		strVal := fmt.Sprintf("%v", v)
		attrs[k] = strVal

		measurements = append(measurements, Measurement{
			Key:   k,
			Value: strVal,
		})
	}

	return SensorState{
		EntityID:     entityID,
		Name:         name,
		State:        resp.State,
		Unit:         unit,
		Attributes:   attrs,
		Measurements: measurements,
		LastUpdated:  resp.LastUpdated,
	}
}
