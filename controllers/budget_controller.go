package controllers

import (
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

	var req dto.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
		return
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

	unreadOnly := c.Query("unread_only") == "true"

	alerts, err := ctrl.service.GetUserAlerts(userID.(uint), unreadOnly)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Alerts retrieved successfully", alerts)
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
