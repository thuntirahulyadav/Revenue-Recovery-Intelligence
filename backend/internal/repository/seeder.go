package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"razorpay-recovery-intelligence/backend/internal/domain"

	"github.com/google/uuid"
)

var defaultFailureReasons = []domain.FailureReason{
	domain.FailureBankTimeout,
	domain.FailureNetworkError,
	domain.FailureInsufficientFunds,
	domain.FailureCardExpired,
	domain.FailurePaymentMethodFailure,
	domain.FailureCustomerAbandonment,
	domain.FailureTechnicalError,
}

var defaultPaymentMethods = []domain.PaymentMethod{
	domain.MethodCard,
	domain.MethodUPI,
	domain.MethodNetbanking,
	domain.MethodWallet,
	domain.MethodEMI,
}

func SeedDatabaseIfEmpty(ctx context.Context, db *DB) error {
	var paymentCount int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(id) FROM payments").Scan(&paymentCount)
	if err == nil && paymentCount >= 500 {
		log.Printf("[Seeder] Database already contains %d payments. Skipping auto-seeding.", paymentCount)
		return nil
	}

	log.Println("[Seeder] Seeding PostgreSQL database with realistic initial dataset...")

	// 1. Merchant
	merchantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	settingsJSON, _ := json.Marshal(domain.MerchantRecoverySettings{
		MaxRetryAttempts:              3,
		MinConfidenceThreshold:        0.65,
		MaxCommAttempts:               2,
		HumanApprovalThreshold:        50000.0,
		HighValueTransactionThreshold: 25000.0,
		AutoExecutionEnabled:          true,
	})

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO merchants (id, name, recovery_settings)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, merchantID, "Razorpay Buildathon Demo Merchant", settingsJSON)
	if err != nil {
		return fmt.Errorf("error creating seed merchant: %w", err)
	}

	// 2. Customers
	customerIDs := make([]uuid.UUID, 150)
	for i := 0; i < 150; i++ {
		cID := uuid.New()
		customerIDs[i] = cID
		succRate := 0.70 + rand.Float64()*0.28
		val := 5000.0 + rand.Float64()*85000.0
		email := fmt.Sprintf("buyer_%d@razorpay-demo.in", i+1)
		phone := fmt.Sprintf("+9198765%05d", i+1)

		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO customers (id, merchant_id, email, phone, historical_success_rate, historical_failure_rate, customer_value)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`, cID, merchantID, email, phone, succRate, 1.0-succRate, val)
	}

	// 3. Payments, Predictions, Decisions, Outcomes
	r := rand.New(rand.NewSource(42))
	numPayments := 1200 // Initial seed batch for instant responsive demo

	for i := 0; i < numPayments; i++ {
		pID := uuid.New()
		cID := customerIDs[r.Intn(len(customerIDs))]
		reason := defaultFailureReasons[r.Intn(len(defaultFailureReasons))]
		method := defaultPaymentMethods[r.Intn(len(defaultPaymentMethods))]
		attempts := 1 + r.Intn(3)

		// Realistic amount
		amt := int64(350 + r.Intn(12500))
		if r.Float64() < 0.08 {
			amt = int64(52000 + r.Intn(45000)) // High value
		}

		createdAt := time.Now().UTC().Add(-time.Duration(r.Intn(14*24)) * time.Hour)

		// Base probability & strategy
		prob := 0.72 - float64(attempts)*0.10
		strat := domain.StrategyRetryLater
		switch reason {
		case domain.FailureBankTimeout:
			strat = domain.StrategyRetryLater
			prob = 0.76 - float64(attempts)*0.08
		case domain.FailureNetworkError:
			strat = domain.StrategyRetryNow
			prob = 0.85 - float64(attempts)*0.10
		case domain.FailureInsufficientFunds:
			strat = domain.StrategyRetryLater
			prob = 0.48 - float64(attempts)*0.09
		case domain.FailureCardExpired:
			strat = domain.StrategySwitchPaymentMethod
			prob = 0.65
		case domain.FailureCustomerAbandonment:
			strat = domain.StrategySendPaymentLink
			prob = 0.62
		case domain.FailurePaymentMethodFailure:
			strat = domain.StrategySwitchPaymentMethod
			prob = 0.70
		case domain.FailureTechnicalError:
			strat = domain.StrategyRetryNow
			prob = 0.72
		}

		if amt >= 50000 {
			strat = domain.StrategyEscalateToHuman
			prob += 0.10
		}
		if attempts >= 3 && prob < 0.40 {
			strat = domain.StrategyStopRecovery
			prob = 0.05
		}

		if prob > 0.95 {
			prob = 0.95
		}
		if prob < 0.05 {
			prob = 0.05
		}

		status := domain.PaymentStatusFailed
		isRecovered := r.Float64() <= prob && strat != domain.StrategyStopRecovery
		if isRecovered {
			status = domain.PaymentStatusRecovered
		}

		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO payments (id, merchant_id, customer_id, amount, currency, payment_method, status, failure_reason, attempt_count, created_at)
			VALUES ($1, $2, $3, $4, 'INR', $5, $6, $7, $8, $9)
		`, pID, merchantID, cID, amt, method, status, reason, attempts, createdAt)

		// Recovery Prediction
		predID := uuid.New()
		shapJSON, _ := json.Marshal([]domain.SHAPFactor{
			{Feature: "failure_reason", Impact: 0.24, Direction: "positive", Description: "Failure characteristic match"},
			{Feature: "customer_success_rate", Impact: 0.16, Direction: "positive", Description: "Customer lifetime success record"},
			{Feature: "attempt_count", Impact: -0.09, Direction: "negative", Description: "Previous attempt penalty"},
		})
		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO recovery_predictions (id, payment_id, recovery_probability, model_version, confidence, shap_factors, created_at)
			VALUES ($1, $2, $3, 'v1.0.0-xgb', $4, $5, $6)
		`, predID, pID, prob, 0.85, shapJSON, createdAt)

		// Recovery Decision
		decID := uuid.New()
		expRev := float64(amt) * prob
		expCost := 8.0
		if strat == domain.StrategySendPaymentLink {
			expCost = 15.0
		} else if strat == domain.StrategyEscalateToHuman {
			expCost = 85.0
		} else if strat == domain.StrategyStopRecovery {
			expCost = 0.0
			expRev = 0.0
		}
		netVal := expRev - expCost
		prio := (float64(amt) / 500.0) * prob
		if prio > 99.9 {
			prio = 99.9
		}

		policyChecksJSON, _ := json.Marshal([]domain.PolicyCheck{
			{Name: "Max Retry Limit", Passed: attempts <= 3, Description: "Attempt count within limit"},
			{Name: "Confidence Threshold", Passed: true, Description: "ML confidence >= 65%"},
			{Name: "Economic Net Viability", Passed: netVal >= 0, Description: "Positive net recovery value"},
		})

		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO recovery_decisions (id, payment_id, strategy, expected_revenue, expected_cost, expected_net_value, priority_score, explanation, policy_status, policy_checks, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'APPROVED', $9, $10)
		`, decID, pID, strat, expRev, expCost, netVal, prio, "AI optimized recovery path", policyChecksJSON, createdAt)

		// Outcome if recovered
		if isRecovered {
			actID := uuid.New()
			_, _ = db.Pool.Exec(ctx, `
				INSERT INTO recovery_actions (id, payment_id, action_type, status, execution_mode, executed_at)
				VALUES ($1, $2, $3, 'EXECUTED', 'SIMULATED', $4)
			`, actID, pID, strat, createdAt.Add(15*time.Minute))

			outID := uuid.New()
			_, _ = db.Pool.Exec(ctx, `
				INSERT INTO recovery_outcomes (id, action_id, payment_id, success, recovered_amount, recovery_cost, net_recovery_value, completed_at)
				VALUES ($1, $2, $3, true, $4, $5, $6, $7)
			`, outID, actID, pID, float64(amt), expCost, float64(amt)-expCost, createdAt.Add(16*time.Minute))
		}
	}

	log.Printf("[Seeder] Successfully seeded database with %d payments, predictions, decisions, and outcomes.", numPayments)
	return nil
}
