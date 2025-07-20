package models

import (
	"time"
)

// Flight represents a single flight
type Flight struct {
	ID            int       `json:"id" db:"id"`
	FlightNumber  string    `json:"flight_number" db:"flight_number"`
	Source        string    `json:"source" db:"source"`
	Destination   string    `json:"destination" db:"destination"`
	DepartureTime time.Time `json:"departure_time" db:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time" db:"arrival_time"`
	TotalSeats    int       `json:"total_seats" db:"total_seats"`
	BookedSeats   int       `json:"booked_seats" db:"booked_seats"`
	Price         float64   `json:"price" db:"price"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// FlightPath represents a complete flight path (can be direct or multi-stop)
type FlightPath struct {
	Flights    []Flight `json:"flights"`
	TotalPrice float64  `json:"total_price"`
	TotalTime  int64    `json:"total_time_minutes"` // in minutes
	Stops      int      `json:"stops"`
}

// SearchRequest represents a flight search request
type SearchRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
	Seats       int    `json:"seats"`
	SortBy      string `json:"sort_by"` // "cheapest" or "fastest"
}

// SearchResponse represents the response for flight search
type SearchResponse struct {
	Paths []FlightPath `json:"paths"`
	Count int          `json:"count"`
}

// FlightValidationRequest represents a flight validation request
type FlightValidationRequest struct {
	FlightID int    `json:"flight_id"`
	Seats    int    `json:"seats"`
	Date     string `json:"date"`
}

// FlightValidationResponse represents the response for flight validation
type FlightValidationResponse struct {
	Valid     bool    `json:"valid"`
	Message   string  `json:"message,omitempty"`
	Price     float64 `json:"price,omitempty"`
	Available int     `json:"available_seats,omitempty"`
}

// SeatUpdateRequest represents a seat update request
type SeatUpdateRequest struct {
	FlightID int    `json:"flight_id"`
	Seats    int    `json:"seats"`
	Date     string `json:"date"`
}

// AvailableSeats returns the number of available seats
func (f *Flight) AvailableSeats() int {
	return f.TotalSeats - f.BookedSeats
}

// CanBook checks if the flight can be booked for the given number of seats
func (f *Flight) CanBook(seats int) bool {
	return f.AvailableSeats() >= seats
}

// CalculateTotalTime calculates total travel time in minutes
func (fp *FlightPath) CalculateTotalTime() {
	if len(fp.Flights) == 0 {
		fp.TotalTime = 0
		return
	}

	firstFlight := fp.Flights[0]
	lastFlight := fp.Flights[len(fp.Flights)-1]

	duration := lastFlight.ArrivalTime.Sub(firstFlight.DepartureTime)
	fp.TotalTime = int64(duration.Minutes())
}

// CalculateTotalPrice calculates total price for all flights
func (fp *FlightPath) CalculateTotalPrice() {
	fp.TotalPrice = 0
	for _, flight := range fp.Flights {
		fp.TotalPrice += flight.Price
	}
}

// CalculateStops calculates number of stops
func (fp *FlightPath) CalculateStops() {
	if len(fp.Flights) <= 1 {
		fp.Stops = 0
	} else {
		fp.Stops = len(fp.Flights) - 1
	}
}
