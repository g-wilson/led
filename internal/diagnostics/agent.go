package diagnostics

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	pingInterval = 2 * time.Minute
	staleAfter   = 5 * time.Minute

	pingGreenMax  = 50 * time.Millisecond
	pingYellowMax = 100 * time.Millisecond
	pingOrangeMax = 200 * time.Millisecond
)

var (
	pingAddress = "8.8.8.8:53"
	pingTimeout = 2 * time.Second
	pingNetwork = "tcp"
)

type PingLevel int

const (
	PingLevelGreen PingLevel = iota
	PingLevelYellow
	PingLevelOrange
	PingLevelRed
)

type Status struct {
	LastHealthyAt time.Time
	LastPing      time.Duration
	LastPingOk    bool
	LastCheckedAt time.Time
}

func (s Status) PingLevel() PingLevel {
	if s.LastPing <= 0 {
		return PingLevelRed
	}
	if s.LastPing <= pingGreenMax {
		return PingLevelGreen
	}
	if s.LastPing <= pingYellowMax {
		return PingLevelYellow
	}
	if s.LastPing <= pingOrangeMax {
		return PingLevelOrange
	}
	return PingLevelRed
}

func (s Status) IsStale(now time.Time) bool {
	if s.LastHealthyAt.IsZero() {
		return true
	}

	return now.Sub(s.LastHealthyAt) > staleAfter
}

type Agent struct {
	mu            sync.RWMutex
	lastHealthyAt time.Time
	lastPing      time.Duration
	lastPingOk    bool
	lastCheckedAt time.Time
}

func New() (*Agent, error) {
	a := &Agent{}
	a.checkOnce()

	go func() {
		ticker := time.NewTicker(pingInterval)
		for range ticker.C {
			a.checkOnce()
		}
	}()

	return a, nil
}

func (a *Agent) GetStatus() Status {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return Status{
		LastHealthyAt: a.lastHealthyAt,
		LastPing:      a.lastPing,
		LastPingOk:    a.lastPingOk,
		LastCheckedAt: a.lastCheckedAt,
	}
}

func (a *Agent) checkOnce() {
	start := time.Now()
	dialer := net.Dialer{Timeout: pingTimeout}
	conn, err := dialer.Dial(pingNetwork, pingAddress)
	elapsed := time.Since(start)

	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastCheckedAt = time.Now()
	if err != nil {
		a.lastPingOk = false
		a.lastPing = 0
		log.Println(fmt.Errorf("diagnostics ping failed: %w", err))
		return
	}
	_ = conn.Close()

	a.lastPingOk = true
	a.lastPing = elapsed
	a.lastHealthyAt = a.lastCheckedAt
}
