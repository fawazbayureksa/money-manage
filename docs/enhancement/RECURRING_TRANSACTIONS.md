# Recurring Transactions Implementation Plan

## Problem Statement

Users have to manually enter the same transactions every month:
- Monthly salary
- Rent/mortgage payments
- Utility bills (electricity, water, internet)
- Subscription services (Netflix, Spotify, gym membership)
- Insurance premiums
- Loan payments

This is tedious and often forgotten, leading to incomplete financial records.

---

## Solution Overview

Automatically create transactions based on user-defined schedules. The system will:
1. Allow users to define recurring transaction templates
2. Automatically generate actual transactions on scheduled dates
3. Send reminders before upcoming debits
4. Handle edge cases (weekends, month variations)

---

## Database Schema

### Table: `recurring_transactions`

```sql
CREATE TABLE recurring_transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Transaction template
    description VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    transaction_type INT NOT NULL COMMENT '1=Income, 2=Expense',
    category_id BIGINT UNSIGNED NOT NULL,
    asset_id BIGINT UNSIGNED NOT NULL,
    
    -- Schedule configuration
    frequency ENUM('daily', 'weekly', 'biweekly', 'monthly', 'quarterly', 'yearly') NOT NULL,
    day_of_week INT NULL COMMENT '0-6 for weekly (0=Sunday)',
    day_of_month INT NULL COMMENT '1-31 for monthly',
    month_of_year INT NULL COMMENT '1-12 for yearly',
    
    -- Date boundaries
    start_date DATE NOT NULL,
    end_date DATE NULL COMMENT 'NULL for indefinite',
    next_occurrence DATE NOT NULL,
    last_generated DATE NULL,
    
    -- Options
    is_active BOOLEAN DEFAULT true,
    auto_create BOOLEAN DEFAULT true COMMENT 'Auto-create or just remind',
    reminder_days_before INT DEFAULT 3 COMMENT 'Days before to send reminder',
    skip_weekends BOOLEAN DEFAULT false COMMENT 'Move to next business day',
    
    -- Metadata
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Foreign keys
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id),
    FOREIGN KEY (asset_id) REFERENCES assets(id),
    
    INDEX idx_next_occurrence (next_occurrence),
    INDEX idx_user_active (user_id, is_active)
);
```

### Table: `recurring_transaction_logs`

```sql
CREATE TABLE recurring_transaction_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    recurring_transaction_id BIGINT UNSIGNED NOT NULL,
    transaction_id BIGINT UNSIGNED NULL COMMENT 'Generated transaction ID',
    scheduled_date DATE NOT NULL,
    status ENUM('pending', 'created', 'skipped', 'failed') NOT NULL,
    error_message TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (recurring_transaction_id) REFERENCES recurring_transactions(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE SET NULL
);
```

---

## API Endpoints

### Create Recurring Transaction
```http
POST /api/v2/recurring-transactions
Authorization: Bearer {token}

{
    "description": "Monthly Salary",
    "amount": 10000000,
    "transaction_type": "Income",
    "category_id": 1,
    "asset_id": 1,
    "frequency": "monthly",
    "day_of_month": 25,
    "start_date": "2026-02-25",
    "end_date": null,
    "auto_create": true,
    "reminder_days_before": 0,
    "notes": "Company XYZ salary"
}
```

### Response
```json
{
    "success": true,
    "message": "Recurring transaction created successfully",
    "data": {
        "id": 1,
        "description": "Monthly Salary",
        "amount": 10000000,
        "transaction_type": "Income",
        "category": {
            "id": 1,
            "name": "Salary"
        },
        "asset": {
            "id": 1,
            "name": "BCA Checking"
        },
        "frequency": "monthly",
        "day_of_month": 25,
        "start_date": "2026-02-25",
        "end_date": null,
        "next_occurrence": "2026-02-25",
        "is_active": true,
        "auto_create": true,
        "reminder_days_before": 0
    }
}
```

### List Recurring Transactions
```http
GET /api/v2/recurring-transactions?is_active=true&frequency=monthly
Authorization: Bearer {token}
```

