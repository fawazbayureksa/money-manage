package controllers

import (
	"net/http"
	"strconv"
	"time"

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

func (ctrl *DebtController) CreateDebt(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse JSON into a map to handle date conversion
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	var req dto.CreateDebtRequest

	// Name
	if name, ok := payload["name"].(string); ok && name != "" {
		req.Name = name
	} else {
		utils.JSONError(c, http.StatusBadRequest, "name is required")
		return
	}

	// DebtType
	if debtType, ok := payload["debt_type"].(string); ok && debtType != "" {
		validTypes := map[string]bool{
			"credit_card": true, "loan": true, "mortgage": true,
			"student_loan": true, "personal": true, "other": true,
		}
		if !validTypes[debtType] {
			utils.JSONError(c, http.StatusBadRequest, "debt_type must be one of: credit_card, loan, mortgage, student_loan, personal, other")
			return
		}
		req.DebtType = debtType
	} else {
		utils.JSONError(c, http.StatusBadRequest, "debt_type is required")
		return
	}

	// OriginalAmount
	if origAmount, ok := payload["original_amount"].(float64); ok {
		req.OriginalAmount = int(origAmount)
	} else {
		utils.JSONError(c, http.StatusBadRequest, "original_amount is required")
		return
	}

	// CurrentBalance
	if currBalance, ok := payload["current_balance"].(float64); ok {
		req.CurrentBalance = int(currBalance)
	} else {
		utils.JSONError(c, http.StatusBadRequest, "current_balance is required")
		return
	}

	// InterestRate (optional)
	if interestRate, ok := payload["interest_rate"].(float64); ok {
		req.InterestRate = interestRate
	}

	// MinimumPayment (optional)
	if minPayment, ok := payload["minimum_payment"].(float64); ok {
		req.MinimumPayment = int(minPayment)
	}

	// DueDay (optional)
	if dueDay, ok := payload["due_day"].(float64); ok {
		req.DueDay = int(dueDay)
	}

	// StartDate
	if startDateStr, ok := payload["start_date"].(string); ok {
		parsed, err := parseDate(startDateStr)
		if err != nil {
			utils.JSONError(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD HH:MM:SS, YYYY-MM-DD, or ISO 8601")
			return
		}
		req.StartDate = utils.CustomTime{Time: parsed}
	} else {
		utils.JSONError(c, http.StatusBadRequest, "start_date is required")
		return
	}

	// LenderName (optional)
	if lenderName, ok := payload["lender_name"].(string); ok {
		req.LenderName = lenderName
	}

	// Notes (optional)
	if notes, ok := payload["notes"].(string); ok {
		req.Notes = notes
	}

	debt, err := ctrl.service.CreateDebt(userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt created successfully", debt)
}

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

	debt, err := ctrl.service.GetDebtByID(uint(id), userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt retrieved successfully", debt)
}

func (ctrl *DebtController) GetDebts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var filter dto.DebtFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetAllDebts(userID.(uint), &filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debts retrieved successfully", result)
}

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
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	debt, err := ctrl.service.UpdateDebt(uint(id), userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt updated successfully", debt)
}

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

// --- Payment endpoints ---

func (ctrl *DebtController) RecordPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	debtID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	// Parse JSON into a map to handle date conversion
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	var req dto.CreateDebtPaymentRequest

	// Amount
	if amount, ok := payload["amount"].(float64); ok {
		req.Amount = int(amount)
		if req.Amount < 1 {
			utils.JSONError(c, http.StatusBadRequest, "amount must be at least 1")
			return
		}
	} else {
		utils.JSONError(c, http.StatusBadRequest, "amount is required")
		return
	}

	// PaymentDate
	if paymentDateStr, ok := payload["payment_date"].(string); ok {
		parsed, err := parseDate(paymentDateStr)
		if err != nil {
			utils.JSONError(c, http.StatusBadRequest, "Invalid payment_date format. Use YYYY-MM-DD HH:MM:SS, YYYY-MM-DD, or ISO 8601")
			return
		}
		req.PaymentDate = utils.CustomTime{Time: parsed}
	} else {
		utils.JSONError(c, http.StatusBadRequest, "payment_date is required")
		return
	}

	// Notes (optional)
	if notes, ok := payload["notes"].(string); ok {
		req.Notes = notes
	}

	payment, err := ctrl.service.RecordPayment(uint(debtID), userID.(uint), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payment recorded successfully", payment)
}

func (ctrl *DebtController) GetPayments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	debtID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	var filter dto.DebtPaymentFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetPayments(uint(debtID), userID.(uint), &filter)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payments retrieved successfully", result)
}

func (ctrl *DebtController) DeletePayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	paymentID, err := strconv.ParseUint(c.Param("payment_id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid payment ID")
		return
	}

	if err := ctrl.service.DeletePayment(uint(paymentID), userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payment deleted successfully", nil)
}

// --- Analytics & Strategy endpoints ---

func (ctrl *DebtController) GetDebtSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	summary, err := ctrl.service.GetDebtSummary(userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Debt summary retrieved successfully", summary)
}

func (ctrl *DebtController) GetPayoffStrategies(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	strategies, err := ctrl.service.GetPayoffStrategies(userID.(uint))
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payoff strategies retrieved successfully", strategies)
}

func (ctrl *DebtController) GetPayoffProjection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	debtID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid debt ID")
		return
	}

	extraPayment := 0
	if extra := c.Query("extra_payment"); extra != "" {
		if val, err := strconv.Atoi(extra); err == nil {
			extraPayment = val
		}
	}

	projection, err := ctrl.service.GetPayoffProjection(uint(debtID), userID.(uint), extraPayment)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Payoff projection retrieved successfully", projection)
}

// --- Milestone endpoints ---

func (ctrl *DebtController) GetMilestones(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	unreadOnly := c.Query("unread_only") == "true"

	milestones, err := ctrl.service.GetMilestones(userID.(uint), unreadOnly)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Milestones retrieved successfully", milestones)
}

func (ctrl *DebtController) MarkMilestoneAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid milestone ID")
		return
	}

	if err := ctrl.service.MarkMilestoneAsRead(uint(id), userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "Milestone marked as read", nil)
}

func (ctrl *DebtController) MarkAllMilestonesAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := ctrl.service.MarkAllMilestonesAsRead(userID.(uint)); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "All milestones marked as read", nil)
}

// --- Helper ---

func parseDate(dateStr string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		parsed, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, dateStr)
			if err != nil {
				return time.Time{}, err
			}
		}
	}
	return parsed, nil
}
