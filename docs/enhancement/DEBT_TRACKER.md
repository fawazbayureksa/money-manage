# Debt Tracker Implementation Plan

## Problem Statement

Users with multiple debts face significant challenges:
- Hard to track multiple loans/credit cards
- Unclear which debt to pay off first
- Interest accumulates without visibility
- No motivation from seeing progress
- Confusion between different payoff strategies

---

## Solution Overview

Comprehensive debt management system that:
1. Tracks all debts with interest calculations
2. Shows payoff progress and projections
3. Recommends optimal payoff strategies
4. Integrates with transaction tracking
5. Celebrates milestones to maintain motivation

---

## Database Schema

### Table: `debts`

```sql
CREATE TABLE debts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Basic info
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    debt_type ENUM('credit_card', 'personal_loan', 'mortgage', 'car_loan', 'student_loan', 'other') NOT NULL,
    creditor_name VARCHAR(100) NULL,
    
    -- Financial details
    original_amount INT NOT NULL,
    current_balance INT NOT NULL,
    interest_rate DECIMAL(5,2) NOT NULL COMMENT 'Annual interest rate as percentage',
    interest_type ENUM('fixed', 'variable') DEFAULT 'fixed',
    
    -- Payment info
    minimum_payment INT NOT NULL,
    payment_due_day INT NOT NULL COMMENT '1-31',
    current_month_paid BOOLEAN DEFAULT false,
    
    -- Dates
    start_date DATE NOT NULL,
    expected_payoff_date DATE NULL COMMENT 'Based on minimum payments',
    actual_payoff_date DATE NULL,
    
    -- Status
    status ENUM('active', 'paid_off', 'defaulted', 'settled') DEFAULT 'active',
    
    -- Optional: Link to asset for payment tracking
    payment_asset_id BIGINT UNSIGNED NULL,
    
    -- Settings
    include_in_net_worth BOOLEAN DEFAULT true,
    auto_track_interest BOOLEAN DEFAULT true COMMENT 'Auto-add interest as transaction',
    
    -- Metadata
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_user_status (user_id, status),
    INDEX idx_payment_due (payment_due_day, status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (payment_asset_id) REFERENCES assets(id) ON DELETE SET NULL
);
```

### Table: `debt_payments`

```sql
CREATE TABLE debt_payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    debt_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Payment details
    amount INT NOT NULL,
    payment_type ENUM('regular', 'extra', 'interest_only', 'payoff', 'adjustment') NOT NULL,
    
    -- Breakdown
    principal_amount INT NOT NULL DEFAULT 0,
    interest_amount INT NOT NULL DEFAULT 0,
    fees_amount INT NOT NULL DEFAULT 0,
    
    -- Balance tracking
    balance_before INT NOT NULL,
    balance_after INT NOT NULL,
    
    -- References
    transaction_id BIGINT UNSIGNED NULL COMMENT 'Link to actual transaction if tracked',
    
    -- Status
    payment_date DATE NOT NULL,
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_debt_date (debt_id, payment_date),
    FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE SET NULL
);
```

### Table: `debt_milestones`

```sql
CREATE TABLE debt_milestones (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    debt_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    milestone_type ENUM('25_percent', '50_percent', '75_percent', 'paid_off', 'custom') NOT NULL,
    description VARCHAR(255) NULL,
    reached_at TIMESTAMP NULL,
    is_celebrated BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

## API Endpoints

### Create Debt
```http
POST /api/v2/debts
Authorization: Bearer {token}

