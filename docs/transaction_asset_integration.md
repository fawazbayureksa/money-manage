# Transaction-Asset Integration Implementation Plan

## Overview
This document provides a step-by-step implementation plan for integrating Transaction and Asset models to enable automatic balance management.

## Current State Analysis

### Transaction Model (`models/transaction.go`)
- Has `BankID` field but no `AssetID`
- No direct relationship with Asset model
- Balance not automatically updated on transaction operations

### Asset Model (`models/assets.go`)
- Has `Balance` field
- Not linked to transactions
- Balance updated manually

---

## Implementation Steps

### Step 1: Update Transaction Model

**File**: `models/transaction.go`

Add `AssetID` field and Asset relation:

```go
type Transaction struct {
    ID              uint            `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
    Description     string          `gorm:"size:200;not null" json:"description"`
    UserID          uint            `gorm:"not null;index;type:int unsigned" json:"user_id"`
    CategoryID      uint            `gorm:"not null;index;type:int unsigned" json:"category_id"`
    BankID          uint            `gorm:"not null;index;type:int unsigned" json:"bank_id"`
    AssetID         uint64          `gorm:"not null;index;type:bigint unsigned" json:"asset_id"`
    Amount          int             `gorm:"not null" json:"amount"`
    TransactionType int             `gorm:"not null" json:"transaction_type"` // 1=income, 2=expense
    Date            utils.CustomTime `gorm:"not null;index;type:datetime" json:"date"`
    CreatedAt       utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
    UpdatedAt       utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`

    // Relations
    User     User     `gorm:"foreignKey:UserID" json:"-"`
    Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Bank     Bank     `gorm:"foreignKey:BankID" json:"bank,omitempty"`
    Asset    Asset    `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}
```

---

### Step 2: Create Database Migration

**File**: `models/transaction_migration.go`

```go
package models

import (
    "gorm.io/gorm"
    "log"
)

func AddAssetIDToTransactions(db *gorm.DB) error {
    // Step 1: Add asset_id column
    err := db.Exec(`
        ALTER TABLE transactions 
        ADD COLUMN asset_id BIGINT UNSIGNED NOT NULL DEFAULT 0
    `).Error

    if err != nil {
        log.Printf("Error adding asset_id column: %v", err)
        return err
    }

    // Step 2: Create index on asset_id
    err = db.Exec(`
        CREATE INDEX idx_transactions_asset_id 
        ON transactions(asset_id)
    `).Error

    if err != nil {
        log.Printf("Error creating index on asset_id: %v", err)
        return err
    }

    // Step 3: Migrate existing data from bank_id to asset_id
    // This assumes each bank has a corresponding asset with matching ID
    err = db.Exec(`
        UPDATE transactions t
        JOIN assets a ON t.bank_id = a.id
        SET t.asset_id = a.id
        WHERE t.asset_id = 0
    `).Error

    if err != nil {
        log.Printf("Error migrating data from bank_id to asset_id: %v", err)
        return err
    }

    // Step 4: Add foreign key constraint
    err = db.Exec(`
        ALTER TABLE transactions 
        ADD CONSTRAINT fk_transactions_asset 
        FOREIGN KEY (asset_id) REFERENCES assets(id)
    `).Error

    if err != nil {
        log.Printf("Error adding foreign key constraint: %v", err)
        return err
    }

    log.Println("Successfully added asset_id to transactions table")
    return nil
}

