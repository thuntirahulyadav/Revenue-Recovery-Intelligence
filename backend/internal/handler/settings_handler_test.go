package handler

import (
	"testing"

	"razorpay-recovery-intelligence/backend/internal/domain"
)

func TestValidateSettings(t *testing.T) {
	valid := domain.MerchantRecoverySettings{
		MaxRetryAttempts: 3, MinConfidenceThreshold: 0.65, MaxCommAttempts: 2,
		HumanApprovalThreshold: 50000, HighValueTransactionThreshold: 25000,
	}
	if err := validateSettings(valid); err != nil {
		t.Fatalf("expected valid settings, got %v", err)
	}
	valid.MinConfidenceThreshold = 1.1
	if err := validateSettings(valid); err == nil {
		t.Fatal("expected out-of-range confidence threshold to be rejected")
	}
}