{
    "name": "BCA Credit Card",
    "debt_type": "credit_card",
    "creditor_name": "Bank BCA",
    "original_amount": 15000000,
    "current_balance": 12000000,
    "interest_rate": 24.0,
    "interest_type": "fixed",
    "minimum_payment": 500000,
    "payment_due_day": 15,
    "start_date": "2025-06-01",
    "payment_asset_id": 1,
    "notes": "Balance transfer from old card"
}
```

### Response
```json
{
    "success": true,
    "message": "Debt added successfully",
    "data": {
        "id": 1,
        "name": "BCA Credit Card",
        "debt_type": "credit_card",
        "original_amount": 15000000,
        "current_balance": 12000000,
        "paid_off_amount": 3000000,
        "paid_off_percentage": 20.0,
        "interest_rate": 24.0,
        "minimum_payment": 500000,
        "payment_due_day": 15,
        "days_until_due": 4,
        "current_month_paid": false,
        "status": "active",
        "projections": {
            "payoff_date_minimum": "2028-06-15",
            "total_interest_minimum": 7200000,
            "months_remaining_minimum": 28
        }
    }
}
```

### List All Debts
```http
GET /api/v2/debts?status=active
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "summary": {
            "total_debt": 27000000,
            "total_minimum_payment": 1500000,
            "total_paid_this_month": 500000,
            "debts_paid_this_month": 1,
            "debts_unpaid_this_month": 2,
            "avg_interest_rate": 18.5,
            "projected_payoff_date": "2028-06-15",
            "total_interest_if_minimum": 12500000
        },
        "debts": [
            {
                "id": 1,
                "name": "BCA Credit Card",
                "debt_type": "credit_card",
                "current_balance": 12000000,
                "interest_rate": 24.0,
                "minimum_payment": 500000,
                "due_day": 15,
                "days_until_due": 4,
                "paid_this_month": false,
                "paid_percentage": 20.0
            },
            {
                "id": 2,
                "name": "Car Loan",
                "debt_type": "car_loan",
                "current_balance": 15000000,
                "interest_rate": 12.0,
                "minimum_payment": 1000000,
                "due_day": 5,
                "days_until_due": 22,
                "paid_this_month": true,
                "paid_percentage": 40.0
            }
        ]
    }
}
```

### Record Payment
```http
POST /api/v2/debts/{id}/payments
Authorization: Bearer {token}

{
    "amount": 1000000,
    "payment_type": "extra",
    "payment_date": "2026-02-11",
    "principal_amount": 900000,
    "interest_amount": 100000,
    "source_asset_id": 1,
    "notes": "Extra payment from bonus"
}
```

### Get Payoff Strategies
```http
GET /api/v2/debts/strategies?extra_payment=500000
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "current_monthly_payment": 1500000,
        "extra_available": 500000,
        "strategies": [
            {
                "name": "Avalanche (Highest Interest First)",
                "description": "Pay minimum on all debts, put extra toward the highest interest debt",
                "payoff_date": "2027-08-15",
                "total_interest": 4800000,
                "interest_saved": 7700000,
                "months_saved": 10,
                "order": [
                    {"debt_id": 1, "name": "BCA Credit Card", "interest_rate": 24.0},
                    {"debt_id": 2, "name": "Car Loan", "interest_rate": 12.0}
                ]
            },
            {
                "name": "Snowball (Smallest Balance First)",
                "description": "Pay minimum on all debts, put extra toward the smallest balance",
                "payoff_date": "2027-10-15",
                "total_interest": 5200000,
                "interest_saved": 7300000,
                "months_saved": 8,
                "order": [
                    {"debt_id": 2, "name": "Car Loan", "balance": 15000000},
                    {"debt_id": 1, "name": "BCA Credit Card", "balance": 12000000}
                ],
                "milestones": [
                    {"date": "2027-02-15", "event": "Car Loan paid off!"}
                ]
            },
            {
                "name": "Minimum Payments Only",
                "description": "Pay only minimum payments on all debts",
                "payoff_date": "2028-06-15",
                "total_interest": 12500000,
                "interest_saved": 0,
                "months_saved": 0
            }
        ],
        "recommendation": {
            "strategy": "Avalanche",
            "reason": "Saves the most money (Rp 7,700,000 in interest) while paying off debts 10 months faster"
        }
    }
}
```

### Get Debt Detail with History
```http
GET /api/v2/debts/{id}?include_payments=true&include_milestones=true
Authorization: Bearer {token}
```

### Update Monthly Balance (Manual Update)
```http
PUT /api/v2/debts/{id}/balance
Authorization: Bearer {token}

{
    "new_balance": 11500000,
    "as_of_date": "2026-02-11"
}
```

### Get Debt Payoff Timeline
```http
GET /api/v2/debts/{id}/timeline?strategy=avalanche&extra_payment=500000
Authorization: Bearer {token}
```

---

## Go Implementation

### Model: `models/debt.go`

```go
package models

import (
    "my-api/utils"
    "gorm.io/gorm"
    "time"
)