### Get Single Recurring Transaction
```http
GET /api/v2/recurring-transactions/{id}
Authorization: Bearer {token}
```

### Update Recurring Transaction
```http
PUT /api/v2/recurring-transactions/{id}
Authorization: Bearer {token}

{
    "amount": 11000000,
    "description": "Monthly Salary (Raised)"
}
```

### Pause/Resume Recurring Transaction
```http
PUT /api/v2/recurring-transactions/{id}/toggle
Authorization: Bearer {token}

{
    "is_active": false
}
```

### Delete Recurring Transaction
```http
DELETE /api/v2/recurring-transactions/{id}
Authorization: Bearer {token}
```

### Skip Next Occurrence
```http
POST /api/v2/recurring-transactions/{id}/skip
Authorization: Bearer {token}

{
    "reason": "Salary delayed this month"
}
```

### Get Upcoming Transactions
```http
GET /api/v2/recurring-transactions/upcoming?days=30
Authorization: Bearer {token}
```

---

## Go Implementation

### Model: `models/recurring_transaction.go`

```go
package models

import (
    "my-api/utils"
    "gorm.io/gorm"
)

type RecurringTransaction struct {
    ID                uint            `gorm:"primaryKey" json:"id"`
    UserID            uint            `gorm:"not null;index" json:"user_id"`
    Description       string          `gorm:"size:255;not null" json:"description"`
    Amount            int             `gorm:"not null" json:"amount"`
    TransactionType   int             `gorm:"not null" json:"transaction_type"`
    CategoryID        uint            `gorm:"not null" json:"category_id"`
    AssetID           uint64          `gorm:"not null" json:"asset_id"`
    
    Frequency         string          `gorm:"type:enum('daily','weekly','biweekly','monthly','quarterly','yearly');not null" json:"frequency"`
    DayOfWeek         *int            `json:"day_of_week,omitempty"`
    DayOfMonth        *int            `json:"day_of_month,omitempty"`
    MonthOfYear       *int            `json:"month_of_year,omitempty"`
    
    StartDate         utils.CustomTime `gorm:"type:date;not null" json:"start_date"`
    EndDate           *utils.CustomTime `gorm:"type:date" json:"end_date,omitempty"`
    NextOccurrence    utils.CustomTime `gorm:"type:date;not null;index" json:"next_occurrence"`
    LastGenerated     *utils.CustomTime `gorm:"type:date" json:"last_generated,omitempty"`
    
    IsActive          bool            `gorm:"default:true" json:"is_active"`
    AutoCreate        bool            `gorm:"default:true" json:"auto_create"`
    ReminderDaysBefore int            `gorm:"default:3" json:"reminder_days_before"`
    SkipWeekends      bool            `gorm:"default:false" json:"skip_weekends"`
    
    Notes             string          `json:"notes,omitempty"`
    
    CreatedAt         utils.CustomTime `json:"created_at"`
    UpdatedAt         utils.CustomTime `json:"updated_at"`
    DeletedAt         gorm.DeletedAt  `gorm:"index" json:"-"`
    
    // Relations
    User              User            `gorm:"foreignKey:UserID" json:"-"`
    Category          Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Asset             Asset           `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

type RecurringTransactionLog struct {
    ID                     uint            `gorm:"primaryKey" json:"id"`
    RecurringTransactionID uint            `gorm:"not null;index" json:"recurring_transaction_id"`
    TransactionID          *uint           `json:"transaction_id,omitempty"`
    ScheduledDate          utils.CustomTime `gorm:"type:date;not null" json:"scheduled_date"`
    Status                 string          `gorm:"type:enum('pending','created','skipped','failed');not null" json:"status"`
    ErrorMessage           string          `json:"error_message,omitempty"`
    CreatedAt              utils.CustomTime `json:"created_at"`
}
```

### Service: `services/recurring_transaction_service.go`

```go
package services

import (
    "errors"
    "my-api/models"
    "my-api/repositories"
    "my-api/utils"
    "time"
)

