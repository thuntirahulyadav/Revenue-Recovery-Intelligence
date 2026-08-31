package domain

import (
	"time"

	"github.com/google/uuid"
)

type RecoveryStrategy string

const (
	StrategyRetryNow            RecoveryStrategy = "RETRY_NOW"
	StrategyRetryLater          RecoveryStrategy = "RETRY_LATER"
	StrategySwitchPaymentMethod RecoveryStrategy = "SWITCH_PAYMENT_METHOD"
	StrategySendPaymentLink     RecoveryStrategy = "SEND_PAYMENT_LINK"
	StrategyEscalateToHuman     RecoveryStrategy = "ESCALATE_TO_HUMAN"
	StrategyStopRecovery        RecoveryStrategy = "STOP_RECOVERY"
)

type PolicyStatus string

const (
	PolicyStatusApproved             PolicyStatus = "APPROVED"
	PolicyStatusRejected             PolicyStatus = "REJECTED"
	PolicyStatusPendingHumanApproval PolicyStatus = "PENDING_HUMAN_APPROVAL"
)

type ExecutionMode string

const (
	ExecutionModeSimulated ExecutionMode = "SIMULATED"
	ExecutionModeMock      ExecutionMode = "MOCK"
	ExecutionModeReal      ExecutionMode = "REAL"
)

type ActionStatus string

const (
	ActionStatusPending   ActionStatus = "PENDING"
	ActionStatusExecuted  ActionStatus = "EXECUTED"
	ActionStatusSimulated ActionStatus = "SIMULATED"
	ActionStatusFailed    ActionStatus = "FAILED"
)

type SHAPFactor struct {
	Feature     string  `json:"feature"`
	Impact      float64 `json:"impact"`
	Direction   string  `json:"direction"` // positive | negative
	Description string  `json:"description"`
}

type PolicyCheck struct {
	Name        string `json:"name"`
	Passed      bool   `json:"passed"`
	Description string `json:"description"`
}

type RecoveryPrediction struct {
	ID                  uuid.UUID    `json:"id"`
	PaymentID           uuid.UUID    `json:"payment_id"`
	RecoveryProbability float64      `json:"recovery_probability"`
	ModelVersion        string       `json:"model_version"`
	Confidence          float64      `json:"confidence"`
	SHAPFactors         []SHAPFactor `json:"shap_factors"`
	CreatedAt           time.Time    `json:"created_at"`
}

type RecoveryDecision struct {
	ID               uuid.UUID        `json:"id"`
	PaymentID        uuid.UUID        `json:"payment_id"`
	Strategy         RecoveryStrategy `json:"strategy"`
	ExpectedRevenue  float64          `json:"expected_revenue"`
	ExpectedCost     float64          `json:"expected_cost"`
	ExpectedNetValue float64          `json:"expected_net_value"`
	PriorityScore    float64          `json:"priority_score"`
	Explanation      string           `json:"explanation"`
	PolicyStatus     PolicyStatus     `json:"policy_status"`
	PolicyChecks     []PolicyCheck    `json:"policy_checks"`
	CreatedAt        time.Time        `json:"created_at"`
}

type RecoveryAction struct {
	ID            uuid.UUID              `json:"id"`
	PaymentID     uuid.UUID              `json:"payment_id"`
	ActionType    string                 `json:"action_type"`
	Status        ActionStatus           `json:"status"`
	ExecutionMode ExecutionMode          `json:"execution_mode"`
	Payload       map[string]interface{} `json:"payload"`
	ExecutedAt    time.Time              `json:"executed_at"`
}

type RecoveryOutcome struct {
	ID               uuid.UUID `json:"id"`
	ActionID         uuid.UUID `json:"action_id"`
	PaymentID        uuid.UUID `json:"payment_id"`
	Success          bool      `json:"success"`
	RecoveredAmount  float64   `json:"recovered_amount"`
	RecoveryCost     float64   `json:"recovery_cost"`
	NetRecoveryValue float64   `json:"net_recovery_value"`
	CompletedAt      time.Time `json:"completed_at"`
}

type FullPaymentRecoveryAnalysis struct {
	Payment              Payment              `json:"payment"`
	Customer             Customer             `json:"customer"`
	Prediction           *RecoveryPrediction  `json:"prediction"`
	Decision             *RecoveryDecision    `json:"decision"`
	Action               *RecoveryAction      `json:"action"`
	Outcome              *RecoveryOutcome     `json:"outcome"`
	AlternativeStrategies []StrategyComparison `json:"alternative_strategies"`
}

type StrategyComparison struct {
	Strategy         RecoveryStrategy `json:"strategy"`
	Probability      float64          `json:"probability"`
	ExpectedCost     float64          `json:"expected_cost"`
	ExpectedRevenue  float64          `json:"expected_revenue"`
	ExpectedNetValue float64          `json:"expected_net_value"`
	IsSelected       bool             `json:"is_selected"`
}

type AuditLog struct {
	ID        uuid.UUID              `json:"id"`
	PaymentID *uuid.UUID             `json:"payment_id,omitempty"`
	EventType string                 `json:"event_type"`
	Actor     string                 `json:"actor"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}
