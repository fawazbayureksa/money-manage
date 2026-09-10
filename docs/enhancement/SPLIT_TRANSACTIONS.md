# Split Transactions Implementation Plan

## Problem Statement

A single purchase often spans multiple categories:
- Supermarket trip: groceries + household items + snacks
- Online shopping: personal items + gifts + office supplies
- Restaurant bill: split between food and entertainment

Without split transactions, users must either:
- Assign everything to one category (inaccurate budgeting)
- Create multiple manual transactions (tedious)

---

## Solution Overview

Allow a single transaction to be split across multiple categories, maintaining accurate budget tracking while keeping the user experience simple.

---

## Database Schema

### Table: `transaction_splits`

```sql
CREATE TABLE transaction_splits (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    transaction_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    amount INT NOT NULL,
    description VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_transaction (transaction_id),
    INDEX idx_category (category_id),
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

### Modify `transactions_v2` table

```sql
ALTER TABLE transactions_v2 
ADD COLUMN is_split BOOLEAN DEFAULT false AFTER transaction_type;
```

---

## API Endpoints

### Create Split Transaction
```http
POST /api/v2/transactions
Authorization: Bearer {token}

{
    "description": "Supermarket Shopping",
    "total_amount": 500000,
    "transaction_type": "Expense",
    "asset_id": 1,
    "date": "2026-02-11",
    "is_split": true,
    "splits": [
        {
            "category_id": 5,
            "amount": 300000,
            "description": "Groceries"
        },
        {
            "category_id": 8,
            "amount": 150000,
            "description": "Household items"
        },
        {
            "category_id": 12,
            "amount": 50000,
            "description": "Snacks"
        }
    ]
}
```

### Response
```json
{
    "success": true,
    "message": "Transaction created successfully",
    "data": {
        "id": 123,
        "description": "Supermarket Shopping",
        "amount": 500000,
        "transaction_type": "Expense",
        "is_split": true,
        "asset": {
            "id": 1,
            "name": "BCA Checking"
        },
        "date": "2026-02-11",
        "splits": [
            {
                "id": 1,
                "category": {"id": 5, "name": "Groceries"},
                "amount": 300000,
                "description": "Groceries"
            },
            {
                "id": 2,
                "category": {"id": 8, "name": "Household"},
                "amount": 150000,
                "description": "Household items"
            },
            {
                "id": 3,
                "category": {"id": 12, "name": "Food & Snacks"},
                "amount": 50000,
                "description": "Snacks"
            }
        ]
    }
}
```

### Update Split Transaction
```http
PUT /api/v2/transactions/{id}
Authorization: Bearer {token}

{
    "splits": [
        {
            "category_id": 5,
            "amount": 280000,
            "description": "Groceries"
        },
        {
            "category_id": 8,
            "amount": 170000,
            "description": "Household items"
        },
        {
            "category_id": 12,
            "amount": 50000,
            "description": "Snacks"
        }
    ]
}
```

### Convert Regular Transaction to Split
```http
POST /api/v2/transactions/{id}/split
Authorization: Bearer {token}

{
    "splits": [
        {
            "category_id": 5,
            "amount": 300000
        },
        {
            "category_id": 8,
            "amount": 200000
        }
    ]
}
```

---

## Go Implementation

### Model Updates

```go
// models/transaction_v2.go

