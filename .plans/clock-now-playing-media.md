# Plan: Now Playing Media Screen

## Context

The LED clock cycles through "pages" on a 5-second ticker. It already integrates with Home Assistant for sensor data (temperature etc. grouped by area). This adds a new page that shows currently-playing media from HA media player entities — useful for seeing what's on a speaker or TV from across the room. When nothing is playing the page still shows but indicates "Nothing playing" rather than skipping entirely, so the page cycle stays predictable.

## HA Connection

Reuse the existing REST API approach (`homeassistant.Client.GetState()`). WebSocket subscriptions would give real-time updates but add significant complexity and are inconsistent with all existing data sources. Polling uses an adaptive interval: 10 seconds while something is playing (to catch track changes quickly), 60 seconds when idle (to reduce unnecessary requests).

## Files to Create / Modify

| Action | Path |
|--------|------|
| Create | `internal/hamediaplayer/agent.go` |
| Create | `clock/pages_mediaplayer.go` |
| Rename | `clock/pages_homeassistant.go` → `clock/pages_hasensors.go` |
| Modify | `config/config.go` |
| Modify | `clock/clock.go` |

## Configuration

Add one field to `config.Settings` (after `HASensors`):

```go
HAMediaPlayers []string `env:"HA_MEDIA_PLAYERS" envSeparator:","`
```

Reuses existing `HA_URL` and `HA_TOKEN`. The media page is only created when all three are set. Example `.env`:

```
HA_URL=http://192.168.1.100:8123
HA_TOKEN=<long-lived-token>
HA_MEDIA_PLAYERS=media_player.living_room_speaker,media_player.kitchen_speaker
```

## Step 1 — `internal/hamediaplayer/agent.go`

Mirrors `internal/hasensors/agent.go` in structure.

**Domain type:**
```go
type MediaPlayerState struct {
    EntityID     string
    FriendlyName string
    State        string   // "playing", "paused", "idle", "off", "unavailable"
    ContentType  string   // "music", "tvshow", "movie", ""
    MediaTitle   string
    MediaArtist  string
    MediaAlbum   string
}
```

**Minimal StateProvider interface** (only GetState, no area template needed):
```go
type StateProvider interface {
    GetState(ctx context.Context, entityID string) (homeassistant.StateResponse, error)
}
```

**Adaptive polling intervals:**
```go
const (
    refreshIntervalPlaying = 10 * time.Second
    refreshIntervalIdle    = 60 * time.Second
    populateCacheTimeout   = 20 * time.Second
)
```

After each `populateCache()` call, check whether any player is currently playing and select the next tick interval accordingly. The goroutine uses `time.NewTimer` (reset after each tick) rather than `time.NewTicker` so the interval can vary dynamically.

**Agent struct:** `ctx`, `client StateProvider`, `entityIDs []string`, `mu sync.RWMutex`, `players map[string]MediaPlayerState`

**New():** validates entityIDs, calls `populateCache()` synchronously, starts adaptive polling goroutine using `time.NewTimer` (not `time.NewTicker` so the interval can change each cycle). After each poll, checks `GetPlayingPlayer()` to choose the next interval. Returns `(*Agent, error)` matching hasensors signature.

```go
go func() {
    for {
        interval := refreshIntervalIdle
        if _, ok := a.GetPlayingPlayer(); ok {
            interval = refreshIntervalPlaying
        }
        timer := time.NewTimer(interval)
        select {
        case <-ctx.Done():
            timer.Stop()
            return
        case <-timer.C:
            a.populateCache()
        }
    }
}()
```

**populateCache():** iterates entityIDs, calls `client.GetState()` with timeout context, calls `toDomain()`, updates map under write lock per entity.

**toDomain():** extracts state from `resp.State`; extracts `friendly_name`, `media_content_type`, `media_title`, `media_artist`, `media_album_name` from `resp.Attributes` via type-assertion to string with ok-check:
```go
strAttr := func(key string) string {
    v, _ := resp.Attributes[key].(string)
    return v
}
```

**Convenience method on MediaPlayerState:**
```go
func (m MediaPlayerState) IsPlaying() bool {
    return m.State == "playing"
}
```

**Public accessor:**
- `GetPlayingPlayer() (MediaPlayerState, bool)` — returns first entity (in configured order, not map order) where `State == "playing"`. Iterates `a.entityIDs` (not `a.players`) to preserve priority order.

No `GetPausedPlayer()` — paused content is not shown.

