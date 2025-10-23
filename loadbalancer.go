package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// LoadBalancer manages the backend servers and routing
type LoadBalancer struct {
	backends      []*Backend
	current       uint64
	healthChecker *HealthChecker
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
	}, nil
}

// Start starts the load balancer and health checker
func (lb *LoadBalancer) Start(ctx context.Context) {
	go lb.healthChecker.Start(ctx)
}

// GetNextBackend returns the next available backend using round-robin
func (lb *LoadBalancer) GetNextBackend() *Backend {
	// Round-robin through backends
	for i := 0; i < len(lb.backends); i++ {
		idx := atomic.AddUint64(&lb.current, 1) % uint64(len(lb.backends))
		backend := lb.backends[idx]

		if backend.IsAlive() {
			return backend
		}
	}

	// If no backend is alive, return the first one anyway
	return lb.backends[0]
}

// ServeHTTP handles incoming requests
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.GetNextBackend()

	if !backend.IsAlive() {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		log.Printf("No healthy backends available")
		return
	}

	start := time.Now()
	backend.ReverseProxy.ServeHTTP(w, r)
	latency := time.Since(start)

	backend.AddRequest(latency)
	log.Printf("Proxied request to %s (latency: %v)", backend.URL.String(), latency)
}

// GetMetrics returns metrics for all backends
func (lb *LoadBalancer) GetMetrics() []BackendMetrics {
	metrics := make([]BackendMetrics, 0, len(lb.backends))
	for _, backend := range lb.backends {
		metrics = append(metrics, backend.GetMertics())
	}
	return metrics
}
