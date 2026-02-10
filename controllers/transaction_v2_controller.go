package controllers

import (
	"my-api/dto"
	"my-api/models"
	"my-api/services"
	"my-api/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TransactionV2Controller struct {
	transactionService services.TransactionV2Service
}

func NewTransactionV2Controller(transactionService services.TransactionV2Service) *TransactionV2Controller {
	return &TransactionV2Controller{
		transactionService: transactionService,
	}
}

func (ctrl *TransactionV2Controller) GetTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// Accept both page_size and limit for backward compatibility
	pageSize := c.Query("page_size")
	if pageSize == "" {
		pageSize = c.DefaultQuery("limit", "10")
	}
	limit, _ := strconv.Atoi(pageSize)

	var startDate, endDate *time.Time
	var transactionType *int
	var categoryID *uint64
	var assetID *uint64

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	if txTypeStr := c.Query("transaction_type"); txTypeStr != "" {
		var txType int
		if txTypeStr == "Income" || txTypeStr == "income" {
			txType = 1
		} else if txTypeStr == "Expense" || txTypeStr == "expense" {
			txType = 2
		} else {
			if parsed, err := strconv.Atoi(txTypeStr); err == nil {
				txType = parsed
			}
		}

		if txType == 1 || txType == 2 {
			transactionType = &txType
		}
	}

	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if catID, err := strconv.ParseUint(catIDStr, 10, 64); err == nil {
			catIDUint := catID
			categoryID = &catIDUint
		}
	}

	if assetIDStr := c.Query("asset_id"); assetIDStr != "" {
		if aID, err := strconv.ParseUint(assetIDStr, 10, 64); err == nil {
			assetID = &aID
		}
	}

	transactions, pagination, err := ctrl.transactionService.GetTransactions(
		userIDUint, page, limit, startDate, endDate, transactionType, categoryID, assetID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Transactions fetched successfully",
		"data":       transactions,
		"pagination": pagination,
	})
}

func (ctrl *TransactionV2Controller) GetTransactionByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid transaction ID"})
		return
	}

	transaction, err := ctrl.transactionService.GetTransactionByID(uint(id), userIDUint)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction fetched successfully",
		"data":    transaction,
	})
}

func (ctrl *TransactionV2Controller) CreateTransaction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	var req dto.CreateTransactionV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		date, err = time.Parse(time.RFC3339, req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid date format. Use YYYY-MM-DD or ISO 8601"})
			return
		}
	}

	transactionType := 1
	if req.TransactionType == "Expense" || req.TransactionType == "expense" {
		transactionType = 2
	}

	transaction := &models.TransactionV2{
		UserID:          userIDUint,
		Description:     req.Description,
		CategoryID:      req.CategoryID,
		AssetID:         req.AssetID,
		Amount:          req.Amount,
		TransactionType: transactionType,
		Date:            utils.CustomTime{Time: date},
		BankID:          0, // Optional for v2
	}

	if err := ctrl.transactionService.CreateTransaction(transaction); err != nil {
		if err.Error() == "insufficient balance" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Insufficient balance in the selected asset"})
			return
		}
		if err.Error() == "asset not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Asset not found"})
			return
		}
		if err.Error() == "unauthorized: asset does not belong to user" {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Asset does not belong to you"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create transaction"})
		return
	}

	created, _ := ctrl.transactionService.GetTransactionByID(transaction.ID, userIDUint)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Transaction created successfully",
		"data":    created,
	})
}

func (ctrl *TransactionV2Controller) UpdateTransaction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid transaction ID"})
		return
	}

	existing, err := ctrl.transactionService.GetTransactionByID(uint(id), userIDUint)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Transaction not found"})
		return
	}

	var req dto.UpdateTransactionV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	oldAmount := existing.Amount
	oldType := existing.TransactionType

	transaction := &models.TransactionV2{
		ID:              uint(id),
		UserID:          userIDUint,
		Description:     existing.Description,
		CategoryID:      0,
		AssetID:         existing.AssetID,
		Amount:          existing.Amount,
		TransactionType: existing.TransactionType,
		Date:            existing.Date,
		BankID:          0,
	}

	if req.Description != nil {
		transaction.Description = *req.Description
	}
	if req.CategoryID != nil {
		transaction.CategoryID = *req.CategoryID
	}
	if req.AssetID != nil {
		transaction.AssetID = *req.AssetID
	}
	if req.Amount != nil {
		transaction.Amount = *req.Amount
	}
	if req.TransactionType != nil {
		if *req.TransactionType == "Expense" || *req.TransactionType == "expense" {
			transaction.TransactionType = 2
		} else {
			transaction.TransactionType = 1
		}
	}
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			date, err = time.Parse(time.RFC3339, *req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid date format"})
				return
			}
		}
		transaction.Date = utils.CustomTime{Time: date}
	}

	if err := ctrl.transactionService.UpdateTransaction(transaction, oldAmount, oldType); err != nil {
		if err.Error() == "insufficient balance" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Insufficient balance in the selected asset"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update transaction"})
		return
	}

	updated, _ := ctrl.transactionService.GetTransactionByID(uint(id), userIDUint)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction updated successfully",
		"data":    updated,
	})
}

func (ctrl *TransactionV2Controller) DeleteTransaction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid transaction ID"})
		return
	}

	if err := ctrl.transactionService.DeleteTransaction(uint(id), userIDUint); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Transaction not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction deleted successfully",
	})
}

func (ctrl *TransactionV2Controller) GetAssetTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	assetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid asset ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// Accept both page_size and limit for backward compatibility
	pageSize := c.Query("page_size")
	if pageSize == "" {
		pageSize = c.DefaultQuery("limit", "50")
	}
	limit, _ := strconv.Atoi(pageSize)

	response, err := ctrl.transactionService.GetAssetTransactions(assetID, userIDUint, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch asset transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Asset transactions fetched successfully",
		"data":    response,
	})
}
