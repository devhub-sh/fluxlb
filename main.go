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

	// TODO: Initialize load balancer here

	/*
		 * @ Start HTTP server
			* with health check endpoint
			* and dashboard endpoint
			* for monitoring backends
			* and load balancing
			* using round-robin algorithm
	*/
	mux := http.NewServeMux()
	mux.HandleFunc("/health", HealthCheckHandler)
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
		log.Printf("FluxLb listening on http://localhost:%d", config.Port)
		log.Printf("Dashboard avaliable at http://localhost:%d/dashbboard", config.Port)
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("FluxLB stopped")

}
