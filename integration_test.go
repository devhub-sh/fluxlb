package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoadBalancerRoundRobin(t *testing.T) {
	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "server1")
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "server2")
	}))
	defer server2.Close()

	// Create config
	config := &Config{
		Port:                8080,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		Backends: []BackendConfig{
			{URL: server1.URL},
			{URL: server2.URL},
		},
	}

	// Create load balancer
	lb, err := NewLoadBalancer(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Test round-robin
	responses := make(map[string]int)
	for i := 0; i < 10; i++ {
		backend := lb.GetNextBackend()
		responses[backend.URL.String()]++
	}

	// Each backend should get approximately equal requests
	if responses[server1.URL] != 5 {
		t.Errorf("Expected 5 requests to server1, got %d", responses[server1.URL])
	}
	if responses[server2.URL] != 5 {
		t.Errorf("Expected 5 requests to server2, got %d", responses[server2.URL])
	}
}

func TestLoadBalancerSkipsUnhealthyBackend(t *testing.T) {
	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "server1")
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "server2")
	}))
	defer server2.Close()

	// Create config
	config := &Config{
		Port:                8080,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		Backends: []BackendConfig{
			{URL: server1.URL},
			{URL: server2.URL},
		},
	}

	// Create load balancer
	lb, err := NewLoadBalancer(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Mark first backend as down
	lb.backends[0].SetAlive(false)

	// All requests should go to second backend
	for i := 0; i < 5; i++ {
		backend := lb.GetNextBackend()
		if backend.URL.String() != server2.URL {
			t.Errorf("Expected request to go to server2, got %s", backend.URL.String())
		}
	}
}

func TestHealthChecker(t *testing.T) {
	// Create a test server with health endpoint
	healthy := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			if healthy {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		}
	}))
	defer server.Close()

	backend, _ := NewBackend(server.URL)
	backends := []*Backend{backend}

	hc := NewHealthChecker(backends, "/health", 1*time.Second)

	// Start health checker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hc.Start(ctx)

	// Wait for health check
	time.Sleep(1500 * time.Millisecond)

	if !backend.IsAlive() {
		t.Error("Backend should be alive")
	}

	// Mark backend as unhealthy
	healthy = false
	time.Sleep(1500 * time.Millisecond)

	if backend.IsAlive() {
		t.Error("Backend should be marked as down")
	}
}

func TestLoadBalancerHTTPHandler(t *testing.T) {
	// Create a test backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from backend")
	}))
	defer backendServer.Close()

	// Create config
	config := &Config{
		Port:                8080,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		Backends: []BackendConfig{
			{URL: backendServer.URL},
		},
	}

	// Create load balancer
	lb, err := NewLoadBalancer(config)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	lb.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}

	body, _ := io.ReadAll(rr.Body)
	if string(body) != "Hello from backend\n" {
		t.Errorf("Unexpected body: %s", string(body))
	}

	// Check metrics
	metrics := lb.GetMetrics()
	if metrics[0].RequestCount != 1 {
		t.Errorf("Expected 1 request, got %d", metrics[0].RequestCount)
	}
}
