package airmatters

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const fetchTimeout = 15 * time.Second

type Agent struct {
	ctx     context.Context
	client  *Client
	options AgentOptions

	mu   sync.RWMutex
	data AirCondition
}

type AgentOptions struct {
	Latitude  string
	Longitude string
	Refresh   int
}

func NewAgent(ctx context.Context, client *Client, options AgentOptions) (*Agent, error) {
	a := &Agent{
		ctx:     ctx,
		client:  client,
		options: options,
	}

	if err := a.populateCache(); err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(time.Duration(options.Refresh) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.populateCache(); err != nil {
					log.Println(fmt.Errorf("error fetching air quality: %w", err))
				}
			}
		}
	}()

	return a, nil
}

func (a *Agent) populateCache() error {
	log.Println("fetching air quality")

	ctx, cancel := context.WithTimeout(a.ctx, fetchTimeout)
	defer cancel()

	data, err := a.client.GetNearbyAirCondition(ctx, a.options.Latitude, a.options.Longitude)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.data = data

	return nil
}

func (a *Agent) Get() AirCondition {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.data
}
