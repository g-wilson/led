package hamediaplayer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/g-wilson/led/internal/homeassistant"
)

// MediaPlayerState represents the current state of a single HA media player entity.
type MediaPlayerState struct {
	EntityID     string
	FriendlyName string
	State        string // "playing", "paused", "idle", "off", "unavailable"
	ContentType  string // "music", "tvshow", "movie", ""
	MediaTitle   string
	MediaArtist  string
	MediaAlbum   string
}

// IsPlaying reports whether the player is actively playing media.
func (m MediaPlayerState) IsPlaying() bool {
	return m.State == "playing"
}

const (
	// refreshIntervalPlaying is the polling interval when at least one player is active.
	refreshIntervalPlaying = 10 * time.Second
	// refreshIntervalIdle is the polling interval when nothing is playing.
	refreshIntervalIdle    = 60 * time.Second
	populateCacheTimeout   = 20 * time.Second
)

// StateProvider abstracts the Home Assistant API client.
type StateProvider interface {
	GetState(ctx context.Context, entityID string) (homeassistant.StateResponse, error)
}

// Agent polls a set of HA media player entities and caches their state.
type Agent struct {
	ctx       context.Context
	client    StateProvider
	entityIDs []string

	mu      sync.RWMutex
	players map[string]MediaPlayerState
}

// New creates an Agent, performs an initial cache population, and starts the
// background polling goroutine. Polling frequency is adaptive: faster while
// something is playing, slower when idle.
func New(ctx context.Context, client StateProvider, entityIDs []string) (*Agent, error) {
	if len(entityIDs) == 0 {
		return nil, fmt.Errorf("at least one entity ID is required")
	}

	a := &Agent{
		ctx:       ctx,
		client:    client,
		entityIDs: entityIDs,
		players:   make(map[string]MediaPlayerState),
	}

	a.populateCache()

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

	return a, nil
}

func (a *Agent) populateCache() {
	log.Println("fetching HA media players")

	ctx, cancel := context.WithTimeout(a.ctx, populateCacheTimeout)
	defer cancel()

	for _, entityID := range a.entityIDs {
		resp, err := a.client.GetState(ctx, entityID)
		if err != nil {
			log.Printf("error fetching HA media player %s: %v", entityID, err)
			continue
		}

		a.mu.Lock()
		a.players[entityID] = toDomain(entityID, resp)
		a.mu.Unlock()
	}
}

// GetPlayingPlayer returns the first configured entity that is actively playing,
// in the order they were provided to New. Returns false if nothing is playing.
func (a *Agent) GetPlayingPlayer() (MediaPlayerState, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, id := range a.entityIDs {
		if p, ok := a.players[id]; ok && p.IsPlaying() {
			return p, true
		}
	}
	return MediaPlayerState{}, false
}

func toDomain(entityID string, resp homeassistant.StateResponse) MediaPlayerState {
	strAttr := func(key string) string {
		v, _ := resp.Attributes[key].(string)
		return v
	}

	return MediaPlayerState{
		EntityID:     entityID,
		State:        resp.State,
		FriendlyName: strAttr("friendly_name"),
		ContentType:  strAttr("media_content_type"),
		MediaTitle:   strAttr("media_title"),
		MediaArtist:  strAttr("media_artist"),
		MediaAlbum:   strAttr("media_album_name"),
	}
}
