# Cash Flow Forecasting Implementation Plan

## Problem Statement

Users don't have visibility into their future financial situation:
- "Will I have enough money at the end of the month?"
- "Can I afford this large purchase next week?"
- "When will I run out of money if I keep spending like this?"
- No early warning for potential cash crunches

---

## Solution Overview

Cash flow forecasting system that:
1. Projects future balance based on recurring transactions
2. Analyzes spending patterns to predict variable expenses
3. Shows upcoming large expenses (rent, bills)
4. Alerts when projected balance goes negative
5. Supports "what-if" scenario planning

---

## API Endpoints

### Get Cash Flow Forecast
```http
GET /api/v2/forecasts?days=30&asset_id=1
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "forecast_period": {
            "start_date": "2026-02-11",
            "end_date": "2026-03-12",
            "days": 30
        },
        "current_balance": 5000000,
        "projected_balance": 2500000,
        "lowest_balance": {
            "amount": 1200000,
            "date": "2026-03-01",
            "reason": "After rent payment"
        },
        "total_projected_income": 10000000,
        "total_projected_expenses": 12500000,
        "warnings": [
            {
                "type": "low_balance",
                "date": "2026-03-01",
                "message": "Balance will drop to Rp 1,200,000 after rent payment",
                "suggestion": "Consider postponing non-essential purchases"
            }
        ],
        "daily_forecast": [
            {
                "date": "2026-02-11",
                "opening_balance": 5000000,
                "projected_income": 0,
                "projected_expenses": 150000,
                "closing_balance": 4850000,
                "events": [
                    {
                        "type": "variable_expense",
                        "description": "Daily average spending",
                        "amount": -150000,
                        "confidence": 0.7
                    }
                ]
            },
            {
                "date": "2026-02-25",
                "opening_balance": 2500000,
                "projected_income": 10000000,
                "projected_expenses": 150000,
                "closing_balance": 12350000,
                "events": [
                    {
                        "type": "recurring_income",
                        "description": "Monthly Salary",
                        "amount": 10000000,
                        "confidence": 0.95
                    }
                ]
            },
            {
                "date": "2026-03-01",
                "opening_balance": 11800000,
                "projected_income": 0,
                "projected_expenses": 3650000,
                "closing_balance": 8150000,
                "events": [
                    {
                        "type": "recurring_expense",
                        "description": "Rent Payment",
                        "amount": -3500000,
                        "confidence": 0.95
                    }
                ]
            }
        ],
        "breakdown": {
            "recurring_income": 10000000,
            "recurring_expenses": 5500000,
            "variable_expenses_estimate": 7000000
        }
    }
}
```

### Get What-If Scenario
```http
POST /api/v2/forecasts/scenario
Authorization: Bearer {token}

{
    "base_asset_id": 1,
    "days": 30,
    "hypothetical_transactions": [
        {
            "description": "New Laptop",
            "amount": -15000000,
            "date": "2026-02-20"
        }
    ],
    "exclude_recurring": [2, 5]  // Recurring transaction IDs to exclude
}
```

### Response
```json
{
    "success": true,
    "data": {
        "base_scenario": {
            "end_balance": 2500000,
            "lowest_balance": 1200000,
            "lowest_date": "2026-03-01"
        },
        "hypothetical_scenario": {
            "end_balance": -12500000,
            "lowest_balance": -12500000,
            "lowest_date": "2026-02-20",
            "days_negative": 22
        },
        "impact": {
            "balance_difference": -15000000,
            "goes_negative": true,
            "first_negative_date": "2026-02-20",
            "recommendation": "This purchase would result in a negative balance. Consider waiting until after your salary on Feb 25, or spreading the cost."
        },
        "alternatives": [
            {
                "description": "Wait until Feb 26",
                "new_end_balance": 7500000,
                "lowest_balance": 1200000
            },
            {
                "description": "Split into 3 payments",
                "new_end_balance": -2500000,
                "lowest_balance": -2500000
            }
        ]
    }
}
```

