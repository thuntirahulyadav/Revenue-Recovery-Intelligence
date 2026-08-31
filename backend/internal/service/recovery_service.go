package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"razorpay-recovery-intelligence/backend/internal/audit"
	"razorpay-recovery-intelligence/backend/internal/domain"
	"razorpay-recovery-intelligence/backend/internal/kafka"
	"razorpay-recovery-intelligence/backend/internal/policy"
	"razorpay-recovery-intelligence/backend/internal/redis"
	"razorpay-recovery-intelligence/backend/internal/repository"

	"github.com/google/uuid"
)

var StrategyCosts = map[domain.RecoveryStrategy]float64{
	domain.StrategyRetryNow:            5.00,  // Gateway API & retry cost
	domain.StrategyRetryLater:          8.00,  // Scheduling, queue & gateway fee
	domain.StrategySwitchPaymentMethod: 12.00, // Smart router & routing fee
	domain.StrategySendPaymentLink:     15.00, // SMS/WhatsApp & hosted checkout fee
	domain.StrategyEscalateToHuman:     85.00, // Merchant operations/agent cost
	domain.StrategyStopRecovery:        0.00,  // 0 cost
}

type RecoveryService struct {
	paymentRepo  *repository.PaymentRepo
	customerRepo *repository.CustomerRepo
	recoveryRepo *repository.RecoveryRepo
	merchantRepo *repository.MerchantRepo
	mlClient     *MLClient
	policyEngine *policy.Engine
	auditLogger  *audit.Logger
	eventBus     *kafka.EventBus
	redisClient  *redis.Client
	pipelineMu   sync.Mutex
}

func NewRecoveryService(
	paymentRepo *repository.PaymentRepo,
	customerRepo *repository.CustomerRepo,
	recoveryRepo *repository.RecoveryRepo,
	merchantRepo *repository.MerchantRepo,
	mlClient *MLClient,
	policyEngine *policy.Engine,
	auditLogger *audit.Logger,
	eventBus *kafka.EventBus,
	redisClient *redis.Client,
) *RecoveryService {
	svc := &RecoveryService{
		paymentRepo:  paymentRepo,
		customerRepo: customerRepo,
		recoveryRepo: recoveryRepo,
		merchantRepo: merchantRepo,
		mlClient:     mlClient,
		policyEngine: policyEngine,
		auditLogger:  auditLogger,
		eventBus:     eventBus,
		redisClient:  redisClient,
	}

	// Register Event Consumers for the 7-Step Kafka Lifecycle
	svc.registerEventHandlers()

	return svc
}

func (s *RecoveryService) registerEventHandlers() {
	// Step 1 -> 2: payment.failed -> enrich and trigger prediction
	s.eventBus.RegisterHandler(domain.EventPaymentFailed, func(ctx context.Context, env domain.EventEnvelope) error {
		paymentID := env.CorrelationID
		s.auditLogger.Log(ctx, &paymentID, string(domain.EventPaymentFailed), "rri.event_bus", map[string]interface{}{
			"event_id": env.EventID,
		})
		_, err := s.OrchestrateRecoveryPipeline(ctx, paymentID)
		return err
	})
}

