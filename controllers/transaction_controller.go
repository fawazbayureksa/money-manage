package controllers

import (
    "my-api/config"
    "my-api/models"
    "github.com/gin-gonic/gin"
    "my-api/utils"
    "net/http"
    "strconv"
    "time"
)

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

func CreateTransaction(c *gin.Context) {

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

    transaction.TransactionType = int(payload["TransactionType"].(float64))

    // Parse date
    dateStr := payload["Date"].(string)
    date, err := time.Parse("2006-01-02", dateStr)

    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid date format")
        return
    }

    transaction.UserID = userIDUint
    transaction.Date   = date

    if err := config.DB.Create(&transaction).Error; err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to create transaction")
        return
    }
    utils.JSONSuccess(c, "Transaction created successfully", transaction)

}