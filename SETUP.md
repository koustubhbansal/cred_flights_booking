# Flight Booking System - Setup Guide

This guide provides detailed setup and usage instructions for the flight booking system.

## Prerequisites

### Required Software
- **Docker Desktop** - Download from [docker.com](https://www.docker.com/products/docker-desktop/)
- **Go 1.22+** - Download from [golang.org/dl](https://golang.org/dl/) or use `brew install go` (Mac)
- **jq** - For JSON formatting in test scripts
  - Mac: `brew install jq`
  - Ubuntu: `sudo apt-get install jq`
  - CentOS: `sudo yum install jq`
- **make** - For simplified commands (usually pre-installed)
- **curl** - Usually pre-installed on most systems

### Verify Installation
```bash
docker --version
docker-compose --version
go version
jq --version
curl --version
```

## Quick Start

### 1. Clone and Setup Project
```bash
# Clone the repository
git clone <repository-url>
cd cred_flights_booking

# Download Go dependencies
go mod tidy
```

### 2. Start All Services
```bash
# Start all services with Docker Compose
make docker-up

# Check service status
docker-compose ps
```

This will start:
- `postgres-flights` (Port 5432) - Flight service database
- `postgres-bookings` (Port 5433) - Booking service database  
- `redis` (Port 6379) - Shared cache
- `flight-service` (Port 8080) - Flight search and management
- `booking-service` (Port 8081) - Booking creation and management
- `payment-service` (Port 8082) - Mock payment processing

### 3. Wait for Services to be Ready
```bash
# Wait 30-60 seconds for services to initialize
sleep 45
```

### 4. Test the System
```bash
# Run automated API tests
make test

# Run stress tests
make stress-test
```

### 5. Development Workflow
```bash
# Show all available commands
make help

# Format code
make fmt

# Run linter
make lint

# Show service logs
make logs

# Restart services
make restart

# Reset database (removes all data)
make db-reset
```

## Service Details

### Flight Service (Port 8080)

**Database**: `flights_db` (PostgreSQL on port 5432)

**Key Features**:
- Multi-stop flight search (up to 3 stops)
- Singleflight cache stampede protection
- Atomic seat count management via Redis
- Search result caching (2-hour TTL)

**Endpoints**:
- `GET /api/flights/search` - Search flights
- `POST /api/flights/validate` - Validate flight availability
- `POST /api/flights/seats/decrement` - Decrement seats (atomic)
- `POST /api/flights/seats/increment` - Increment seats (atomic)

**Cache Keys**:
- Search results: `flight_search:{source}:{destination}:{date}`
- Seat counts: `flight_seats:{flight_id}:{date}`

### Booking Service (Port 8081)

**Database**: `bookings_db` (PostgreSQL on port 5433)

**Key Features**:
- Temporary bookings in Redis (15-minute expiry)
- HTTP communication with flight service
- Payment integration via HTTP
- Automatic rollback on failures

**Endpoints**:
- `POST /api/bookings` - Create booking
- `GET /api/bookings/{id}` - Get booking details
- `PUT /api/bookings/{id}/cancel` - Cancel booking

**Cache Keys**:
- Temporary bookings: `temp_booking:{user_id}:{flight_id}`
- Confirmed bookings: `booking:{booking_id}`

### Payment Service (Port 8082)

**Features**:
- Mock payment processing
- Configurable failure rates
- Timeout simulation
- Success/failure scenarios

**Endpoints**:
- `POST /api/payments/process` - Process payment

## API Usage Examples

### Flight Search

```bash
# Search for direct flights
curl "http://localhost:8080/api/flights/search?source=DEL&destination=BOM&date=2024-02-15&seats=2&sort_by=cheapest"

# Search for fastest flights
curl "http://localhost:8080/api/flights/search?source=DEL&destination=BLR&date=2024-02-15&seats=1&sort_by=fastest"
```

### Flight Validation

```bash
# Validate flight availability
curl -X POST "http://localhost:8080/api/flights/validate" \
  -H "Content-Type: application/json" \
  -d '{"flight_id": 1, "seats": 2, "date": "2024-02-15"}'
```

### Booking Creation

```bash
# Create a booking
curl -X POST "http://localhost:8081/api/bookings" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "flight_id": 1,
    "seats": 2,
    "date": "2024-02-15"
  }'
```

### Payment Processing

```bash
# Process payment
curl -X POST "http://localhost:8082/api/payments/process" \
  -H "Content-Type: application/json" \
  -d '{"booking_id": 1, "amount": 17000, "user_id": 1, "payment_type": "credit_card"}'
```

## Caching Strategy

### Flight Search Cache
- **Key**: `flight_search:{source}:{destination}:{date}`
- **TTL**: 2 hours
- **Content**: All flights for the route (not filtered by seats)
- **Protection**: Singleflight prevents cache stampede

### Seat Count Cache
- **Key**: `flight_seats:{flight_id}:{date}`
- **TTL**: 1 hour
- **Content**: Available seats count
- **Operations**: Atomic INCR/DECR with Lua script validation

### Temporary Booking Cache
- **Key**: `temp_booking:{user_id}:{flight_id}`
- **TTL**: 15 minutes
- **Content**: Temporary booking details during payment processing
- **Cleanup**: Automatic expiry or manual cleanup on success/failure

## Error Handling

### Booking Flow Failures

1. **Flight Validation Failure**: Returns error immediately
2. **Seat Reservation Failure**: Cleans up temporary booking
3. **Payment Failure**: Reverts seat count and cleans up temporary booking
4. **Database Failure**: Reverts all changes (seats, temporary booking)

### Cache Consistency

- Seat counts are always reverted on booking failure
- Temporary bookings are cleaned up on success or failure
- Search cache is independent and doesn't affect booking consistency

## Development Commands

The project includes a comprehensive Makefile with the following commands:

### Build & Run
- `make build` - Build all services
- `make run` - Run services locally (requires PostgreSQL and Redis)
- `make clean` - Clean build artifacts

### Docker Operations
- `make docker-up` - Start all services with Docker Compose
- `make docker-down` - Stop all services
- `make restart` - Restart all services
- `make logs` - Show service logs

### Testing
- `make test` - Run API tests (requires jq)
- `make stress-test` - Run stress tests

### Development
- `make deps` - Install dependencies
- `make fmt` - Format code
- `make lint` - Run linter (requires golangci-lint)
- `make dev-setup` - Complete development setup

### Database
- `make db-reset` - Reset database (removes all data)

### Help
- `make help` - Show all available commands

## Monitoring and Debugging

### Check Service Logs

```bash
# View all service logs
make logs

# View specific service logs
docker-compose logs -f flight-service
docker-compose logs -f booking-service
docker-compose logs -f payment-service
```

### Check Redis Cache

```bash
# Connect to Redis
docker exec -it cred_flights_booking-redis-1 redis-cli

# List all keys
KEYS *

# Get specific cache values
GET "flight_search:DEL:BOM:2024-02-15"
GET "flight_seats:1:2024-02-15"
```

### Check Database Data

```bash
# Connect to flights database
docker exec -it cred_flights_booking-postgres-flights-1 psql -U postgres -d flights_db

# Connect to bookings database
docker exec -it cred_flights_booking-postgres-bookings-1 psql -U postgres -d bookings_db
```

## Development

### Local Development Setup

```bash
# Start only dependencies
docker-compose up -d postgres-flights postgres-bookings redis

# Run services locally
make run
```

### Environment Variables

**Flight Service**:
- `DB_HOST=localhost` (or `postgres-flights` in Docker)
- `DB_PORT=5432`
- `DB_NAME=flights_db`
- `REDIS_HOST=localhost` (or `redis` in Docker)

**Booking Service**:
- `DB_HOST=localhost` (or `postgres-bookings` in Docker)
- `DB_PORT=5432`
- `DB_NAME=bookings_db`
- `FLIGHT_SERVICE_URL=http://localhost:8080`
- `PAYMENT_SERVICE_URL=http://localhost:8082`

## Troubleshooting

### Common Issues

1. **Service Connection Errors**:
   - Check if all services are running: `docker-compose ps`
   - Verify network connectivity: `docker network ls`

2. **Database Connection Issues**:
   - Check database logs: `docker-compose logs postgres-flights`
   - Verify database initialization: Check if tables exist

3. **Redis Connection Issues**:
   - Check Redis logs: `docker-compose logs redis`
   - Test Redis connectivity: `docker exec -it cred_flights_booking-redis-1 redis-cli ping`

4. **Booking Failures**:
   - Check seat availability in cache
   - Verify payment service is responding
   - Check booking service logs for HTTP call failures

5. **Test Script Issues**:
   - Ensure `jq` is installed: `jq --version`
   - Use `make test` which handles script permissions automatically
   - Check if services are ready before running tests

### Performance Tuning

1. **Cache TTL Adjustment**:
   - Search cache: 2 hours (good balance between performance and freshness)
   - Seat cache: 1 hour (frequent updates)
   - Temporary bookings: 15 minutes (payment processing time)

2. **Database Optimization**:
   - Indexes are created automatically
   - Connection pooling is configured
   - Consider read replicas for high load

3. **Concurrency Handling**:
   - Singleflight prevents cache stampede
   - Atomic operations in Redis prevent race conditions
   - Proper rollback mechanisms ensure consistency

## Security Considerations

1. **Database Security**:
   - Use strong passwords in production
   - Enable SSL connections
   - Implement proper access controls

2. **API Security**:
   - Add authentication/authorization
   - Implement rate limiting
   - Use HTTPS in production

3. **Cache Security**:
   - Secure Redis access
   - Implement cache key validation
   - Monitor for cache poisoning attacks 