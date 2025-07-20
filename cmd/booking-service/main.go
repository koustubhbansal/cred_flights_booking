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
	log.Println("Starting Booking Service...")

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

	// Get service URLs from environment
	flightServiceURL := os.Getenv("FLIGHT_SERVICE_URL")
	if flightServiceURL == "" {
		flightServiceURL = "http://localhost:8080"
	}

	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "http://localhost:8082"
	}

	bookingService := services.NewBookingServiceV2(db, cache, flightServiceURL, paymentServiceURL)

	// Initialize handlers
	bookingHandlers := handlers.NewBookingHandlers(bookingService)

	// Create HTTP server with Go 1.22 ServeMux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("POST /api/bookings", bookingHandlers.CreateBooking)
	mux.HandleFunc("GET /api/bookings/{id}", bookingHandlers.GetBooking)
	mux.HandleFunc("PUT /api/bookings/{id}/cancel", bookingHandlers.CancelBooking)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"booking-service"}`))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Booking Service listening on port 8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Booking Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Booking Service exited")
}
