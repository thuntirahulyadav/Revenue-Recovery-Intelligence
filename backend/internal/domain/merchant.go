package domain

import (
	"time"

	"github.com/google/uuid"
)

type MerchantRecoverySettings struct {
	MaxRetryAttempts              int     `json:"max_retry_attempts"`
	MinConfidenceThreshold        float64 `json:"min_confidence_threshold"`
	MaxCommAttempts               int     `json:"max_comm_attempts"`
	HumanApprovalThreshold        float64 `json:"human_approval_threshold"`
	HighValueTransactionThreshold float64 `json:"high_value_transaction_threshold"`
	AutoExecutionEnabled          bool    `json:"auto_execution_enabled"`
}

type Merchant struct {
	ID               uuid.UUID                `json:"id"`
	Name             string                   `json:"name"`
	RecoverySettings MerchantRecoverySettings `json:"recovery_settings"`
	CreatedAt        time.Time                `json:"created_at"`
}
