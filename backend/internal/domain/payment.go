package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusRecovered  PaymentStatus = "RECOVERED"
	PaymentStatusRecovering PaymentStatus = "RECOVERING"
	PaymentStatusAbandoned  PaymentStatus = "ABANDONED"
)

type FailureReason string

const (
	FailureBankTimeout          FailureReason = "BANK_TIMEOUT"
	FailureNetworkError         FailureReason = "NETWORK_ERROR"
	FailureInsufficientFunds    FailureReason = "INSUFFICIENT_FUNDS"
	FailureCardExpired          FailureReason = "CARD_EXPIRED"
	FailurePaymentMethodFailure FailureReason = "PAYMENT_METHOD_FAILURE"
	FailureCustomerAbandonment  FailureReason = "CUSTOMER_ABANDONMENT"
	FailureTechnicalError       FailureReason = "TECHNICAL_ERROR"
)

type PaymentMethod string

const (
	MethodCard       PaymentMethod = "card"
	MethodUPI        PaymentMethod = "upi"
	MethodNetbanking PaymentMethod = "netbanking"
	MethodWallet     PaymentMethod = "wallet"
	MethodEMI        PaymentMethod = "emi"
)

type Payment struct {
	ID            uuid.UUID     `json:"id"`
	MerchantID    uuid.UUID     `json:"merchant_id"`
	CustomerID    uuid.UUID     `json:"customer_id"`
	Amount        int64         `json:"amount"` // Transaction amount in INR units
	Currency      string        `json:"currency"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Status        PaymentStatus `json:"status"`
	FailureReason FailureReason `json:"failure_reason"`
	AttemptCount  int           `json:"attempt_count"`
	CreatedAt     time.Time     `json:"created_at"`
}

type PaymentFailedEventPayload struct {
	PaymentID     uuid.UUID     `json:"payment_id"`
	MerchantID    uuid.UUID     `json:"merchant_id"`
	CustomerID    uuid.UUID     `json:"customer_id"`
	Amount        int64         `json:"amount"`
	Currency      string        `json:"currency"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	FailureReason FailureReason `json:"failure_reason"`
	AttemptCount  int           `json:"attempt_count"`
	Timestamp     time.Time     `json:"timestamp"`
}
