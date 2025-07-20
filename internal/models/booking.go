package models

import (
	"time"
)

// Booking represents a flight booking
type Booking struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	FlightID    int       `json:"flight_id" db:"flight_id"`
	Seats       int       `json:"seats" db:"seats"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	Status      string    `json:"status" db:"status"`
	PaymentID   string    `json:"payment_id,omitempty" db:"payment_id"`
	Date        string    `json:"date" db:"date"` // Flight date
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Flight      *Flight   `json:"flight,omitempty" db:"-"`
}

// BookingRequest represents a booking request
type BookingRequest struct {
	UserID   int    `json:"user_id"`
	FlightID int    `json:"flight_id"`
	Seats    int    `json:"seats"`
	Date     string `json:"date"`
}

// TempBooking represents a temporary booking in cache
type TempBooking struct {
	UserID      int       `json:"user_id"`
	FlightID    int       `json:"flight_id"`
	Seats       int       `json:"seats"`
	TotalAmount float64   `json:"total_amount"`
	Date        string    `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// BookingResponse represents the response for booking
type BookingResponse struct {
	BookingID   int     `json:"booking_id"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"total_amount"`
	PaymentID   string  `json:"payment_id,omitempty"`
	Message     string  `json:"message,omitempty"`
}

// BookingStatus constants
const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusFailed    = "failed"
	BookingStatusCancelled = "cancelled"
)

// IsValidStatus checks if the booking status is valid
func (b *Booking) IsValidStatus() bool {
	validStatuses := []string{
		BookingStatusPending,
		BookingStatusConfirmed,
		BookingStatusFailed,
		BookingStatusCancelled,
	}

	for _, status := range validStatuses {
		if b.Status == status {
			return true
		}
	}
	return false
}

// CanCancel checks if the booking can be cancelled
func (b *Booking) CanCancel() bool {
	return b.Status == BookingStatusPending || b.Status == BookingStatusConfirmed
}