// OrchestrateRecoveryPipeline implements the full closed loop:
// DETECT -> ENRICH -> PREDICT -> PRIORITIZE -> SELECT STRATEGY -> VALIDATE POLICY
func (s *RecoveryService) OrchestrateRecoveryPipeline(ctx context.Context, paymentID uuid.UUID) (*domain.FullPaymentRecoveryAnalysis, error) {
	// Payment events may be delivered more than once. Serialize processing in this
	// process and reuse the existing decision instead of creating duplicate records.
	s.pipelineMu.Lock()
	defer s.pipelineMu.Unlock()

	// 1. Fetch payment
	payment, err := s.paymentRepo.GetPaymentByID(ctx, paymentID)
	if err != nil || payment == nil {
		return nil, fmt.Errorf("payment %s not found: %w", paymentID, err)
	}
	if existing, err := s.existingAnalysis(ctx, payment); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	// 2. Enrich with customer context
	customer, err := s.customerRepo.GetOrCreateCustomer(ctx, payment.MerchantID, payment.CustomerID)
	if err != nil {
		customer = &domain.Customer{
			ID:                    payment.CustomerID,
			MerchantID:            payment.MerchantID,
			HistoricalSuccessRate: 0.85,
			HistoricalFailureRate: 0.15,
			CustomerValue:         15000.00,
		}
	}

	// Emit payment.enriched event
	_ = s.eventBus.Publish(ctx, domain.NewEventEnvelope(
		domain.EventPaymentEnriched,
		paymentID,
		map[string]interface{}{
			"payment_id":  payment.ID,
			"customer_id": customer.ID,
			"amount":      payment.Amount,
		},
	))

	// 3. ML Prediction
	now := time.Now()
	mlReq := MLPredictionRequest{
		TransactionAmount:   float64(payment.Amount),
		PaymentMethod:       string(payment.PaymentMethod),
		FailureReason:       string(payment.FailureReason),
		AttemptCount:        payment.AttemptCount,
		CustomerSuccessRate: customer.HistoricalSuccessRate,
		CustomerFailureRate: customer.HistoricalFailureRate,
		CustomerValue:       customer.CustomerValue,
		HourOfDay:           now.Hour(),
		DayOfWeek:           int(now.Weekday()),
	}

	predResp, err := s.mlClient.PredictRecovery(ctx, mlReq)
	if err != nil {
		return nil, fmt.Errorf("ML prediction failed: %w", err)
	}

	pred := &domain.RecoveryPrediction{
		ID:                  uuid.New(),
		PaymentID:           payment.ID,
		RecoveryProbability: predResp.RecoveryProbability,
		ModelVersion:        predResp.ModelVersion,
		Confidence:          predResp.Confidence,
		SHAPFactors:         predResp.SHAPFactors,
		CreatedAt:           time.Now().UTC(),
	}
	if err := s.recoveryRepo.SavePrediction(ctx, pred); err != nil {
		return nil, fmt.Errorf("save recovery prediction: %w", err)
	}

	_ = s.eventBus.Publish(ctx, domain.NewEventEnvelope(
		domain.EventRecoveryPredicted,
		paymentID,
		pred,
	))

	settings, err := s.merchantRepo.GetRecoverySettings(ctx, payment.MerchantID)
	if err != nil {
		return nil, err
	}

	// 4. Economic Strategy Engine: evaluate all candidate strategies
	candidates := s.evaluateStrategyCandidates(payment, pred, customer, settings)
	bestCandidate := candidates[0]
	for _, c := range candidates {
		if c.ExpectedNetValue > bestCandidate.ExpectedNetValue {
			bestCandidate = c
		}
	}

	// Calculate priority score (0.0 to 100.0)
	// Priority = (Amount / 1000) * Recovery Probability * (1 - (AttemptCount * 0.12))
	basePrio := (float64(payment.Amount) / 500.0) * pred.RecoveryProbability * (1.0 - (float64(payment.AttemptCount) * 0.10))
	if basePrio > 99.9 {
		basePrio = 99.9
	}
	if basePrio < 1.0 && bestCandidate.Strategy != domain.StrategyStopRecovery {
		basePrio = 5.0
	}

	explanation := s.buildExplanation(payment, pred, bestCandidate)

	dec := &domain.RecoveryDecision{
		ID:               uuid.New(),
		PaymentID:        payment.ID,
		Strategy:         bestCandidate.Strategy,
		ExpectedRevenue:  bestCandidate.ExpectedRevenue,
		ExpectedCost:     bestCandidate.ExpectedCost,
		ExpectedNetValue: bestCandidate.ExpectedNetValue,
		PriorityScore:    round(basePrio, 2),
		Explanation:      explanation,
		PolicyStatus:     domain.PolicyStatusApproved,
		CreatedAt:        time.Now().UTC(),
	}

	// 5. Policy Engine Validation
	policyStatus, policyChecks := s.policyEngine.ValidatePolicy(payment, pred, dec, settings)
	dec.PolicyStatus = policyStatus
	dec.PolicyChecks = policyChecks

	if err := s.recoveryRepo.SaveDecision(ctx, dec); err != nil {
		return nil, fmt.Errorf("save recovery decision: %w", err)
	}

	_ = s.eventBus.Publish(ctx, domain.NewEventEnvelope(
		domain.EventRecoveryDecisionCreated,
		paymentID,
		dec,
	))

	s.auditLogger.Log(ctx, &payment.ID, string(domain.EventRecoveryDecisionCreated), "rri.strategy_engine", map[string]interface{}{
		"strategy":           dec.Strategy,
		"expected_net_value": dec.ExpectedNetValue,
		"policy_status":      dec.PolicyStatus,
	})

	analysis := &domain.FullPaymentRecoveryAnalysis{
		Payment:               *payment,
		Customer:              *customer,
		Prediction:            pred,
		Decision:              dec,
		AlternativeStrategies: candidates,
	}
	if settings.AutoExecutionEnabled && dec.PolicyStatus == domain.PolicyStatusApproved && dec.Strategy != domain.StrategyStopRecovery {
		outcome, err := s.ExecuteRecovery(ctx, paymentID, domain.ExecutionModeSimulated, "rri.auto_execution")
		if err != nil {
			return nil, fmt.Errorf("auto execute approved recovery: %w", err)
		}
		analysis.Outcome = outcome
		analysis.Action, _ = s.recoveryRepo.GetActionByPaymentID(ctx, paymentID)
	}
	return analysis, nil
}

