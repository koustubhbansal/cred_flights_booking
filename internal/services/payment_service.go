package services

import (
	"context"
	"log"
	"math/rand"
	"time"

	"cred_flights_booking/internal/models"

	"github.com/google/uuid"
)

// PaymentService handles payment processing
type PaymentService struct {
	// Mock configuration for different scenarios
	failureRate    float64       // Percentage of payments that should fail
	timeoutRate    float64       // Percentage of payments that should timeout
	processingTime time.Duration // Average processing time
}

// NewPaymentService creates a new payment service
func NewPaymentService() *PaymentService {
	return &PaymentService{
		failureRate:    0.15,            // 15% failure rate
		timeoutRate:    0.05,            // 5% timeout rate
		processingTime: 2 * time.Second, // 2 seconds average processing time
	}
}

// ProcessPayment processes a payment request with mock scenarios
func (ps *PaymentService) ProcessPayment(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
	log.Printf("Processing payment for booking %d, amount: %.2f", req.BookingID, req.Amount)

	// Validate payment type
	if !models.IsValidPaymentType(req.PaymentType) {
		return &models.PaymentResponse{
			PaymentID:   "",
			Status:      models.PaymentStatusFailed,
			Message:     "Invalid payment type",
			BookingID:   req.BookingID,
			Amount:      req.Amount,
			ProcessedAt: time.Now(),
		}, nil
	}

	// Simulate processing time
	processingTime := ps.processingTime + time.Duration(rand.Intn(3000))*time.Millisecond

	// Check for timeout scenario
	select {
	case <-ctx.Done():
		return &models.PaymentResponse{
			PaymentID:   "",
			Status:      models.PaymentStatusTimeout,
			Message:     "Payment processing timeout",
			BookingID:   req.BookingID,
			Amount:      req.Amount,
			ProcessedAt: time.Now(),
		}, nil
	case <-time.After(processingTime):
		// Continue processing
	}

	// Simulate random scenarios
	rand.Seed(time.Now().UnixNano())
	randomValue := rand.Float64()

	// Determine payment outcome
	var status string
	var message string

	switch {
	case randomValue < ps.timeoutRate:
		// Timeout scenario
		status = models.PaymentStatusTimeout
		message = "Payment gateway timeout"

	case randomValue < ps.timeoutRate+ps.failureRate:
		// Failure scenario
		status = models.PaymentStatusFailed
		message = ps.getRandomFailureMessage()

	default:
		// Success scenario
		status = models.PaymentStatusSuccess
		message = "Payment processed successfully"
	}

	// Generate payment ID
	paymentID := ""
	if status == models.PaymentStatusSuccess {
		paymentID = uuid.New().String()
	}

	response := &models.PaymentResponse{
		PaymentID:   paymentID,
		Status:      status,
		Message:     message,
		BookingID:   req.BookingID,
		Amount:      req.Amount,
		ProcessedAt: time.Now(),
	}

	log.Printf("Payment processed for booking %d: %s - %s", req.BookingID, status, message)
	return response, nil
}

// getRandomFailureMessage returns a random failure message
func (ps *PaymentService) getRandomFailureMessage() string {
	failureMessages := []string{
		"Insufficient funds",
		"Card declined",
		"Invalid card number",
		"Expired card",
		"CVV mismatch",
		"Bank declined transaction",
		"Fraud detection alert",
		"Daily limit exceeded",
		"Card blocked",
		"Network error",
	}

	return failureMessages[rand.Intn(len(failureMessages))]
}

// SetFailureRate sets the failure rate for testing
func (ps *PaymentService) SetFailureRate(rate float64) {
	if rate >= 0 && rate <= 1 {
		ps.failureRate = rate
	}
}

// SetTimeoutRate sets the timeout rate for testing
func (ps *PaymentService) SetTimeoutRate(rate float64) {
	if rate >= 0 && rate <= 1 {
		ps.timeoutRate = rate
	}
}

// SetProcessingTime sets the processing time for testing
func (ps *PaymentService) SetProcessingTime(duration time.Duration) {
	ps.processingTime = duration
}

// SimulatePaymentFailure simulates a payment failure for testing
func (ps *PaymentService) SimulatePaymentFailure(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
	originalFailureRate := ps.failureRate
	originalTimeoutRate := ps.timeoutRate

	ps.failureRate = 1.0 // 100% failure rate
	ps.timeoutRate = 0.0 // 0% timeout rate

	defer func() {
		ps.failureRate = originalFailureRate
		ps.timeoutRate = originalTimeoutRate
	}()

	return ps.ProcessPayment(ctx, req)
}

// SimulatePaymentTimeout simulates a payment timeout for testing
func (ps *PaymentService) SimulatePaymentTimeout(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
	originalTimeoutRate := ps.timeoutRate
	ps.timeoutRate = 1.0 // 100% timeout rate
	defer func() { ps.timeoutRate = originalTimeoutRate }()

	return ps.ProcessPayment(ctx, req)
}

// SimulatePaymentSuccess simulates a successful payment for testing
func (ps *PaymentService) SimulatePaymentSuccess(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
	originalFailureRate := ps.failureRate
	originalTimeoutRate := ps.timeoutRate

	ps.failureRate = 0.0 // 0% failure rate
	ps.timeoutRate = 0.0 // 0% timeout rate

	defer func() {
		ps.failureRate = originalFailureRate
		ps.timeoutRate = originalTimeoutRate
	}()

	return ps.ProcessPayment(ctx, req)
}
