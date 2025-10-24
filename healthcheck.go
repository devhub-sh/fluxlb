package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

/*
 * @ HealthChecker periodically checks the health of backend servers
 */
type HealthChecker struct {
	backends []*Backend
	path     string
	interval time.Duration
	mu       sync.RWMutex
}

/*
 * @ NewHealthChecker creates a new health checker
 */
func NewHealthChecker(backends []*Backend, path string, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		backends: backends,
		path:     path,
		interval: interval,
	}
}

/*
 * @ Start begins the health check routine
 */
func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	// Perform initial health check
	hc.checkAll()

	for {
		select {
		case <-ticker.C:
			hc.checkAll()
		case <-ctx.Done():
			return
		}
	}
}

// checkAll checks the health of all backends
func (hc *HealthChecker) checkAll() {
	hc.mu.RLock()
	backends := make([]*Backend, len(hc.backends))
	copy(backends, hc.backends)
	hc.mu.RUnlock()

	for _, backend := range backends {
		go hc.check(backend)
	}
}

// AddBackend adds a backend to the health checker
func (hc *HealthChecker) AddBackend(backend *Backend) {
	hc.mu.Lock()
	hc.backends = append(hc.backends, backend)
	hc.mu.Unlock()

	// Perform immediate health check
	go hc.check(backend)
}

// RemoveBackend removes a backend from the health checker
func (hc *HealthChecker) RemoveBackend(backend *Backend) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for i, b := range hc.backends {
		if b == backend {
			hc.backends = append(hc.backends[:i], hc.backends[i+1:]...)
			return
		}
	}
}

// check performs a health check on a single backend
func (hc *HealthChecker) check(backend *Backend) {
	url := backend.URL.String() + hc.path

	// Validate URL scheme to prevent SSRF attacks
	if backend.URL.Scheme != "http" && backend.URL.Scheme != "https" {
		backend.SetAlive(false)
		log.Printf("Backend %s has invalid scheme: %s", backend.URL.String(), backend.URL.Scheme)
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		backend.SetAlive(false)
		log.Printf("Backend %s is DOWN: %v", backend.URL.String(), err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		backend.SetAlive(true)
		log.Printf("Backend %s is UP", backend.URL.String())
	} else {
		backend.SetAlive(false)
		log.Printf("Backend %s is DOWN (status %d)", backend.URL.String(), resp.StatusCode)
	}
}
