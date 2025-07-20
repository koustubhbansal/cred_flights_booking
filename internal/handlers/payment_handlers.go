package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cred_flights_booking/internal/models"
	"cred_flights_booking/internal/services"
)

// PaymentHandlers handles payment-related HTTP requests
type PaymentHandlers struct {
	paymentService *services.PaymentService
}

// NewPaymentHandlers creates new payment handlers
func NewPaymentHandlers(paymentService *services.PaymentService) *PaymentHandlers {
	return &PaymentHandlers{
		paymentService: paymentService,
	}
}

// ProcessPayment handles payment processing requests
func (ph *PaymentHandlers) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BookingID <= 0 || req.Amount <= 0 || req.UserID <= 0 {
		http.Error(w, "Invalid booking ID, amount, or user ID", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Process payment
	response, err := ph.paymentService.ProcessPayment(ctx, &req)
	if err != nil {
		log.Printf("Payment processing error: %v", err)
		http.Error(w, "Payment processing failed", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")

	// Set appropriate status code based on payment result
	statusCode := http.StatusOK
	if response.Status == models.PaymentStatusFailed {
		statusCode = http.StatusBadRequest
	} else if response.Status == models.PaymentStatusTimeout {
		statusCode = http.StatusRequestTimeout
	}

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Payment processed: BookingID=%d, Status=%s", req.BookingID, response.Status)
}

// SimulatePaymentFailure handles payment failure simulation requests
func (ph *PaymentHandlers) SimulatePaymentFailure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BookingID <= 0 || req.Amount <= 0 || req.UserID <= 0 {
		http.Error(w, "Invalid booking ID, amount, or user ID", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Simulate payment failure
	response, err := ph.paymentService.SimulatePaymentFailure(ctx, &req)
	if err != nil {
		log.Printf("Payment failure simulation error: %v", err)
		http.Error(w, "Payment failure simulation failed", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Payment failure simulated: BookingID=%d", req.BookingID)
}

// SimulatePaymentTimeout handles payment timeout simulation requests
func (ph *PaymentHandlers) SimulatePaymentTimeout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BookingID <= 0 || req.Amount <= 0 || req.UserID <= 0 {
		http.Error(w, "Invalid booking ID, amount, or user ID", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Simulate payment timeout
	response, err := ph.paymentService.SimulatePaymentTimeout(ctx, &req)
	if err != nil {
		log.Printf("Payment timeout simulation error: %v", err)
		http.Error(w, "Payment timeout simulation failed", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusRequestTimeout)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Payment timeout simulated: BookingID=%d", req.BookingID)
}

// SimulatePaymentSuccess handles payment success simulation requests
func (ph *PaymentHandlers) SimulatePaymentSuccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BookingID <= 0 || req.Amount <= 0 || req.UserID <= 0 {
		http.Error(w, "Invalid booking ID, amount, or user ID", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Simulate payment success
	response, err := ph.paymentService.SimulatePaymentSuccess(ctx, &req)
	if err != nil {
		log.Printf("Payment success simulation error: %v", err)
		http.Error(w, "Payment success simulation failed", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Payment success simulated: BookingID=%d", req.BookingID)
}