type RecurringTransactionService interface {
    Create(userID uint, req *CreateRecurringTransactionRequest) (*models.RecurringTransaction, error)
    GetByID(id, userID uint) (*models.RecurringTransaction, error)
    GetAll(userID uint, filter *RecurringTransactionFilter) ([]models.RecurringTransaction, int64, error)
    Update(id, userID uint, req *UpdateRecurringTransactionRequest) (*models.RecurringTransaction, error)
    Delete(id, userID uint) error
    Toggle(id, userID uint, isActive bool) error
    SkipNext(id, userID uint, reason string) error
    GetUpcoming(userID uint, days int) ([]UpcomingTransaction, error)
    ProcessDueTransactions() error
}

type recurringTransactionService struct {
    repo        repositories.RecurringTransactionRepository
    txService   TransactionV2Service
    assetRepo   repositories.AssetRepository
}

func (s *recurringTransactionService) ProcessDueTransactions() error {
    today := time.Now().Truncate(24 * time.Hour)
    
    // Get all active recurring transactions due today or earlier
    dueTransactions, err := s.repo.FindDue(today)
    if err != nil {
        return err
    }
    
    for _, rt := range dueTransactions {
        if rt.AutoCreate {
            // Create the actual transaction
            tx := &models.TransactionV2{
                UserID:          rt.UserID,
                Description:     rt.Description,
                Amount:          rt.Amount,
                TransactionType: rt.TransactionType,
                CategoryID:      rt.CategoryID,
                AssetID:         rt.AssetID,
                Date:            utils.CustomTime{Time: today},
            }
            
            err := s.txService.CreateTransaction(tx)
            status := "created"
            var errorMsg string
            
            if err != nil {
                status = "failed"
                errorMsg = err.Error()
            }
            
            // Log the result
            s.repo.CreateLog(&models.RecurringTransactionLog{
                RecurringTransactionID: rt.ID,
                TransactionID:          &tx.ID,
                ScheduledDate:          rt.NextOccurrence,
                Status:                 status,
                ErrorMessage:           errorMsg,
            })
        }
        
        // Calculate and update next occurrence
        nextDate := s.calculateNextOccurrence(rt)
        s.repo.UpdateNextOccurrence(rt.ID, today, nextDate)
    }
    
    return nil
}

func (s *recurringTransactionService) calculateNextOccurrence(rt models.RecurringTransaction) time.Time {
    current := rt.NextOccurrence.Time
    
    switch rt.Frequency {
    case "daily":
        return current.AddDate(0, 0, 1)
    case "weekly":
        return current.AddDate(0, 0, 7)
    case "biweekly":
        return current.AddDate(0, 0, 14)
    case "monthly":
        return current.AddDate(0, 1, 0)
    case "quarterly":
        return current.AddDate(0, 3, 0)
    case "yearly":
        return current.AddDate(1, 0, 0)
    default:
        return current.AddDate(0, 1, 0)
    }
}
```

### Background Job: `jobs/recurring_transaction_job.go`

```go
package jobs

import (
    "my-api/services"
    "my-api/utils"
    "time"
)

type RecurringTransactionJob struct {
    service services.RecurringTransactionService
}

func NewRecurringTransactionJob(service services.RecurringTransactionService) *RecurringTransactionJob {
    return &RecurringTransactionJob{service: service}
}

// Run executes the job - should be called daily
func (j *RecurringTransactionJob) Run() {
    utils.LogInfo("Starting recurring transaction processing...")
    
    if err := j.service.ProcessDueTransactions(); err != nil {
        utils.LogErrorf("Failed to process recurring transactions: %v", err)
        return
    }
    
    utils.LogInfo("Recurring transaction processing completed")
}

