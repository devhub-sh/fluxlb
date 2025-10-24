package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancer manages the backend servers and routing
type LoadBalancer struct {
	backends      []*Backend
	current       uint64
	healthChecker *HealthChecker
	config        *Config
	mu            sync.RWMutex
}

// NewLoadBalancer creates a new load balancer instance
func NewLoadBalancer(config *Config) (*LoadBalancer, error) {
	backends := make([]*Backend, 0, len(config.Backends))

	for _, bc := range config.Backends {
		backend, err := NewBackend(bc.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to create backend %s: %w", bc.URL, err)
		}
		backends = append(backends, backend)
		log.Printf("Added backend: %s", bc.URL)
	}

	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends configured")
	}

	healthChecker := NewHealthChecker(backends, config.HealthCheckPath, config.HealthCheckInterval)

	return &LoadBalancer{
		backends:      backends,
		healthChecker: healthChecker,
		config:        config,
	}, nil
}

// Start starts the load balancer and health checker
func (lb *LoadBalancer) Start(ctx context.Context) {
	go lb.healthChecker.Start(ctx)
}

// GetNextBackend returns the next available backend using smart round-robin
// with smallest time-quanta-based scheduling
func (lb *LoadBalancer) GetNextBackend() *Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.backends) == 0 {
		return nil
	}

	// Find backends with the smallest score
	// Include backends within 20% of the best score for fair distribution
	var bestBackends []*Backend
	var bestScore float64 = -1

	for _, backend := range lb.backends {
		if !backend.IsAlive() {
			continue
		}

		// Calculate score: lower is better
		// Score = (time_quanta * (1 + connections)) + avg_latency
		timeQuanta := float64(backend.GetTimeQuanta().Nanoseconds())
		connections := float64(backend.GetActiveConnections())
		avgLatency := float64(backend.GetMetrics().AvgLatency.Nanoseconds())

		score := (timeQuanta * (1 + connections)) + avgLatency

		if bestScore == -1 || score < bestScore {
			bestScore = score
		}
	}

	// Collect all backends within 20% of the best score
	threshold := bestScore * 1.2
	for _, backend := range lb.backends {
		if !backend.IsAlive() {
			continue
		}

		timeQuanta := float64(backend.GetTimeQuanta().Nanoseconds())
		connections := float64(backend.GetActiveConnections())
		avgLatency := float64(backend.GetMetrics().AvgLatency.Nanoseconds())
		score := (timeQuanta * (1 + connections)) + avgLatency

		if score <= threshold {
			bestBackends = append(bestBackends, backend)
		}
	}

	// If we have backends with the same best score, use round-robin among them
	if len(bestBackends) > 0 {
		idx := atomic.AddUint64(&lb.current, 1) % uint64(len(bestBackends))
		return bestBackends[idx]
	}

	// Fallback to simple round-robin if all backends are down
	for i := 0; i < len(lb.backends); i++ {
		idx := atomic.AddUint64(&lb.current, 1) % uint64(len(lb.backends))
		backend := lb.backends[idx]
		if backend.IsAlive() {
			return backend
		}
	}

	// Return first backend as last resort
	if len(lb.backends) > 0 {
		return lb.backends[0]
	}

	return nil
}

// ServeHTTP handles incoming requests
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.GetNextBackend()

	if backend == nil || !backend.IsAlive() {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		log.Printf("No healthy backends available")
		return
	}

	backend.IncrementConnections()
	defer backend.DecrementConnections()

	start := time.Now()
	backend.ReverseProxy.ServeHTTP(w, r)
	latency := time.Since(start)

	backend.AddRequest(latency)
	log.Printf("Proxied request to %s (latency: %v)", backend.URL.String(), latency)
}

// GetMetrics returns metrics for all backends
func (lb *LoadBalancer) GetMetrics() []BackendMetrics {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	metrics := make([]BackendMetrics, 0, len(lb.backends))
	for _, backend := range lb.backends {
		metrics = append(metrics, backend.GetMetrics())
	}
	return metrics
}

// AddBackend adds a new backend to the load balancer
func (lb *LoadBalancer) AddBackend(urlStr string) error {
	backend, err := NewBackend(urlStr)
	if err != nil {
		return fmt.Errorf("failed to create backend %s: %w", urlStr, err)
	}

	lb.mu.Lock()
	lb.backends = append(lb.backends, backend)
	lb.mu.Unlock()

	// Add to health checker
	lb.healthChecker.AddBackend(backend)

	log.Printf("Added backend: %s", urlStr)
	return nil
}

// RemoveBackend removes a backend from the load balancer
func (lb *LoadBalancer) RemoveBackend(urlStr string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, backend := range lb.backends {
		if backend.URL.String() == urlStr {
			// Remove from backends slice
			lb.backends = append(lb.backends[:i], lb.backends[i+1:]...)

			// Remove from health checker
			lb.healthChecker.RemoveBackend(backend)

			log.Printf("Removed backend: %s", urlStr)
			return nil
		}
	}

	return fmt.Errorf("backend not found: %s", urlStr)
}

// GetBackends returns a copy of the backends list
func (lb *LoadBalancer) GetBackends() []*Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	backends := make([]*Backend, len(lb.backends))
	copy(backends, lb.backends)
	return backends
}
