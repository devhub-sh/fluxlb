package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: test_backend <port>")
	}

	port := os.Args[1]

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate processing time
		fmt.Fprintf(w, "Response from backend on port %s\n", port)
		log.Printf("Handled request from %s", r.RemoteAddr)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	addr := ":" + port
	log.Printf("Test backend starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
