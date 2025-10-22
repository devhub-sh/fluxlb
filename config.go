package main

import (
	"encoding/json"
	"os"
)

type BackendConfig struct {
	URL string `json:"url"`
}
type Config struct {
	Port                int             `json:"port"`
	HealthCheckPath     string          `json:"health_check_path"`
	HealthCheckInterval int             `json:"health_check_interval"`
	Backends            []BackendConfig `json:"backends"`
}

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

	// config.HealthCheckInterval = config.HealthCheckInterval * time.Second
	return &config, nil
}
