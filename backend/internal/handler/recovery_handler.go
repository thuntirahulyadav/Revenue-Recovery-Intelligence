package handler

import (
	"net/http"
	"strconv"
	"time"

	"razorpay-recovery-intelligence/backend/internal/domain"
	"razorpay-recovery-intelligence/backend/internal/kafka"
	"razorpay-recovery-intelligence/backend/internal/repository"
	"razorpay-recovery-intelligence/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecoveryHandler struct {
	recoveryService *service.RecoveryService
	paymentRepo     *repository.PaymentRepo
	customerRepo    *repository.CustomerRepo
	recoveryRepo    *repository.RecoveryRepo
	eventBus        *kafka.EventBus
}

func NewRecoveryHandler(
	recoveryService *service.RecoveryService,
	paymentRepo *repository.PaymentRepo,
	customerRepo *repository.CustomerRepo,
	recoveryRepo *repository.RecoveryRepo,
	eventBus *kafka.EventBus,
) *RecoveryHandler {
	return &RecoveryHandler{
		recoveryService: recoveryService,
		paymentRepo:     paymentRepo,
		customerRepo:    customerRepo,
		recoveryRepo:    recoveryRepo,
		eventBus:        eventBus,
	}
}

type IngestFailedPaymentRequest struct {
	PaymentID     *string `json:"payment_id"`
	MerchantID    *string `json:"merchant_id"`
	CustomerID    *string `json:"customer_id"`
	Amount        int64   `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	FailureReason string  `json:"failure_reason" binding:"required"`
	AttemptCount  int     `json:"attempt_count"`
}

// IngestPaymentFailed godoc
// POST /api/v1/events/payment-failed
func (h *RecoveryHandler) IngestPaymentFailed(c *gin.Context) {
	var req IngestFailedPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST_PAYLOAD",
				"message": err.Error(),
			},
		})
		return
	}

	paymentID := uuid.New()
	if req.PaymentID != nil && *req.PaymentID != "" {
		if parsed, err := uuid.Parse(*req.PaymentID); err == nil {
			paymentID = parsed
		}
	}

	merchantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if req.MerchantID != nil && *req.MerchantID != "" {
		if parsed, err := uuid.Parse(*req.MerchantID); err == nil {
			merchantID = parsed
		}
	}

	customerID := uuid.New()
	if req.CustomerID != nil && *req.CustomerID != "" {
		if parsed, err := uuid.Parse(*req.CustomerID); err == nil {
			customerID = parsed
		}
	}

	// Ensure customer exists (create if necessary)
	_, err := h.customerRepo.GetOrCreateCustomer(c.Request.Context(), merchantID, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create or retrieve customer: " + err.Error(),
			},
		})
		return
	}

	currency := "INR"
	if req.Currency != "" {
		currency = req.Currency
	}

	attemptCount := 1
	if req.AttemptCount > 0 {
		attemptCount = req.AttemptCount
	}

	payment := &domain.Payment{
		ID:            paymentID,
		MerchantID:    merchantID,
		CustomerID:    customerID,
		Amount:        req.Amount,
		Currency:      currency,
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		Status:        domain.PaymentStatusFailed,
		FailureReason: domain.FailureReason(req.FailureReason),
		AttemptCount:  attemptCount,
		CreatedAt:     time.Now().UTC(),
	}

	if err := h.paymentRepo.CreatePayment(c.Request.Context(), payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to persist failed payment event: " + err.Error(),
			},
		})
		return
	}

	// Publish payment.failed event onto Kafka / EventBus
	envelope := domain.NewEventEnvelope(
		domain.EventPaymentFailed,
		paymentID,
		payment,
	)
	_ = h.eventBus.Publish(c.Request.Context(), envelope)

	// Synchronously run pipeline so immediate response has complete recovery decision
	analysis, err := h.recoveryService.OrchestrateRecoveryPipeline(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"payment_id": paymentID,
				"status":     "INGESTED",
				"message":    "Payment failure recorded, background recovery analysis queued.",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    analysis,
		"meta": gin.H{
			"correlation_id": paymentID,
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// GetPaymentRecovery godoc
// GET /api/v1/payments/:payment_id/recovery
func (h *RecoveryHandler) GetPaymentRecovery(c *gin.Context) {
	idStr := c.Param("payment_id")
	paymentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PAYMENT_ID",
				"message": "Payment ID must be a valid UUID",
			},
		})
		return
	}

	payment, err := h.paymentRepo.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil || payment == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PAYMENT_NOT_FOUND",
				"message": "No payment found with ID " + idStr,
			},
		})
		return
	}

	customer, _ := h.customerRepo.GetCustomerByID(c.Request.Context(), payment.CustomerID)
	if customer == nil {
		customer = &domain.Customer{
			ID:                    payment.CustomerID,
			MerchantID:            payment.MerchantID,
			HistoricalSuccessRate: 0.85,
			HistoricalFailureRate: 0.15,
			CustomerValue:         15000.0,
		}
	}

	pred, _ := h.recoveryRepo.GetPredictionByPaymentID(c.Request.Context(), paymentID)
	dec, _ := h.recoveryRepo.GetDecisionByPaymentID(c.Request.Context(), paymentID)
	action, _ := h.recoveryRepo.GetActionByPaymentID(c.Request.Context(), paymentID)
	outcome, _ := h.recoveryRepo.GetOutcomeByPaymentID(c.Request.Context(), paymentID)

	// If no prediction yet, run recovery pipeline on demand
	if pred == nil || dec == nil {
		analysis, err := h.recoveryService.OrchestrateRecoveryPipeline(c.Request.Context(), paymentID)
		if err == nil && analysis != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    analysis,
			})
			return
		}
	}

	analysis := &domain.FullPaymentRecoveryAnalysis{
		Payment:    *payment,
		Customer:   *customer,
		Prediction: pred,
		Decision:   dec,
		Action:     action,
		Outcome:    outcome,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
	})
}

// GetOpportunities godoc
// GET /api/v1/recovery/opportunities
func (h *RecoveryHandler) GetOpportunities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	minProb, _ := strconv.ParseFloat(c.DefaultQuery("min_probability", "0"), 64)
	minAmt, _ := strconv.ParseInt(c.DefaultQuery("min_amount", "0"), 10, 64)

	filter := repository.OpportunitiesFilter{
		FailureReason:  c.Query("failure_reason"),
		Strategy:       c.Query("strategy"),
		MinProbability: minProb,
		MinAmount:      minAmt,
		Status:         c.Query("status"),
		Search:         c.Query("search"),
		SortBy:         c.DefaultQuery("sort_by", "priority_score"),
		SortOrder:      c.DefaultQuery("sort_order", "DESC"),
		Page:           page,
		Limit:          limit,
	}

	result, err := h.paymentRepo.GetRecoveryOpportunities(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "QUERY_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Items,
		"meta": gin.H{
			"total_count": result.TotalCount,
			"page":        result.Page,
			"limit":       result.Limit,
			"total_pages": result.TotalPages,
		},
	})
}

// ApproveRecovery godoc
// POST /api/v1/recovery/:payment_id/approve
func (h *RecoveryHandler) ApproveRecovery(c *gin.Context) {
	idStr := c.Param("payment_id")
	paymentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PAYMENT_ID",
				"message": "Payment ID must be a valid UUID",
			},
		})
		return
	}

	actor := "merchant.admin"
	dec, err := h.recoveryService.ApproveRecovery(c.Request.Context(), paymentID, actor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "APPROVAL_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dec,
		"meta": gin.H{
			"message": "Policy approval granted. Recovery action ready for execution.",
		},
	})
}

type ExecuteRecoveryRequest struct {
	ExecutionMode string `json:"execution_mode"` // SIMULATED | MOCK | REAL
}

// ExecuteRecovery godoc
// POST /api/v1/recovery/:payment_id/execute
func (h *RecoveryHandler) ExecuteRecovery(c *gin.Context) {
	idStr := c.Param("payment_id")
	paymentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PAYMENT_ID",
				"message": "Payment ID must be a valid UUID",
			},
		})
		return
	}

	var req ExecuteRecoveryRequest
	_ = c.ShouldBindJSON(&req)

	mode := domain.ExecutionModeSimulated
	if req.ExecutionMode == string(domain.ExecutionModeMock) {
		mode = domain.ExecutionModeMock
	} else if req.ExecutionMode == string(domain.ExecutionModeReal) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "REAL_EXECUTION_UNAVAILABLE",
				"message": "Live gateway execution is not configured. Use SIMULATED or MOCK mode.",
			},
		})
		return
	}

	actor := "merchant.operator"
	outcome, err := h.recoveryService.ExecuteRecovery(c.Request.Context(), paymentID, mode, actor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "EXECUTION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    outcome,
		"meta": gin.H{
			"execution_mode": mode,
			"message":        "Recovery action executed successfully in " + string(mode) + " mode.",
		},
	})
}