## Step 2 — `clock/pages_mediaplayer.go`

**Layout on 64×32 display** (font is 4px/char → 16 chars per row):

Use the `huegradient` package for media field colours, matching the pattern in `pages_hasensors.go`. Define a gradient with a warm starting hue and wide step so the three rows are visually distinct:

```go
var mediaGradient = huegradient.Gradient{BaseHue: 160, Step: 110}
```

This yields (at Oklch L=0.75, C=0.12): teal ≈160° for artist (i=0), blue-violet ≈270° for title (i=1), warm orange ≈20° for album (i=2). Three perceptually-uniform colours with strong hue separation.

| Y offset | Content | Color |
|----------|---------|-------|
| Y=5 | "Now Playing" | Pink (215,0,88) — same pink as all page titles |
| Y=12 | Artist name | `mediaGradient.Color(0)` — teal |
| Y=18 | Track title | `mediaGradient.Color(1)` — blue-violet |
| Y=24 | Album name | `mediaGradient.Color(2)` — warm orange |
| Y=12 | "Nothing playing" | Muted (100,100,100) when idle or paused |

**truncate16() helper:**
```go
func truncate16(s string) string {
    r := []rune(s)
    if len(r) > 16 { return string(r[:16]) }
    return s
}
```

**renderNowPlaying() method on ClockRenderer:**
1. Draw "Now Playing" title at Y=5 in pink
2. Call `r.mediaPlayer.GetPlayingPlayer()` → if found, render fields with active colors
3. Else render "Nothing playing" at Y=12 in muted gray (paused state is not shown)

Each field (artist, title, album) is checked for non-empty before rendering — TV/podcast content may only have a title.

**renderMediaInfo() private helper:** renders each non-empty field using `mediaGradient.Color(i)` (i=0 artist, i=1 title, i=2 album). Import `huegradient` the same way `pages_hasensors.go` does.

## Step 3 — `clock/clock.go` Changes

**Add field to ClockRenderer struct:**
```go
mediaPlayer *hamediaplayer.Agent
```

**Add import:** `"github.com/g-wilson/led/internal/hamediaplayer"`

**Add Phase 3 block in `New()`** after Phase 2 (HA sensors) block, before `startPageIterator`:
```go
if cfg.HAURL != "" && cfg.HAToken != "" && len(cfg.HAMediaPlayers) > 0 {
    haClient := homeassistant.New(cfg.HAURL, cfg.HAToken, nil)
    mediaAgent, err := hamediaplayer.New(ctx, haClient, cfg.HAMediaPlayers)
    if err != nil {
        log.Printf("media player agent unavailable, skipping now playing page: %v", err)
    } else {
        r.mediaPlayer = mediaAgent
        r.pages = append(r.pages, r.renderNowPlaying)
    }
}
```

Note: if both `HA_SENSORS` and `HA_MEDIA_PLAYERS` are set, two separate `*homeassistant.Client` instances are created — this is intentional (stateless HTTP wrappers, low poll frequency, no shared state issues) and consistent with the existing codebase style.

## Rename Existing HA Sensors Page

Rename `clock/pages_homeassistant.go` → `clock/pages_hasensors.go`. No code changes needed beyond the filename — the function names (`renderArea`, `shortenSensorName`) and the `sensorGradient` var are already sensor-specific. This is a `git mv` operation.

## Edge Cases

- **TV/podcast (no artist):** `MediaArtist` is empty → Y=12 row is skipped, only title shows at Y=18
- **Paused content:** not shown — treated the same as idle/off, shows "Nothing playing"
- **Multiple players active:** `GetPlayingPlayer()` returns first in configured order — `HA_MEDIA_PLAYERS` order acts as a priority list; if two are playing simultaneously only the first is shown
- **Cold start / cache empty:** "Nothing playing" shows until first poll completes (same behaviour as HA sensors)
- **Only sensors configured, not media players:** `len(cfg.HAMediaPlayers) == 0` skips Phase 3 entirely

## Verification

1. Build with `go build ./...` — confirms no compile errors
2. Run `cmd/debug/main.go` with `HA_URL`, `HA_TOKEN`, and `HA_MEDIA_PLAYERS` set — produces `output.png` for each frame
3. Set a media player to "playing" in HA, wait 10s, verify `output.png` shows artist/title
4. Stop media, wait 60s, verify "Nothing playing" renders
5. Build without `HA_MEDIA_PLAYERS` set — verify the page is absent from rotation