// Alternative: Using GORM AutoMigrate
func MigrateTransactionWithAsset(db *gorm.DB) error {
    // Ensure the updated Transaction struct is migrated
    return db.AutoMigrate(&Transaction{})
}
```

**File**: `models/assets_migration.go` (update existing)

```go
// Update the existing migration to include the transaction relationship
func MigrateAssets(db *gorm.DB) error {
    if err := db.AutoMigrate(&Asset{}); err != nil {
        return err
    }
    
    // Add transactions back-reference
    err := db.Exec(`
        ALTER TABLE assets 
        ADD COLUMN transactions []Transaction
    `).Error
    
    // Note: This may not work in all databases, 
    // so we'll handle transactions relationship in GORM instead
    return err
}
```

---

### Step 3: Update Transaction DTO

**File**: `dto/analytics_dto.go` (or create `dto/transaction_dto.go`)

Update `TransactionResponse` to include asset information:

```go
type TransactionResponse struct {
    ID              uint              `json:"id"`
    Description     string            `json:"description"`
    Amount          int               `json:"amount"`
    TransactionType int               `json:"transaction_type"`
    Date            utils.CustomTime  `json:"date"`
    CategoryName    string            `json:"category_name"`
    BankName        string            `json:"bank_name"`
    AssetName       string            `json:"asset_name,omitempty"`
    AssetBalance    float64           `json:"asset_balance,omitempty"`
    AssetCurrency   string            `json:"asset_currency,omitempty"`
}
```

Create request DTO for creating transactions:

```go
type CreateTransactionRequest struct {
    Description     string    `json:"Description" binding:"required"`
    CategoryID      uint      `json:"CategoryID" binding:"required"`
    AssetID         uint64    `json:"AssetID" binding:"required"`
    Amount          string    `json:"Amount" binding:"required"`
    TransactionType string    `json:"TransactionType" binding:"required,oneof=Income Expense"`
    Date            string    `json:"Date" binding:"required"`
}
```

---

### Step 4: Update Transaction Repository

**File**: `repositories/transaction_repository.go`

Update methods to include Asset preloading and filtering:

```go
type TransactionRepository interface {
    GetAll(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID, assetID *uint64) ([]models.Transaction, int64, error)
    GetByID(id, userID uint) (*models.Transaction, error)
    GetByIDWithAsset(id, userID uint) (*models.Transaction, error)
    Create(transaction *models.Transaction) error
    CreateWithBalanceUpdate(transaction *models.Transaction) error
    Update(transaction *models.Transaction) error
    UpdateWithBalanceUpdate(transaction *models.Transaction, oldAmount int, oldType int) error
    Delete(id, userID uint) error
    DeleteWithBalanceRollback(id, userID uint) error
}

func (r *transactionRepository) GetAll(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID, assetID *uint64) ([]models.Transaction, int64, error) {
    var transactions []models.Transaction
    var total int64

    query := r.db.Model(&models.Transaction{}).Where("user_id = ?", userID)

    // Apply filters
    if startDate != nil {
        query = query.Where("date >= ?", startDate)
    }
    if endDate != nil {
        query = query.Where("date <= ?", endDate)
    }
    if transactionType != nil {
        query = query.Where("transaction_type = ?", *transactionType)
    }
    if categoryID != nil {
        query = query.Where("category_id = ?", *categoryID)
    }
    if bankID != nil {
        query = query.Where("bank_id = ?", *bankID)
    }
    if assetID != nil {
        query = query.Where("asset_id = ?", *assetID)
    }

    // Count total
    query.Count(&total)

    // Apply pagination
    offset := (page - 1) * limit
    err := query.
        Preload("Category").
        Preload("Bank").
        Preload("Asset").
        Order("date DESC, id DESC").
        Limit(limit).
        Offset(offset).
        Find(&transactions).Error

    return transactions, total, err
}

func (r *transactionRepository) GetByIDWithAsset(id, userID uint) (*models.Transaction, error) {
    var transaction models.Transaction
    err := r.db.
        Preload("Category").
        Preload("Bank").
        Preload("Asset").
        Where("id = ? AND user_id = ?", id, userID).
        First(&transaction).Error
    
    if err != nil {
        return nil, err
    }
    return &transaction, nil
}

func (r *transactionRepository) CreateWithBalanceUpdate(transaction *models.Transaction) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // Lock asset row for update to prevent race conditions
        var asset models.Asset
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&asset, transaction.AssetID).Error; err != nil {
            return err
        }

        // Validate sufficient balance for expense
        if transaction.TransactionType == 2 && asset.Balance < float64(transaction.Amount) {
            return errors.New("insufficient balance")
        }

        // Update asset balance
        if transaction.TransactionType == 1 { // Income
            asset.Balance += float64(transaction.Amount)
        } else { // Expense
            asset.Balance -= float64(transaction.Amount)
        }

        // Save updated asset
        if err := tx.Save(&asset).Error; err != nil {
            return err
        }

        // Create transaction
        return tx.Create(transaction).Error
    })
}

