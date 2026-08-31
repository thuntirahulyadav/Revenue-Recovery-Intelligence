package handler

import (
	"net/http"

	"razorpay-recovery-intelligence/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	recoveryRepo *repository.RecoveryRepo
}

func NewDashboardHandler(recoveryRepo *repository.RecoveryRepo) *DashboardHandler {
	return &DashboardHandler{recoveryRepo: recoveryRepo}
}

// GetOverview godoc
// GET /api/v1/dashboard/overview
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	overview, err := h.recoveryRepo.GetDashboardOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DASHBOARD_METRICS_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overview,
	})
}
