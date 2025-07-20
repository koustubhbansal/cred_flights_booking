package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cred_flights_booking/internal/database"
	"cred_flights_booking/internal/handlers"
	"cred_flights_booking/internal/services"
)

func main() {
	log.Println("Starting Flight Service...")

	// Initialize database connection
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	cache, err := database.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	// Initialize services
	flightService := services.NewFlightService(db, cache)

	// Initialize handlers
	flightHandlers := handlers.NewFlightHandlers(flightService)

	// Create HTTP server with Go 1.22 ServeMux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /api/flights/search", flightHandlers.SearchFlights)
	mux.HandleFunc("GET /api/flights/{id}", flightHandlers.GetFlight)
	mux.HandleFunc("POST /api/flights/validate", flightHandlers.ValidateFlight)
	mux.HandleFunc("POST /api/flights/seats/decrement", flightHandlers.DecrementSeats)
	mux.HandleFunc("POST /api/flights/seats/increment", flightHandlers.IncrementSeats)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"flight-service"}`))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Flight Service listening on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Flight Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Flight Service exited")
}
