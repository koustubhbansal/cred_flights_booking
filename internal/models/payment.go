package models

import (
	"time"
)

// PaymentRequest represents a payment request
type PaymentRequest struct {
	BookingID   int     `json:"booking_id"`
	Amount      float64 `json:"amount"`
	UserID      int     `json:"user_id"`
	PaymentType string  `json:"payment_type"` // "credit_card", "debit_card", "upi", etc.
}

// PaymentResponse represents the response for payment processing
type PaymentResponse struct {
	PaymentID   string    `json:"payment_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
	BookingID   int       `json:"booking_id"`
	Amount      float64   `json:"amount"`
	ProcessedAt time.Time `json:"processed_at"`
}

// PaymentStatus constants
const (
	PaymentStatusSuccess = "success"
	PaymentStatusFailed  = "failed"
	PaymentStatusTimeout = "timeout"
	PaymentStatusPending = "pending"
)

// PaymentType constants
const (
	PaymentTypeCreditCard = "credit_card"
	PaymentTypeDebitCard  = "debit_card"
	PaymentTypeUPI        = "upi"
	PaymentTypeNetBanking = "net_banking"
)

// IsValidPaymentType checks if the payment type is valid
func IsValidPaymentType(paymentType string) bool {
	validTypes := []string{
		PaymentTypeCreditCard,
		PaymentTypeDebitCard,
		PaymentTypeUPI,
		PaymentTypeNetBanking,
	}

	for _, t := range validTypes {
		if paymentType == t {
			return true
		}
	}
	return false
}

// IsValidPaymentStatus checks if the payment status is valid
func IsValidPaymentStatus(status string) bool {
	validStatuses := []string{
		PaymentStatusSuccess,
		PaymentStatusFailed,
		PaymentStatusTimeout,
		PaymentStatusPending,
	}

	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}