type Debt struct {
    ID              uint            `gorm:"primaryKey" json:"id"`
    UserID          uint            `gorm:"not null;index" json:"user_id"`
    Name            string          `gorm:"size:100;not null" json:"name"`
    Description     string          `json:"description,omitempty"`
    DebtType        string          `gorm:"type:enum('credit_card','personal_loan','mortgage','car_loan','student_loan','other');not null" json:"debt_type"`
    CreditorName    string          `gorm:"size:100" json:"creditor_name,omitempty"`
    
    OriginalAmount  int             `gorm:"not null" json:"original_amount"`
    CurrentBalance  int             `gorm:"not null" json:"current_balance"`
    InterestRate    float64         `gorm:"type:decimal(5,2);not null" json:"interest_rate"`
    InterestType    string          `gorm:"type:enum('fixed','variable');default:'fixed'" json:"interest_type"`
    
    MinimumPayment  int             `gorm:"not null" json:"minimum_payment"`
    PaymentDueDay   int             `gorm:"not null" json:"payment_due_day"`
    CurrentMonthPaid bool           `gorm:"default:false" json:"current_month_paid"`
    
    StartDate       utils.CustomTime `gorm:"type:date;not null" json:"start_date"`
    ExpectedPayoffDate *utils.CustomTime `gorm:"type:date" json:"expected_payoff_date,omitempty"`
    ActualPayoffDate *utils.CustomTime `gorm:"type:date" json:"actual_payoff_date,omitempty"`
    
    Status          string          `gorm:"type:enum('active','paid_off','defaulted','settled');default:'active'" json:"status"`
    PaymentAssetID  *uint64         `json:"payment_asset_id,omitempty"`
    
    IncludeInNetWorth bool          `gorm:"default:true" json:"include_in_net_worth"`
    AutoTrackInterest bool          `gorm:"default:true" json:"auto_track_interest"`
    
    Notes           string          `json:"notes,omitempty"`
    CreatedAt       utils.CustomTime `json:"created_at"`
    UpdatedAt       utils.CustomTime `json:"updated_at"`
    DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
    
    // Relations
    Payments        []DebtPayment   `gorm:"foreignKey:DebtID" json:"payments,omitempty"`
    Milestones      []DebtMilestone `gorm:"foreignKey:DebtID" json:"milestones,omitempty"`
}

// Computed fields
func (d *Debt) PaidOffAmount() int {
    return d.OriginalAmount - d.CurrentBalance
}

func (d *Debt) PaidOffPercentage() float64 {
    if d.OriginalAmount == 0 {
        return 0
    }
    return float64(d.PaidOffAmount()) / float64(d.OriginalAmount) * 100
}

func (d *Debt) DaysUntilDue() int {
    now := time.Now()
    dueDate := time.Date(now.Year(), now.Month(), d.PaymentDueDay, 0, 0, 0, 0, now.Location())
    
    // If due date has passed this month, calculate for next month
    if now.After(dueDate) {
        dueDate = dueDate.AddDate(0, 1, 0)
    }
    
    return int(dueDate.Sub(now).Hours() / 24)
}

func (d *Debt) MonthlyInterest() int {
    return int(float64(d.CurrentBalance) * (d.InterestRate / 100 / 12))
}

type DebtPayment struct {
    ID              uint            `gorm:"primaryKey" json:"id"`
    DebtID          uint            `gorm:"not null;index" json:"debt_id"`
    UserID          uint            `gorm:"not null" json:"user_id"`
    Amount          int             `gorm:"not null" json:"amount"`
    PaymentType     string          `gorm:"type:enum('regular','extra','interest_only','payoff','adjustment');not null" json:"payment_type"`
    PrincipalAmount int             `gorm:"default:0" json:"principal_amount"`
    InterestAmount  int             `gorm:"default:0" json:"interest_amount"`
    FeesAmount      int             `gorm:"default:0" json:"fees_amount"`
    BalanceBefore   int             `gorm:"not null" json:"balance_before"`
    BalanceAfter    int             `gorm:"not null" json:"balance_after"`
    TransactionID   *uint           `json:"transaction_id,omitempty"`
    PaymentDate     utils.CustomTime `gorm:"type:date;not null" json:"payment_date"`
    Notes           string          `json:"notes,omitempty"`
    CreatedAt       utils.CustomTime `json:"created_at"`
}

