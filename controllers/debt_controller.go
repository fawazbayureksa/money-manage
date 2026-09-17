package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"my-api/dto"
	"my-api/services"
	"my-api/utils"
)

type DebtController struct {
	service services.DebtService
}

func NewDebtController(service services.DebtService) *DebtController {
	return &DebtController{service: service}
}

// ─── Create Debt ──────────────────────────────────────────────────────────────

// POST /api/v2/debts
func (ctrl *DebtController) CreateDebt(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.CreateDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	debt, err := ctrl.service.CreateDebt(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt added successfully", debt)
}

// ─── List Debts ───────────────────────────────────────────────────────────────

// GET /api/v2/debts?status=active
func (ctrl *DebtController) GetAllDebts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	status := c.DefaultQuery("status", "")

	result, err := ctrl.service.GetAllDebts(userID.(uint), status)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debts retrieved successfully", result)
}

// ─── Get Debt Detail ──────────────────────────────────────────────────────────

// GET /api/v2/debts/:id?include_payments=true&include_milestones=true
func (ctrl *DebtController) GetDebt(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	includePayments := c.Query("include_payments") == "true"
	includeMilestones := c.Query("include_milestones") == "true"

	debt, err := ctrl.service.GetDebt(uint(id), userID.(uint), includePayments, includeMilestones)
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt retrieved successfully", debt)
}

// ─── Update Debt ──────────────────────────────────────────────────────────────

// PUT /api/v2/debts/:id
func (ctrl *DebtController) UpdateDebt(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	var req dto.UpdateDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	debt, err := ctrl.service.UpdateDebt(uint(id), userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt updated successfully", debt)
}

// ─── Delete Debt ──────────────────────────────────────────────────────────────

// DELETE /api/v2/debts/:id
func (ctrl *DebtController) DeleteDebt(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	if err := ctrl.service.DeleteDebt(uint(id), userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt deleted successfully", nil)
}

// ─── Record Payment ───────────────────────────────────────────────────────────

// POST /api/v2/debts/:id/payments
func (ctrl *DebtController) RecordPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	var req dto.RecordDebtPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	result, err := ctrl.service.RecordPayment(uint(id), userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payment recorded successfully", result)
}

// ─── Update Balance ───────────────────────────────────────────────────────────

// PUT /api/v2/debts/:id/balance
func (ctrl *DebtController) UpdateBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	var req dto.UpdateDebtBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if err := ctrl.service.UpdateBalance(uint(id), userID.(uint), req.NewBalance); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Balance updated successfully", nil)
}

// ─── Payoff Strategies ────────────────────────────────────────────────────────

// GET /api/v2/debts/strategies?extra_payment=500000
func (ctrl *DebtController) GetPayoffStrategies(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	extraPayment := 0
	if ep := c.Query("extra_payment"); ep != "" {
		if val, err := strconv.Atoi(ep); err == nil {
			extraPayment = val
		}
	}

	result, err := ctrl.service.GetPayoffStrategies(userID.(uint), extraPayment)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payoff strategies retrieved successfully", result)
}

// ─── Payoff Timeline ──────────────────────────────────────────────────────────

// GET /api/v2/debts/:id/timeline?strategy=avalanche&extra_payment=500000
func (ctrl *DebtController) GetPayoffTimeline(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	strategy := c.DefaultQuery("strategy", "minimum")
	extraPayment := 0
	if ep := c.Query("extra_payment"); ep != "" {
		if val, err := strconv.Atoi(ep); err == nil {
			extraPayment = val
		}
	}

	result, err := ctrl.service.GetPayoffTimeline(uint(id), userID.(uint), strategy, extraPayment)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payoff timeline retrieved successfully", result)
}
