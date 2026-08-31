package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"razorpay-recovery-intelligence/backend/internal/domain"
	"razorpay-recovery-intelligence/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	db *repository.DB
}

func NewSettingsHandler(db *repository.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

// GetSettings godoc
// GET /api/v1/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	query := `
		SELECT recovery_settings
		FROM merchants
		LIMIT 1
	`
	var settingsJSON []byte
	err := h.db.Pool.QueryRow(c.Request.Context(), query).Scan(&settingsJSON)

	defaultSettings := domain.MerchantRecoverySettings{
		MaxRetryAttempts:              3,
		MinConfidenceThreshold:        0.65,
		MaxCommAttempts:               2,
		HumanApprovalThreshold:        50000.0,
		HighValueTransactionThreshold: 25000.0,
		AutoExecutionEnabled:          true,
	}

	if err == nil && len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &defaultSettings)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    defaultSettings,
	})
}

// UpdateSettings godoc
// PUT /api/v1/settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var settings domain.MerchantRecoverySettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SETTINGS_PAYLOAD",
				"message": err.Error(),
			},
		})
		return
	}
	if err := validateSettings(settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_SETTINGS", "message": err.Error()}})
		return
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "SERIALIZATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	query := `
		UPDATE merchants
		SET recovery_settings = $1
		WHERE id = (SELECT id FROM merchants LIMIT 1)
	`
	result, err := h.db.Pool.Exec(c.Request.Context(), query, settingsJSON)
	if err != nil || result.RowsAffected() != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "SETTINGS_UPDATE_FAILED", "message": "Merchant settings could not be saved"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"meta": gin.H{
			"message": "Merchant recovery settings updated successfully.",
		},
	})
}

func validateSettings(s domain.MerchantRecoverySettings) error {
	if s.MaxRetryAttempts < 1 || s.MaxRetryAttempts > 10 {
		return fmt.Errorf("max_retry_attempts must be between 1 and 10")
	}
	if s.MinConfidenceThreshold < 0 || s.MinConfidenceThreshold > 1 {
		return fmt.Errorf("min_confidence_threshold must be between 0 and 1")
	}
	if s.MaxCommAttempts < 0 || s.MaxCommAttempts > 10 {
		return fmt.Errorf("max_comm_attempts must be between 0 and 10")
	}
	if s.HumanApprovalThreshold < 0 || s.HighValueTransactionThreshold < 0 {
		return fmt.Errorf("amount thresholds cannot be negative")
	}
	return nil
}
