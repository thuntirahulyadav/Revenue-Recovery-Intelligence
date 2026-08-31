package handler

import (
	"net/http"
	"strconv"

	"razorpay-recovery-intelligence/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SimulationHandler struct {
	simulationService *service.SimulationService
}

func NewSimulationHandler(simulationService *service.SimulationService) *SimulationHandler {
	return &SimulationHandler{simulationService: simulationService}
}

// CompareSimulation godoc
// GET /api/v1/simulation/compare
func (h *SimulationHandler) CompareSimulation(c *gin.Context) {
	sampleSize, _ := strconv.Atoi(c.DefaultQuery("sample_size", "2500"))
	res, err := h.simulationService.RunComparisonSimulation(c.Request.Context(), sampleSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "SIMULATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}