func (r *transactionRepository) UpdateWithBalanceUpdate(transaction *models.Transaction, oldAmount int, oldType int) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // Get asset with lock
        var asset models.Asset
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&asset, transaction.AssetID).Error; err != nil {
            return err
        }

        // Revert old transaction effect
        if oldType == 1 { // Income
            asset.Balance -= float64(oldAmount)
        } else { // Expense
            asset.Balance += float64(oldAmount)
        }

        // Validate sufficient balance for new expense
        if transaction.TransactionType == 2 && asset.Balance < float64(transaction.Amount) {
            return errors.New("insufficient balance")
        }

        // Apply new transaction effect
        if transaction.TransactionType == 1 { // Income
            asset.Balance += float64(transaction.Amount)
        } else { // Expense
            asset.Balance -= float64(transaction.Amount)
        }

        // Save updated asset
        if err := tx.Save(&asset).Error; err != nil {
            return err
        }

        // Update transaction
        return tx.Save(transaction).Error
    })
}

func (r *transactionRepository) DeleteWithBalanceRollback(id, userID uint) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // Get transaction first to get amount and type
        var transaction models.Transaction
        if err := tx.Where("id = ? AND user_id = ?", id, userID).
            First(&transaction).Error; err != nil {
            return err
        }

        // Get asset with lock
        var asset models.Asset
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&asset, transaction.AssetID).Error; err != nil {
            return err
        }

        // Rollback balance (reverse the transaction)
        if transaction.TransactionType == 1 { // Income - subtract
            asset.Balance -= float64(transaction.Amount)
        } else { // Expense - add back
            asset.Balance += float64(transaction.Amount)
        }

        // Save updated asset
        if err := tx.Save(&asset).Error; err != nil {
            return err
        }

        // Delete transaction
        return tx.Delete(&transaction).Error
    })
}
```

---

### Step 5: Update Transaction Service

**File**: `services/transaction_service.go`

```go
type TransactionService interface {
    GetTransactions(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID, assetID *uint64) ([]dto.TransactionResponse, *dto.PaginationResponse, error)
    GetTransactionByID(id, userID uint) (*dto.TransactionResponse, error)
    CreateTransaction(transaction *models.Transaction) error
    CreateTransactionWithBalanceUpdate(transaction *models.Transaction) error
    UpdateTransaction(transaction *models.Transaction) error
    UpdateTransactionWithBalanceUpdate(transaction *models.Transaction, oldAmount int, oldType int) error
    DeleteTransaction(id, userID uint) error
    DeleteTransactionWithBalanceRollback(id, userID uint) error
}

type transactionService struct {
    transactionRepo repositories.TransactionRepository
    assetRepo       repositories.AssetRepository
}

func NewTransactionService(transactionRepo repositories.TransactionRepository, assetRepo repositories.AssetRepository) TransactionService {
    return &transactionService{
        transactionRepo: transactionRepo,
        assetRepo:       assetRepo,
    }
}

func (s *transactionService) GetTransactions(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID, assetID *uint64) ([]dto.TransactionResponse, *dto.PaginationResponse, error) {
    transactions, total, err := s.transactionRepo.GetAll(userID, page, limit, startDate, endDate, transactionType, categoryID, bankID, assetID)
    if err != nil {
        return nil, nil, err
    }

    // Convert to DTO with asset information
    transactionResponses := make([]dto.TransactionResponse, len(transactions))
    for i, t := range transactions {
        assetName := ""
        assetBalance := 0.0
        assetCurrency := ""
        
        if t.Asset.ID != 0 {
            assetName = t.Asset.Name
            assetBalance = t.Asset.Balance
            assetCurrency = t.Asset.Currency
        }
        
        transactionResponses[i] = dto.TransactionResponse{
            ID:              t.ID,
            Description:     t.Description,
            Amount:          t.Amount,
            TransactionType: t.TransactionType,
            Date:            t.Date,
            CategoryName:    t.Category.CategoryName,
            BankName:        t.Bank.BankName,
            AssetName:       assetName,
            AssetBalance:    assetBalance,
            AssetCurrency:   assetCurrency,
        }
    }

    // Create pagination response
    totalPages := int(total) / limit
    if int(total)%limit != 0 {
        totalPages++
    }

    pagination := &dto.PaginationResponse{
        Page:       page,
        PageSize:   limit,
        TotalItems: total,
        TotalPages: totalPages,
    }

    return transactionResponses, pagination, nil
}

