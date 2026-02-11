# Smart Budget Alerts Implementation Plan

## Problem Statement

Current budget alerts are reactive - users only find out they've overspent after it happens. Real-world problems:
- Only aware of overspending after the fact
- No warning when spending velocity is too high
- No context on which purchases pushed them over
- No daily/weekly spending awareness
- Missed opportunities for course correction

---

## Solution Overview

Proactive, intelligent alert system that:
1. Warns before hitting budget limits (velocity alerts)
2. Sends daily/weekly spending summaries
3. Identifies unusual spending patterns
4. Provides actionable insights with each alert
5. Respects user notification preferences

---

## Alert Types

### 1. **Threshold Alerts** (Existing, Enhanced)
- Alert at 50%, 75%, 90%, 100% of budget
- Now includes: which transactions contributed most

### 2. **Velocity Alerts** (New)
- "At your current pace, you'll exceed budget by day 20"
- Early warning based on spending rate vs. days remaining

### 3. **Daily Summary** (New)
- End-of-day spending recap
- Comparison to daily average

### 4. **Weekly Report** (New)
- Week-over-week comparison
- Top spending categories
- Budget health overview

### 5. **Anomaly Alerts** (New)
- "This purchase is 3x your typical spending in this category"
- Unusual vendor or amount detection

### 6. **Goal-Related Alerts** (New)
- "You're behind on your Emergency Fund goal"
- "Congratulations! You've reached 50% of your Japan Trip goal"

---

## Database Schema

### Table: `notification_preferences`

