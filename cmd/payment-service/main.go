package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cred_flights_booking/internal/handlers"
	"cred_flights_booking/internal/services"
)

func main() {
	log.Println("Starting Payment Service...")

	// Initialize services
	paymentService := services.NewPaymentService()

	// Initialize handlers
	paymentHandlers := handlers.NewPaymentHandlers(paymentService)

	// Create HTTP server with Go 1.22 ServeMux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("POST /api/payments/process", paymentHandlers.ProcessPayment)
	mux.HandleFunc("POST /api/payments/simulate/failure", paymentHandlers.SimulatePaymentFailure)
	mux.HandleFunc("POST /api/payments/simulate/timeout", paymentHandlers.SimulatePaymentTimeout)
	mux.HandleFunc("POST /api/payments/simulate/success", paymentHandlers.SimulatePaymentSuccess)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"payment-service"}`))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8082",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Payment Service listening on port 8082")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Payment Service...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Payment Service exited")
}
