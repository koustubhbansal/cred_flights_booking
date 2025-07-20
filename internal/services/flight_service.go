package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"cred_flights_booking/internal/database"
	"cred_flights_booking/internal/models"
	"golang.org/x/sync/singleflight"
)

// FlightService handles flight-related operations
type FlightService struct {
	db    *database.DB
	cache *database.RedisClient
	// Singleflight group to prevent cache stampede
	searchGroup singleflight.Group
}

// NewFlightService creates a new flight service
func NewFlightService(db *database.DB, cache *database.RedisClient) *FlightService {
	return &FlightService{
		db:          db,
		cache:       cache,
		searchGroup: singleflight.Group{},
	}
}

// SearchFlights searches for flights with improved caching strategy
func (fs *FlightService) SearchFlights(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	// Generate cache key for search results (src, dest, date only)
	cacheKey := database.GenerateSearchCacheKey(req.Source, req.Destination, req.Date)

	// Try to get cached search results
	var cachedFlights []models.Flight
	if err := fs.cache.GetJSON(ctx, cacheKey, &cachedFlights); err == nil {
		log.Printf("Cache hit for search key: %s", cacheKey)
		// Filter flights based on available seats and sort
		paths := fs.filterAndSortFlights(cachedFlights, req.Seats, req.SortBy)
		return &models.SearchResponse{
			Paths: paths,
			Count: len(paths),
		}, nil
	}

	// Cache miss - use singleflight to prevent stampede
	searchKey := fmt.Sprintf("%s:%s:%s", req.Source, req.Destination, req.Date)
	flights, err, _ := fs.searchGroup.Do(searchKey, func() (interface{}, error) {
		return fs.searchFlightsFromDB(ctx, req.Source, req.Destination, req.Date)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search flights: %w", err)
	}

	flightList := flights.([]models.Flight)

	// Cache the search results for 2 hours
	if err := fs.cache.SetJSON(ctx, cacheKey, flightList, 2*time.Hour); err != nil {
		log.Printf("Failed to cache search results: %v", err)
	}

	// Filter flights based on available seats and sort
	paths := fs.filterAndSortFlights(flightList, req.Seats, req.SortBy)

	response := &models.SearchResponse{
		Paths: paths,
		Count: len(paths),
	}

	return response, nil
}

// searchFlightsFromDB searches flights from database (called by singleflight)
func (fs *FlightService) searchFlightsFromDB(ctx context.Context, source, destination, date string) ([]models.Flight, error) {
	// Parse date
	searchDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Search for flights
	paths, err := fs.findFlightPaths(ctx, source, destination, searchDate, 1) // Use 1 seat for search
	if err != nil {
		return nil, fmt.Errorf("failed to find flight paths: %w", err)
	}

	// Extract all unique flights from paths
	flightMap := make(map[int]models.Flight)
	for _, path := range paths {
		for _, flight := range path.Flights {
			flightMap[flight.ID] = flight
		}
	}

	// Convert to slice
	var flights []models.Flight
	for _, flight := range flightMap {
		flights = append(flights, flight)
	}

	return flights, nil
}

// filterAndSortFlights filters flights based on available seats and sorts them
func (fs *FlightService) filterAndSortFlights(flights []models.Flight, requestedSeats int, sortBy string) []models.FlightPath {
	var validPaths []models.FlightPath

	// Check seat availability for each flight
	for _, flight := range flights {
		availableSeats, err := fs.getAvailableSeats(context.Background(), flight.ID, flight.DepartureTime.Format("2006-01-02"))
		if err != nil {
			log.Printf("Failed to get available seats for flight %d: %v", flight.ID, err)
			continue
		}

		if availableSeats >= requestedSeats {
			path := models.FlightPath{
				Flights: []models.Flight{flight},
			}
			path.CalculateTotalPrice()
			path.CalculateTotalTime()
			path.CalculateStops()
			validPaths = append(validPaths, path)
		}
	}

	// Sort paths
	fs.sortFlightPaths(validPaths, sortBy)

	// Limit to top 20
	if len(validPaths) > 20 {
		validPaths = validPaths[:20]
	}

	return validPaths
}

// getAvailableSeats gets available seats from cache or database
func (fs *FlightService) getAvailableSeats(ctx context.Context, flightID int, date string) (int, error) {
	cacheKey := database.GenerateSeatCacheKey(flightID, date)

	// Try cache first
	if seats, err := fs.cache.Get(ctx, cacheKey).Int(); err == nil {
		return seats, nil
	}

	// Cache miss - get from database
	query := `
		SELECT total_seats - booked_seats
		FROM flights 
		WHERE id = $1 AND DATE(departure_time) = $2
	`

	var availableSeats int
	err := fs.db.QueryRowContext(ctx, query, flightID, date).Scan(&availableSeats)
	if err != nil {
		return 0, fmt.Errorf("failed to get available seats: %w", err)
	}

	// Cache the result for 1 hour
	if err := fs.cache.Set(ctx, cacheKey, availableSeats, time.Hour).Err(); err != nil {
		log.Printf("Failed to cache seat count: %v", err)
	}

	return availableSeats, nil
}

// ValidateFlight validates if a flight can be booked
func (fs *FlightService) ValidateFlight(ctx context.Context, flightID, seats int, date string) (*models.FlightValidationResponse, error) {
	// Get flight details
	query := `
		SELECT id, flight_number, source, destination, departure_time, arrival_time,
		       total_seats, booked_seats, price, created_at
		FROM flights 
		WHERE id = $1
	`

	var flight models.Flight
	err := fs.db.QueryRowContext(ctx, query, flightID).Scan(
		&flight.ID, &flight.FlightNumber, &flight.Source, &flight.Destination,
		&flight.DepartureTime, &flight.ArrivalTime, &flight.TotalSeats,
		&flight.BookedSeats, &flight.Price, &flight.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &models.FlightValidationResponse{
				Valid:   false,
				Message: "Flight not found",
			}, nil
		}
		return nil, fmt.Errorf("failed to query flight: %w", err)
	}

	// Get available seats from cache
	availableSeats, err := fs.getAvailableSeats(ctx, flightID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get available seats: %w", err)
	}

	canBook := availableSeats >= seats

	response := &models.FlightValidationResponse{
		Valid:     canBook,
		Price:     flight.Price * float64(seats),
		Available: availableSeats,
	}

	if !canBook {
		response.Message = fmt.Sprintf("Not enough seats available. Requested: %d, Available: %d", seats, availableSeats)
	}

	return response, nil
}

