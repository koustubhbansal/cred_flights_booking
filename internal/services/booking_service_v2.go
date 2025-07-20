package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"cred_flights_booking/internal/database"
	"cred_flights_booking/internal/models"
)

// BookingServiceV2 handles booking-related operations with improved architecture
type BookingServiceV2 struct {
	db                *database.DB
	cache             *database.RedisClient
	flightServiceURL  string
	paymentServiceURL string
	httpClient        *http.Client
}

// NewBookingServiceV2 creates a new booking service
func NewBookingServiceV2(db *database.DB, cache *database.RedisClient, flightServiceURL, paymentServiceURL string) *BookingServiceV2 {
	return &BookingServiceV2{
		db:                db,
		cache:             cache,
		flightServiceURL:  flightServiceURL,
		paymentServiceURL: paymentServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateBooking creates a new booking with improved flow
func (bs *BookingServiceV2) CreateBooking(ctx context.Context, req *models.BookingRequest) (*models.BookingResponse, error) {
	log.Printf("Creating booking for user %d, flight %d, seats %d", req.UserID, req.FlightID, req.Seats)

	// Step 1: Validate flight availability via Flight Service
	validation, err := bs.validateFlightViaHTTP(ctx, req.FlightID, req.Seats, req.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to validate flight: %w", err)
	}

	if !validation.Valid {
		return &models.BookingResponse{
			Status:  models.BookingStatusFailed,
			Message: validation.Message,
		}, nil
	}

	// Step 2: Create temporary booking in Redis
	tempBooking := &models.TempBooking{
		UserID:      req.UserID,
		FlightID:    req.FlightID,
		Seats:       req.Seats,
		TotalAmount: validation.Price,
		Date:        req.Date,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(15 * time.Minute), // 15 minutes expiry
	}

	tempBookingKey := database.GenerateTempBookingCacheKey(req.UserID, req.FlightID)
	if err := bs.cache.SetJSON(ctx, tempBookingKey, tempBooking, 15*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to create temporary booking: %w", err)
	}

	// Step 3: Decrement seats in Flight Service
	if err := bs.decrementSeatsViaHTTP(ctx, req.FlightID, req.Seats, req.Date); err != nil {
		// Clean up temporary booking
		bs.cache.Delete(ctx, tempBookingKey)
		return &models.BookingResponse{
			Status:  models.BookingStatusFailed,
			Message: fmt.Sprintf("Failed to reserve seats: %v", err),
		}, nil
	}

	// Step 4: Process payment
	paymentReq := &models.PaymentRequest{
		BookingID:   req.UserID, // Use user ID as temporary booking ID
		Amount:      validation.Price,
		UserID:      req.UserID,
		PaymentType: "credit_card", // Default payment type
	}

	paymentResp, err := bs.processPayment(ctx, paymentReq)
	if err != nil {
		// Payment failed - revert seat count and clean up
		bs.revertBookingOnFailure(ctx, req.FlightID, req.Seats, req.Date, tempBookingKey)
		return &models.BookingResponse{
			Status:  models.BookingStatusFailed,
			Message: fmt.Sprintf("Payment failed: %v", err),
		}, nil
	}

	// Step 5: Handle payment result
	var bookingStatus string
	switch paymentResp.Status {
	case models.PaymentStatusSuccess:
		bookingStatus = models.BookingStatusConfirmed
		// Create permanent booking in database
		bookingID, err := bs.createPermanentBooking(ctx, req, validation.Price, paymentResp.PaymentID)
		if err != nil {
			// Revert everything on database failure
			bs.revertBookingOnFailure(ctx, req.FlightID, req.Seats, req.Date, tempBookingKey)
			return &models.BookingResponse{
				Status:  models.BookingStatusFailed,
				Message: fmt.Sprintf("Failed to create booking: %v", err),
			}, nil
		}
		// Remove temporary booking
		bs.cache.Delete(ctx, tempBookingKey)

		return &models.BookingResponse{
			BookingID:   bookingID,
			Status:      bookingStatus,
			TotalAmount: validation.Price,
			PaymentID:   paymentResp.PaymentID,
			Message:     "Booking created successfully",
		}, nil

	case models.PaymentStatusFailed, models.PaymentStatusTimeout:
		bookingStatus = models.BookingStatusFailed
		// Revert seat count and clean up
		bs.revertBookingOnFailure(ctx, req.FlightID, req.Seats, req.Date, tempBookingKey)
		return &models.BookingResponse{
			Status:      bookingStatus,
			TotalAmount: validation.Price,
			Message:     paymentResp.Message,
		}, nil

	default:
		bookingStatus = models.BookingStatusPending
		// Keep temporary booking for retry
		return &models.BookingResponse{
			Status:      bookingStatus,
			TotalAmount: validation.Price,
			Message:     "Payment pending, please retry",
		}, nil
	}
}

// validateFlightViaHTTP validates flight via HTTP call to Flight Service
func (bs *BookingServiceV2) validateFlightViaHTTP(ctx context.Context, flightID, seats int, date string) (*models.FlightValidationResponse, error) {
	reqBody := models.FlightValidationRequest{
		FlightID: flightID,
		Seats:    seats,
		Date:     date,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validation request: %w", err)
	}

	url := fmt.Sprintf("%s/api/flights/validate", bs.flightServiceURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := bs.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make validation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("validation request failed with status: %d", resp.StatusCode)
	}

	var validation models.FlightValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		return nil, fmt.Errorf("failed to decode validation response: %w", err)
	}

	return &validation, nil
}

// decrementSeatsViaHTTP decrements seats via HTTP call to Flight Service
func (bs *BookingServiceV2) decrementSeatsViaHTTP(ctx context.Context, flightID, seats int, date string) error {
	reqBody := models.SeatUpdateRequest{
		FlightID: flightID,
		Seats:    seats,
		Date:     date,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal seat update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/flights/seats/decrement", bs.flightServiceURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := bs.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make seat decrement request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("seat decrement request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// incrementSeatsViaHTTP increments seats via HTTP call to Flight Service
func (bs *BookingServiceV2) incrementSeatsViaHTTP(ctx context.Context, flightID, seats int, date string) error {
	reqBody := models.SeatUpdateRequest{
		FlightID: flightID,
		Seats:    seats,
		Date:     date,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal seat update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/flights/seats/increment", bs.flightServiceURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := bs.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make seat increment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("seat increment request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// revertBookingOnFailure reverts seat count and cleans up temporary booking
func (bs *BookingServiceV2) revertBookingOnFailure(ctx context.Context, flightID, seats int, date, tempBookingKey string) {
	// Increment seats back
	if err := bs.incrementSeatsViaHTTP(ctx, flightID, seats, date); err != nil {
		log.Printf("Failed to revert seat count for flight %d: %v", flightID, err)
	}

	// Remove temporary booking
	if err := bs.cache.Delete(ctx, tempBookingKey); err != nil {
		log.Printf("Failed to remove temporary booking: %v", err)
	}

	log.Printf("Reverted booking failure for flight %d, seats %d", flightID, seats)
}

// createPermanentBooking creates a permanent booking in the database
func (bs *BookingServiceV2) createPermanentBooking(ctx context.Context, req *models.BookingRequest, totalAmount float64, paymentID string) (int, error) {
	query := `
		INSERT INTO bookings (user_id, flight_id, seats, total_amount, status, payment_id, date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var bookingID int
	err := bs.db.QueryRowContext(ctx, query, req.UserID, req.FlightID, req.Seats, totalAmount, models.BookingStatusConfirmed, paymentID, req.Date).Scan(&bookingID)
	if err != nil {
		return 0, fmt.Errorf("failed to create booking: %w", err)
	}

	// Cache the booking
	booking := &models.Booking{
		ID:          bookingID,
		UserID:      req.UserID,
		FlightID:    req.FlightID,
		Seats:       req.Seats,
		TotalAmount: totalAmount,
		Status:      models.BookingStatusConfirmed,
		PaymentID:   paymentID,
		Date:        req.Date,
		CreatedAt:   time.Now(),
	}

	cacheKey := database.GenerateBookingCacheKey(bookingID)
	if err := bs.cache.SetJSON(ctx, cacheKey, booking, 30*time.Minute); err != nil {
		log.Printf("Failed to cache booking: %v", err)
	}

	return bookingID, nil
}

// processPayment processes payment through the payment service
func (bs *BookingServiceV2) processPayment(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	url := fmt.Sprintf("%s/api/payments/process", bs.paymentServiceURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := bs.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make payment request: %w", err)
	}
	defer resp.Body.Close()

	var paymentResp models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return &paymentResp, nil
}

// GetBooking retrieves a booking by ID
func (bs *BookingServiceV2) GetBooking(ctx context.Context, bookingID int) (*models.Booking, error) {
	// Check cache first
	cacheKey := database.GenerateBookingCacheKey(bookingID)
	var booking models.Booking
	if err := bs.cache.GetJSON(ctx, cacheKey, &booking); err == nil {
		return &booking, nil
	}

	// Query from database
	query := `
		SELECT id, user_id, flight_id, seats, total_amount, status, payment_id, date, created_at
		FROM bookings
		WHERE id = $1
	`

	err := bs.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID, &booking.UserID, &booking.FlightID, &booking.Seats, &booking.TotalAmount,
		&booking.Status, &booking.PaymentID, &booking.Date, &booking.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to query booking: %w", err)
	}

	// Cache the result
	if err := bs.cache.SetJSON(ctx, cacheKey, booking, 30*time.Minute); err != nil {
		log.Printf("Failed to cache booking: %v", err)
	}

	return &booking, nil
}

// CancelBooking cancels a booking
func (bs *BookingServiceV2) CancelBooking(ctx context.Context, bookingID int) error {
	// Get booking first
	booking, err := bs.GetBooking(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}

	if !booking.CanCancel() {
		return fmt.Errorf("booking cannot be cancelled in current status: %s", booking.Status)
	}

	// Update booking status
	query := `UPDATE bookings SET status = $1 WHERE id = $2`
	_, err = bs.db.ExecContext(ctx, query, models.BookingStatusCancelled, bookingID)
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	// Increment seats back in Flight Service using the actual flight date
	if err := bs.incrementSeatsViaHTTP(ctx, booking.FlightID, booking.Seats, booking.Date); err != nil {
		log.Printf("Failed to increment seats on cancellation: %v", err)
		// Don't return error here as the booking is already cancelled in database
	}

	// Remove from cache
	cacheKey := database.GenerateBookingCacheKey(bookingID)
	bs.cache.Delete(ctx, cacheKey)

	return nil
}