type DebtMilestone struct {
    ID            uint            `gorm:"primaryKey" json:"id"`
    DebtID        uint            `gorm:"not null;index" json:"debt_id"`
    UserID        uint            `gorm:"not null" json:"user_id"`
    MilestoneType string          `gorm:"type:enum('25_percent','50_percent','75_percent','paid_off','custom');not null" json:"milestone_type"`
    Description   string          `gorm:"size:255" json:"description,omitempty"`
    ReachedAt     *time.Time      `json:"reached_at,omitempty"`
    IsCelebrated  bool            `gorm:"default:false" json:"is_celebrated"`
    CreatedAt     utils.CustomTime `json:"created_at"`
}
```

### Service: `services/debt_service.go`

```go
package services

import (
    "errors"
    "math"
    "my-api/models"
    "my-api/repositories"
    "sort"
    "time"
)

type DebtService interface {
    CreateDebt(userID uint, req *CreateDebtRequest) (*DebtResponse, error)
    GetDebt(id, userID uint) (*DebtDetailResponse, error)
    GetAllDebts(userID uint, status string) (*DebtsListResponse, error)
    UpdateDebt(id, userID uint, req *UpdateDebtRequest) (*DebtResponse, error)
    DeleteDebt(id, userID uint) error
    RecordPayment(debtID, userID uint, req *DebtPaymentRequest) (*DebtPaymentResponse, error)
    UpdateBalance(debtID, userID uint, newBalance int) error
    GetPayoffStrategies(userID uint, extraPayment int) (*StrategiesResponse, error)
    GetPayoffTimeline(debtID, userID uint, strategy string, extraPayment int) (*TimelineResponse, error)
    ProcessMonthlyInterest() error
    CheckMilestones(debt *models.Debt) error
}

type debtService struct {
    repo          repositories.DebtRepository
    assetRepo     repositories.AssetRepository
    txService     TransactionV2Service
    notifyService NotificationService
}

type PayoffStrategy struct {
    Name            string             `json:"name"`
    Description     string             `json:"description"`
    PayoffDate      time.Time          `json:"payoff_date"`
    TotalInterest   int                `json:"total_interest"`
    InterestSaved   int                `json:"interest_saved"`
    MonthsSaved     int                `json:"months_saved"`
    Order           []DebtOrder        `json:"order"`
    Milestones      []PayoffMilestone  `json:"milestones,omitempty"`
}

type DebtOrder struct {
    DebtID       uint    `json:"debt_id"`
    Name         string  `json:"name"`
    Balance      int     `json:"balance,omitempty"`
    InterestRate float64 `json:"interest_rate,omitempty"`
}

func (s *debtService) GetPayoffStrategies(userID uint, extraPayment int) (*StrategiesResponse, error) {
    debts, err := s.repo.FindActiveDebts(userID)
    if err != nil {
        return nil, err
    }
    
    if len(debts) == 0 {
        return nil, errors.New("no active debts found")
    }
    
    totalMinimum := 0
    for _, d := range debts {
        totalMinimum += d.MinimumPayment
    }
    
    strategies := make([]PayoffStrategy, 0)
    
    // 1. Minimum payments only
    minStrategy := s.calculatePayoff(debts, 0, "minimum")
    strategies = append(strategies, minStrategy)
    
    if extraPayment > 0 {
        // 2. Avalanche (highest interest first)
        avalancheStrategy := s.calculatePayoff(debts, extraPayment, "avalanche")
        avalancheStrategy.InterestSaved = minStrategy.TotalInterest - avalancheStrategy.TotalInterest
        avalancheStrategy.MonthsSaved = s.monthsBetween(avalancheStrategy.PayoffDate, minStrategy.PayoffDate)
        strategies = append(strategies, avalancheStrategy)
        
        // 3. Snowball (smallest balance first)
        snowballStrategy := s.calculatePayoff(debts, extraPayment, "snowball")
        snowballStrategy.InterestSaved = minStrategy.TotalInterest - snowballStrategy.TotalInterest
        snowballStrategy.MonthsSaved = s.monthsBetween(snowballStrategy.PayoffDate, minStrategy.PayoffDate)
        strategies = append(strategies, snowballStrategy)
    }
    
    // Determine recommendation
    var recommendation StrategyRecommendation
    if extraPayment > 0 {
        // Avalanche saves more money
        recommendation = StrategyRecommendation{
            Strategy: "Avalanche",
            Reason:   fmt.Sprintf("Saves the most money (Rp %d in interest) while paying off debts %d months faster",
                strategies[1].InterestSaved, strategies[1].MonthsSaved),
        }
    } else {
        recommendation = StrategyRecommendation{
            Strategy: "Add Extra Payments",
            Reason:   "Consider adding extra payments to accelerate debt payoff and save on interest",
        }
    }
    
    return &StrategiesResponse{
        CurrentMonthlyPayment: totalMinimum,
        ExtraAvailable:        extraPayment,
        Strategies:            strategies,
        Recommendation:        recommendation,
    }, nil
}

