package controllers

import (
    "my-api/config"
    "my-api/models"
    "my-api/services"
    "github.com/gin-gonic/gin"
    "my-api/utils"
    "net/http"
    "strconv"
    "time"
)

type TransactionController struct {
    transactionService services.TransactionService
    budgetService      services.BudgetService
}

func NewTransactionController(transactionService services.TransactionService, budgetService services.BudgetService) *TransactionController {
    return &TransactionController{
        transactionService: transactionService,
        budgetService:      budgetService,
    }
}

func GetInitialData(c *gin.Context) {
	var banks []models.Bank
    var categories []models.Category
    var users []models.User

    config.DB.Find(&banks)
    config.DB.Find(&categories)
    config.DB.Find(&users)

    data := gin.H{
        "banks": banks,
        "categories": categories,
        "users": users,
    }

    utils.JSONSuccess(c, "Initial data successfully fetched", data)
}

func (ctrl *TransactionController) GetTransactions(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }

    userIDUint, ok := userID.(uint)
    if !ok {
        utils.JSONError(c, http.StatusInternalServerError, "Invalid user ID")
        return
    }

    // Parse query parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    // Accept both page_size and limit for backward compatibility
    pageSize := c.Query("page_size")
    if pageSize == "" {
        pageSize = c.DefaultQuery("limit", "10")
    }
    limit, _ := strconv.Atoi(pageSize)

    // Parse optional filters
    var startDate, endDate *time.Time
    var transactionType *int
    var categoryID, bankID *uint

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
        // Handle both string ("Income"/"Expense") and numeric (1/2) values
        var txType int
        if txTypeStr == "Income" || txTypeStr == "income" {
            txType = 1
        } else if txTypeStr == "Expense" || txTypeStr == "expense" {
            txType = 2
        } else {
            // Try to parse as integer
            if parsed, err := strconv.Atoi(txTypeStr); err == nil {
                txType = parsed
            }
        }
        
        // Only set if valid type (1 or 2)
        if txType == 1 || txType == 2 {
            transactionType = &txType
        }
    }

    if catIDStr := c.Query("category_id"); catIDStr != "" {
        if catID, err := strconv.ParseUint(catIDStr, 10, 32); err == nil {
            catIDUint := uint(catID)
            categoryID = &catIDUint
        }
    }

    if bankIDStr := c.Query("bank_id"); bankIDStr != "" {
        if bID, err := strconv.ParseUint(bankIDStr, 10, 32); err == nil {
            bIDUint := uint(bID)
            bankID = &bIDUint
        }
    }

    transactions, pagination, err := ctrl.transactionService.GetTransactions(
        userIDUint, page, limit, startDate, endDate, transactionType, categoryID, bankID,
    )

    if err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to fetch transactions")
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Transactions fetched successfully",
        "data":    transactions,
        "pagination": pagination,
    })
}

func (ctrl *TransactionController) GetTransactionByID(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }

    userIDUint, ok := userID.(uint)
    if !ok {
        utils.JSONError(c, http.StatusInternalServerError, "Invalid user ID")
        return
    }

    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid transaction ID")
        return
    }

    transaction, err := ctrl.transactionService.GetTransactionByID(uint(id), userIDUint)
    if err != nil {
        utils.JSONError(c, http.StatusNotFound, "Transaction not found")
        return
    }

    utils.JSONSuccess(c, "Transaction fetched successfully", transaction)
}

func (ctrl *TransactionController) CreateTransaction(c *gin.Context) {

    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }

    // Convert userID to uint (since we set it as uint in middleware)
    userIDUint, ok := userID.(uint)
    if !ok {
        utils.JSONError(c, http.StatusInternalServerError, "Invalid user ID")
        return 
    }

    var transaction models.Transaction

    // First bind the JSON to a map to handle custom types
    var payload map[string]interface{}

    if err := c.ShouldBindJSON(&payload); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid request payload")
        return
    }

    // Set the UserID from the token
    transaction.UserID = userIDUint

    // Manually convert types
    transaction.Description = payload["Description"].(string)
    transaction.CategoryID = uint(payload["CategoryID"].(float64))
    transaction.BankID = uint(payload["BankID"].(float64))

    
    // Convert string amount to int
    if amountStr, ok := payload["Amount"].(string); ok {
        amount, err := strconv.Atoi(amountStr)
        if err != nil {
            utils.JSONError(c, http.StatusBadRequest, "Invalid amount format")
            return
        }
        transaction.Amount = amount
    } else if amount, ok := payload["Amount"].(float64); ok {
        // If sent as number instead of string
        transaction.Amount = int(amount)
    }

    // Handle TransactionType - can be string or number
    if txTypeStr, ok := payload["TransactionType"].(string); ok {
        // Convert string to number: "Income" = 1, "Expense" = 2
        if txTypeStr == "Income" {
            transaction.TransactionType = 1
        } else if txTypeStr == "Expense" {
            transaction.TransactionType = 2
        } else {
            utils.JSONError(c, http.StatusBadRequest, "Invalid transaction type. Must be 'Income' or 'Expense'")
            return
        }
    } else if txType, ok := payload["TransactionType"].(float64); ok {
        // If sent as number directly
        transaction.TransactionType = int(txType)
    }

    // Parse date - support multiple formats
    dateStr := payload["Date"].(string)
    var date utils.CustomTime
    var err error
    
    // Try datetime format first (YYYY-MM-DD HH:MM:SS)
    parsed, err := time.Parse("2006-01-02 15:04:05", dateStr)
    if err != nil {
        // Try ISO 8601 format (2025-12-26T06:54:44.955Z)
        parsed, err = time.Parse(time.RFC3339, dateStr)
        if err != nil {
            // Try simple date format (2006-01-02)
            parsed, err = time.Parse("2006-01-02", dateStr)
            if err != nil {
                utils.JSONError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD HH:MM:SS, ISO 8601 or YYYY-MM-DD")
                return
            }
        }
    }

    date.Time = parsed
    transaction.UserID = userIDUint
    transaction.Date   = date

    if err := ctrl.transactionService.CreateTransaction(&transaction); err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to create transaction")
        return
    }

    // Check budget alerts if this is an expense transaction
    if transaction.TransactionType == 2 {
        ctrl.budgetService.CheckBudgetAlerts(userIDUint)
    }

    utils.JSONSuccess(c, "Transaction created successfully", transaction)

}
func (ctrl *TransactionController) DeleteTransaction(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }

    userIDUint, ok := userID.(uint)
    if !ok {
        utils.JSONError(c, http.StatusInternalServerError, "Invalid user ID")
        return
    }

    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid transaction ID")
        return
    }

    if err := ctrl.transactionService.DeleteTransaction(uint(id), userIDUint); err != nil {
        utils.JSONError(c, http.StatusNotFound, "Transaction not found or unauthorized")
        return
    }

    utils.JSONSuccess(c, "Transaction deleted successfully", nil)
}
