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

// FlightHandlers handles flight-related HTTP requests
type FlightHandlers struct {
	flightService *services.FlightService
}

// NewFlightHandlers creates new flight handlers
func NewFlightHandlers(flightService *services.FlightService) *FlightHandlers {
	return &FlightHandlers{
		flightService: flightService,
	}
}

// SearchFlights handles flight search requests
func (fh *FlightHandlers) SearchFlights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	source := r.URL.Query().Get("source")
	destination := r.URL.Query().Get("destination")
	date := r.URL.Query().Get("date")
	seatsStr := r.URL.Query().Get("seats")
	sortBy := r.URL.Query().Get("sort_by")

	// Validate required parameters
	if source == "" || destination == "" || date == "" || seatsStr == "" {
		http.Error(w, "Missing required parameters: source, destination, date, seats", http.StatusBadRequest)
		return
	}

	// Parse seats
	seats, err := strconv.Atoi(seatsStr)
	if err != nil || seats <= 0 {
		http.Error(w, "Invalid seats parameter", http.StatusBadRequest)
		return
	}

	// Validate sort order
	if sortBy != "" && sortBy != "cheapest" && sortBy != "fastest" {
		http.Error(w, "Invalid sort_by parameter. Must be 'cheapest' or 'fastest'", http.StatusBadRequest)
		return
	}

	// Set default sort order
	if sortBy == "" {
		sortBy = "cheapest"
	}

	// Create search request
	req := &models.SearchRequest{
		Source:      source,
		Destination: destination,
		Date:        date,
		Seats:       seats,
		SortBy:      sortBy,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Search flights
	response, err := fh.flightService.SearchFlights(ctx, req)
	if err != nil {
		log.Printf("Flight search error: %v", err)
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
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

	log.Printf("Flight search completed: %d paths found", response.Count)
}

// GetFlight handles getting flight details
func (fh *FlightHandlers) GetFlight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract flight ID from URL path
	// Assuming URL pattern: /api/flights/{id}
	// You might need to adjust this based on your routing setup
	flightIDStr := r.URL.Query().Get("id")
	if flightIDStr == "" {
		http.Error(w, "Missing flight ID", http.StatusBadRequest)
		return
	}

	flightID, err := strconv.Atoi(flightIDStr)
	if err != nil || flightID <= 0 {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	// Get flight details (you'll need to implement this in FlightService)
	// For now, we'll return a placeholder
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":   "Flight details endpoint not implemented yet",
		"flight_id": flightID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ValidateFlight handles flight validation requests
func (fh *FlightHandlers) ValidateFlight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.FlightValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.FlightID <= 0 || req.Seats <= 0 || req.Date == "" {
		http.Error(w, "Invalid flight ID, seats, or date", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Validate flight
	response, err := fh.flightService.ValidateFlight(ctx, req.FlightID, req.Seats, req.Date)
	if err != nil {
		log.Printf("Flight validation error: %v", err)
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusInternalServerError)
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

	log.Printf("Flight validation completed for flight %d: %v", req.FlightID, response.Valid)
}

// DecrementSeats handles seat decrement requests
func (fh *FlightHandlers) DecrementSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.SeatUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.FlightID <= 0 || req.Seats <= 0 || req.Date == "" {
		http.Error(w, "Invalid flight ID, seats, or date", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Decrement seats
	err := fh.flightService.DecrementSeats(ctx, req.FlightID, req.Seats, req.Date)
	if err != nil {
		log.Printf("Seat decrement error: %v", err)
		http.Error(w, fmt.Sprintf("Seat decrement failed: %v", err), http.StatusBadRequest)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":    "Seats decremented successfully",
		"flight_id":  req.FlightID,
		"seats":      req.Seats,
		"date":       req.Date,
		"updated_at": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Seats decremented for flight %d: %d seats", req.FlightID, req.Seats)
}

// IncrementSeats handles seat increment requests
func (fh *FlightHandlers) IncrementSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.SeatUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.FlightID <= 0 || req.Seats <= 0 || req.Date == "" {
		http.Error(w, "Invalid flight ID, seats, or date", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Increment seats
	err := fh.flightService.IncrementSeats(ctx, req.FlightID, req.Seats, req.Date)
	if err != nil {
		log.Printf("Seat increment error: %v", err)
		http.Error(w, fmt.Sprintf("Seat increment failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message":    "Seats incremented successfully",
		"flight_id":  req.FlightID,
		"seats":      req.Seats,
		"date":       req.Date,
		"updated_at": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Seats incremented for flight %d: %d seats", req.FlightID, req.Seats)
}