func (s *debtService) calculatePayoff(debts []models.Debt, extraPayment int, strategy string) PayoffStrategy {
    // Create copies to simulate
    debtsCopy := make([]simulatedDebt, len(debts))
    for i, d := range debts {
        debtsCopy[i] = simulatedDebt{
            ID:             d.ID,
            Name:           d.Name,
            Balance:        d.CurrentBalance,
            InterestRate:   d.InterestRate,
            MinimumPayment: d.MinimumPayment,
        }
    }
    
    // Sort based on strategy
    switch strategy {
    case "avalanche":
        sort.Slice(debtsCopy, func(i, j int) bool {
            return debtsCopy[i].InterestRate > debtsCopy[j].InterestRate
        })
    case "snowball":
        sort.Slice(debtsCopy, func(i, j int) bool {
            return debtsCopy[i].Balance < debtsCopy[j].Balance
        })
    }
    
    currentDate := time.Now()
    totalInterest := 0
    milestones := make([]PayoffMilestone, 0)
    
    // Simulate month by month
    for {
        allPaidOff := true
        availableExtra := extraPayment
        
        for i := range debtsCopy {
            if debtsCopy[i].Balance <= 0 {
                continue
            }
            allPaidOff = false
            
            // Calculate monthly interest
            monthlyInterest := int(float64(debtsCopy[i].Balance) * (debtsCopy[i].InterestRate / 100 / 12))
            totalInterest += monthlyInterest
            debtsCopy[i].Balance += monthlyInterest
            
            // Apply minimum payment
            payment := debtsCopy[i].MinimumPayment
            if debtsCopy[i].Balance < payment {
                payment = debtsCopy[i].Balance
            }
            debtsCopy[i].Balance -= payment
            
            // Apply extra payment to target debt (first in sorted order with balance)
            if availableExtra > 0 && debtsCopy[i].Balance > 0 {
                extraApplied := availableExtra
                if debtsCopy[i].Balance < extraApplied {
                    extraApplied = debtsCopy[i].Balance
                }
                debtsCopy[i].Balance -= extraApplied
                availableExtra -= extraApplied
            }
            
            // Check for payoff milestone
            if debtsCopy[i].Balance <= 0 && strategy != "minimum" {
                milestones = append(milestones, PayoffMilestone{
                    Date:  currentDate,
                    Event: fmt.Sprintf("%s paid off!", debtsCopy[i].Name),
                })
            }
        }
        
        if allPaidOff {
            break
        }
        
        currentDate = currentDate.AddDate(0, 1, 0)
        
        // Safety limit
        if currentDate.Year() > time.Now().Year()+30 {
            break
        }
    }
    
    // Build order
    order := make([]DebtOrder, len(debtsCopy))
    for i, d := range debtsCopy {
        order[i] = DebtOrder{
            DebtID:       d.ID,
            Name:         d.Name,
            Balance:      debts[i].CurrentBalance, // Original balance
            InterestRate: d.InterestRate,
        }
    }
    
    var name, description string
    switch strategy {
    case "avalanche":
        name = "Avalanche (Highest Interest First)"
        description = "Pay minimum on all debts, put extra toward the highest interest debt"
    case "snowball":
        name = "Snowball (Smallest Balance First)"
        description = "Pay minimum on all debts, put extra toward the smallest balance"
    default:
        name = "Minimum Payments Only"
        description = "Pay only minimum payments on all debts"
    }
    
    return PayoffStrategy{
        Name:          name,
        Description:   description,
        PayoffDate:    currentDate,
        TotalInterest: totalInterest,
        Order:         order,
        Milestones:    milestones,
    }
}

