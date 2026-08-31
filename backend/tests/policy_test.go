package tests

import (
	"testing"

	"razorpay-recovery-intelligence/backend/internal/domain"
	"razorpay-recovery-intelligence/backend/internal/policy"

	"github.com/google/uuid"
)

func TestPolicyEngine_MaxRetries(t *testing.T) {
	engine := policy.NewEngine()

	payment := &domain.Payment{
		ID:           uuid.New(),
		Amount:       2500,
		AttemptCount: 4,
	}

	prediction := &domain.RecoveryPrediction{
		Confidence:          0.85,
		RecoveryProbability: 0.70,
	}

	decision := &domain.RecoveryDecision{
		Strategy:         domain.StrategyRetryLater,
		ExpectedNetValue: 1500,
	}

	settings := domain.MerchantRecoverySettings{
		MaxRetryAttempts:       3,
		MinConfidenceThreshold: 0.65,
	}

	status, checks := engine.ValidatePolicy(payment, prediction, decision, settings)

	if status != domain.PolicyStatusRejected {
		t.Errorf("Expected PolicyStatusRejected for attempt_count 4 (max 3), got %s", status)
	}

	foundRetryCheck := false
	for _, c := range checks {
		if c.Name == "Max Retry Limit" {
			foundRetryCheck = true
			if c.Passed {
				t.Errorf("Expected Max Retry Limit check to fail, but it passed")
			}
		}
	}
	if !foundRetryCheck {
		t.Errorf("Expected to find 'Max Retry Limit' check in policy checks")
	}
}

func TestPolicyEngine_HighValueGate(t *testing.T) {
	engine := policy.NewEngine()

	payment := &domain.Payment{
		ID:           uuid.New(),
		Amount:       75000,
		AttemptCount: 1,
	}

	prediction := &domain.RecoveryPrediction{
		Confidence:          0.90,
		RecoveryProbability: 0.85,
	}

	decision := &domain.RecoveryDecision{
		Strategy:         domain.StrategyEscalateToHuman,
		ExpectedNetValue: 62000,
	}

	settings := domain.MerchantRecoverySettings{
		MaxRetryAttempts:        3,
		MinConfidenceThreshold:  0.65,
		HumanApprovalThreshold: 50000,
	}

	status, _ := engine.ValidatePolicy(payment, prediction, decision, settings)

	if status != domain.PolicyStatusPendingHumanApproval {
		t.Errorf("Expected PolicyStatusPendingHumanApproval for transaction above threshold, got %s", status)
	}
}
