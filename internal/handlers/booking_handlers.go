package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"cred_flights_booking/internal/models"
	"cred_flights_booking/internal/services"
)

// BookingHandlers handles booking-related HTTP requests
type BookingHandlers struct {
	bookingService *services.BookingServiceV2
}

// NewBookingHandlers creates new booking handlers
func NewBookingHandlers(bookingService *services.BookingServiceV2) *BookingHandlers {
	return &BookingHandlers{
		bookingService: bookingService,
	}
}

// CreateBooking handles booking creation requests
func (bh *BookingHandlers) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID <= 0 || req.FlightID <= 0 || req.Seats <= 0 || req.Date == "" {
		http.Error(w, "Invalid user ID, flight ID, seats, or date", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second) // Longer timeout for booking
	defer cancel()

	// Create booking
	response, err := bh.bookingService.CreateBooking(ctx, &req)
	if err != nil {
		log.Printf("Booking creation error: %v", err)
		http.Error(w, fmt.Sprintf("Booking failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")

	// Set appropriate status code based on booking result
	statusCode := http.StatusOK
	if response.Status == models.BookingStatusFailed {
		statusCode = http.StatusBadRequest
	}

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Booking creation completed: ID=%d, Status=%s", response.BookingID, response.Status)
}

// GetBooking handles getting booking details
func (bh *BookingHandlers) GetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract booking ID from URL path
	bookingIDStr := r.URL.Query().Get("id")
	if bookingIDStr == "" {
		http.Error(w, "Missing booking ID", http.StatusBadRequest)
		return
	}

	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil || bookingID <= 0 {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get booking
	booking, err := bh.bookingService.GetBooking(ctx, bookingID)
	if err != nil {
		log.Printf("Get booking error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get booking: %v", err), http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(booking); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Booking retrieved: ID=%d", bookingID)
}

// CancelBooking handles booking cancellation requests
func (bh *BookingHandlers) CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract booking ID from URL path
	bookingIDStr := r.URL.Query().Get("id")
	if bookingIDStr == "" {
		http.Error(w, "Missing booking ID", http.StatusBadRequest)
		return
	}

	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil || bookingID <= 0 {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Cancel booking
	err = bh.bookingService.CancelBooking(ctx, bookingID)
	if err != nil {
		log.Printf("Cancel booking error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to cancel booking: %v", err), http.StatusBadRequest)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":      "Booking cancelled successfully",
		"booking_id":   bookingID,
		"cancelled_at": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Booking cancelled: ID=%d", bookingID)
}