func (s *debtService) RecordPayment(debtID, userID uint, req *DebtPaymentRequest) (*DebtPaymentResponse, error) {
    debt, err := s.repo.FindByID(debtID, userID)
    if err != nil {
        return nil, errors.New("debt not found")
    }
    
    if debt.Status != "active" {
        return nil, errors.New("cannot record payment on inactive debt")
    }
    
    balanceBefore := debt.CurrentBalance
    balanceAfter := balanceBefore - req.PrincipalAmount
    
    if balanceAfter < 0 {
        balanceAfter = 0
    }
    
    payment := &models.DebtPayment{
        DebtID:          debtID,
        UserID:          userID,
        Amount:          req.Amount,
        PaymentType:     req.PaymentType,
        PrincipalAmount: req.PrincipalAmount,
        InterestAmount:  req.InterestAmount,
        FeesAmount:      req.FeesAmount,
        BalanceBefore:   balanceBefore,
        BalanceAfter:    balanceAfter,
        PaymentDate:     req.PaymentDate,
        Notes:           req.Notes,
    }
    
    // Create linked transaction if source asset specified
    if req.SourceAssetID != nil {
        tx := &models.TransactionV2{
            UserID:          userID,
            Description:     fmt.Sprintf("Debt payment: %s", debt.Name),
            Amount:          req.Amount,
            TransactionType: 2, // Expense
            AssetID:         *req.SourceAssetID,
            Date:            req.PaymentDate,
        }
        if err := s.txService.CreateTransaction(tx); err == nil {
            payment.TransactionID = &tx.ID
        }
    }
    
    if err := s.repo.CreatePayment(payment); err != nil {
        return nil, err
    }
    
    // Update debt balance
    debt.CurrentBalance = balanceAfter
    debt.CurrentMonthPaid = true
    
    // Check if paid off
    if balanceAfter == 0 {
        debt.Status = "paid_off"
        now := time.Now()
        debt.ActualPayoffDate = &utils.CustomTime{Time: now}
    }
    
    s.repo.Update(debt)
    
    // Check milestones
    s.CheckMilestones(debt)
    
    return &DebtPaymentResponse{
        PaymentID:     payment.ID,
        Amount:        payment.Amount,
        BalanceBefore: balanceBefore,
        BalanceAfter:  balanceAfter,
        IsPaidOff:     balanceAfter == 0,
    }, nil
}

