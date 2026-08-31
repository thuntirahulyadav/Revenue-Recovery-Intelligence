package tests

import (
	"testing"

	"razorpay-recovery-intelligence/backend/internal/domain"
	"razorpay-recovery-intelligence/backend/internal/service"
)

func TestStrategyCost_Calculations(t *testing.T) {
	// Expected Net Value = (Amount * P_strategy) - Cost
	amount := 10000.0
	prob := 0.80
	cost := service.StrategyCosts[domain.StrategyRetryLater] // 8.00

	expectedRev := amount * prob
	expectedNet := expectedRev - cost

	if expectedRev != 8000.0 {
		t.Errorf("Expected gross revenue 8000.0, got %f", expectedRev)
	}

	if expectedNet != 7992.0 {
		t.Errorf("Expected net recovery value 7992.0, got %f", expectedNet)
	}
}

func TestStrategyCost_StopRecoveryZeroCost(t *testing.T) {
	cost := service.StrategyCosts[domain.StrategyStopRecovery]
	if cost != 0.0 {
		t.Errorf("Expected STOP_RECOVERY cost to be 0.0, got %f", cost)
	}
}