### Get Balance Projection Graph Data
```http
GET /api/v2/forecasts/projection-graph?days=90&asset_id=1
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "points": [
            {"date": "2026-02-11", "balance": 5000000, "is_actual": true},
            {"date": "2026-02-15", "balance": 4400000, "is_actual": false},
            {"date": "2026-02-25", "balance": 12100000, "is_actual": false},
            {"date": "2026-03-01", "balance": 8100000, "is_actual": false},
            {"date": "2026-03-25", "balance": 15500000, "is_actual": false}
        ],
        "confidence_bands": {
            "upper": [
                {"date": "2026-02-15", "balance": 4600000},
                {"date": "2026-02-25", "balance": 12500000}
            ],
            "lower": [
                {"date": "2026-02-15", "balance": 4200000},
                {"date": "2026-02-25", "balance": 11700000}
            ]
        },
        "key_events": [
            {"date": "2026-02-25", "event": "Salary", "amount": 10000000},
            {"date": "2026-03-01", "event": "Rent", "amount": -3500000}
        ]
    }
}
```

---

## Database Schema

### Table: `forecast_history` (Optional - for tracking accuracy)

```sql
CREATE TABLE forecast_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    asset_id BIGINT UNSIGNED NOT NULL,
    forecast_date DATE NOT NULL COMMENT 'Date the forecast was made',
    target_date DATE NOT NULL COMMENT 'Date being forecasted',
    predicted_balance INT NOT NULL,
    actual_balance INT NULL COMMENT 'Filled in when date passes',
    variance INT NULL COMMENT 'actual - predicted',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_dates (user_id, forecast_date, target_date),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
);
```

---

## Go Implementation

### Service: `services/forecast_service.go`