func (s *transactionService) GetTransactionByID(id, userID uint) (*dto.TransactionResponse, error) {
    transaction, err := s.transactionRepo.GetByIDWithAsset(id, userID)
    if err != nil {
        return nil, err
    }

    assetName := ""
    assetBalance := 0.0
    assetCurrency := ""
    
    if transaction.Asset.ID != 0 {
        assetName = transaction.Asset.Name
        assetBalance = transaction.Asset.Balance
        assetCurrency = transaction.Asset.Currency
    }

    response := &dto.TransactionResponse{
        ID:              transaction.ID,
        Description:     transaction.Description,
        Amount:          transaction.Amount,
        TransactionType: transaction.TransactionType,
        Date:            transaction.Date,
        CategoryName:    transaction.Category.CategoryName,
        BankName:        transaction.Bank.BankName,
        AssetName:       assetName,
        AssetBalance:    assetBalance,
        AssetCurrency:   assetCurrency,
    }

    return response, nil
}

func (s *transactionService) CreateTransactionWithBalanceUpdate(transaction *models.Transaction) error {
    // Validate asset ownership
    asset, err := s.assetRepo.GetByID(transaction.AssetID)
    if err != nil {
        return errors.New("asset not found")
    }
    
    if asset.UserID != uint64(transaction.UserID) {
        return errors.New("unauthorized: asset does not belong to user")
    }

    // Create transaction with automatic balance update
    return s.transactionRepo.CreateWithBalanceUpdate(transaction)
}

func (s *transactionService) UpdateTransactionWithBalanceUpdate(transaction *models.Transaction, oldAmount int, oldType int) error {
    // Validate asset ownership
    asset, err := s.assetRepo.GetByID(transaction.AssetID)
    if err != nil {
        return errors.New("asset not found")
    }
    
    if asset.UserID != uint64(transaction.UserID) {
        return errors.New("unauthorized: asset does not belong to user")
    }

    // Update transaction with balance recalculation
    return s.transactionRepo.UpdateWithBalanceUpdate(transaction, oldAmount, oldType)
}

func (s *transactionService) DeleteTransactionWithBalanceRollback(id, userID uint) error {
    // Delete transaction and rollback balance
    return s.transactionRepo.DeleteWithBalanceRollback(id, userID)
}

// Keep old methods for backward compatibility if needed
func (s *transactionService) CreateTransaction(transaction *models.Transaction) error {
    return s.transactionRepo.Create(transaction)
}

func (s *transactionService) UpdateTransaction(transaction *models.Transaction) error {
    return s.transactionRepo.Update(transaction)
}

