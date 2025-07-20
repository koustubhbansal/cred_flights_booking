.PHONY: help build run test clean docker-build docker-up docker-down stress-test deps fmt lint logs restart db-reset dev-setup

# Default target
help:
	@echo "Flight Booking System - Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build         - Build all services"
	@echo "  run           - Run services locally (requires PostgreSQL and Redis)"
	@echo "  clean         - Clean build artifacts"
	@echo ""
	@echo "Docker Operations:"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-up     - Start all services with Docker Compose"
	@echo "  docker-down   - Stop all services"
	@echo "  restart       - Restart all services"
	@echo "  logs          - Show service logs"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run API tests"
	@echo "  stress-test   - Run stress tests"
	@echo ""
	@echo "Development:"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  dev-setup     - Complete development setup"
	@echo ""
	@echo "Database:"
	@echo "  db-reset      - Reset database (removes all data)"

# Build all services
build:
	@echo "Building Flight Service..."
	go build -o bin/flight-service ./cmd/flight-service
	@echo "Building Booking Service..."
	go build -o bin/booking-service ./cmd/booking-service
	@echo "Building Payment Service..."
	go build -o bin/payment-service ./cmd/payment-service
	@echo "Building Stress Test..."
	go build -o bin/stress-test ./cmd/stress-test

# Run services locally (requires PostgreSQL and Redis)
run: build
	@echo "Starting services locally..."
	@echo "Make sure PostgreSQL and Redis are running!"
	@echo "Flight Service: http://localhost:8080"
	@echo "Booking Service: http://localhost:8081"
	@echo "Payment Service: http://localhost:8082"
	@echo ""
	@echo "Starting Flight Service..."
	./bin/flight-service &
	@echo "Starting Booking Service..."
	./bin/booking-service &
	@echo "Starting Payment Service..."
	./bin/payment-service &
	@echo ""
	@echo "All services started. Press Ctrl+C to stop."
	@wait

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	@if [ ! -f "./Dockerfile.flight" ] || [ ! -f "./Dockerfile.booking" ] || [ ! -f "./Dockerfile.payment" ]; then \
		echo "Error: One or more Dockerfiles not found"; \
		exit 1; \
	fi
	docker build -f Dockerfile.flight -t flight-service .
	docker build -f Dockerfile.booking -t booking-service .
	docker build -f Dockerfile.payment -t payment-service .
	@echo "Docker images built successfully!"

# Start services with Docker Compose
docker-up:
	@echo "Starting services with Docker Compose..."
	@if [ ! -f "./docker-compose.yml" ]; then \
		echo "Error: docker-compose.yml not found"; \
		exit 1; \
	fi
	@if ! command -v docker-compose >/dev/null 2>&1; then \
		echo "Error: docker-compose not found. Please install Docker Compose."; \
		exit 1; \
	fi
	docker-compose up -d
	@echo "Services started!"
	@echo "Flight Service: http://localhost:8080"
	@echo "Booking Service: http://localhost:8081"
	@echo "Payment Service: http://localhost:8082"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

# Stop services
docker-down:
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped!"

# Run API tests
test:
	@echo "Running API tests..."
	@if [ ! -f "./scripts/test-api.sh" ]; then \
		echo "Error: test script not found at ./scripts/test-api.sh"; \
		exit 1; \
	fi
	@if command -v jq >/dev/null 2>&1; then \
		chmod +x ./scripts/test-api.sh; \
		./scripts/test-api.sh; \
	else \
		echo "jq is required for API tests. Please install jq first."; \
		exit 1; \
	fi

# Run stress tests
stress-test: build
	@echo "Running stress tests..."
	@if [ ! -f "./bin/stress-test" ]; then \
		echo "Error: stress test binary not found. Run 'make build' first."; \
		exit 1; \
	fi
	@echo "Make sure all services are running!"
	./bin/stress-test

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	@echo "Clean completed!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	@echo "Dependencies installed!"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted!"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Please install it first."; \
		exit 1; \
	fi

# Show logs
logs:
	@echo "Showing service logs..."
	docker-compose logs -f

# Restart services
restart: docker-down docker-up
	@echo "Services restarted!"

# Database operations
db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres redis
	@echo "Database reset completed!"

# Development setup
dev-setup: deps docker-up
	@echo "Development environment setup completed!"
	@echo "Services are running at:"
	@echo "  Flight Service: http://localhost:8080"
	@echo "  Booking Service: http://localhost:8081"
	@echo "  Payment Service: http://localhost:8082" 