func (s *RecoveryService) existingAnalysis(ctx context.Context, payment *domain.Payment) (*domain.FullPaymentRecoveryAnalysis, error) {
	pred, err := s.recoveryRepo.GetPredictionByPaymentID(ctx, payment.ID)
	if err != nil {
		return nil, err
	}
	dec, err := s.recoveryRepo.GetDecisionByPaymentID(ctx, payment.ID)
	if err != nil {
		return nil, err
	}
	if pred == nil || dec == nil {
		return nil, nil
	}
	customer, err := s.customerRepo.GetOrCreateCustomer(ctx, payment.MerchantID, payment.CustomerID)
	if err != nil {
		return nil, err
	}
	action, _ := s.recoveryRepo.GetActionByPaymentID(ctx, payment.ID)
	outcome, _ := s.recoveryRepo.GetOutcomeByPaymentID(ctx, payment.ID)
	return &domain.FullPaymentRecoveryAnalysis{Payment: *payment, Customer: *customer, Prediction: pred, Decision: dec, Action: action, Outcome: outcome}, nil
}

func (s *RecoveryService) evaluateStrategyCandidates(
	payment *domain.Payment,
	pred *domain.RecoveryPrediction,
	customer *domain.Customer,
	settings domain.MerchantRecoverySettings,
) []domain.StrategyComparison {
	strategies := []domain.RecoveryStrategy{
		domain.StrategyRetryNow,
		domain.StrategyRetryLater,
		domain.StrategySwitchPaymentMethod,
		domain.StrategySendPaymentLink,
		domain.StrategyEscalateToHuman,
		domain.StrategyStopRecovery,
	}

	results := make([]domain.StrategyComparison, 0, len(strategies))
	amount := float64(payment.Amount)

	for _, strat := range strategies {
		pStrat := s.estimateStrategyProbability(strat, payment, pred.RecoveryProbability, settings)
		cost := StrategyCosts[strat]
		expRev := amount * pStrat
		netVal := expRev - cost

		results = append(results, domain.StrategyComparison{
			Strategy:         strat,
			Probability:      round(pStrat, 4),
			ExpectedCost:     cost,
			ExpectedRevenue:  round(expRev, 2),
			ExpectedNetValue: round(netVal, 2),
			IsSelected:       false,
		})
	}

	return results
}

