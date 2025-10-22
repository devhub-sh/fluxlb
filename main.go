package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("FluxLB starting with %d backends on port %d", len(config.Backends), config.Port)

	// Create load balancer
	lb, err := NewLoadBalancer(config)
	if err != nil {
		log.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create dashboard
	dashboard, err := NewDashboard(lb)
	if err != nil {
		log.Fatalf("Failed to create dashboard: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health checker
	lb.Start(ctx)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/dashboard", dashboard.ServeHTTP)
	mux.HandleFunc("/api/metrics", dashboard.ServeMetricsAPI)
	mux.HandleFunc("/", lb.ServeHTTP)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("FluxLB listening on http://localhost:%d", config.Port)
		log.Printf("Dashboard available at http://localhost:%d/dashboard", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// Graceful shutdown
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("FluxLB stopped")
}