```sql
CREATE TABLE notification_preferences (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,
    
    -- Alert types
    budget_threshold_alerts BOOLEAN DEFAULT true,
    budget_velocity_alerts BOOLEAN DEFAULT true,
    daily_summary BOOLEAN DEFAULT false,
    weekly_report BOOLEAN DEFAULT true,
    anomaly_alerts BOOLEAN DEFAULT true,
    goal_alerts BOOLEAN DEFAULT true,
    bill_reminders BOOLEAN DEFAULT true,
    
    -- Threshold customization
    alert_thresholds JSON DEFAULT '[50, 75, 90, 100]',
    velocity_warn_days INT DEFAULT 7 COMMENT 'Days before month-end to start velocity checks',
    
    -- Delivery preferences
    push_enabled BOOLEAN DEFAULT true,
    email_enabled BOOLEAN DEFAULT false,
    email_address VARCHAR(255) NULL,
    
    -- Timing preferences
    daily_summary_time TIME DEFAULT '21:00:00' COMMENT 'Time to send daily summary',
    weekly_report_day INT DEFAULT 0 COMMENT '0=Sunday, 6=Saturday',
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    
    -- Quiet hours
    quiet_hours_enabled BOOLEAN DEFAULT false,
    quiet_hours_start TIME DEFAULT '22:00:00',
    quiet_hours_end TIME DEFAULT '07:00:00',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Table: `notifications` (Enhanced)

```sql
CREATE TABLE notifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    
    -- Notification details
    type ENUM('budget_threshold', 'budget_velocity', 'daily_summary', 'weekly_report', 'anomaly', 'goal_progress', 'goal_achieved', 'bill_reminder', 'system') NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    
    -- Rich data payload
    data JSON NULL COMMENT 'Additional context data',
    
    -- Reference links
    reference_type VARCHAR(50) NULL COMMENT 'budget, goal, transaction, etc.',
    reference_id BIGINT UNSIGNED NULL,
    
    -- Status
    is_read BOOLEAN DEFAULT false,
    is_pushed BOOLEAN DEFAULT false,
    is_emailed BOOLEAN DEFAULT false,
    
    -- Priority
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP NULL,
    
    INDEX idx_user_unread (user_id, is_read),
    INDEX idx_user_type (user_id, type),
    INDEX idx_created (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

## API Endpoints

### Get Notification Preferences
```http
GET /api/v2/notifications/preferences
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "budget_threshold_alerts": true,
        "budget_velocity_alerts": true,
        "daily_summary": false,
        "weekly_report": true,
        "anomaly_alerts": true,
        "goal_alerts": true,
        "alert_thresholds": [50, 75, 90, 100],
        "daily_summary_time": "21:00",
        "weekly_report_day": 0,
        "timezone": "Asia/Jakarta",
        "quiet_hours": {
            "enabled": false,
            "start": "22:00",
            "end": "07:00"
        }
    }
}
```

### Update Notification Preferences
```http
PUT /api/v2/notifications/preferences
Authorization: Bearer {token}

{
    "daily_summary": true,
    "daily_summary_time": "20:00",
    "alert_thresholds": [60, 80, 95, 100]
}
```

### Get All Notifications
```http
GET /api/v2/notifications?unread_only=false&type=budget_threshold&page=1&limit=20
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "unread_count": 5,
        "notifications": [
            {
                "id": 42,
                "type": "budget_velocity",
                "title": "⚠️ Budget Warning: Food & Dining",
                "message": "At your current pace, you'll exceed your Rp 2,000,000 budget by Feb 20. Consider reducing spending by Rp 50,000/day.",
                "priority": "high",
                "is_read": false,
                "created_at": "2026-02-11T10:30:00Z",
                "data": {
                    "budget_id": 1,
                    "category_name": "Food & Dining",
                    "current_spent": 1200000,
                    "budget_amount": 2000000,
                    "days_remaining": 17,
                    "projected_overspend": 400000,
                    "daily_reduction_needed": 50000
                }
            },
            {
                "id": 41,
                "type": "anomaly",
                "title": "🔍 Unusual Purchase Detected",
                "message": "Your Rp 500,000 purchase at 'Electronics Store' is 4x your typical spending in Shopping.",
                "priority": "medium",
                "is_read": false,
                "created_at": "2026-02-11T09:15:00Z",
                "data": {
                    "transaction_id": 456,
                    "amount": 500000,
                    "typical_amount": 125000,
                    "category": "Shopping"
                }
            }
        ]
    },
    "pagination": {
        "page": 1,
        "page_size": 20,
        "total": 42
    }
}
```

### Mark Notification as Read
```http
PUT /api/v2/notifications/{id}/read
Authorization: Bearer {token}
```

### Mark All as Read
```http
PUT /api/v2/notifications/read-all
Authorization: Bearer {token}
```

### Get Unread Count
```http
GET /api/v2/notifications/unread-count
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "total_unread": 5,
        "by_type": {
            "budget_threshold": 1,
            "budget_velocity": 2,
            "anomaly": 1,
            "goal_progress": 1
        }
    }
}
```

---

## Go Implementation

### Enhanced Budget Alert Checking

```go
package services

import (
    "fmt"
    "my-api/models"
    "my-api/repositories"
    "time"
)

type SmartAlertService interface {
    CheckBudgetAlerts(userID uint, transaction *models.TransactionV2) error
    GenerateDailySummary(userID uint) error
    GenerateWeeklyReport(userID uint) error
    CheckVelocityAlerts(userID uint) error
    DetectAnomalies(userID uint, transaction *models.TransactionV2) error
}

type smartAlertService struct {
    budgetRepo       repositories.BudgetRepository
    notificationRepo repositories.NotificationRepository
    transactionRepo  repositories.TransactionV2Repository
    preferencesRepo  repositories.NotificationPreferencesRepository
}

func (s *smartAlertService) CheckBudgetAlerts(userID uint, transaction *models.TransactionV2) error {
    // Only check for expense transactions
    if transaction.TransactionType != 2 {
        return nil
    }
    
    prefs, err := s.preferencesRepo.GetByUserID(userID)
    if err != nil {
        // Use defaults if no preferences
        prefs = &models.NotificationPreferences{
            BudgetThresholdAlerts: true,
            BudgetVelocityAlerts:  true,
            AnomalyAlerts:         true,
        }
    }
    
    // 1. Check threshold alerts
    if prefs.BudgetThresholdAlerts {
        s.checkThresholdAlerts(userID, transaction, prefs)
    }
    
    // 2. Check velocity alerts
    if prefs.BudgetVelocityAlerts {
        s.checkVelocityAlerts(userID, transaction)
    }
    
    // 3. Check for anomalies
    if prefs.AnomalyAlerts {
        s.detectAnomalies(userID, transaction)
    }
    
    return nil
}

func (s *smartAlertService) checkThresholdAlerts(userID uint, tx *models.TransactionV2, prefs *models.NotificationPreferences) {
    budgets, _ := s.budgetRepo.FindActiveBudgetsByCategory(userID, tx.CategoryID)
    
    for _, budget := range budgets {
        spent, _ := s.budgetRepo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
        percentage := float64(spent) / float64(budget.Amount) * 100
        
        thresholds := []int{50, 75, 90, 100} // Default thresholds
        if prefs.AlertThresholds != nil {
            // Parse custom thresholds from JSON
        }
        
        for _, threshold := range thresholds {
            if percentage >= float64(threshold) {
                // Check if alert already sent for this threshold
                if s.hasAlertBeenSent(budget.ID, threshold) {
                    continue
                }
                
                var title, message string
                var priority string
                
                if percentage >= 100 {
                    title = fmt.Sprintf("🚨 Budget Exceeded: %s", budget.Category.CategoryName)
                    message = fmt.Sprintf("You've spent Rp %d of your Rp %d budget (%.0f%%).", spent, budget.Amount, percentage)
                    priority = "urgent"
                } else {
                    title = fmt.Sprintf("⚠️ Budget Alert: %s", budget.Category.CategoryName)
                    message = fmt.Sprintf("You've used %d%% of your %s budget (Rp %d of Rp %d).", threshold, budget.Category.CategoryName, spent, budget.Amount)
                    priority = "high"
                }
                
                // Get top contributing transactions
                topTransactions, _ := s.transactionRepo.GetTopTransactionsByCategory(
                    userID, tx.CategoryID, budget.StartDate.Time, budget.EndDate.Time, 3,
                )
                
                s.createNotification(&models.Notification{
                    UserID:        userID,
                    Type:          "budget_threshold",
                    Title:         title,
                    Message:       message,
                    Priority:      priority,
                    ReferenceType: "budget",
                    ReferenceID:   &budget.ID,
                    Data: map[string]interface{}{
                        "budget_id":        budget.ID,
                        "category_name":    budget.Category.CategoryName,
                        "spent_amount":     spent,
                        "budget_amount":    budget.Amount,
                        "percentage":       percentage,
                        "threshold":        threshold,
                        "top_transactions": topTransactions,
                        "remaining":        budget.Amount - spent,
                    },
                })
            }
        }
    }
}

func (s *smartAlertService) checkVelocityAlerts(userID uint, tx *models.TransactionV2) {
    budgets, _ := s.budgetRepo.FindActiveBudgetsByCategory(userID, tx.CategoryID)
    
    for _, budget := range budgets {
        spent, _ := s.budgetRepo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
        
        now := time.Now()
        daysElapsed := now.Sub(budget.StartDate.Time).Hours() / 24
        daysRemaining := budget.EndDate.Time.Sub(now).Hours() / 24
        totalDays := budget.EndDate.Time.Sub(budget.StartDate.Time).Hours() / 24
        
        if daysElapsed < 3 || daysRemaining < 3 {
            continue // Not enough data or too close to end
        }
        
        // Calculate daily spending rate
        dailyRate := float64(spent) / daysElapsed
        
        // Project total spending at current rate
        projectedTotal := dailyRate * totalDays
        
        // Only alert if projected to exceed
        if projectedTotal > float64(budget.Amount)*1.1 { // 10% buffer
            projectedOverspend := int(projectedTotal) - budget.Amount
            dailyReduction := float64(projectedOverspend) / daysRemaining
            
            // Check if we already sent this alert today
            if s.hasVelocityAlertToday(budget.ID) {
                continue
            }
            
            title := fmt.Sprintf("⚠️ Spending Pace Alert: %s", budget.Category.CategoryName)
            message := fmt.Sprintf(
                "At your current pace (Rp %.0f/day), you'll exceed your budget by Rp %d. Reduce by Rp %.0f/day to stay on track.",
                dailyRate, projectedOverspend, dailyReduction,
            )
            
            s.createNotification(&models.Notification{
                UserID:        userID,
                Type:          "budget_velocity",
                Title:         title,
                Message:       message,
                Priority:      "high",
                ReferenceType: "budget",
                ReferenceID:   &budget.ID,
                Data: map[string]interface{}{
                    "budget_id":            budget.ID,
                    "category_name":        budget.Category.CategoryName,
                    "current_spent":        spent,
                    "budget_amount":        budget.Amount,
                    "daily_rate":           dailyRate,
                    "days_remaining":       daysRemaining,
                    "projected_total":      projectedTotal,
                    "projected_overspend":  projectedOverspend,
                    "daily_reduction":      dailyReduction,
                },
            })
        }
    }
}

func (s *smartAlertService) detectAnomalies(userID uint, tx *models.TransactionV2) {
    // Get average transaction amount for this category
    avgAmount, _ := s.transactionRepo.GetAverageTransactionAmount(userID, tx.CategoryID, 90) // Last 90 days
    
    if avgAmount == 0 {
        return // No history to compare
    }
    
    // Alert if transaction is 3x or more than average
    ratio := float64(tx.Amount) / float64(avgAmount)
    
    if ratio >= 3.0 {
        title := "🔍 Unusual Purchase Detected"
        message := fmt.Sprintf(
            "Your Rp %d purchase is %.1fx your typical spending of Rp %d in this category.",
            tx.Amount, ratio, avgAmount,
        )
        
        s.createNotification(&models.Notification{
            UserID:        userID,
            Type:          "anomaly",
            Title:         title,
            Message:       message,
            Priority:      "medium",
            ReferenceType: "transaction",
            ReferenceID:   &tx.ID,
            Data: map[string]interface{}{
                "transaction_id":  tx.ID,
                "amount":          tx.Amount,
                "typical_amount":  avgAmount,
                "ratio":           ratio,
                "category":        tx.Category.CategoryName,
                "description":     tx.Description,
            },
        })
    }
}

func (s *smartAlertService) GenerateDailySummary(userID uint) error {
    today := time.Now().Truncate(24 * time.Hour)
    yesterday := today.AddDate(0, 0, -1)
    
    // Get today's spending
    todaySpent, _ := s.transactionRepo.GetTotalSpent(userID, today, today.AddDate(0, 0, 1))
    yesterdaySpent, _ := s.transactionRepo.GetTotalSpent(userID, yesterday, today)
    
    // Get daily average (last 30 days)
    avgDaily, _ := s.transactionRepo.GetDailyAverage(userID, 30)
    
    // Get top categories today
    topCategories, _ := s.transactionRepo.GetSpendingByCategory(userID, today, today.AddDate(0, 0, 1), 3)
    
    var comparison string
    if todaySpent > avgDaily {
        comparison = fmt.Sprintf("Rp %d above", todaySpent-avgDaily)
    } else {
        comparison = fmt.Sprintf("Rp %d below", avgDaily-todaySpent)
    }
    
    title := fmt.Sprintf("📊 Daily Summary: Rp %d spent today", todaySpent)
    message := fmt.Sprintf(
        "Today: Rp %d (%s your daily average of Rp %d). Yesterday: Rp %d.",
        todaySpent, comparison, avgDaily, yesterdaySpent,
    )
    
    return s.createNotification(&models.Notification{
        UserID:   userID,
        Type:     "daily_summary",
        Title:    title,
        Message:  message,
        Priority: "low",
        Data: map[string]interface{}{
            "today_spent":     todaySpent,
            "yesterday_spent": yesterdaySpent,
            "daily_average":   avgDaily,
            "top_categories":  topCategories,
            "date":            today.Format("2006-01-02"),
        },
    })
}
```

---

## Frontend Integration Guide

### Notification Center

```
┌─────────────────────────────────────────┐
│  🔔 Notifications              Mark All │
├─────────────────────────────────────────┤
│                                         │
│  TODAY                                  │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🚨 Budget Exceeded: Food        │   │
│  │ You've spent Rp 2,100,000 of    │   │
│  │ your Rp 2,000,000 budget (105%) │   │
│  │                                 │   │
│  │ Top spending:                   │   │
│  │ • Starbucks: Rp 250,000         │   │
│  │ • Restaurant: Rp 180,000        │   │
│  │                                 │   │
│  │ 10:30 AM                        │   │
│  └─────────────────────────────────┘   │
│  ● (unread indicator)                   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🔍 Unusual Purchase Detected    │   │
│  │ Your Rp 500,000 at Electronics  │   │
│  │ Store is 4x your typical...     │   │
│  │                                 │   │
│  │ 9:15 AM                         │   │
│  └─────────────────────────────────┘   │
│  ●                                      │
│                                         │
│  YESTERDAY                              │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📊 Daily Summary: Rp 450,000    │   │
│  │ Below your daily average...     │   │
│  │                                 │   │
│  │ 9:00 PM                         │   │
│  └─────────────────────────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

### Notification Preferences Screen

```
┌─────────────────────────────────────────┐
│  ← Notification Settings                │
├─────────────────────────────────────────┤
│                                         │
│  BUDGET ALERTS                          │
│                                         │
│  Budget threshold alerts         [ON]   │
│  Get notified at spending milestones    │
│                                         │
│  Alert at: 50% 75% 90% 100%      [Edit] │
│                                         │
│  Spending pace warnings          [ON]   │
│  Early warning when overspending likely │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  SUMMARIES                              │
│                                         │
│  Daily spending summary          [OFF]  │
│  Time: 9:00 PM                          │
│                                         │
│  Weekly financial report         [ON]   │
│  Day: Sunday                            │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  SMART ALERTS                           │
│                                         │
│  Unusual spending detection      [ON]   │
│  Savings goal updates            [ON]   │
│  Bill reminders                  [ON]   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  QUIET HOURS                            │
│                                         │
│  Enable quiet hours              [OFF]  │
│  From 10:00 PM to 7:00 AM               │
│                                         │
└─────────────────────────────────────────┘
```

### Budget Card with Velocity Warning

```
┌─────────────────────────────────────────┐
│  🍔 Food & Dining                       │
│                                         │
│  Rp 1,200,000 / Rp 2,000,000           │
│  ████████████░░░░░░░░░░ 60%            │
│                                         │
│  ⚠️ At current pace: Rp 2,400,000      │
│     Reduce by Rp 50,000/day to          │
│     stay on budget                      │
│                                         │
│  17 days remaining                      │
└─────────────────────────────────────────┘
```

---

## Migration File

```sql
-- db/migrations/YYYYMMDDHHMMSS_create_smart_alerts.up.sql

CREATE TABLE notification_preferences (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,
    budget_threshold_alerts BOOLEAN DEFAULT true,
    budget_velocity_alerts BOOLEAN DEFAULT true,
    daily_summary BOOLEAN DEFAULT false,
    weekly_report BOOLEAN DEFAULT true,
    anomaly_alerts BOOLEAN DEFAULT true,
    goal_alerts BOOLEAN DEFAULT true,
    bill_reminders BOOLEAN DEFAULT true,
    alert_thresholds JSON DEFAULT '[50, 75, 90, 100]',
    velocity_warn_days INT DEFAULT 7,
    push_enabled BOOLEAN DEFAULT true,
    email_enabled BOOLEAN DEFAULT false,
    email_address VARCHAR(255) NULL,
    daily_summary_time TIME DEFAULT '21:00:00',
    weekly_report_day INT DEFAULT 0,
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    quiet_hours_enabled BOOLEAN DEFAULT false,
    quiet_hours_start TIME DEFAULT '22:00:00',
    quiet_hours_end TIME DEFAULT '07:00:00',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE notifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    type ENUM('budget_threshold', 'budget_velocity', 'daily_summary', 'weekly_report', 'anomaly', 'goal_progress', 'goal_achieved', 'bill_reminder', 'system') NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSON NULL,
    reference_type VARCHAR(50) NULL,
    reference_id BIGINT UNSIGNED NULL,
    is_read BOOLEAN DEFAULT false,
    is_pushed BOOLEAN DEFAULT false,
    is_emailed BOOLEAN DEFAULT false,
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP NULL,
    INDEX idx_user_unread (user_id, is_read),
    INDEX idx_user_type (user_id, type),
    INDEX idx_created (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Migrate existing budget_alerts to notifications
INSERT INTO notifications (user_id, type, title, message, reference_type, reference_id, is_read, priority, created_at)
SELECT 
    user_id, 
    'budget_threshold',
    CONCAT('Budget Alert: ', percentage, '%'),
    message,
    'budget',
    budget_id,
    is_read,
    CASE 
        WHEN percentage >= 100 THEN 'urgent'
        WHEN percentage >= 90 THEN 'high'
        ELSE 'medium'
    END,
    created_at
FROM budget_alerts;
```

---

## Background Jobs

```go
// jobs/notification_jobs.go

package jobs

import (
    "my-api/services"
    "my-api/utils"
    "time"
)

type NotificationJobs struct {
    alertService services.SmartAlertService
    userService  services.UserService
}

// RunDailySummaries - Run at configured time for each user
func (j *NotificationJobs) RunDailySummaries() {
    users, _ := j.userService.GetUsersWithDailySummaryEnabled()
    
    for _, user := range users {
        // Check if it's the right time for this user (based on their timezone)
        if j.isUserSummaryTime(user) {
            go j.alertService.GenerateDailySummary(user.ID)
        }
    }
}

// RunWeeklyReports - Run on configured day for each user
func (j *NotificationJobs) RunWeeklyReports() {
    users, _ := j.userService.GetUsersWithWeeklyReportEnabled()
    
    for _, user := range users {
        if j.isUserReportDay(user) {
            go j.alertService.GenerateWeeklyReport(user.ID)
        }
    }
}

// StartScheduler
func (j *NotificationJobs) StartScheduler() {
    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        for range ticker.C {
            j.RunDailySummaries()
            j.RunWeeklyReports()
        }
    }()
}
```

---

## Integration Points

1. **After Transaction Create (V2 Controller)**
   - Call `smartAlertService.CheckBudgetAlerts(userID, transaction)`

2. **Daily Background Job**
   - Generate daily summaries
   - Check velocity alerts for all active budgets

3. **Weekly Background Job**
   - Generate weekly reports

4. **Frontend Webhook/Polling**
   - Poll `/api/v2/notifications/unread-count` every minute
   - Or implement WebSocket for real-time notifications
