# Plan: Dynamic Clock Pages

## Overview

Refactor the clock package to support dynamic pages alongside the existing static ones. Each Home Assistant area returned by the sensors agent at startup becomes an additional page in the rotation.

The core change is replacing the `pages []string` / switch-statement pattern with a slice of functions. Every page — static or dynamic — is a `func(*image.RGBA) error`. Dynamic area pages are closures that capture the area name at construction time and call a shared render method.

## Approach

A named function type replaces the string slice:

```go
type page func(c *image.RGBA) error
```

All pages become values of this type. Static pages are method expressions on `ClockRenderer`. Dynamic pages are closures constructed in `New()`, each capturing one area name and delegating to `r.renderArea`.

`DrawFrame()` replaces its switch statement with a single call:

```go
return r.pages[r.currentPage](c)
```

## Changes

### `clock/clock.go`

#### Struct

Add the `page` type alias and two fields to `ClockRenderer`:

```go
type page func(c *image.RGBA) error

type ClockRenderer struct {
    font         *fopix.Drawer
    weather      *weather.Agent
    diagnostics  *diagnostics.Agent
    sensors      *hasensors.Agent   // new — nil if HA env vars not set
    location     *time.Location
    pages        []page             // was []string
    currentPage  int
    pageInterval time.Duration
    debug        bool
}
```

#### `New()`

Two-phase page construction after the existing agent setup:

```go
// Phase 1: static pages
r.pages = []page{
    r.renderToday,
    r.renderTomorrow,
    r.renderDaylight,
    r.renderCountdown,
    r.renderDiag,
}

// Phase 2: dynamic area pages (skipped entirely if HA env vars not set)
haURL := os.Getenv("HA_URL")
haToken := os.Getenv("HA_TOKEN")
haEntityIDs := strings.Split(os.Getenv("HA_SENSORS"), ",")

if haURL != "" && haToken != "" && len(haEntityIDs) > 0 {
    haClient := homeassistant.New(haURL, haToken, nil)
    sensorsAgent, err := hasensors.New(haClient, haEntityIDs)
    if err != nil {
        log.Printf("sensors agent unavailable, skipping area pages: %v", err)
    } else {
        r.sensors = sensorsAgent
        for _, areaName := range sensorsAgent.GetAreas() {
            areaName := areaName
            r.pages = append(r.pages, func(c *image.RGBA) error {
                return r.renderArea(c, areaName)
            })
        }
    }
}
```

`hasensors.New()` calls `fetchAreas()` synchronously before returning, so area names are immediately available. If `fetchAreas()` failed internally, `GetAreas()` returns an empty slice and the loop adds nothing — static pages are unaffected.

### `internal/hasensors/agent.go`

Replace `GetSensorsByArea() []AreaSensors` with two focused methods:

```go
// GetAreas returns the names of all configured areas.
func (a *Agent) GetAreas() []string {
    a.mu.RLock()
    defer a.mu.RUnlock()

    areas := make([]string, 0, len(a.areas))
    for _, ag := range a.areas {
        areas = append(areas, ag.Area)
    }
    return areas
}

// GetArea returns the current sensor states for a single area.
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
```

`GetSensorsByArea() []AreaSensors` is removed — no caller needs area names and sensor states together at the same time.

#### `DrawFrame()`

Remove the switch statement. The page render call becomes:

```go
return r.pages[r.currentPage](c)
```

#### Static page methods

Extract each switch case to a named method. Signatures must match the `page` type:

```go
func (r *ClockRenderer) renderToday(c *image.RGBA) error { ... }
func (r *ClockRenderer) renderTomorrow(c *image.RGBA) error { ... }
func (r *ClockRenderer) renderDaylight(c *image.RGBA) error { ... }
func (r *ClockRenderer) renderCountdown(c *image.RGBA) error { ... }
func (r *ClockRenderer) renderDiag(c *image.RGBA) error { ... }
```

The method bodies are the existing switch case bodies verbatim, with `return nil` appended.

#### Area render method (simple placeholder)

```go
func (r *ClockRenderer) renderArea(c *image.RGBA, area string) error {
    r.addText(c, image.Point{X: 0, Y: 8}, area, color.RGBA{215, 0, 88, 255})

    if as, ok := r.sensors.GetArea(area); ok {
        for i, s := range as.Sensors {
            y := 16 + (i * 8)
            r.addText(c, image.Point{X: 0, Y: y}, fmt.Sprintf("%s %s%s", s.Name, s.State, s.Unit), color.RGBA{200, 200, 200, 255})
        }
    }

    return nil
}
```

This is intentionally minimal. The layout will be iterated in a follow-up.

## Files Modified

1. `internal/hasensors/agent.go` — replace `GetSensorsByArea` with `GetAreas` and `GetArea`
2. `clock/clock.go` — all clock changes

## Files Created

None.

## Design Notes

- `startPageIterator` is unchanged — it indexes into `r.pages` regardless of element type.
- The `sensors` field is nil when HA env vars are absent. `renderArea` is only ever called via closures constructed when `sensors` is non-nil, so there is no nil dereference risk.
- Area pages are appended after static pages. The existing page order is preserved.
- If a future phase requires area pages to be refreshed at runtime (e.g. new HA areas added), that would require rebuilding the `pages` slice under a lock. That is out of scope here.