type TransactionV2 struct {
    ID              uint                `gorm:"primaryKey" json:"id"`
    UserID          uint                `gorm:"not null;index" json:"user_id"`
    Description     string              `gorm:"size:255" json:"description"`
    Amount          int                 `gorm:"not null" json:"amount"`
    TransactionType int                 `gorm:"not null" json:"transaction_type"`
    IsSplit         bool                `gorm:"default:false" json:"is_split"`
    CategoryID      uint                `json:"category_id,omitempty"` // Primary category (for non-split)
    AssetID         uint64              `gorm:"not null" json:"asset_id"`
    Date            utils.CustomTime    `gorm:"type:date" json:"date"`
    // ... other fields
    
    // Relations
    Category        *Category           `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Splits          []TransactionSplit  `gorm:"foreignKey:TransactionID" json:"splits,omitempty"`
}

type TransactionSplit struct {
    ID            uint            `gorm:"primaryKey" json:"id"`
    TransactionID uint            `gorm:"not null;index" json:"transaction_id"`
    CategoryID    uint            `gorm:"not null" json:"category_id"`
    Amount        int             `gorm:"not null" json:"amount"`
    Description   string          `gorm:"size:255" json:"description,omitempty"`
    CreatedAt     utils.CustomTime `json:"created_at"`
    
    // Relations
    Category      Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}
```

### Service Updates

```go
// services/transaction_v2_service.go

func (s *transactionV2Service) CreateTransaction(tx *models.TransactionV2) error {
    // Validate splits sum equals total amount
    if tx.IsSplit {
        if len(tx.Splits) < 2 {
            return errors.New("split transaction must have at least 2 splits")
        }
        
        splitTotal := 0
        for _, split := range tx.Splits {
            if split.Amount <= 0 {
                return errors.New("split amounts must be positive")
            }
            splitTotal += split.Amount
        }
        
        if splitTotal != tx.Amount {
            return fmt.Errorf("splits total (%d) must equal transaction amount (%d)", splitTotal, tx.Amount)
        }
        
        // Clear single category since it's split
        tx.CategoryID = 0
    }
    
    // ... existing balance logic ...
    
    // Use transaction for atomicity
    return s.db.Transaction(func(dbTx *gorm.DB) error {
        if err := dbTx.Create(tx).Error; err != nil {
            return err
        }
        
        // Create splits
        if tx.IsSplit {
            for i := range tx.Splits {
                tx.Splits[i].TransactionID = tx.ID
                if err := dbTx.Create(&tx.Splits[i]).Error; err != nil {
                    return err
                }
            }
        }
        
        return nil
    })
}
```

### Budget Integration

Split transactions need to affect budgets for EACH category:

```go
// services/budget_service.go

func (s *budgetService) CheckBudgetAlertsForTransaction(userID uint, tx *models.TransactionV2) error {
    if tx.TransactionType != 2 { // Only expenses
        return nil
    }
    
    if tx.IsSplit {
        // Check each split category
        for _, split := range tx.Splits {
            s.checkCategoryBudget(userID, split.CategoryID, split.Amount)
        }
    } else {
        s.checkCategoryBudget(userID, tx.CategoryID, tx.Amount)
    }
    
    return nil
}
```

### Analytics Integration

```go
// repositories/analytics_repository.go

func (r *analyticsRepository) GetSpendingByCategory(userID uint, startDate, endDate time.Time) ([]CategorySpending, error) {
    var results []CategorySpending
    
    // Include both regular transactions and split amounts
    err := r.db.Raw(`
        SELECT 
            c.id as category_id,
            c.category_name,
            COALESCE(regular.amount, 0) + COALESCE(splits.amount, 0) as total_amount
        FROM categories c
        LEFT JOIN (
            SELECT category_id, SUM(amount) as amount
            FROM transactions_v2
            WHERE user_id = ? 
              AND transaction_type = 2 
              AND is_split = false
              AND date BETWEEN ? AND ?
            GROUP BY category_id
        ) regular ON c.id = regular.category_id
        LEFT JOIN (
            SELECT ts.category_id, SUM(ts.amount) as amount
            FROM transaction_splits ts
            JOIN transactions_v2 t ON ts.transaction_id = t.id
            WHERE t.user_id = ? 
              AND t.transaction_type = 2
              AND t.date BETWEEN ? AND ?
            GROUP BY ts.category_id
        ) splits ON c.id = splits.category_id
        WHERE COALESCE(regular.amount, 0) + COALESCE(splits.amount, 0) > 0
        ORDER BY total_amount DESC
    `, userID, startDate, endDate, userID, startDate, endDate).Scan(&results).Error
    
    return results, err
}
```

---

## Frontend Integration Guide

### Split Transaction Entry

```
┌─────────────────────────────────────────┐
│  ← New Transaction                      │
├─────────────────────────────────────────┤
│                                         │
│  Description                            │
│  ┌─────────────────────────────────┐   │
│  │ Supermarket Shopping            │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Total Amount                           │
│  ┌─────────────────────────────────┐   │
│  │ Rp 500,000                      │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ☑️ Split across categories             │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  SPLITS                                 │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🛒 Groceries                 ✕  │   │
│  │ Rp 300,000                      │   │
│  │ Note: Weekly groceries          │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🏠 Household                 ✕  │   │
│  │ Rp 150,000                      │   │
│  │ Note: Cleaning supplies         │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🍿 Snacks                    ✕  │   │
│  │ Rp 50,000                       │   │
│  └─────────────────────────────────┘   │
│                                         │
│  [ + Add Split ]                        │
│                                         │
│  ─────────────────────────────────────  │
│  Total: Rp 500,000 ✅ Balanced          │
│  ─────────────────────────────────────  │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │           💾 Save               │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Transaction Detail with Splits

```
┌─────────────────────────────────────────┐
│  ← Transaction Details              ⋮   │
├─────────────────────────────────────────┤
│                                         │
│  📦 Supermarket Shopping                │
│                                         │
│  -Rp 500,000                            │
│  Feb 11, 2026                           │
│  BCA Checking                           │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  SPLIT BREAKDOWN                        │
│                                         │
│  🛒 Groceries                           │
│  Rp 300,000 (60%)                       │
│  Weekly groceries                       │
│                                         │
│  🏠 Household                           │
│  Rp 150,000 (30%)                       │
│  Cleaning supplies                      │
│                                         │
│  🍿 Snacks                              │
│  Rp 50,000 (10%)                        │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  [ Edit ] [ Delete ]                    │
└─────────────────────────────────────────┘
```

---

## Migration Files

```sql
-- db/migrations/YYYYMMDDHHMMSS_add_split_transactions.up.sql

ALTER TABLE transactions_v2 ADD COLUMN is_split BOOLEAN DEFAULT false AFTER transaction_type;

CREATE TABLE transaction_splits (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    transaction_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    amount INT NOT NULL,
    description VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_transaction (transaction_id),
    INDEX idx_category (category_id),
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

---

## Validation Rules

1. Split amounts must be positive integers
2. Sum of splits must equal transaction amount exactly
3. Minimum 2 splits for a split transaction
4. Each split must have a valid category
5. Converting: existing category becomes first split with full amount
