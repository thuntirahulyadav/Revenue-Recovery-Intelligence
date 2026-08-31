package policy

import (
	"fmt"

	"razorpay-recovery-intelligence/backend/internal/domain"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) ValidatePolicy(
	payment *domain.Payment,
	prediction *domain.RecoveryPrediction,
	decision *domain.RecoveryDecision,
	settings domain.MerchantRecoverySettings,
) (domain.PolicyStatus, []domain.PolicyCheck) {
	checks := make([]domain.PolicyCheck, 0)
	status := domain.PolicyStatusApproved

	// 1. Max Retry Attempts Check
	maxRetries := settings.MaxRetryAttempts
	if maxRetries <= 0 {
		maxRetries = 3
	}
	retryPassed := payment.AttemptCount <= maxRetries
	retryDesc := fmt.Sprintf("Attempt count (%d) within allowed limit (max %d)", payment.AttemptCount, maxRetries)
	if !retryPassed {
		retryDesc = fmt.Sprintf("Attempt count (%d) exceeded allowed limit (max %d)", payment.AttemptCount, maxRetries)
		status = domain.PolicyStatusRejected
	}
	checks = append(checks, domain.PolicyCheck{
		Name:        "Max Retry Limit",
		Passed:      retryPassed,
		Description: retryDesc,
	})

	// 2. Minimum ML Confidence Gate
	minConf := settings.MinConfidenceThreshold
	if minConf <= 0.0 {
		minConf = 0.65
	}
	confPassed := prediction.Confidence >= minConf || decision.Strategy == domain.StrategyStopRecovery
	confDesc := fmt.Sprintf("ML Confidence (%.0f%%) meets threshold (%.0f%%)", prediction.Confidence*100, minConf*100)
	if !confPassed {
		confDesc = fmt.Sprintf("ML Confidence (%.0f%%) below policy threshold (%.0f%%)", prediction.Confidence*100, minConf*100)
		if status == domain.PolicyStatusApproved {
			status = domain.PolicyStatusPendingHumanApproval
		}
	}
	checks = append(checks, domain.PolicyCheck{
		Name:        "Confidence Threshold",
		Passed:      confPassed,
		Description: confDesc,
	})

	// 3. High Value Transaction Human Gate
	humanThresh := settings.HumanApprovalThreshold
	if humanThresh <= 0 {
		humanThresh = 50000.0
	}
	highValPassed := float64(payment.Amount) < humanThresh
	highValDesc := fmt.Sprintf("Transaction value (₹%d) below human review threshold (₹%.0f)", payment.Amount, humanThresh)
	if !highValPassed && decision.Strategy != domain.StrategyStopRecovery {
		highValDesc = fmt.Sprintf("High-value transaction (₹%d) triggers mandatory human approval gate (₹%.0f)", payment.Amount, humanThresh)
		status = domain.PolicyStatusPendingHumanApproval
	}
	checks = append(checks, domain.PolicyCheck{
		Name:        "High Value Authorization",
		Passed:      highValPassed,
		Description: highValDesc,
	})

	// 4. Positive Net Value Check
	netValPassed := decision.ExpectedNetValue > 0 || decision.Strategy == domain.StrategyStopRecovery
	netValDesc := fmt.Sprintf("Expected net recovery value (₹%.2f) is economically viable", decision.ExpectedNetValue)
	if !netValPassed {
		netValDesc = fmt.Sprintf("Negative expected net recovery (₹%.2f) violates economic viability rule", decision.ExpectedNetValue)
		status = domain.PolicyStatusRejected
	}
	checks = append(checks, domain.PolicyCheck{
		Name:        "Economic Net Viability",
		Passed:      netValPassed,
		Description: netValDesc,
	})

	return status, checks
}