// DecrementSeats decrements available seats in cache (atomic operation)
func (fs *FlightService) DecrementSeats(ctx context.Context, flightID int, seats int, date string) error {
	cacheKey := database.GenerateSeatCacheKey(flightID, date)

	// Use Lua script for atomic decrement with validation
	luaScript := `
		local current = redis.call('GET', KEYS[1])
		if not current then
			return {err = 'Seat count not found in cache'}
		end
		local available = tonumber(current)
		local requested = tonumber(ARGV[1])
		if available < requested then
			return {err = 'Not enough seats available'}
		end
		redis.call('DECRBY', KEYS[1], requested)
		return {ok = available - requested}
	`

	result, err := fs.cache.Eval(ctx, luaScript, []string{cacheKey}, seats).Result()
	if err != nil {
		return fmt.Errorf("failed to decrement seats: %w", err)
	}

	if resultMap, ok := result.([]interface{}); ok && len(resultMap) > 0 {
		if errMsg, ok := resultMap[0].(string); ok && errMsg == "err" {
			return fmt.Errorf("seat decrement failed: %v", resultMap[1])
		}
	}

	log.Printf("Decremented %d seats for flight %d on %s", seats, flightID, date)
	return nil
}

// IncrementSeats increments available seats in cache (atomic operation)
func (fs *FlightService) IncrementSeats(ctx context.Context, flightID int, seats int, date string) error {
	cacheKey := database.GenerateSeatCacheKey(flightID, date)

	// Use atomic increment
	if err := fs.cache.IncrBy(ctx, cacheKey, int64(seats)).Err(); err != nil {
		return fmt.Errorf("failed to increment seats: %w", err)
	}

	log.Printf("Incremented %d seats for flight %d on %s", seats, flightID, date)
	return nil
}