// StartScheduler starts a background scheduler
func (j *RecurringTransactionJob) StartScheduler() {
    go func() {
        // Run immediately on startup
        j.Run()
        
        // Then run daily at midnight
        for {
            now := time.Now()
            next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
            duration := next.Sub(now)
            
            time.Sleep(duration)
            j.Run()
        }
    }()
}
```

---

## Frontend Integration Guide

### Recurring Transaction List Screen

```
┌─────────────────────────────────────────┐
│  Recurring Transactions              ⚙️ │
├─────────────────────────────────────────┤
│                                         │
│  📅 UPCOMING THIS MONTH                 │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 💰 Monthly Salary               │   │
│  │ +Rp 10,000,000 • Feb 25         │   │
│  │ BCA Checking                    │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🏠 Rent Payment                 │   │
│  │ -Rp 3,500,000 • Mar 1           │   │
│  │ BCA Checking                    │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📺 Netflix Subscription         │   │
│  │ -Rp 186,000 • Mar 5             │   │
│  │ Credit Card                     │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  🔄 ALL RECURRING (5)                   │
│                                         │
│  [+ Add New]                            │
└─────────────────────────────────────────┘
```

### Create/Edit Form

```
┌─────────────────────────────────────────┐
│  ← New Recurring Transaction            │
├─────────────────────────────────────────┤
│                                         │
│  Description                            │
│  ┌─────────────────────────────────┐   │
│  │ Netflix Subscription            │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Amount                                 │
│  ┌─────────────────────────────────┐   │
│  │ Rp 186,000                      │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Type                                   │
│  ┌────────────┐ ┌────────────┐         │
│  │  Income    │ │  Expense ✓ │         │
│  └────────────┘ └────────────┘         │
│                                         │
│  Frequency                              │
│  ┌─────────────────────────────────┐   │
│  │ Monthly                      ▼  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Day of Month                           │
│  ┌─────────────────────────────────┐   │
│  │ 5                               │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Category                               │
│  ┌─────────────────────────────────┐   │
│  │ Entertainment               ▼   │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Wallet                                 │
│  ┌─────────────────────────────────┐   │
│  │ Credit Card                 ▼   │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ☑️ Auto-create transaction             │
│  ☑️ Remind me 3 days before             │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │           💾 Save               │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

---

## Edge Cases to Handle

1. **Month-end dates**: If day_of_month is 31 but month only has 30 days, use last day of month
2. **Weekend handling**: Option to skip to next business day
3. **Insufficient balance**: Log as failed, notify user, don't create transaction
4. **Paused transactions**: Don't generate but track for resumption
5. **End date reached**: Automatically deactivate and notify user
6. **Backdated start**: If start_date is in current period, evaluate if should generate

---

## Migration File

```sql
-- db/migrations/YYYYMMDDHHMMSS_create_recurring_transactions.up.sql

CREATE TABLE recurring_transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    description VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    transaction_type INT NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    asset_id BIGINT UNSIGNED NOT NULL,
    frequency ENUM('daily', 'weekly', 'biweekly', 'monthly', 'quarterly', 'yearly') NOT NULL,
    day_of_week INT NULL,
    day_of_month INT NULL,
    month_of_year INT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    next_occurrence DATE NOT NULL,
    last_generated DATE NULL,
    is_active BOOLEAN DEFAULT true,
    auto_create BOOLEAN DEFAULT true,
    reminder_days_before INT DEFAULT 3,
    skip_weekends BOOLEAN DEFAULT false,
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_next_occurrence (next_occurrence),
    INDEX idx_user_active (user_id, is_active),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id),
    FOREIGN KEY (asset_id) REFERENCES assets(id)
);

CREATE TABLE recurring_transaction_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    recurring_transaction_id BIGINT UNSIGNED NOT NULL,
    transaction_id BIGINT UNSIGNED NULL,
    scheduled_date DATE NOT NULL,
    status ENUM('pending', 'created', 'skipped', 'failed') NOT NULL,
    error_message TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (recurring_transaction_id) REFERENCES recurring_transactions(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE SET NULL
);
```

---

## Testing Scenarios

1. ✅ Create monthly recurring transaction
2. ✅ Create weekly recurring transaction with specific day
3. ✅ Process due transactions (auto-create enabled)
4. ✅ Process due transactions (auto-create disabled - reminder only)
5. ✅ Skip next occurrence
6. ✅ Handle insufficient balance gracefully
7. ✅ Pause and resume recurring transaction
8. ✅ End date reached - auto-deactivate
9. ✅ Month-end date handling (Feb 30 → Feb 28/29)
10. ✅ Weekend skip to next business day