func (s *RecoveryService) estimateStrategyProbability(
	strat domain.RecoveryStrategy,
	payment *domain.Payment,
	baseMLProb float64,
	settings domain.MerchantRecoverySettings,
) float64 {
	switch strat {
	case domain.StrategyStopRecovery:
		return 0.0

	case domain.StrategyRetryNow:
		if payment.FailureReason == domain.FailureNetworkError || payment.FailureReason == domain.FailureTechnicalError {
			return min(0.92, baseMLProb*1.15)
		}
		if payment.FailureReason == domain.FailureInsufficientFunds || payment.FailureReason == domain.FailureCardExpired {
			return 0.04
		}
		return baseMLProb * 0.70

	case domain.StrategyRetryLater:
		if payment.FailureReason == domain.FailureBankTimeout || payment.FailureReason == domain.FailureInsufficientFunds {
			return min(0.88, baseMLProb*1.20)
		}
		if payment.FailureReason == domain.FailureCardExpired {
			return 0.05
		}
		return baseMLProb * 0.95

	case domain.StrategySwitchPaymentMethod:
		if payment.FailureReason == domain.FailureCardExpired || payment.FailureReason == domain.FailurePaymentMethodFailure {
			return min(0.85, max(0.60, baseMLProb*1.25))
		}
		return baseMLProb * 0.80

	case domain.StrategySendPaymentLink:
		if payment.FailureReason == domain.FailureCustomerAbandonment || payment.FailureReason == domain.FailureCardExpired {
			return min(0.80, max(0.55, baseMLProb*1.20))
		}
		return baseMLProb * 0.75

	case domain.StrategyEscalateToHuman:
		threshold := settings.HighValueTransactionThreshold
		if threshold <= 0 {
			threshold = 25000
		}
		if float64(payment.Amount) >= threshold {
			return min(0.95, baseMLProb*1.30)
		}
		return baseMLProb * 1.05
	}

	return baseMLProb
}

func (s *RecoveryService) buildExplanation(
	payment *domain.Payment,
	pred *domain.RecoveryPrediction,
	best domain.StrategyComparison,
) string {
	if best.Strategy == domain.StrategyStopRecovery {
		return fmt.Sprintf("Recovery aborted: high attempt count (%d) and low probability (%.0f%%) would produce negative net ROI.", payment.AttemptCount, pred.RecoveryProbability*100)
	}

	switch best.Strategy {
	case domain.StrategyRetryNow:
		return fmt.Sprintf("Immediate retry recommended for transient %s. High success probability (%.0f%%) with minimal overhead.", payment.FailureReason, best.Probability*100)
	case domain.StrategyRetryLater:
		return fmt.Sprintf("Delayed retry recommended for %s to allow issuer/balance resolution. Expected net recovery ₹%.2f.", payment.FailureReason, best.ExpectedNetValue)
	case domain.StrategySwitchPaymentMethod:
		return fmt.Sprintf("Alternate payment routing recommended due to %s. Expected recovery %.0f%%.", payment.FailureReason, best.Probability*100)
	case domain.StrategySendPaymentLink:
		return fmt.Sprintf("Instant Payment Link dispatch recommended to re-engage customer after %s.", payment.FailureReason)
	case domain.StrategyEscalateToHuman:
		return fmt.Sprintf("High transaction value (₹%d) warrants priority agent recovery outreach.", payment.Amount)
	}

	return fmt.Sprintf("AI selected %s with expected net value ₹%.2f (ROI: %.1fx cost).", best.Strategy, best.ExpectedNetValue, best.ExpectedNetValue/max(1.0, best.ExpectedCost))
}

// ApproveRecovery manually clears a pending human approval gate
func (s *RecoveryService) ApproveRecovery(ctx context.Context, paymentID uuid.UUID, actor string) (*domain.RecoveryDecision, error) {
	dec, err := s.recoveryRepo.GetDecisionByPaymentID(ctx, paymentID)
	if err != nil || dec == nil {
		return nil, fmt.Errorf("recovery decision not found for payment %s", paymentID)
	}
	if dec.PolicyStatus != domain.PolicyStatusPendingHumanApproval {
		return nil, fmt.Errorf("only decisions pending human approval can be approved (current status: %s)", dec.PolicyStatus)
	}

	dec.PolicyStatus = domain.PolicyStatusApproved
	if err := s.recoveryRepo.SaveDecision(ctx, dec); err != nil {
		return nil, err
	}

	_ = s.eventBus.Publish(ctx, domain.NewEventEnvelope(
		domain.EventRecoveryPolicyApproved,
		paymentID,
		dec,
	))

	s.auditLogger.Log(ctx, &paymentID, string(domain.EventRecoveryPolicyApproved), actor, map[string]interface{}{
		"approved_by": actor,
		"strategy":    dec.Strategy,
	})

	return dec, nil
}

