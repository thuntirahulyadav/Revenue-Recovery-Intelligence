package service

import (
	"testing"

	"razorpay-recovery-intelligence/backend/internal/domain"
)

func TestHighValueThresholdControlsHumanEscalation(t *testing.T) {
	svc := &RecoveryService{}
	payment := &domain.Payment{Amount: 30000}

	probability := svc.estimateStrategyProbability(
		domain.StrategyEscalateToHuman,
		payment,
		0.50,
		domain.MerchantRecoverySettings{HighValueTransactionThreshold: 40000},
	)
	if probability != 0.525 {
		t.Fatalf("expected normal human strategy probability below the merchant threshold, got %v", probability)
	}

	probability = svc.estimateStrategyProbability(
		domain.StrategyEscalateToHuman,
		payment,
		0.50,
		domain.MerchantRecoverySettings{HighValueTransactionThreshold: 25000},
	)
	if probability != 0.65 {
		t.Fatalf("expected high-value uplift at the merchant threshold, got %v", probability)
	}
}

func TestOnlyApprovedDecisionsAreExecutable(t *testing.T) {
	for _, status := range []domain.PolicyStatus{domain.PolicyStatusRejected, domain.PolicyStatusPendingHumanApproval} {
		if err := validateExecutableDecision(&domain.RecoveryDecision{PolicyStatus: status}); err == nil {
			t.Fatalf("expected %s decision to be blocked", status)
		}
	}
	if err := validateExecutableDecision(&domain.RecoveryDecision{PolicyStatus: domain.PolicyStatusApproved}); err != nil {
		t.Fatalf("expected approved decision to be executable: %v", err)
	}
}
