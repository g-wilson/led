# Plan: Home Assistant Sensor Integration

## Overview

Add support for reading sensor data from a Home Assistant server via its REST API. This introduces two new packages following the existing weather/tomorrowio pattern:

- `internal/homeassistant` — HTTP client for the HA REST API
- `internal/hasensors` — Background-polling agent that caches sensor states

The clock package will not be modified in this phase.

## Architecture

```
.env                          (HA_URL, HA_TOKEN, HA_SENSORS, HA_REFRESH)
  │
  ▼
clock/clock.go                (reads env, constructs client + agent)
  │
  ├─► internal/homeassistant/
  │     ├── client.go          (HTTP client: GET /api/states/<entity_id>)
  │     └── types.go           (API response types)
  │
  └─► internal/hasensors/
        └── agent.go           (background poller, exposes sensor data)
```

## Sensor Configuration

Sensor entity IDs will be configured via a single comma-separated environment variable:

```
HA_SENSORS=sensor.living_room_temperature,sensor.living_room_humidity,sensor.electricity_usage
```

This keeps configuration simple and consistent with the existing env-var pattern. The agent will split this string on commas to get the list of entity IDs to poll.

## Environment Variables

| Variable | Required | Example | Description |
|----------|----------|---------|-------------|
| `HA_URL` | Yes | `http://192.168.1.100:8123` | Base URL of the HA instance |
| `HA_TOKEN` | Yes | `eyJ0eXAi...` | Long-lived access token |
| `HA_SENSORS` | Yes | `sensor.temp,sensor.humidity` | Comma-separated entity IDs |

## Package 1: `internal/homeassistant`

### Files

#### `internal/homeassistant/client.go`

HTTP client for the HA REST API. Follows the same structure as `internal/tomorrowio/client.go`.

```go
package homeassistant

type Client struct {
    baseURL string
    token   string
    client  *http.Client
}

func New(baseURL, token string, client *http.Client) *Client
func (c *Client) GetState(entityID string) (StateResponse, error)
```

- `New` — accepts base URL, token, and optional `*http.Client` (defaults to 10s timeout, same as tomorrowio)
- `GetState` — performs `GET {baseURL}/api/states/{entityID}` with `Authorization: Bearer {token}` header. Parses JSON response into `StateResponse`. Handles HTTP error codes following the same pattern as tomorrowio (5xx → generic error, 4xx → parsed error or status text).

#### `internal/homeassistant/types.go`

API response types matching the HA REST API JSON schema.

```go
package homeassistant

type StateResponse struct {
    EntityID    string                 `json:"entity_id"`
    State       string                 `json:"state"`
    Attributes  map[string]interface{} `json:"attributes"`
    LastChanged string                 `json:"last_changed"`
    LastUpdated string                 `json:"last_updated"`
}
```

`Attributes` is `map[string]interface{}` because HA sensor attributes vary per entity type — they can contain strings, numbers, booleans, and nested objects. This is intentionally untyped at the API layer.

## Package 2: `internal/hasensors`

### Files

#### `internal/hasensors/agent.go`

Background-polling agent. Follows the same Agent pattern as `internal/weather/agent.go` and `internal/diagnostics/agent.go`.

### Domain Types

```go
package hasensors

// Measurement represents a single key/value/unit tuple from a sensor.
type Measurement struct {
    Key   string  // attribute name, e.g. "temperature"
    Value string  // attribute value as string
    Unit  string  // unit from unit_of_measurement, or empty
}

// SensorState represents the current state of a single HA sensor.
type SensorState struct {
    EntityID     string        // e.g. "sensor.living_room_temperature"
    Name         string        // friendly_name from attributes
    State        string        // primary state value
    Unit         string        // unit_of_measurement from attributes (for the primary state)
    Attributes   map[string]string // all non-meta attributes as string key-value pairs
    Measurements []Measurement // derived list of all measurements (state + attributes)
    LastUpdated  string        // from HA API
}
```

The `Measurements` slice provides a uniform way to iterate over all data a sensor exposes:
- The first entry is always the primary state (Key = entity_id suffix after the dot, Value = state, Unit = unit_of_measurement)
- Subsequent entries come from the attributes map, excluding meta-attributes like `friendly_name`, `unit_of_measurement`, `icon`, `device_class`, `state_class`, and `entity_picture`

### Provider Interface

```go
type StateProvider interface {
    GetState(entityID string) (homeassistant.StateResponse, error)
}
```

This decouples the agent from the concrete HTTP client, following the `DayWeatherProvider` pattern.

### Agent

```go
type Agent struct {
    client    StateProvider
    entityIDs []string
    mu        sync.RWMutex
    sensors   map[string]SensorState // keyed by entity_id
}

type AgentOptions struct {
    EntityIDs []string // parsed from HA_SENSORS
    Refresh   int      // seconds between polls
}

func New(client StateProvider, options AgentOptions) (*Agent, error)
func (a *Agent) GetSensor(entityID string) (SensorState, bool)
func (a *Agent) GetAllSensors() []SensorState
```

- `New` — validates options, performs initial fetch of all sensors, starts background goroutine with `time.NewTicker`
- `populateCache` — iterates over `entityIDs`, calls `client.GetState()` for each, converts `StateResponse` to `SensorState` using a `toDomain` conversion function, stores in map under write lock. Errors are logged but don't stop other sensors from being fetched.
- `GetSensor` — returns a copy of `SensorState` for a given entity ID plus a `bool` indicating if it was found. Thread-safe via `RLock`.
- `GetAllSensors` — returns a slice of all cached `SensorState` values. Thread-safe via `RLock`.

### Domain Conversion

A `toDomain(entityID string, resp homeassistant.StateResponse) SensorState` function converts the raw API response to the domain type:

1. Extracts `friendly_name` from attributes (falls back to entity_id if missing)
2. Extracts `unit_of_measurement` from attributes
3. Builds the `Measurements` slice:
   - First entry: primary state value with unit
   - Then iterates remaining attributes, skipping meta-attributes (`friendly_name`, `unit_of_measurement`, `icon`, `device_class`, `state_class`, `entity_picture`), and converts each value to string via `fmt.Sprintf`
4. Builds `Attributes` map with all non-meta attributes as strings

## Integration Point (clock package — future phase)

For context on how this will eventually be consumed (not implemented now):

```go
// in clock.go New():
haClient := homeassistant.New(os.Getenv("HA_URL"), os.Getenv("HA_TOKEN"), nil)
haEntityIDs := strings.Split(os.Getenv("HA_SENSORS"), ",")
haRefresh, _ := strconv.ParseInt(os.Getenv("HA_REFRESH"), 10, 32)
if haRefresh == 0 { haRefresh = 60 }
haSensorsAgent, err := hasensors.New(haClient, hasensors.AgentOptions{
    EntityIDs: haEntityIDs,
    Refresh:   int(haRefresh),
})
```

The `ClockRenderer` struct would gain a `sensors *hasensors.Agent` field, and new pages would call `r.sensors.GetSensor(id)` or `r.sensors.GetAllSensors()` to render sensor data. But that's out of scope for this phase.

## Files to Create

1. `internal/homeassistant/client.go` — HA REST API HTTP client
2. `internal/homeassistant/types.go` — API response structs
3. `internal/hasensors/agent.go` — background polling agent with domain types

## Files to Modify

None. The clock package integration is deferred to a future phase.

## Implementation Order

1. Create `internal/homeassistant/types.go` (API response types)
2. Create `internal/homeassistant/client.go` (HTTP client)
3. Create `internal/hasensors/agent.go` (domain types, provider interface, agent, domain conversion)
4. Verify the project compiles with `go build ./...`
