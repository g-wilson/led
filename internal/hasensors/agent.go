package hasensors

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/g-wilson/led/internal/homeassistant"
)

// AreaSensors represents an area and the current state of its sensors.
type AreaSensors struct {
	Area    string
	Sensors []SensorState
}

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

const refreshInterval = 1 * time.Minute

// StateProvider abstracts the Home Assistant API client.
type StateProvider interface {
	GetState(entityID string) (homeassistant.StateResponse, error)
	RunTemplateAreaSensors() ([]homeassistant.AreaSensorsResponse, error)
}

type Agent struct {
	client    StateProvider
	entityIDs []string

	mu      sync.RWMutex
	sensors map[string]SensorState
	areas   []homeassistant.AreaSensorsResponse
}

func New(client StateProvider, entityIDs []string) (*Agent, error) {
	if len(entityIDs) == 0 {
		return nil, fmt.Errorf("at least one entity ID is required")
	}

	a := &Agent{
		client:    client,
		entityIDs: entityIDs,
		sensors:   make(map[string]SensorState),
	}

	a.fetchAreas()
	a.populateCache()

	go func() {
		ticker := time.NewTicker(refreshInterval)
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

func (a *Agent) GetAreas() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	areas := make([]string, 0, len(a.areas))
	for _, ag := range a.areas {
		areas = append(areas, ag.Area)
	}
	return areas
}

func (a *Agent) GetArea(area string) (AreaSensors, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, ag := range a.areas {
		if ag.Area != area {
			continue
		}
		as := AreaSensors{
			Area:    ag.Area,
			Sensors: make([]SensorState, 0, len(ag.Entities)),
		}
		for _, eid := range ag.Entities {
			if s, ok := a.sensors[eid]; ok {
				as.Sensors = append(as.Sensors, s)
			}
		}
		return as, true
	}

	return AreaSensors{}, false
}

func (a *Agent) fetchAreas() {
	log.Println("fetching HA area groupings")

	allAreas, err := a.client.RunTemplateAreaSensors()
	if err != nil {
		log.Println(fmt.Errorf("error fetching HA areas: %w", err))
		return
	}

	// Build a set of configured entity IDs for fast lookup
	configured := make(map[string]bool, len(a.entityIDs))
	for _, id := range a.entityIDs {
		configured[id] = true
	}

	// Filter each area's entities to only those in our configured list,
	// tracking which configured entities were found in an area.
	assigned := make(map[string]bool)
	var filtered []homeassistant.AreaSensorsResponse

	for _, ag := range allAreas {
		var matched []string
		for _, eid := range ag.Entities {
			if configured[eid] {
				matched = append(matched, eid)
				assigned[eid] = true
			}
		}
		if len(matched) > 0 {
			filtered = append(filtered, homeassistant.AreaSensorsResponse{
				Area:     ag.Area,
				Entities: matched,
			})
		}
	}

	// Log an error for each configured entity not found in any area and
	// exclude it from future polling.
	var validIDs []string
	for _, id := range a.entityIDs {
		if assigned[id] {
			validIDs = append(validIDs, id)
		} else {
			log.Println(fmt.Errorf("HA sensor %s not found in any area, skipping", id))
		}
	}

	a.areas = filtered
	a.entityIDs = validIDs
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