func (s *transactionService) DeleteTransaction(id, userID uint) error {
    return s.transactionRepo.Delete(id, userID)
}
```

---

### Step 6: Update Transaction Controller

**File**: `controllers/transaction_controller.go`

```go
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
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    // Parse optional filters
    var startDate, endDate *time.Time
    var transactionType *int
    var categoryID *uint
    var bankID *uint
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

    if assetIDStr := c.Query("asset_id"); assetIDStr != "" {
        if aID, err := strconv.ParseUint(assetIDStr, 10, 64); err == nil {
            assetID = &aID
        }
    }

    transactions, pagination, err := ctrl.transactionService.GetTransactions(
        userIDUint, page, limit, startDate, endDate, transactionType, categoryID, bankID, assetID,
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

func (ctrl *TransactionController) CreateTransaction(c *gin.Context) {
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

    var transaction models.Transaction

    // Bind JSON
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid request payload")
        return
    }

    // Set the UserID from the token
    transaction.UserID = userIDUint

    // Parse fields
    transaction.Description = payload["Description"].(string)
    transaction.CategoryID = uint(payload["CategoryID"].(float64))
    transaction.AssetID = uint64(payload["AssetID"].(float64))
    
    // Convert bank_id (optional for backward compatibility)
    if bankID, ok := payload["BankID"]; ok {
        transaction.BankID = uint(bankID.(float64))
    }
    
    // Convert amount
    if amountStr, ok := payload["Amount"].(string); ok {
        amount, err := strconv.Atoi(amountStr)
        if err != nil {
            utils.JSONError(c, http.StatusBadRequest, "Invalid amount format")
            return
        }
        transaction.Amount = amount
    } else if amount, ok := payload["Amount"].(float64); ok {
        transaction.Amount = int(amount)
    }

    // Handle TransactionType
    if txTypeStr, ok := payload["TransactionType"].(string); ok {
        if txTypeStr == "Income" {
            transaction.TransactionType = 1
        } else if txTypeStr == "Expense" {
            transaction.TransactionType = 2
        } else {
            utils.JSONError(c, http.StatusBadRequest, "Invalid transaction type. Must be 'Income' or 'Expense'")
            return
        }
    } else if txType, ok := payload["TransactionType"].(float64); ok {
        transaction.TransactionType = int(txType)
    }

    // Parse date
    dateStr := payload["Date"].(string)
    var date utils.CustomTime
    var err error
    
    parsed, err := time.Parse("2006-01-02 15:04:05", dateStr)
    if err != nil {
        parsed, err = time.Parse(time.RFC3339, dateStr)
        if err != nil {
            parsed, err = time.Parse("2006-01-02", dateStr)
            if err != nil {
                utils.JSONError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD HH:MM:SS, ISO 8601 or YYYY-MM-DD")
                return
            }
        }
    }

    date.Time = parsed
    transaction.Date = date

    // Create transaction with balance update
    if err := ctrl.transactionService.CreateTransactionWithBalanceUpdate(&transaction); err != nil {
        if err.Error() == "insufficient balance" {
            utils.JSONError(c, http.StatusBadRequest, "Insufficient balance in the selected asset")
            return
        }
        if err.Error() == "unauthorized: asset does not belong to user" {
            utils.JSONError(c, http.StatusForbidden, "Asset does not belong to you")
            return
        }
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

    if err := ctrl.transactionService.DeleteTransactionWithBalanceRollback(uint(id), userIDUint); err != nil {
        utils.JSONError(c, http.StatusNotFound, "Transaction not found or unauthorized")
        return
    }

    utils.JSONSuccess(c, "Transaction deleted successfully", nil)
}
```

---

### Step 7: Update Routes (if needed)

**File**: `routes/routes.go`

No major changes needed if using existing routes. Ensure controller is properly initialized:

```go
func SetupRoutes(r *gin.Engine) {
    // ... existing code ...
    
    // Initialize services with asset repo
    transactionService := services.NewTransactionService(
        repositories.NewTransactionRepository(config.DB),
        repositories.NewAssetRepository(config.DB),
    )
    
    transactionController := controllers.NewTransactionController(
        transactionService,
        budgetService,
    )
    
    // Transaction routes
    api.GET("/transactions", transactionController.GetTransactions)
    api.GET("/transactions/:id", transactionController.GetTransactionByID)
    api.POST("/transactions", transactionController.CreateTransaction)
    api.DELETE("/transactions/:id", transactionController.DeleteTransaction)
}
```

---

### Step 8: Add New API Endpoints (Optional but Recommended)

**File**: `controllers/assets_controller.go`

Add method to get transactions for a specific asset:

```go
func (ctrl *AssetsController) GetAssetTransactions(c *gin.Context) {
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

    assetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid asset ID")
        return
    }

    // Verify asset ownership
    asset, err := ctrl.assetService.GetAsset(userIDUint, uint(assetID))
    if err != nil {
        utils.JSONError(c, http.StatusNotFound, "Asset not found or unauthorized")
        return
    }

    // Get pagination parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

    // Get transactions for this asset
    transactions, pagination, err := ctrl.transactionService.GetTransactions(
        userIDUint, page, limit, nil, nil, nil, nil, nil, &assetID,
    )

    if err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to fetch transactions")
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Asset transactions fetched successfully",
        "data": gin.H{
            "asset":        asset,
            "transactions": transactions,
            "pagination":   pagination,
        },
    })
}
```

---

## Migration Execution

### 1. Create Migration Script

**File**: `migrations/add_asset_id_to_transactions.go`

```go
package main

import (
    "log"
    "my-api/config"
    "my-api/models"
)

func main() {
    // Initialize database connection
    db, err := config.ConnectDB()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Run migration
    if err := models.AddAssetIDToTransactions(db); err != nil {
        log.Fatal("Migration failed:", err)
    }

    log.Println("Migration completed successfully")
}
```

### 2. Run Migration

```bash
go run migrations/add_asset_id_to_transactions.go
```

---

## Testing

### Unit Test Example

**File**: `services/transaction_service_test.go`

```go
package services

import (
    "testing"
    "my-api/models"
    "github.com/stretchr/testify/assert"
)