func (s *debtService) CheckMilestones(debt *models.Debt) error {
    percentPaid := debt.PaidOffPercentage()
    
    // Check each milestone
    milestones := []struct {
        Type       string
        Threshold  float64
        Message    string
    }{
        {"25_percent", 25.0, "You've paid off 25% of %s! Keep going! 💪"},
        {"50_percent", 50.0, "Halfway there! 50% of %s is paid off! 🎉"},
        {"75_percent", 75.0, "Amazing! 75% of %s is gone! Almost there! 🚀"},
        {"paid_off", 100.0, "Congratulations! %s is completely paid off! 🏆"},
    }
    
    for _, m := range milestones {
        if percentPaid >= m.Threshold {
            // Check if milestone already reached
            existing, _ := s.repo.FindMilestone(debt.ID, m.Type)
            if existing == nil {
                // Create milestone
                milestone := &models.DebtMilestone{
                    DebtID:        debt.ID,
                    UserID:        debt.UserID,
                    MilestoneType: m.Type,
                    Description:   fmt.Sprintf(m.Message, debt.Name),
                    ReachedAt:     new(time.Time),
                }
                *milestone.ReachedAt = time.Now()
                
                s.repo.CreateMilestone(milestone)
                
                // Send notification
                s.notifyService.CreateNotification(&models.Notification{
                    UserID:   debt.UserID,
                    Type:     "goal_achieved",
                    Title:    "Debt Milestone Reached! 🎉",
                    Message:  fmt.Sprintf(m.Message, debt.Name),
                    Priority: "medium",
                })
            }
        }
    }
    
    return nil
}
```

---

## Frontend Integration Guide

### Debts Overview Screen

```
┌─────────────────────────────────────────┐
│  💳 Debt Tracker                        │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Total Debt                      │   │
│  │ Rp 27,000,000                   │   │
│  │                                 │   │
│  │ Monthly Payments: Rp 1,500,000  │   │
│  │ Avg Interest Rate: 18.5%        │   │
│  │                                 │   │
│  │ Freedom Date: Jun 2028          │   │
│  │ (at minimum payments)           │   │
│  └─────────────────────────────────┘   │
│                                         │
│  📅 DUE THIS MONTH                      │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 💳 BCA Credit Card        ⚠️    │   │
│  │ Due in 4 days (Feb 15)          │   │
│  │ Rp 12,000,000 • 24% APR         │   │
│  │ Min: Rp 500,000                 │   │
│  │ [████████░░░░░░░░░░░░] 20%      │   │
│  │                                 │   │
│  │ [ Pay Now ]                     │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🚗 Car Loan              ✅     │   │
│  │ Paid this month                 │   │
│  │ Rp 15,000,000 • 12% APR         │   │
│  │ Min: Rp 1,000,000               │   │
│  │ [████████████░░░░░░░░] 40%      │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  💡 PAYOFF STRATEGIES                   │
│                                         │
│  With extra Rp 500,000/month:           │
│  • Avalanche: Free in Oct 2027          │
│    (save Rp 7.7M in interest)           │
│                                         │
│  [ View Strategies ]                    │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  [ + Add Debt ]                         │
└─────────────────────────────────────────┘
```

### Payoff Strategy Comparison

```
┌─────────────────────────────────────────┐
│  ← Payoff Strategies                    │
├─────────────────────────────────────────┤
│                                         │
│  Extra monthly payment available:       │
│  ┌─────────────────────────────────┐   │
│  │ Rp 500,000                      │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                         │
│  ⭐ RECOMMENDED                         │
│  ┌─────────────────────────────────┐   │
│  │ 🏔️ Avalanche Method             │   │
│  │ Pay highest interest first      │   │
│  │                                 │   │
│  │ Debt-free: Aug 2027             │   │
│  │ Total interest: Rp 4,800,000    │   │
│  │ You save: Rp 7,700,000          │   │
│  │ 10 months faster                │   │
│  │                                 │   │
│  │ Order:                          │   │
│  │ 1. BCA Card (24%)               │   │
│  │ 2. Car Loan (12%)               │   │
│  │                                 │   │
│  │ [ Apply This Strategy ]         │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ ⛄ Snowball Method              │   │
│  │ Pay smallest balance first      │   │
│  │                                 │   │
│  │ Debt-free: Oct 2027             │   │
│  │ Total interest: Rp 5,200,000    │   │
│  │ You save: Rp 7,300,000          │   │
│  │ 8 months faster                 │   │
│  │                                 │   │
│  │ 💡 Quick wins to stay motivated │   │
│  │                                 │   │
│  │ [ Apply This Strategy ]         │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📉 Minimum Only                 │   │
│  │                                 │   │
│  │ Debt-free: Jun 2028             │   │
│  │ Total interest: Rp 12,500,000   │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

---

## Background Jobs

```go
// jobs/debt_jobs.go

// ProcessMonthlyInterest runs on the 1st of each month
// Adds interest to each debt balance
func (j *DebtJobs) ProcessMonthlyInterest() {
    debts, _ := j.debtRepo.FindAllActiveWithAutoTrack()
    
    for _, debt := range debts {
        interest := debt.MonthlyInterest()
        if interest > 0 {
            j.debtRepo.AddInterest(debt.ID, interest)
            // Optionally create a transaction record
        }
        
        // Reset current_month_paid flag
        j.debtRepo.ResetMonthlyPaidFlag(debt.ID)
    }
}

// SendPaymentReminders runs daily
func (j *DebtJobs) SendPaymentReminders() {
    // Find debts with payment due in 3 days
    debts, _ := j.debtRepo.FindDebtsWithDueDateIn(3)
    
    for _, debt := range debts {
        if !debt.CurrentMonthPaid {
            j.notifyService.CreateNotification(&models.Notification{
                UserID:  debt.UserID,
                Type:    "bill_reminder",
                Title:   fmt.Sprintf("💳 %s payment due in 3 days", debt.Name),
                Message: fmt.Sprintf("Minimum payment of Rp %d is due on day %d", debt.MinimumPayment, debt.PaymentDueDay),
            })
        }
    }
}
```

---

## Benefits

1. **Visibility**: See all debts in one place with clear progress
2. **Strategy**: Know the optimal payoff order to save money
3. **Motivation**: Celebrate milestones to stay on track
4. **Integration**: Auto-track payments from regular transactions
5. **Prevention**: Get reminders before payments are due