// findFlightPaths finds all possible flight paths (direct and multi-stop)
func (fs *FlightService) findFlightPaths(ctx context.Context, source, destination string, date time.Time, seats int) ([]models.FlightPath, error) {
	var paths []models.FlightPath

	// Find direct flights
	directFlights, err := fs.findDirectFlights(ctx, source, destination, date, seats)
	if err != nil {
		return nil, err
	}

	for _, flight := range directFlights {
		path := models.FlightPath{
			Flights: []models.Flight{flight},
		}
		path.CalculateTotalPrice()
		path.CalculateTotalTime()
		path.CalculateStops()
		paths = append(paths, path)
	}

	// Find multi-stop flights (up to 3 stops)
	for stops := 1; stops <= 3; stops++ {
		multiStopPaths, err := fs.findMultiStopFlights(ctx, source, destination, date, seats, stops)
		if err != nil {
			log.Printf("Error finding %d-stop flights: %v", stops, err)
			continue
		}
		paths = append(paths, multiStopPaths...)
	}

	return paths, nil
}

// findDirectFlights finds direct flights between source and destination
func (fs *FlightService) findDirectFlights(ctx context.Context, source, destination string, date time.Time, seats int) ([]models.Flight, error) {
	query := `
		SELECT id, flight_number, source, destination, departure_time, arrival_time, 
		       total_seats, booked_seats, price, created_at
		FROM flights 
		WHERE source = $1 AND destination = $2 
		  AND DATE(departure_time) = $3 
		  AND (total_seats - booked_seats) >= $4
		ORDER BY departure_time
	`

	rows, err := fs.db.QueryContext(ctx, query, source, destination, date, seats)
	if err != nil {
		return nil, fmt.Errorf("failed to query direct flights: %w", err)
	}
	defer rows.Close()

	var flights []models.Flight
	for rows.Next() {
		var flight models.Flight
		err := rows.Scan(
			&flight.ID, &flight.FlightNumber, &flight.Source, &flight.Destination,
			&flight.DepartureTime, &flight.ArrivalTime, &flight.TotalSeats,
			&flight.BookedSeats, &flight.Price, &flight.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flight: %w", err)
		}
		flights = append(flights, flight)
	}

	return flights, nil
}

// findMultiStopFlights finds multi-stop flights using recursive CTE
func (fs *FlightService) findMultiStopFlights(ctx context.Context, source, destination string, date time.Time, seats int, maxStops int) ([]models.FlightPath, error) {
	// Build the recursive CTE query
	query := fs.buildMultiStopQuery(maxStops)

	rows, err := fs.db.QueryContext(ctx, query, source, destination, date, seats)
	if err != nil {
		return nil, fmt.Errorf("failed to query multi-stop flights: %w", err)
	}
	defer rows.Close()

	var paths []models.FlightPath
	pathMap := make(map[string]models.FlightPath)

	for rows.Next() {
		var flightIDs []int
		var flightNumbers []string
		var sources []string
		var destinations []string
		var departureTimes []time.Time
		var arrivalTimes []time.Time
		var totalSeats []int
		var bookedSeats []int
		var prices []float64
		var createdAt []time.Time

		err := rows.Scan(
			&flightIDs, &flightNumbers, &sources, &destinations,
			&departureTimes, &arrivalTimes, &totalSeats, &bookedSeats,
			&prices, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan multi-stop flight: %w", err)
		}

		// Build flight path
		var flights []models.Flight
		for i := range flightIDs {
			flight := models.Flight{
				ID:            flightIDs[i],
				FlightNumber:  flightNumbers[i],
				Source:        sources[i],
				Destination:   destinations[i],
				DepartureTime: departureTimes[i],
				ArrivalTime:   arrivalTimes[i],
				TotalSeats:    totalSeats[i],
				BookedSeats:   bookedSeats[i],
				Price:         prices[i],
				CreatedAt:     createdAt[i],
			}
			flights = append(flights, flight)
		}

		// Create unique key for this path
		pathKey := fs.generatePathKey(flights)
		if _, exists := pathMap[pathKey]; !exists {
			path := models.FlightPath{Flights: flights}
			path.CalculateTotalPrice()
			path.CalculateTotalTime()
			path.CalculateStops()
			pathMap[pathKey] = path
		}
	}

	// Convert map to slice
	for _, path := range pathMap {
		paths = append(paths, path)
	}

	return paths, nil
}

