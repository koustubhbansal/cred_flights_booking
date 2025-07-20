# Flight Booking and Management System

A microservices-based flight booking system with search, booking, and payment capabilities.

## System Architecture

The system consists of three main services with separate databases and shared Redis cache:

1. **Flight Service** - Handles flight search with support for multi-stop flights (up to 3 stops)
   - Owns its own PostgreSQL database (flights only)
   - Implements singleflight cache stampede protection
   - Exposes HTTP endpoints for seat validation and management
   - Uses Redis for search result caching and seat count management

2. **Booking Service** - Manages booking creation, payment processing, and booking lifecycle
   - Owns its own PostgreSQL database (bookings only)
   - Uses HTTP calls to Flight Service for flight validation and seat management
   - Implements temporary bookings in Redis cache
   - Handles payment integration via HTTP calls to Payment Service

3. **Payment Service** - Mock payment processing with configurable success/failure rates
   - Simulates real payment processing with configurable failure and timeout rates
   - Used by Booking Service via HTTP calls

## Features

- **Flight Search**: Direct and multi-stop flights (up to 3 stops)
- **Sorting**: By price (cheapest) and duration (fastest)
- **Caching**: Redis-based caching for flight search results with singleflight protection
- **Booking Flow**: Complete booking process with payment integration
- **Concurrent Handling**: Support for concurrent searches and bookings
- **Error Scenarios**: Payment failure, timeout, and other edge cases
- **Stress Testing**: Load testing for search and booking endpoints
- **Atomic Operations**: Lua scripts for seat count management

## Tech Stack

- **Language**: Go 1.22+
- **Database**: PostgreSQL (separate databases for flights and bookings)
- **Cache**: Redis (shared between services)
- **Containerization**: Docker & Docker Compose
- **API**: RESTful APIs using Go's net/http and ServeMux

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.22+ (for stress testing)
- jq (for API testing script)
- make (for simplified commands)

### Basic Setup
1. **Clone and setup**:
   ```bash
   git clone <repository>
   cd cred_flights_booking
   go mod tidy
   ```

2. **Start services**:
   ```bash
   make docker-up
   ```

3. **Test the system**:
   ```bash
   # Run API tests
   make test
   
   # Run stress tests
   make stress-test
   ```

**For detailed setup instructions, see [SETUP.md](SETUP.md)**

## Available Commands

The project includes a comprehensive Makefile for simplified operations:

```bash
# Show all available commands
make help

# Build all services
make build

# Start services with Docker Compose
make docker-up

# Stop services
make docker-down

# Restart services
make restart

# Show service logs
make logs

# Run API tests
make test

# Run stress tests
make stress-test

# Complete development setup
make dev-setup

# Clean build artifacts
make clean

# Format code
make fmt

# Run linter
make lint

# Reset database (removes all data)
make db-reset
```

## API Endpoints

### Flight Service (Port 8080)
- `GET /api/flights/search` - Search flights with filters
- `GET /api/flights/{id}` - Get flight details
- `POST /api/flights/validate` - Validate flight availability
- `POST /api/flights/seats/decrement` - Decrement available seats (atomic)
- `POST /api/flights/seats/increment` - Increment available seats (atomic)

### Booking Service (Port 8081)
- `POST /api/bookings` - Create a new booking
- `GET /api/bookings/{id}` - Get booking details
- `PUT /api/bookings/{id}/cancel` - Cancel booking

### Payment Service (Port 8082)
- `POST /api/payments/process` - Process payment (mock)

## Database Schema

### Flights Table
```sql
CREATE TABLE flights (
    id SERIAL PRIMARY KEY,
    flight_number VARCHAR(20) NOT NULL,
    source VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_time TIMESTAMP NOT NULL,
    arrival_time TIMESTAMP NOT NULL,
    total_seats INTEGER NOT NULL,
    booked_seats INTEGER DEFAULT 0,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Bookings Table
```sql
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    flight_id INTEGER NOT NULL,
    seats INTEGER NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payment_id VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Note**: The booking service has its own database and communicates with the flight service via HTTP for flight validation and seat management.

## Testing

### API Testing
```bash
# Run automated API tests
make test
```

### Stress Testing
The system includes comprehensive stress testing for:
- Concurrent flight searches
- Concurrent bookings
- Payment failure scenarios
- Payment timeout scenarios

Run stress tests with:
```bash
make stress-test
```

## Development

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- jq (for JSON processing in tests)

### Local Development
```bash
# Complete development setup (dependencies + services)
make dev-setup

# Or manually:
# Start dependencies
docker-compose up -d postgres-flights postgres-bookings redis

# Run services
make run
```

## Documentation

- **[SETUP.md](SETUP.md)** - Detailed setup and usage guide
- **[API Examples](SETUP.md#api-usage-examples)** - Complete API usage examples
- **[Troubleshooting](SETUP.md#troubleshooting)** - Common issues and solutions
- **[Monitoring](SETUP.md#monitoring-and-debugging)** - How to monitor and debug the system