```go
package services

import (
    "my-api/models"
    "my-api/repositories"
    "time"
)

type ForecastService interface {
    GetForecast(userID uint, assetID *uint64, days int) (*ForecastResponse, error)
    GetScenario(userID uint, req *ScenarioRequest) (*ScenarioResponse, error)
    GetProjectionGraph(userID uint, assetID *uint64, days int) (*ProjectionGraphResponse, error)
}

type forecastService struct {
    assetRepo      repositories.AssetRepository
    recurringRepo  repositories.RecurringTransactionRepository
    transactionRepo repositories.TransactionV2Repository
}

func NewForecastService(
    assetRepo repositories.AssetRepository,
    recurringRepo repositories.RecurringTransactionRepository,
    transactionRepo repositories.TransactionV2Repository,
) ForecastService {
    return &forecastService{
        assetRepo:      assetRepo,
        recurringRepo:  recurringRepo,
        transactionRepo: transactionRepo,
    }
}

type DailyForecast struct {
    Date           time.Time         `json:"date"`
    OpeningBalance int               `json:"opening_balance"`
    ProjectedIncome int              `json:"projected_income"`
    ProjectedExpenses int            `json:"projected_expenses"`
    ClosingBalance int               `json:"closing_balance"`
    Events         []ForecastEvent   `json:"events"`
}

type ForecastEvent struct {
    Type        string  `json:"type"` // recurring_income, recurring_expense, variable_expense
    Description string  `json:"description"`
    Amount      int     `json:"amount"`
    Confidence  float64 `json:"confidence"`
    RecurringID *uint   `json:"recurring_id,omitempty"`
}

func (s *forecastService) GetForecast(userID uint, assetID *uint64, days int) (*ForecastResponse, error) {
    // Get current balance
    var currentBalance int64
    if assetID != nil {
        asset, err := s.assetRepo.FindByID(*assetID, userID)
        if err != nil {
            return nil, err
        }
        currentBalance = asset.Balance
    } else {
        // Total across all assets
        currentBalance, _ = s.assetRepo.GetTotalBalance(userID)
    }
    
    // Get recurring transactions
    recurringTxs, _ := s.recurringRepo.FindActiveByUser(userID, assetID)
    
    // Calculate average daily variable spending (non-recurring expenses)
    avgDailyExpense, _ := s.transactionRepo.GetAverageDailyVariableExpense(userID, assetID, 60) // Last 60 days
    
    // Build daily forecast
    startDate := time.Now().Truncate(24 * time.Hour)
    endDate := startDate.AddDate(0, 0, days)
    
    dailyForecasts := make([]DailyForecast, 0, days)
    balance := int(currentBalance)
    lowestBalance := balance
    lowestDate := startDate
    
    var totalProjectedIncome, totalProjectedExpenses int
    var warnings []ForecastWarning
    
    for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
        dayForecast := DailyForecast{
            Date:           d,
            OpeningBalance: balance,
            Events:         make([]ForecastEvent, 0),
        }
        
        dayIncome := 0
        dayExpense := 0
        
        // Check recurring transactions for this day
        for _, rt := range recurringTxs {
            if s.recurringOccursOn(rt, d) {
                event := ForecastEvent{
                    Description: rt.Description,
                    Confidence:  0.95,
                    RecurringID: &rt.ID,
                }
                
                if rt.TransactionType == 1 { // Income
                    event.Type = "recurring_income"
                    event.Amount = rt.Amount
                    dayIncome += rt.Amount
                } else { // Expense
                    event.Type = "recurring_expense"
                    event.Amount = -rt.Amount
                    dayExpense += rt.Amount
                }
                
                dayForecast.Events = append(dayForecast.Events, event)
            }
        }
        
        // Add variable expense estimate
        if avgDailyExpense > 0 {
            dayForecast.Events = append(dayForecast.Events, ForecastEvent{
                Type:        "variable_expense",
                Description: "Estimated daily spending",
                Amount:      -avgDailyExpense,
                Confidence:  0.7,
            })
            dayExpense += avgDailyExpense
        }
        
        dayForecast.ProjectedIncome = dayIncome
        dayForecast.ProjectedExpenses = dayExpense
        balance = balance + dayIncome - dayExpense
        dayForecast.ClosingBalance = balance
        
        totalProjectedIncome += dayIncome
        totalProjectedExpenses += dayExpense
        
        // Track lowest balance
        if balance < lowestBalance {
            lowestBalance = balance
            lowestDate = d
            
            // Check for significant events on this day
            for _, event := range dayForecast.Events {
                if event.Type == "recurring_expense" {
                    // Generate warning
                    warnings = append(warnings, ForecastWarning{
                        Type:       "low_balance",
                        Date:       d,
                        Message:    fmt.Sprintf("Balance will drop to Rp %d after %s", balance, event.Description),
                        Suggestion: "Consider postponing non-essential purchases",
                    })
                    break
                }
            }
        }
        
        dailyForecasts = append(dailyForecasts, dayForecast)
    }
    
    return &ForecastResponse{
        ForecastPeriod: ForecastPeriod{
            StartDate: startDate,
            EndDate:   endDate,
            Days:      days,
        },
        CurrentBalance:   int(currentBalance),
        ProjectedBalance: balance,
        LowestBalance: LowestBalanceInfo{
            Amount: lowestBalance,
            Date:   lowestDate,
            Reason: s.getLowestBalanceReason(dailyForecasts, lowestDate),
        },
        TotalProjectedIncome:   totalProjectedIncome,
        TotalProjectedExpenses: totalProjectedExpenses,
        Warnings:               warnings,
        DailyForecast:          dailyForecasts,
        Breakdown: ForecastBreakdown{
            RecurringIncome:          s.calculateRecurringIncome(recurringTxs, days),
            RecurringExpenses:        s.calculateRecurringExpenses(recurringTxs, days),
            VariableExpensesEstimate: avgDailyExpense * days,
        },
    }, nil
}

func (s *forecastService) recurringOccursOn(rt models.RecurringTransaction, date time.Time) bool {
    // Check if before start date or after end date
    if date.Before(rt.StartDate.Time) {
        return false
    }
    if rt.EndDate != nil && date.After(rt.EndDate.Time) {
        return false
    }
    
    switch rt.Frequency {
    case "daily":
        return true
    case "weekly":
        return int(date.Weekday()) == *rt.DayOfWeek
    case "biweekly":
        weeks := int(date.Sub(rt.StartDate.Time).Hours() / 24 / 7)
        return weeks%2 == 0 && int(date.Weekday()) == *rt.DayOfWeek
    case "monthly":
        return date.Day() == *rt.DayOfMonth
    case "quarterly":
        monthDiff := (date.Year()-rt.StartDate.Time.Year())*12 + int(date.Month()-rt.StartDate.Time.Month())
        return monthDiff%3 == 0 && date.Day() == *rt.DayOfMonth
    case "yearly":
        return int(date.Month()) == *rt.MonthOfYear && date.Day() == *rt.DayOfMonth
    }
    
    return false
}

func (s *forecastService) GetScenario(userID uint, req *ScenarioRequest) (*ScenarioResponse, error) {
    // Get base forecast
    baseForecast, err := s.GetForecast(userID, &req.BaseAssetID, req.Days)
    if err != nil {
        return nil, err
    }
    
    // Apply hypothetical transactions
    hypotheticalBalance := baseForecast.CurrentBalance
    lowestBalance := hypotheticalBalance
    lowestDate := time.Now()
    daysNegative := 0
    var firstNegativeDate *time.Time
    
    // Sort hypothetical transactions by date
    // ... sorting logic ...
    
    for _, df := range baseForecast.DailyForecast {
        dayBalance := df.ClosingBalance
        
        // Apply any hypothetical transactions for this day
        for _, ht := range req.HypotheticalTransactions {
            htDate, _ := time.Parse("2006-01-02", ht.Date)
            if htDate.Equal(df.Date) {
                dayBalance += ht.Amount
            }
        }
        
        if dayBalance < lowestBalance {
            lowestBalance = dayBalance
            lowestDate = df.Date
        }
        
        if dayBalance < 0 {
            daysNegative++
            if firstNegativeDate == nil {
                firstNegativeDate = &df.Date
            }
        }
        
        hypotheticalBalance = dayBalance
    }
    
    // Generate recommendation
    var recommendation string
    if lowestBalance < 0 {
        recommendation = "This scenario would result in a negative balance. Consider adjusting timing or amounts."
    } else if lowestBalance < baseForecast.CurrentBalance/10 {
        recommendation = "This would leave you with very low reserves. Consider building more buffer."
    } else {
        recommendation = "This scenario appears financially feasible."
    }
    
    return &ScenarioResponse{
        BaseScenario: ScenarioSummary{
            EndBalance:    baseForecast.ProjectedBalance,
            LowestBalance: baseForecast.LowestBalance.Amount,
            LowestDate:    baseForecast.LowestBalance.Date,
        },
        HypotheticalScenario: ScenarioSummary{
            EndBalance:    hypotheticalBalance,
            LowestBalance: lowestBalance,
            LowestDate:    lowestDate,
            DaysNegative:  daysNegative,
        },
        Impact: ScenarioImpact{
            BalanceDifference:  hypotheticalBalance - baseForecast.ProjectedBalance,
            GoesNegative:       lowestBalance < 0,
            FirstNegativeDate:  firstNegativeDate,
            Recommendation:     recommendation,
        },
    }, nil
}
```

