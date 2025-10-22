package main

import (
	"encoding/json"
	"os"
	"time"
)

// Config represents the load balancer configuration
type Config struct {
	Port                int             `json:"port"`
	HealthCheckPath     string          `json:"health_check_path"`
	HealthCheckInterval time.Duration   `json:"health_check_interval_seconds"`
	Backends            []BackendConfig `json:"backends"`
}

// BackendConfig represents a backend server configuration
type BackendConfig struct {
	URL string `json:"url"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// Convert seconds to duration
	config.HealthCheckInterval = config.HealthCheckInterval * time.Second

	return &config, nil
}
