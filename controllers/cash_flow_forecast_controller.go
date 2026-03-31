package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
)

type CashFlowForecastController struct {
	forecastService services.CashFlowForecastService
}

func NewCashFlowForecastController(forecastService services.CashFlowForecastService) *CashFlowForecastController {
	return &CashFlowForecastController{forecastService: forecastService}
}

// GetForecast godoc
// GET /api/v2/cash-flow/forecast?days=30&asset_id=1
func (ctrl *CashFlowForecastController) GetForecast(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.CashFlowForecastRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.forecastService.GetForecast(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Cash flow forecast retrieved successfully", result)
}

// GetScenario godoc
// POST /api/v2/cash-flow/scenarios
func (ctrl *CashFlowForecastController) GetScenario(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.CashFlowScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.forecastService.GetScenario(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Cash flow scenario calculated successfully", result)
}
