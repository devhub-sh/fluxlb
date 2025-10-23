package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

/*
 * @ HealthChecker periodically checks the health of backend servers
 */
type HealthChecker struct {
	backends []*Backend
	path     string
	interval time.Duration
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
	for _, backend := range hc.backends {
		go hc.check(backend)
	}
}

// check performs a health check on a single backend
func (hc *HealthChecker) check(backend *Backend) {
	url := backend.URL.String() + hc.path
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
