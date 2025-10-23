package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	/*
		 * @ Load configuration
		 	*  from file specified by command line flag
			 *  default to config.json
	*/
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Printf("FluxLB starting with %d backends on port %d", len(config.Backends), config.Port)

	/*
		 * @ Initialize load balancer
			* with backends from configuration
	*/

	lb, err := NewLoadBalancer(config)
	if err != nil {
		log.Fatalf("Failed to create load balancer: %v", err)
	}

	dashboard, err := NewDashboard(lb)
	if err != nil {
		log.Fatalf("Failed to create dashboard: %v", err)
	}

	// Initialize auth manager
	authManager := NewAuthManager(&config.Auth)

	// Initialize API handler
	apiHandler := NewAPIHandler(lb, authManager)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lb.Start(ctx)

	/*
		 * @ Start HTTP server
			* with health check endpoint
			* and dashboard endpoint
			* for monitoring backends
			* and load balancing
			* using smart round-robin algorithm
	*/
	mux := http.NewServeMux()

	// Serve React static files if available (only the static subdirectory)
	reactBuildPath := "frontend/build"
	reactStaticPath := filepath.Join(reactBuildPath, "static")
	if _, err := os.Stat(reactStaticPath); err == nil {
		fs := http.FileServer(http.Dir(reactBuildPath))
		// Only serve /static/ paths from the React build
		mux.Handle("/static/", http.StripPrefix("/", fs))
	}

	// Public endpoints
	mux.HandleFunc("/api/login", apiHandler.HandleLogin)
	mux.HandleFunc("/health", HealthCheckHandler)

	// Protected endpoints
	mux.HandleFunc("/dashboard", authManager.AuthMiddleware(dashboard.ServeHTTP))
	mux.HandleFunc("/api/metrics", authManager.AuthMiddleware(dashboard.ServeMetricsAPI))
	mux.HandleFunc("/api/logout", authManager.AuthMiddleware(apiHandler.HandleLogout))
	mux.HandleFunc("/api/backends/add", authManager.AuthMiddleware(apiHandler.HandleAddBackend))
	mux.HandleFunc("/api/backends/remove", authManager.AuthMiddleware(apiHandler.HandleRemoveBackend))
	mux.HandleFunc("/api/backends", authManager.AuthMiddleware(apiHandler.HandleGetBackends))

	// Load balancer proxy (unprotected for actual traffic)
	mux.HandleFunc("/", lb.ServeHTTP)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	/*
		 * @ Start server in a goroutine
			* to allow graceful shutdown
	*/
	go func() {
		if config.EnableHTTPS {
			log.Printf("FluxLB HTTPS listening on https://localhost:%d", config.HTTPSPort)
			log.Printf("Dashboard available at https://localhost:%d/dashboard", config.HTTPSPort)

			// Create TLS config
			tlsConfig := &tls.Config{
				MinVersion: tls.VersionTLS12,
			}

			httpsServer := &http.Server{
				Addr:         fmt.Sprintf(":%d", config.HTTPSPort),
				Handler:      mux,
				TLSConfig:    tlsConfig,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			}

			if err := httpsServer.ListenAndServeTLS(config.CertFile, config.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTPS Server Error: %v", err)
			}
		}
	}()

	go func() {
		log.Printf("FluxLB HTTP listening on http://localhost:%d", config.Port)
		log.Printf("Dashboard available at http://localhost:%d/dashboard", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Error: %v", err)
		}
	}()

	/*
		 * @ Wait for interrupt signal
			* to gracefully shutdown the server
	*/
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)
	<-signChan

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("FluxLB stopped")
}
