package main

import (
	"testing"
	"time"
)

func TestNewBackend(t *testing.T) {
	backend, err := NewBackend("http://localhost:8081")
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	if backend.URL.String() != "http://localhost:8081" {
		t.Errorf("Expected URL http://localhost:8081, got %s", backend.URL.String())
	}

	if !backend.IsAlive() {
		t.Error("New backend should be alive by default")
	}

	if backend.RequestCount != 0 {
		t.Errorf("New backend should have 0 requests, got %d", backend.RequestCount)
	}
}

func TestBackendAliveStatus(t *testing.T) {
	backend, _ := NewBackend("http://localhost:8081")

	backend.SetAlive(false)
	if backend.IsAlive() {
		t.Error("Backend should be marked as not alive")
	}

	backend.SetAlive(true)
	if !backend.IsAlive() {
		t.Error("Backend should be marked as alive")
	}
}

func TestBackendAddRequest(t *testing.T) {
	backend, _ := NewBackend("http://localhost:8081")

	latency := 10 * time.Millisecond
	backend.AddRequest(latency)

	if backend.RequestCount != 1 {
		t.Errorf("Expected request count 1, got %d", backend.RequestCount)
	}

	if backend.TotalLatency != latency {
		t.Errorf("Expected total latency %v, got %v", latency, backend.TotalLatency)
	}
}

func TestBackendMetrics(t *testing.T) {
	backend, _ := NewBackend("http://localhost:8081")

	// Add some requests
	backend.AddRequest(10 * time.Millisecond)
	backend.AddRequest(20 * time.Millisecond)
	backend.AddRequest(30 * time.Millisecond)

	metrics := backend.GetMetrics()

	if metrics.RequestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", metrics.RequestCount)
	}

	expectedAvg := 20 * time.Millisecond
	if metrics.AvgLatency != expectedAvg {
		t.Errorf("Expected average latency %v, got %v", expectedAvg, metrics.AvgLatency)
	}

	if !metrics.Alive {
		t.Error("Backend should be alive")
	}
}

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}

	if config.HealthCheckPath != "/health" {
		t.Errorf("Expected health check path /health, got %s", config.HealthCheckPath)
	}

	if len(config.Backends) != 3 {
		t.Errorf("Expected 3 backends, got %d", len(config.Backends))
	}
}

func TestLoadConfigInvalidFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
}
