# Testing Strategy and Validation

This document outlines the comprehensive testing strategy for the Flight Booking System, including response validation, expected value checks, and error scenario testing.

## �� Testing Overview

The Flight Booking System includes a robust testing infrastructure with:

- ✅ **Comprehensive response validation**
- ✅ **Expected value checks**
- ✅ **Error scenario testing**
- ✅ **Business logic validation**
- ✅ **Detailed test reporting**
- ✅ **Performance metrics**

## 📋 Test Scripts

### **1. Enhanced API Test Script (`scripts/test-api.sh`)**

#### **Features:**
- **Color-coded output** for easy reading
- **HTTP status code validation**
- **Response field validation**
- **Error scenario testing**
- **Test summary with pass/fail counts**

#### **Validation Examples:**

```bash
# Flight Search Validation
run_test "Flight Search - DEL to BOM" "200" "count" "1" \
    'curl -s -w "HTTP %{http_code}" "http://localhost:8080/api/flights/search?source=DEL&destination=BOM&date=2024-02-15&seats=2&sort_by=cheapest"'

# Booking Creation Validation
run_test "Booking Creation - Valid Request" "200" "booking_id" "1" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8081/api/bookings" -H "Content-Type: application/json" -d '"'"'{"user_id": 1, "flight_id": 1, "seats": 2, "date": "2024-02-15"}'"'"''

# Payment Status Validation
run_test "Payment Failure Simulation" "200" "status" "failed" \
    'curl -s -w "HTTP %{http_code}" -X POST "http://localhost:8082/api/payments/simulate/failure" -H "Content-Type: application/json" -d '"'"'{"booking_id": 2, "amount": 1000.0, "user_id": 2, "payment_type": "credit_card"}'"'"''
```

#### **Test Categories:**

| **Category** | **Tests** | **Validation** |
|--------------|-----------|----------------|
| **Flight Service** | 4 tests | Status codes, response fields, error handling |
| **Booking Service** | 4 tests | Booking creation, retrieval, error scenarios |
| **Payment Service** | 4 tests | Success, failure, timeout scenarios |
| **Error Scenarios** | 3 tests | Invalid endpoints, JSON, missing fields |

### **2. Enhanced Stress Test (`cmd/stress-test/main.go`)**

#### **Features:**
- **Response validation** for each request
- **Expected field checking**
- **Performance metrics**
- **Detailed error reporting**
- **Concurrent testing with validation**

#### **Validation Examples:**

```go
// Flight Search Validation
expectedFields := map[string]interface{}{
    "count": float64(1), // Should have at least one path
}
result := st.validateResponse(fmt.Sprintf("Flight Search User %d", userID), resp, http.StatusOK, expectedFields)

// Booking Validation
expectedFields := map[string]interface{}{}
expectedStatus := http.StatusOK
if bookingReq.FlightID > 10 { // Assume flights > 10 don't exist
    expectedStatus = http.StatusBadRequest
    expectedFields["status"] = "failed"
} else {
    expectedFields["booking_id"] = float64(1) // Should have a booking ID
}
result := st.validateResponse(fmt.Sprintf("Booking User %d", userID), resp, expectedStatus, expectedFields)
```

## 🔍 Validation Types

### **1. HTTP Status Code Validation**
```bash
# Expected: 200 OK
# Actual: 200 OK ✅ PASS
# Actual: 400 Bad Request ❌ FAIL
```

### **2. Response Field Validation**
```bash
# Expected: {"status": "success"}
# Actual: {"status": "success"} ✅ PASS
# Actual: {"status": "failed"} ❌ FAIL
```

### **3. Business Logic Validation**
```bash
# Flight search should return at least 1 path
# Booking creation should return a booking ID
# Payment failure should return "failed" status
```

### **4. Error Scenario Validation**
```bash
# Invalid flight ID should return 400
# Missing required fields should return 400
# Invalid JSON should return 400
```

## 📊 Test Results

### **Sample Output:**

```
=== Flight Booking System API Tests with Validation ===

🔵 Running: Flight Search - DEL to BOM
✅ PASSED: Flight Search - DEL to BOM

🔵 Running: Flight Validation - Valid Flight
✅ PASSED: Flight Validation - Valid Flight

🔵 Running: Booking Creation - Valid Request
✅ PASSED: Booking Creation - Valid Request

🔵 Running: Payment Failure Simulation
✅ PASSED: Payment Failure Simulation

=== Test Summary ===
Total Tests: 15
Passed: 15
Failed: 0

🎉 All tests passed!
```

### **Stress Test Output:**