// buildMultiStopQuery builds the recursive CTE query for multi-stop flights
func (fs *FlightService) buildMultiStopQuery(maxStops int) string {
	return fmt.Sprintf(`
		WITH RECURSIVE flight_paths AS (
			-- Base case: direct flights
			SELECT 
				id, flight_number, source, destination, departure_time, arrival_time,
				total_seats, booked_seats, price, created_at,
				1 as stops,
				ARRAY[id] as flight_ids,
				ARRAY[flight_number] as flight_numbers,
				ARRAY[source] as sources,
				ARRAY[destination] as destinations,
				ARRAY[departure_time] as departure_times,
				ARRAY[arrival_time] as arrival_times,
				ARRAY[total_seats] as total_seats_array,
				ARRAY[booked_seats] as booked_seats_array,
				ARRAY[price] as prices,
				ARRAY[created_at] as created_ats
			FROM flights 
			WHERE source = $1 AND DATE(departure_time) = $3
			  AND (total_seats - booked_seats) >= $4
			
			UNION ALL
			
			-- Recursive case: add connecting flights
			SELECT 
				f.id, f.flight_number, f.source, f.destination, f.departure_time, f.arrival_time,
				f.total_seats, f.booked_seats, f.price, f.created_at,
				fp.stops + 1,
				fp.flight_ids || f.id,
				fp.flight_numbers || f.flight_number,
				fp.sources || f.source,
				fp.destinations || f.destination,
				fp.departure_times || f.departure_time,
				fp.arrival_times || f.arrival_time,
				fp.total_seats_array || f.total_seats,
				fp.booked_seats_array || f.booked_seats,
				fp.prices || f.price,
				fp.created_ats || f.created_at
			FROM flight_paths fp
			JOIN flights f ON fp.destinations[array_length(fp.destinations, 1)] = f.source
			WHERE fp.stops < %d
			  AND f.destination = $2
			  AND DATE(f.departure_time) = $3
			  AND (f.total_seats - f.booked_seats) >= $4
			  AND f.departure_time > fp.arrival_times[array_length(fp.arrival_times, 1)]
			  AND f.departure_time <= fp.arrival_times[array_length(fp.arrival_times, 1)] + INTERVAL '4 hours'
		)
		SELECT 
			flight_ids, flight_numbers, sources, destinations,
			departure_times, arrival_times, total_seats_array, booked_seats_array,
			prices, created_ats
		FROM flight_paths
		WHERE destinations[array_length(destinations, 1)] = $2
		ORDER BY stops, prices[1]
	`, maxStops)
}

// generatePathKey generates a unique key for a flight path
func (fs *FlightService) generatePathKey(flights []models.Flight) string {
	var keys []string
	for _, flight := range flights {
		keys = append(keys, fmt.Sprintf("%d", flight.ID))
	}
	return strings.Join(keys, "-")
}

// sortFlightPaths sorts flight paths by the specified criteria
func (fs *FlightService) sortFlightPaths(paths []models.FlightPath, sortBy string) {
	switch sortBy {
	case "cheapest":
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].TotalPrice < paths[j].TotalPrice
		})
	case "fastest":
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].TotalTime < paths[j].TotalTime
		})
	default:
		// Default to cheapest
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].TotalPrice < paths[j].TotalPrice
		})
	}
}
