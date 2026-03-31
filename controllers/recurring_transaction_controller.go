package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
)

type RecurringTransactionController struct {
	service services.RecurringTransactionService
}

func NewRecurringTransactionController(service services.RecurringTransactionService) *RecurringTransactionController {
	return &RecurringTransactionController{service: service}
}

func (ctrl *RecurringTransactionController) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.CreateRecurringTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.service.Create(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Recurring transaction created successfully", result)
}

func (ctrl *RecurringTransactionController) GetByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	result, err := ctrl.service.GetByID(id, userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.JSONSuccess(c, "Recurring transaction retrieved successfully", result)
}

func (ctrl *RecurringTransactionController) GetAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var filter dto.RecurringTransactionFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetAll(userID.(uint), &filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Recurring transactions retrieved successfully", result)
}

func (ctrl *RecurringTransactionController) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req dto.UpdateRecurringTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.service.Update(id, userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Recurring transaction updated successfully", result)
}

func (ctrl *RecurringTransactionController) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := ctrl.service.Delete(id, userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Recurring transaction deleted successfully", nil)
}