// ExecuteRecovery executes or simulates the recovery strategy action
func (s *RecoveryService) ExecuteRecovery(ctx context.Context, paymentID uuid.UUID, mode domain.ExecutionMode, actor string) (*domain.RecoveryOutcome, error) {
	payment, err := s.paymentRepo.GetPaymentByID(ctx, paymentID)
	if err != nil || payment == nil {
		return nil, fmt.Errorf("payment %s not found", paymentID)
	}

	dec, err := s.recoveryRepo.GetDecisionByPaymentID(ctx, paymentID)
	if err != nil || dec == nil {
		return nil, fmt.Errorf("no decision found for payment %s", paymentID)
	}
	if err := validateExecutableDecision(dec); err != nil {
		return nil, err
	}
	if existing, err := s.recoveryRepo.GetOutcomeByPaymentID(ctx, paymentID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	if dec.Strategy == domain.StrategyStopRecovery {
		_ = s.paymentRepo.UpdatePaymentStatus(ctx, paymentID, domain.PaymentStatusAbandoned, payment.AttemptCount)
		return nil, fmt.Errorf("recovery stopped by strategy policy")
	}

	// 1. Create Recovery Action
	actionID := uuid.New()
	action := &domain.RecoveryAction{
		ID:            actionID,
		PaymentID:     paymentID,
		ActionType:    string(dec.Strategy),
		Status:        domain.ActionStatusExecuted,
		ExecutionMode: mode,
		Payload: map[string]interface{}{
			"strategy":           dec.Strategy,
			"gateway_action":     "rzp_simulate_action_" + string(dec.Strategy),
			"target_amount":      payment.Amount,
			"execution_mode":     mode,
			"simulated_provider": "Razorpay Test Sandbox",
		},
		ExecutedAt: time.Now().UTC(),
	}
	if mode == domain.ExecutionModeSimulated {
		action.Status = domain.ActionStatusSimulated
	}

	_ = s.recoveryRepo.SaveAction(ctx, action)

	_ = s.eventBus.Publish(ctx, domain.NewEventEnvelope(
		domain.EventRecoveryActionExecuted,
		paymentID,
		action,
	))

	// 2. Measure & Record Outcome (Simulate realistic stochastic execution)
	pred, _ := s.recoveryRepo.GetPredictionByPaymentID(ctx, paymentID)
	prob := 0.70
	if pred != nil {
		prob = pred.RecoveryProbability
	}

	// Random outcome based on probability
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	isSuccess := r.Float64() <= prob

	recoveredAmount := 0.0
	cost := dec.ExpectedCost
	if isSuccess {
		recoveredAmount = float64(payment.Amount)
		_ = s.paymentRepo.UpdatePaymentStatus(ctx, paymentID, domain.PaymentStatusRecovered, payment.AttemptCount+1)
	} else {
		_ = s.paymentRepo.UpdatePaymentStatus(ctx, paymentID, domain.PaymentStatusFailed, payment.AttemptCount+1)
	}

	outcome := &domain.RecoveryOutcome{
		ID:               uuid.New(),
		ActionID:         actionID,
		PaymentID:        paymentID,
		Success:          isSuccess,
		RecoveredAmount:  recoveredAmount,
		RecoveryCost:     cost,
		NetRecoveryValue: recoveredAmount - cost,
		CompletedAt:      time.Now().UTC(),
	}

	_ = s.recoveryRepo.SaveOutcome(ctx, outcome)

	_ = s.eventBus.Publish(ctx, domain.NewEventEnvelope(
		domain.EventRecoveryOutcomeRecorded,
		paymentID,
		outcome,
	))

	s.auditLogger.Log(ctx, &paymentID, string(domain.EventRecoveryOutcomeRecorded), actor, map[string]interface{}{
		"success":            isSuccess,
		"recovered_amount":   recoveredAmount,
		"net_recovery_value": outcome.NetRecoveryValue,
		"execution_mode":     mode,
	})

	return outcome, nil
}

func validateExecutableDecision(dec *domain.RecoveryDecision) error {
	if dec == nil {
		return fmt.Errorf("recovery decision is missing")
	}
	if dec.PolicyStatus != domain.PolicyStatusApproved {
		return fmt.Errorf("recovery decision is not approved (current status: %s)", dec.PolicyStatus)
	}
	return nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