func TestCreateTransactionWithBalanceUpdate(t *testing.T) {
    // Setup test database
    // Create test asset with initial balance
    asset := &models.Asset{
        UserID:    1,
        Name:      "Test Wallet",
        Balance:   1000.00,
        Currency:  "USD",
    }
    assetRepo.CreateAsset(asset)

    service := NewTransactionService(transactionRepo, assetRepo)

    // Test income transaction
    tx := &models.Transaction{
        UserID:          1,
        AssetID:         asset.ID,
        Description:     "Test Income",
        Amount:          500,
        TransactionType: 1,
    }

    err := service.CreateTransactionWithBalanceUpdate(tx)
    assert.NoError(t, err)
    
    // Verify balance increased
    updatedAsset, _ := assetRepo.GetByID(asset.ID)
    assert.Equal(t, 1500.00, updatedAsset.Balance)

    // Test expense transaction
    tx2 := &models.Transaction{
        UserID:          1,
        AssetID:         asset.ID,
        Description:     "Test Expense",
        Amount:          200,
        TransactionType: 2,
    }

    err = service.CreateTransactionWithBalanceUpdate(tx2)
    assert.NoError(t, err)
    
    // Verify balance decreased
    updatedAsset, _ = assetRepo.GetByID(asset.ID)
    assert.Equal(t, 1300.00, updatedAsset.Balance)
}

func TestInsufficientBalance(t *testing.T) {
    asset := &models.Asset{
        UserID:    1,
        Name:      "Low Balance Wallet",
        Balance:   50.00,
        Currency:  "USD",
    }
    assetRepo.CreateAsset(asset)

    service := NewTransactionService(transactionRepo, assetRepo)

    tx := &models.Transaction{
        UserID:          1,
        AssetID:         asset.ID,
        Description:     "Overdraft",
        Amount:          100,
        TransactionType: 2,
    }

    err := service.CreateTransactionWithBalanceUpdate(tx)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient balance")
}

func TestDeleteTransactionWithRollback(t *testing.T) {
    // Create asset and transaction
    asset := &models.Asset{
        UserID:    1,
        Name:      "Test Wallet",
        Balance:   1000.00,
        Currency:  "USD",
    }
    assetRepo.CreateAsset(asset)

    tx := &models.Transaction{
        UserID:          1,
        AssetID:         asset.ID,
        Description:     "Test",
        Amount:          100,
        TransactionType: 1,
    }

    service := NewTransactionService(transactionRepo, assetRepo)
    service.CreateTransactionWithBalanceUpdate(tx)

    // Delete transaction
    err := service.DeleteTransactionWithBalanceRollback(tx.ID, 1)
    assert.NoError(t, err)
    
    // Verify balance rolled back
    updatedAsset, _ := assetRepo.GetByID(asset.ID)
    assert.Equal(t, 1000.00, updatedAsset.Balance)
}
```

---

## API Usage Examples

### Create Transaction with Asset

```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "Description": "Salary Deposit",
    "Amount": "3000",
    "TransactionType": "Income",
    "Date": "2025-01-15",
    "CategoryID": 3,
    "AssetID": 1
  }'
```

### Get Transactions with Asset Filter

```bash
curl -X GET "http://localhost:8080/api/transactions?asset_id=1&page=1&limit=20" \
  -H "Authorization: Bearer <token>"
```

### Get Transactions for Specific Asset

```bash
curl -X GET "http://localhost:8080/api/assets/1/transactions?page=1&limit=50" \
  -H "Authorization: Bearer <token>"
```

---

## Benefits

### Automatic Balance Management
- Asset balance updates automatically on transaction creation
- Balance rolls back when transactions are deleted
- Prevents overdrafts with balance validation

### Better Data Integrity
- Transaction-asset relationship enforced at database level
- Atomic operations prevent partial updates
- Row locking prevents race conditions

### Enhanced Reporting
- Filter transactions by asset
- See asset-specific transaction history
- Track balance changes over time

---

## Rollback Plan

If issues arise after implementation:

1. **Disable automatic balance updates**: Use old service methods without balance updates
2. **Data correction**: Write script to recalculate balances from transaction history
3. **Rollback migration**: Remove `asset_id` column from transactions table

```sql
-- Rollback migration
ALTER TABLE transactions DROP FOREIGN KEY fk_transactions_asset;
ALTER TABLE transactions DROP INDEX idx_transactions_asset_id;
ALTER TABLE transactions DROP COLUMN asset_id;
```

---

## Next Steps

After completing this integration:

1. Monitor balance accuracy for 1-2 weeks
2. Collect user feedback on automatic balance updates
3. Consider implementing additional features from enhancement plan:
   - Asset transfers
   - Recurring transactions
   - Per-asset budgets