---

## Frontend Integration Guide

### Forecast Dashboard Widget

```
┌─────────────────────────────────────────┐
│  📈 30-Day Forecast                     │
├─────────────────────────────────────────┤
│                                         │
│  Current Balance                        │
│  Rp 5,000,000                           │
│        │                                │
│        │    ╭──── Feb 25: Salary        │
│        │   ╱                            │
│        │  ╱                             │
│        ╰╮╱                              │
│  1.2M ──┼───── Mar 1: Rent              │
│         │ ╲                             │
│         │  ╲                            │
│         │   ╲                           │
│         │    ╰────                      │
│  Rp 2,500,000 projected                 │
│  ────────────────────────────────────── │
│  Feb 11    Feb 25    Mar 1    Mar 12    │
│                                         │
│  ⚠️ Lowest: Rp 1.2M on Mar 1            │
│                                         │
│  [ View Details ]                       │
└─────────────────────────────────────────┘
```

### Full Forecast Screen

```
┌─────────────────────────────────────────┐
│  ← Cash Flow Forecast                   │
├─────────────────────────────────────────┤
│                                         │
│  ┌────────────────────────────────┐    │
│  │ [30 days] [60 days] [90 days]  │    │
│  └────────────────────────────────┘    │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Wallet: [All Wallets ▼]         │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌──────────────┬──────────────┐       │
│  │ NOW          │ IN 30 DAYS   │       │
│  │ Rp 5,000,000 │ Rp 2,500,000 │       │
│  └──────────────┴──────────────┘       │
│                                         │
│  [========== GRAPH ==========]          │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  📊 BREAKDOWN                           │
│                                         │
│  Expected Income      +Rp 10,000,000    │
│  ├ Salary               +Rp 10,000,000  │
│                                         │
│  Expected Expenses    -Rp 12,500,000    │
│  ├ Rent                 -Rp 3,500,000   │
│  ├ Utilities            -Rp 500,000     │
│  ├ Subscriptions        -Rp 500,000     │
│  └ Variable spending    -Rp 8,000,000   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  ⚠️ WARNINGS                            │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Mar 1: Balance drops to Rp 1.2M │   │
│  │ after rent payment              │   │
│  │                                 │   │
│  │ 💡 Consider postponing large    │   │
│  │    purchases until after        │   │
│  │    salary on Feb 25             │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  🤔 WHAT IF...                          │
│                                         │
│  [ + Plan a purchase ]                  │
└─────────────────────────────────────────┘
```

