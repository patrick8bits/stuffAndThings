package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/patrick8bits/stuffAndThings/internal/health"
)

func main() {
	// Start a private listener for our simulated CDN servers.
	listener, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		log.Fatal(err)
	}

	simulator := &http.Server{
		Handler:           http.HandlerFunc(health.DemoHandler),
		ReadHeaderTimeout: 2 * time.Second,
	}

	defer simulator.Close()

	go func() {
		err := simulator.Serve(listener)

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Simulator error: %v", err)
		}
	}()

	simulatorURL := "http://" + listener.Addr().String()

	log.Println("CDN simulator starting at", simulatorURL)

	targets := health.DemoTargets(simulatorURL)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil

	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}

	// Configure our website and API routes.
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./web"))

	mux.Handle("/", fileServer)
	mux.HandleFunc("/api/sort", sortHandler)
	mux.HandleFunc("/api/sort/source", sortSourceHandler)
	mux.HandleFunc("/api/health", newHealthHandler(client, targets))
	mux.HandleFunc("/api/health/source", healthSourceHandler)

	log.Println("Server starting at http://localhost:8080")

	err = http.ListenAndServe("127.0.0.1:8080", mux)

	if err != nil {
		log.Fatal(err)
	}
}
