package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Backend represents a backend server with its metrics
type Backend struct {
	URL          *url.URL
	Alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy

	// Metrics
	RequestCount int64
	TotalLatency time.Duration
	StartTime    time.Time
}

// NewBackend creates a new backend instance
func NewBackend(urlStr string) (*Backend, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return &Backend{
		URL:          url,
		Alive:        true,
		ReverseProxy: httputil.NewSingleHostReverseProxy(url),
		StartTime:    time.Now(),
	}, nil
}

// SetAlive sets the alive status of the backend
func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

// IsAlive returns whether the backend is alive
func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

// AddRequest increments the request count and adds latency
func (b *Backend) AddRequest(latency time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.RequestCount++
	b.TotalLatency += latency
}

// GetMetrics returns a copy of the backend metrics
func (b *Backend) GetMetrics() BackendMetrics {
	b.mu.RLock()
	defer b.mu.RUnlock()

	uptime := time.Since(b.StartTime)
	var avgLatency time.Duration
	if b.RequestCount > 0 {
		avgLatency = b.TotalLatency / time.Duration(b.RequestCount)
	}

	return BackendMetrics{
		URL:          b.URL.String(),
		Alive:        b.Alive,
		RequestCount: b.RequestCount,
		AvgLatency:   avgLatency,
		Uptime:       uptime,
	}
}

// BackendMetrics represents the metrics for a backend
type BackendMetrics struct {
	URL          string        `json:"url"`
	Alive        bool          `json:"alive"`
	RequestCount int64         `json:"request_count"`
	AvgLatency   time.Duration `json:"avg_latency_ns"`
	Uptime       time.Duration `json:"uptime_ns"`
}