```
=== Flight Search Stress Test ===
Starting flight search stress test with 10 concurrent users for 30s
Flight search test completed:
  Total requests: 150
  Successful: 148
  Failed: 2
  Success rate: 98.67%

=== Test Summary ===
Total Tests: 150
Passed: 148
Failed: 2
Success Rate: 98.67%

❌ 2 tests failed!
```

## 🚀 Running Tests

### **1. API Tests:**
```bash
# Run API tests
make test
```

### **2. Stress Tests:**
```bash
# Run stress tests
make stress-test
```

### **3. Both Tests:**
```bash
# Run both test suites
make test && make stress-test
```

### **4. Development Workflow:**
```bash
# Show all available commands
make help

# Format code before testing
make fmt

# Run linter
make lint

# Show service logs during testing
make logs
```

## 🎯 Test Coverage

### **API Endpoints Tested:**

| **Service** | **Endpoint** | **Method** | **Validation** |
|-------------|--------------|------------|----------------|
| Flight | `/api/flights/search` | GET | Response structure, count field |
| Flight | `/api/flights/validate` | POST | Valid/invalid flight responses |
| Flight | `/health` | GET | Service health status |
| Booking | `/api/bookings` | POST | Booking creation, ID validation |
| Booking | `/api/bookings` | GET | Booking retrieval |
| Booking | `/health` | GET | Service health status |
| Payment | `/api/payments/process` | POST | Payment processing |
| Payment | `/api/payments/simulate/*` | POST | Success/failure/timeout scenarios |
| Payment | `/health` | GET | Service health status |

### **Error Scenarios Tested:**

| **Scenario** | **Expected Response** | **Validation** |
|--------------|----------------------|----------------|
| Invalid endpoint | 404 Not Found | HTTP status code |
| Invalid JSON | 400 Bad Request | HTTP status code |
| Missing fields | 400 Bad Request | HTTP status code |
| Invalid flight ID | 400 Bad Request | Response status field |
| Non-existent flight | Valid=false | Response validation field |

## 🔧 Customization

### **Adding New Tests:**

#### **API Test Script:**
```bash
# Add new test to scripts/test-api.sh
run_test "Test Name" "expected_status" "field_name" "expected_value" \
    'curl -s -w "HTTP %{http_code}" "http://localhost:8080/api/endpoint"'
```

#### **Stress Test:**
```go
// Add new validation to cmd/stress-test/main.go
expectedFields := map[string]interface{}{
    "field_name": "expected_value",
}
result := st.validateResponse("Test Name", resp, http.StatusOK, expectedFields)
```

### **Modifying Expected Values:**

#### **Change Expected Response:**
```bash
# Update expected value in test script
run_test "Test Name" "200" "status" "new_expected_value" \
    'curl command'
```

#### **Change Validation Logic:**
```go
// Update validation logic in stress test
if someCondition {
    expectedFields["status"] = "success"
} else {
    expectedFields["status"] = "failed"
}
```

## 📈 Performance Metrics

### **Measured Metrics:**
- **Response Time**: Duration of each request
- **Success Rate**: Percentage of successful requests
- **Error Rate**: Percentage of failed requests
- **Concurrent Users**: Number of simultaneous requests
- **Throughput**: Requests per second

### **Sample Metrics:**
```
Flight Search Test:
  Total requests: 150
  Average response time: 45ms
  Success rate: 98.67%
  Throughput: 5.0 req/sec

Booking Test:
  Total bookings: 75
  Average response time: 120ms
  Success rate: 96.00%
  Throughput: 2.5 req/sec
```

## 🎉 Key Features

### **Comprehensive Testing Infrastructure:**

1. **🔍 Response Validation**: Every response is validated against expected values
2. **📊 Detailed Reporting**: Clear pass/fail status with error details
3. **🎯 Business Logic Testing**: Validates actual business requirements
4. **⚡ Performance Monitoring**: Tracks response times and throughput
5. **🛡️ Error Scenario Coverage**: Tests both success and failure paths
6. **🔄 Continuous Validation**: Real-time validation during stress testing

### **Testing Capabilities:**

| **Feature** | **Description** |
|-------------|-----------------|
| **HTTP Status Validation** | Validates correct HTTP status codes |
| **Response Field Validation** | Checks specific fields in JSON responses |
| **Business Logic Validation** | Ensures business rules are followed |
| **Error Scenario Testing** | Tests error handling and edge cases |
| **Performance Metrics** | Measures response times and throughput |
| **Concurrent Testing** | Tests system under load |
| **Detailed Reporting** | Provides comprehensive test results |

This testing infrastructure ensures the Flight Booking System is thoroughly validated and reliable in production environments. 