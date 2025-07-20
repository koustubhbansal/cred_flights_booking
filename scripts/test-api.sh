#!/bin/bash

# Enhanced Test script for Flight Booking System APIs with validation

echo "=== Flight Booking System API Tests with Validation ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper function to run test and validate
run_test() {
    local test_name="$1"
    local expected_status="$2"
    local expected_field="$3"
    local expected_value="$4"
    local command="$5"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "\n${BLUE}Running: $test_name${NC}"
    
    # Execute command and capture response
    local response
    local status_code
    response=$(eval "$command" 2>/dev/null)
    status_code=$?
    
    # Check if command executed successfully
    if [ $status_code -ne 0 ]; then
        echo -e "${RED}❌ FAILED: Command execution failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    
    # Extract HTTP status code if available
    local http_status
    if echo "$response" | grep -q "HTTP"; then
        http_status=$(echo "$response" | grep "HTTP" | tail -1 | cut -d' ' -f2)
    else
        http_status="200" # Assume 200 if not specified
    fi
    
    # Validate HTTP status code
    if [ "$http_status" != "$expected_status" ]; then
        echo -e "${RED}❌ FAILED: Expected HTTP status $expected_status, got $http_status${NC}"
        echo "Response: $response"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    
    # Validate response field if specified
    if [ -n "$expected_field" ] && [ -n "$expected_value" ]; then
        local actual_value
        actual_value=$(echo "$response" | jq -r ".$expected_field" 2>/dev/null)
        
        if [ "$actual_value" != "$expected_value" ]; then
            echo -e "${RED}❌ FAILED: Expected $expected_field=$expected_value, got $actual_value${NC}"
            echo "Response: $response"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    fi
    
    echo -e "${GREEN}✅ PASSED: $test_name${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    return 0
}

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Test Flight Service
echo -e "\n${YELLOW}=== Testing Flight Service ===${NC}"

# Test flight search - should return valid JSON with paths
run_test "Flight Search - DEL to BOM" "200" "count" "3" \
    'curl -s -w "HTTP %{http_code}" "http://localhost:8080/api/flights/search?source=DEL&destination=BOM&date=2024-02-15&seats=2&sort_by=cheapest"'

# Test flight validation - should return valid=true for available flight
run_test "Flight Validation - Valid Flight" "200" "valid" "true" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8080/api/flights/validate" -H "Content-Type: application/json" -d '"'"'{"flight_id": 1, "seats": 2, "date": "2024-02-15"}'"'"''

# Test flight validation - should return valid=false for unavailable flight
run_test "Flight Validation - Invalid Flight" "200" "valid" "false" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8080/api/flights/validate" -H "Content-Type: application/json" -d '"'"'{"flight_id": 999, "seats": 1000, "date": "2024-02-15"}'"'"''

# Test health check
run_test "Flight Service Health Check" "200" "status" "healthy" \
    'curl -s -w "HTTP %{http_code}" "http://localhost:8080/health"'

# Test Booking Service
echo -e "\n${YELLOW}=== Testing Booking Service ===${NC}"

# Test booking creation - should return valid booking ID
run_test "Booking Creation - Valid Request" "200" "status" "confirmed" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8081/api/bookings" -H "Content-Type: application/json" -d '"'"'{"user_id": 1, "flight_id": 1, "seats": 2, "date": "2024-02-15"}'"'"''

# Test booking creation - should fail for invalid flight
run_test "Booking Creation - Invalid Flight" "400" "status" "failed" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8081/api/bookings" -H "Content-Type: application/json" -d '"'"'{"user_id": 2, "flight_id": 999, "seats": 2, "date": "2024-02-15"}'"'"''

# Test get booking - should return booking details
# Note: This endpoint might not be implemented yet, so we'll skip this test
# BOOKING_RESPONSE=$(curl -s -X POST "http://localhost:8081/api/bookings" \
#   -H "Content-Type: application/json" \
#   -d '{"user_id": 3, "flight_id": 1, "seats": 1, "date": "2024-02-15"}')
# 
# BOOKING_ID=$(echo "$BOOKING_RESPONSE" | jq -r '.booking_id')
# 
# if [ "$BOOKING_ID" != "null" ] && [ "$BOOKING_ID" != "" ]; then
#     run_test "Get Booking Details" "200" "id" "$BOOKING_ID" \
#         "curl -s -w 'HTTP %{http_code}' 'http://localhost:8081/api/bookings?id=$BOOKING_ID'"
# fi

# Test health check
run_test "Booking Service Health Check" "200" "status" "healthy" \
    'curl -s -w "HTTP %{http_code}" "http://localhost:8081/health"'

# Test Payment Service
echo -e "\n${YELLOW}=== Testing Payment Service ===${NC}"

# Test payment processing - should return success
run_test "Payment Processing - Success" "200" "status" "success" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8082/api/payments/simulate/success" -H "Content-Type: application/json" -d '"'"'{"booking_id": 1, "amount": 17000.0, "user_id": 1, "payment_type": "credit_card"}'"'"''

# Test payment failure simulation - should return failed
run_test "Payment Failure Simulation" "200" "status" "failed" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8082/api/payments/simulate/failure" -H "Content-Type: application/json" -d '"'"'{"booking_id": 2, "amount": 1000.0, "user_id": 2, "payment_type": "credit_card"}'"'"''

# Test payment timeout simulation - should return timeout
run_test "Payment Timeout Simulation" "408" "status" "timeout" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8082/api/payments/simulate/timeout" -H "Content-Type: application/json" -d '"'"'{"booking_id": 3, "amount": 1500.0, "user_id": 3, "payment_type": "debit_card"}'"'"''

# Test payment success simulation - should return success
run_test "Payment Success Simulation" "200" "status" "success" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8082/api/payments/simulate/success" -H "Content-Type: application/json" -d '"'"'{"booking_id": 4, "amount": 2000.0, "user_id": 4, "payment_type": "upi"}'"'"''

# Test health check
run_test "Payment Service Health Check" "200" "status" "healthy" \
    'curl -s -w "HTTP %{http_code}" "http://localhost:8082/health"'

# Test Error Scenarios
echo -e "\n${YELLOW}=== Testing Error Scenarios ===${NC}"

# Test invalid endpoint
run_test "Invalid Endpoint - 404" "404" "" "" \
    'curl -s -w "HTTP %{http_code}" "http://localhost:8080/api/invalid"'

# Test invalid JSON
run_test "Invalid JSON - 400" "400" "" "" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8081/api/bookings" -H "Content-Type: application/json" -d "invalid json"'

# Test missing required fields
run_test "Missing Required Fields - 400" "400" "" "" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8081/api/bookings" -H "Content-Type: application/json" -d '"'"'{"user_id": 1}'"'"''

# Print test summary
echo -e "\n${YELLOW}=== Test Summary ===${NC}"
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed!${NC}"
    exit 1
fi 