### What-If Scenario Screen

```
┌─────────────────────────────────────────┐
│  ← What-If Scenario                     │
├─────────────────────────────────────────┤
│                                         │
│  What purchase are you planning?        │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ New Laptop                      │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Amount                                 │
│  ┌─────────────────────────────────┐   │
│  │ Rp 15,000,000                   │   │
│  └─────────────────────────────────┘   │
│                                         │
│  When?                                  │
│  ┌─────────────────────────────────┐   │
│  │ Feb 20, 2026                 📅 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  📊 IMPACT ANALYSIS                     │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │     Without        With         │   │
│  │     Purchase       Purchase     │   │
│  │                                 │   │
│  │     Rp 2.5M    →   -Rp 12.5M   │   │
│  │     (Mar 12)       (Mar 12)     │   │
│  │                                 │   │
│  │     ✅ OK          ❌ NEGATIVE  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ⚠️ This purchase would put your       │
│     balance negative for 22 days       │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  💡 ALTERNATIVES                        │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ ⏰ Wait until Feb 26            │   │
│  │    End balance: Rp 7,500,000    │   │
│  │    [Apply this]                 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📅 Split into 3 payments        │   │
│  │    End balance: -Rp 2,500,000   │   │
│  │    [Apply this]                 │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

---

## Key Benefits

1. **Prevents Overdrafts**: Users see future cash crunches before they happen
2. **Better Decision Making**: "Can I afford this?" becomes answerable
3. **Visual Timeline**: See when money comes in and goes out
4. **Scenario Planning**: Test purchases before committing
5. **Pattern Recognition**: Uses historical data for variable expense estimates

---

## Implementation Notes

### Confidence Levels
- **Recurring transactions**: 95% confidence (scheduled and predictable)
- **Variable expenses**: 60-70% confidence (based on historical averages)
- **Adjust over time**: Track forecast accuracy and improve models

### Edge Cases
- New users with no history: Use conservative defaults
- Irregular income: Allow manual override for variable income
- One-time known expenses: Let users add manual events

### Performance
- Cache forecasts for 1 hour
- Recalculate on transaction create/update
- Limit to 90 days maximum for performance
