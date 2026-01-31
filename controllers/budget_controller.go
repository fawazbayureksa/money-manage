package controllers

import (
	"github.com/gin-gonic/gin"
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
	"net/http"
	"strconv"
	"time"
)

type BudgetController struct {
	service services.BudgetService
}

func NewBudgetController(service services.BudgetService) *BudgetController {
	return &BudgetController{service: service}
}

func (ctrl *BudgetController) CreateBudget(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse JSON into a map first to handle date conversion
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Create the request struct
	var req dto.CreateBudgetRequest

	// Manually map fields with proper type conversion
	if categoryID, ok := payload["category_id"].(float64); ok {
		req.CategoryID = uint(categoryID)
	} else {
		utils.JSONError(c, http.StatusBadRequest, "category_id is required")
		return
	}

	amountVal, ok := payload["amount"]
	if !ok {
		utils.JSONError(c, http.StatusBadRequest, "amount is required")
		return
	}

	switch v := amountVal.(type) {
	case float64:
		req.Amount = int(v)
	case int:
		req.Amount = v
	case int64:
		req.Amount = int(v)
	default:
		utils.JSONError(c, http.StatusBadRequest, "amount must be a number")
		return
	}

	if period, ok := payload["period"].(string); ok {
		req.Period = period
	} else {
		utils.JSONError(c, http.StatusBadRequest, "period is required")
		return
	}

	// Parse start_date
	if startDateStr, ok := payload["start_date"].(string); ok {
		parsed, err := time.Parse("2006-01-02 15:04:05", startDateStr)
		if err != nil {
			// Try simple date format
			parsed, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				// Try ISO 8601 format
				parsed, err = time.Parse(time.RFC3339, startDateStr)
				if err != nil {
					utils.JSONError(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD HH:MM:SS, YYYY-MM-DD or ISO 8601")
					return
				}
			}
		}
		req.StartDate = utils.CustomTime{Time: parsed}
	} else {
		utils.JSONError(c, http.StatusBadRequest, "start_date is required")
		return
	}

	if alertAt, ok := payload["alert_at"].(float64); ok {
		req.AlertAt = int(alertAt)
	}

	if description, ok := payload["description"].(string); ok {
		req.Description = description
	}

	budget, err := ctrl.service.CreateBudget(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Budget created successfully", budget)
}

func (ctrl *BudgetController) GetBudget(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	budget, err := ctrl.service.GetBudgetByID(uint(id), userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.JSONSuccess(c, "Budget retrieved successfully", budget)
}

func (ctrl *BudgetController) GetBudgets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var filter dto.BudgetFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetAllBudgets(userID.(uint), &filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Budgets retrieved successfully", result)
}

func (ctrl *BudgetController) UpdateBudget(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	var req dto.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	budget, err := ctrl.service.UpdateBudget(uint(id), userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Budget updated successfully", budget)
}

func (ctrl *BudgetController) DeleteBudget(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	if err := ctrl.service.DeleteBudget(uint(id), userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Budget deleted successfully", nil)
}

func (ctrl *BudgetController) GetBudgetStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	budgets, err := ctrl.service.GetBudgetStatus(userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Budget status retrieved successfully", budgets)
}

func (ctrl *BudgetController) GetAlerts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var filter dto.AlertFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetUserAlertsPaginated(userID.(uint), &filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Alerts retrieved successfully", result)
}

func (ctrl *BudgetController) MarkAlertAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	if err := ctrl.service.MarkAlertAsRead(uint(id), userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Alert marked as read", nil)
}

func (ctrl *BudgetController) MarkAllAlertsAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := ctrl.service.MarkAllAlertsAsRead(userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "All alerts marked as read", nil)
}
