package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID                    uuid.UUID `json:"id"`
	MerchantID            uuid.UUID `json:"merchant_id"`
	Email                 string    `json:"email"`
	Phone                 string    `json:"phone"`
	HistoricalSuccessRate float64   `json:"historical_success_rate"`
	HistoricalFailureRate float64   `json:"historical_failure_rate"`
	CustomerValue         float64   `json:"customer_value"`
	CreatedAt             time.Time `json:"created_at"`
}
