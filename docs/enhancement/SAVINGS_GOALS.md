# Savings Goals Implementation Plan

## Problem Statement

Users struggle to save money for specific purposes because:
- No clear target or deadline creates lack of motivation
- Money gets mixed with regular spending
- No visibility on progress toward goals
- Hard to allocate amounts across multiple goals
- Emergency fund often neglected

---

## Solution Overview

Goal-based savings system that:
1. Allows users to create savings goals with targets and deadlines
2. Tracks contributions toward each goal
3. Shows progress visually
4. Suggests monthly amounts to reach goals on time
5. Can link to dedicated wallets/assets or be virtual

---

## Database Schema

### Table: `savings_goals`

```sql
CREATE TABLE savings_goals (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    target_amount INT NOT NULL,
    current_amount INT DEFAULT 0,
    deadline DATE NULL COMMENT 'Optional deadline',
    
    -- Goal configuration
    priority ENUM('low', 'medium', 'high', 'critical') DEFAULT 'medium',
    goal_type ENUM('general', 'emergency_fund', 'vacation', 'purchase', 'investment', 'education', 'retirement', 'other') DEFAULT 'general',
    icon VARCHAR(50) NULL COMMENT 'Emoji or icon identifier',
    color VARCHAR(7) NULL COMMENT 'Hex color code',
    
    -- Linked asset (optional)
    linked_asset_id BIGINT UNSIGNED NULL COMMENT 'If linked, tracks actual asset balance',
    is_virtual BOOLEAN DEFAULT true COMMENT 'Virtual = tracked separately from assets',
    
    -- Auto-save configuration
    auto_save_enabled BOOLEAN DEFAULT false,
    auto_save_amount INT NULL,
    auto_save_frequency ENUM('weekly', 'biweekly', 'monthly') NULL,
    auto_save_asset_id BIGINT UNSIGNED NULL COMMENT 'Source wallet for auto-save',
    
    -- Status
    status ENUM('active', 'paused', 'completed', 'cancelled') DEFAULT 'active',
    completed_at TIMESTAMP NULL,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (linked_asset_id) REFERENCES assets(id) ON DELETE SET NULL,
    FOREIGN KEY (auto_save_asset_id) REFERENCES assets(id) ON DELETE SET NULL,
    
    INDEX idx_user_status (user_id, status),
    INDEX idx_deadline (deadline)
);
```

### Table: `savings_contributions`

```sql
CREATE TABLE savings_contributions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    savings_goal_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    
    amount INT NOT NULL COMMENT 'Positive for deposit, negative for withdrawal',
    contribution_type ENUM('manual', 'auto', 'transfer_in', 'transfer_out', 'adjustment') NOT NULL,
    
    -- Optional link to actual transaction
    transaction_id BIGINT UNSIGNED NULL,
    source_asset_id BIGINT UNSIGNED NULL,
    
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (savings_goal_id) REFERENCES savings_goals(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE SET NULL,
    FOREIGN KEY (source_asset_id) REFERENCES assets(id) ON DELETE SET NULL,
    
    INDEX idx_goal_date (savings_goal_id, created_at)
);
```

---

## API Endpoints

### Create Savings Goal
```http
POST /api/v2/goals
Authorization: Bearer {token}

{
    "name": "Emergency Fund",
    "description": "3 months of expenses",
    "target_amount": 30000000,
    "deadline": "2026-12-31",
    "priority": "critical",
    "goal_type": "emergency_fund",
    "icon": "🛡️",
    "color": "#FF6B6B",
    "is_virtual": true,
    "auto_save_enabled": true,
    "auto_save_amount": 1000000,
    "auto_save_frequency": "monthly",
    "auto_save_asset_id": 1
}
```

### Response
```json
{
    "success": true,
    "message": "Goal created successfully",
    "data": {
        "id": 1,
        "name": "Emergency Fund",
        "target_amount": 30000000,
        "current_amount": 0,
        "progress_percentage": 0,
        "deadline": "2026-12-31",
        "days_remaining": 323,
        "suggested_monthly_amount": 930000,
        "on_track": false,
        "priority": "critical",
        "goal_type": "emergency_fund",
        "icon": "🛡️",
        "color": "#FF6B6B",
        "status": "active"
    }
}
```

### List All Goals
```http
GET /api/v2/goals?status=active&sort=priority
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "summary": {
            "total_goals": 3,
            "active_goals": 3,
            "total_target": 80000000,
            "total_saved": 12500000,
            "overall_progress": 15.6
        },
        "goals": [
            {
                "id": 1,
                "name": "Emergency Fund",
                "icon": "🛡️",
                "target_amount": 30000000,
                "current_amount": 5000000,
                "progress_percentage": 16.7,
                "deadline": "2026-12-31",
                "days_remaining": 323,
                "on_track": true,
                "priority": "critical"
            },
            {
                "id": 2,
                "name": "Japan Trip 2027",
                "icon": "✈️",
                "target_amount": 25000000,
                "current_amount": 7500000,
                "progress_percentage": 30,
                "deadline": "2027-03-01",
                "days_remaining": 383,
                "on_track": true,
                "priority": "medium"
            }
        ]
    }
}
```

### Add Contribution
```http
POST /api/v2/goals/{id}/contributions
Authorization: Bearer {token}

{
    "amount": 500000,
    "source_asset_id": 1,
    "notes": "Bonus from freelance project"
}
```

### Withdraw from Goal
```http
POST /api/v2/goals/{id}/withdraw
Authorization: Bearer {token}

{
    "amount": 200000,
    "destination_asset_id": 1,
    "notes": "Needed for unexpected expense"
}
```

### Get Goal Details with Contribution History
```http
GET /api/v2/goals/{id}?include_contributions=true
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "goal": {
            "id": 1,
            "name": "Emergency Fund",
            "target_amount": 30000000,
            "current_amount": 5000000,
            "progress_percentage": 16.7,
            "deadline": "2026-12-31",
            "monthly_contribution_needed": 2500000,
            "is_on_track": true
        },
        "contributions": [
            {
                "id": 5,
                "amount": 500000,
                "type": "manual",
                "date": "2026-02-10T10:00:00Z",
                "notes": "Bonus from freelance project",
                "running_total": 5000000
            },
            {
                "id": 4,
                "amount": 1000000,
                "type": "auto",
                "date": "2026-02-01T00:00:00Z",
                "running_total": 4500000
            }
        ],
        "projection": {
            "current_trajectory_date": "2028-06-15",
            "needed_monthly_for_deadline": 2500000,
            "current_monthly_average": 1250000
        }
    }
}
```

### Update Goal
```http
PUT /api/v2/goals/{id}
Authorization: Bearer {token}

{
    "target_amount": 35000000,
    "deadline": "2027-06-30"
}
```

### Pause/Resume Goal
```http
PUT /api/v2/goals/{id}/status
Authorization: Bearer {token}

{
    "status": "paused"
}
```

### Delete Goal
```http
DELETE /api/v2/goals/{id}
Authorization: Bearer {token}
```

### Get Allocation Suggestion
```http
GET /api/v2/goals/allocation-suggestion?monthly_savings=5000000
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": {
        "total_available": 5000000,
        "suggested_allocation": [
            {
                "goal_id": 1,
                "goal_name": "Emergency Fund",
                "priority": "critical",
                "suggested_amount": 2500000,
                "reason": "High priority, behind schedule"
            },
            {
                "goal_id": 2,
                "goal_name": "Japan Trip 2027",
                "priority": "medium",
                "suggested_amount": 1500000,
                "reason": "On track, maintain pace"
            },
            {
                "goal_id": 3,
                "goal_name": "New Laptop",
                "priority": "low",
                "suggested_amount": 1000000,
                "reason": "Flexible deadline"
            }
        ],
        "recommendation": "Prioritizing emergency fund due to critical status"
    }
}
```

---

## Go Implementation

### Model: `models/savings_goal.go`

```go
package models

import (
    "my-api/utils"
    "gorm.io/gorm"
    "time"
)

type SavingsGoal struct {
    ID             uint            `gorm:"primaryKey" json:"id"`
    UserID         uint            `gorm:"not null;index" json:"user_id"`
    Name           string          `gorm:"size:100;not null" json:"name"`
    Description    string          `json:"description,omitempty"`
    TargetAmount   int             `gorm:"not null" json:"target_amount"`
    CurrentAmount  int             `gorm:"default:0" json:"current_amount"`
    Deadline       *utils.CustomTime `gorm:"type:date" json:"deadline,omitempty"`
    
    Priority       string          `gorm:"type:enum('low','medium','high','critical');default:'medium'" json:"priority"`
    GoalType       string          `gorm:"type:enum('general','emergency_fund','vacation','purchase','investment','education','retirement','other');default:'general'" json:"goal_type"`
    Icon           string          `gorm:"size:50" json:"icon,omitempty"`
    Color          string          `gorm:"size:7" json:"color,omitempty"`
    
    LinkedAssetID  *uint64         `json:"linked_asset_id,omitempty"`
    IsVirtual      bool            `gorm:"default:true" json:"is_virtual"`
    
    AutoSaveEnabled   bool         `gorm:"default:false" json:"auto_save_enabled"`
    AutoSaveAmount    *int         `json:"auto_save_amount,omitempty"`
    AutoSaveFrequency *string      `json:"auto_save_frequency,omitempty"`
    AutoSaveAssetID   *uint64      `json:"auto_save_asset_id,omitempty"`
    
    Status         string          `gorm:"type:enum('active','paused','completed','cancelled');default:'active'" json:"status"`
    CompletedAt    *time.Time      `json:"completed_at,omitempty"`
    
    CreatedAt      utils.CustomTime `json:"created_at"`
    UpdatedAt      utils.CustomTime `json:"updated_at"`
    DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
    
    // Relations
    LinkedAsset    *Asset          `gorm:"foreignKey:LinkedAssetID" json:"linked_asset,omitempty"`
    Contributions  []SavingsContribution `gorm:"foreignKey:SavingsGoalID" json:"contributions,omitempty"`
}

// Computed fields
func (g *SavingsGoal) ProgressPercentage() float64 {
    if g.TargetAmount == 0 {
        return 0
    }
    return float64(g.CurrentAmount) / float64(g.TargetAmount) * 100
}

func (g *SavingsGoal) DaysRemaining() *int {
    if g.Deadline == nil {
        return nil
    }
    days := int(time.Until(g.Deadline.Time).Hours() / 24)
    return &days
}

func (g *SavingsGoal) MonthlyAmountNeeded() *int {
    if g.Deadline == nil {
        return nil
    }
    remaining := g.TargetAmount - g.CurrentAmount
    if remaining <= 0 {
        zero := 0
        return &zero
    }
    months := time.Until(g.Deadline.Time).Hours() / 24 / 30
    if months <= 0 {
        return &remaining
    }
    monthly := int(float64(remaining) / months)
    return &monthly
}

type SavingsContribution struct {
    ID              uint            `gorm:"primaryKey" json:"id"`
    SavingsGoalID   uint            `gorm:"not null;index" json:"savings_goal_id"`
    UserID          uint            `gorm:"not null" json:"user_id"`
    Amount          int             `gorm:"not null" json:"amount"`
    ContributionType string         `gorm:"type:enum('manual','auto','transfer_in','transfer_out','adjustment');not null" json:"contribution_type"`
    TransactionID   *uint           `json:"transaction_id,omitempty"`
    SourceAssetID   *uint64         `json:"source_asset_id,omitempty"`
    Notes           string          `json:"notes,omitempty"`
    CreatedAt       utils.CustomTime `json:"created_at"`
    
    // Relations
    SourceAsset     *Asset          `gorm:"foreignKey:SourceAssetID" json:"source_asset,omitempty"`
}
```

### Service: `services/savings_goal_service.go`

```go
package services

import (
    "errors"
    "my-api/models"
    "my-api/repositories"
    "time"
)

type SavingsGoalService interface {
    CreateGoal(userID uint, req *CreateGoalRequest) (*GoalResponse, error)
    GetGoal(id, userID uint) (*GoalDetailResponse, error)
    GetAllGoals(userID uint, filter *GoalFilter) (*GoalListResponse, error)
    UpdateGoal(id, userID uint, req *UpdateGoalRequest) (*GoalResponse, error)
    DeleteGoal(id, userID uint) error
    AddContribution(goalID, userID uint, req *ContributionRequest) (*ContributionResponse, error)
    Withdraw(goalID, userID uint, req *WithdrawRequest) (*ContributionResponse, error)
    UpdateStatus(id, userID uint, status string) error
    GetAllocationSuggestion(userID uint, monthlyAmount int) (*AllocationSuggestion, error)
    ProcessAutoSave() error
}

type savingsGoalService struct {
    repo      repositories.SavingsGoalRepository
    assetRepo repositories.AssetRepository
    txService TransactionV2Service
}

func (s *savingsGoalService) AddContribution(goalID, userID uint, req *ContributionRequest) (*ContributionResponse, error) {
    goal, err := s.repo.FindByID(goalID, userID)
    if err != nil {
        return nil, errors.New("goal not found")
    }
    
    if goal.Status != "active" {
        return nil, errors.New("cannot contribute to inactive goal")
    }
    
    // If source asset specified, validate and create transaction
    if req.SourceAssetID != nil {
        asset, err := s.assetRepo.FindByID(*req.SourceAssetID, userID)
        if err != nil {
            return nil, errors.New("source asset not found")
        }
        
        if asset.Balance < int64(req.Amount) {
            return nil, errors.New("insufficient balance in source asset")
        }
        
        // Deduct from source asset
        s.assetRepo.UpdateBalance(*req.SourceAssetID, -int64(req.Amount))
    }
    
    // Create contribution record
    contribution := &models.SavingsContribution{
        SavingsGoalID:    goalID,
        UserID:           userID,
        Amount:           req.Amount,
        ContributionType: "manual",
        SourceAssetID:    req.SourceAssetID,
        Notes:            req.Notes,
    }
    
    if err := s.repo.CreateContribution(contribution); err != nil {
        return nil, err
    }
    
    // Update goal current amount
    newAmount := goal.CurrentAmount + req.Amount
    s.repo.UpdateCurrentAmount(goalID, newAmount)
    
    // Check if goal is now complete
    if newAmount >= goal.TargetAmount {
        s.repo.UpdateStatus(goalID, "completed")
    }
    
    return &ContributionResponse{
        ID:            contribution.ID,
        Amount:        contribution.Amount,
        Type:          contribution.ContributionType,
        NewBalance:    newAmount,
        GoalProgress:  float64(newAmount) / float64(goal.TargetAmount) * 100,
        IsCompleted:   newAmount >= goal.TargetAmount,
    }, nil
}

func (s *savingsGoalService) GetAllocationSuggestion(userID uint, monthlyAmount int) (*AllocationSuggestion, error) {
    goals, err := s.repo.FindActiveGoals(userID)
    if err != nil {
        return nil, err
    }
    
    if len(goals) == 0 {
        return &AllocationSuggestion{
            TotalAvailable: monthlyAmount,
            Allocations:    []GoalAllocation{},
            Recommendation: "No active savings goals. Consider creating one!",
        }, nil
    }
    
    // Priority weights
    priorityWeight := map[string]float64{
        "critical": 4.0,
        "high":     3.0,
        "medium":   2.0,
        "low":      1.0,
    }
    
    // Calculate urgency score for each goal
    type scoredGoal struct {
        Goal   models.SavingsGoal
        Score  float64
        Needed int
    }
    
    scoredGoals := make([]scoredGoal, 0)
    var totalWeight float64
    
    for _, goal := range goals {
        remaining := goal.TargetAmount - goal.CurrentAmount
        if remaining <= 0 {
            continue
        }
        
        weight := priorityWeight[goal.Priority]
        
        // Increase urgency if deadline approaching
        if goal.Deadline != nil {
            daysLeft := time.Until(goal.Deadline.Time).Hours() / 24
            if daysLeft < 90 {
                weight *= 1.5
            } else if daysLeft < 180 {
                weight *= 1.2
            }
        }
        
        totalWeight += weight
        scoredGoals = append(scoredGoals, scoredGoal{
            Goal:   goal,
            Score:  weight,
            Needed: remaining,
        })
    }
    
    // Allocate proportionally based on score
    allocations := make([]GoalAllocation, 0)
    remainingAmount := monthlyAmount
    
    for i, sg := range scoredGoals {
        var amount int
        if i == len(scoredGoals)-1 {
            amount = remainingAmount // Last goal gets remainder
        } else {
            amount = int(float64(monthlyAmount) * (sg.Score / totalWeight))
            if amount > sg.Needed {
                amount = sg.Needed
            }
        }
        
        remainingAmount -= amount
        
        reason := "Balanced allocation"
        if sg.Goal.Priority == "critical" {
            reason = "High priority goal"
        } else if sg.Goal.Deadline != nil {
            days := int(time.Until(sg.Goal.Deadline.Time).Hours() / 24)
            if days < 90 {
                reason = "Deadline approaching soon"
            }
        }
        
        allocations = append(allocations, GoalAllocation{
            GoalID:          sg.Goal.ID,
            GoalName:        sg.Goal.Name,
            Priority:        sg.Goal.Priority,
            SuggestedAmount: amount,
            Reason:          reason,
        })
    }
    
    return &AllocationSuggestion{
        TotalAvailable: monthlyAmount,
        Allocations:    allocations,
        Recommendation: "Allocation based on priority and deadline urgency",
    }, nil
}
```

---

## Frontend Integration Guide

### Goals Overview Screen

```
┌─────────────────────────────────────────┐
│  💰 Savings Goals                       │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  Total Saved                    │   │
│  │  Rp 12,500,000 / Rp 80,000,000  │   │
│  │  ████████░░░░░░░░░░░░░░ 15.6%   │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  🛡️ Emergency Fund           CRITICAL   │
│  ████████░░░░░░░░░░░░░░░░░░░░░ 16.7%   │
│  Rp 5,000,000 / Rp 30,000,000           │
│  📅 323 days left • Rp 2.5M/month needed│
│                                         │
│  ───────────────────────────────────    │
│                                         │
│  ✈️ Japan Trip 2027            MEDIUM   │
│  ██████████░░░░░░░░░░░░░░░░░░░░ 30%     │
│  Rp 7,500,000 / Rp 25,000,000           │
│  📅 383 days left • ✅ On track!        │
│                                         │
│  ───────────────────────────────────    │
│                                         │
│  💻 New MacBook                   LOW    │
│  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0%     │
│  Rp 0 / Rp 25,000,000                   │
│  No deadline set                        │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │      + Create New Goal          │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Goal Detail Screen

```
┌─────────────────────────────────────────┐
│  ←  Emergency Fund              ⋮       │
├─────────────────────────────────────────┤
│                                         │
│         🛡️                              │
│                                         │
│    Rp 5,000,000                         │
│    ────────────                         │
│    of Rp 30,000,000                     │
│                                         │
│  ████████░░░░░░░░░░░░░░░░░░░░░ 16.7%   │
│                                         │
│  ┌──────────────┬──────────────┐       │
│  │ 📅 Deadline  │ 💵 Monthly   │       │
│  │ Dec 31, 2026 │ Rp 2,500,000 │       │
│  │ 323 days     │ needed       │       │
│  └──────────────┴──────────────┘       │
│                                         │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                         │
│  📊 PROGRESS PROJECTION                 │
│                                         │
│  At current pace: Complete by Jun 2028  │
│  To finish on time: +Rp 1.25M/month     │
│                                         │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                         │
│  📜 CONTRIBUTION HISTORY                │
│                                         │
│  Feb 10  +Rp 500,000      Rp 5,000,000  │
│          Freelance bonus                │
│                                         │
│  Feb 1   +Rp 1,000,000    Rp 4,500,000  │
│          Auto-save                      │
│                                         │
│  Jan 15  +Rp 2,000,000    Rp 3,500,000  │
│          Initial deposit                │
│                                         │
│  ┌───────────────┐ ┌───────────────┐   │
│  │  + Add Money  │ │  − Withdraw   │   │
│  └───────────────┘ └───────────────┘   │
└─────────────────────────────────────────┘
```

### Allocation Suggestion Modal

```
┌─────────────────────────────────────────┐
│  💡 Smart Allocation                    │
├─────────────────────────────────────────┤
│                                         │
│  How much can you save this month?      │
│  ┌─────────────────────────────────┐   │
│  │ Rp 5,000,000                    │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  SUGGESTED ALLOCATION:                  │
│                                         │
│  🛡️ Emergency Fund                      │
│  Rp 2,500,000 (50%)                     │
│  "High priority, behind schedule"       │
│                                         │
│  ✈️ Japan Trip                          │
│  Rp 1,500,000 (30%)                     │
│  "On track, maintain pace"              │
│                                         │
│  💻 New Laptop                          │
│  Rp 1,000,000 (20%)                     │
│  "Flexible deadline"                    │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │     Apply This Allocation       │   │
│  └─────────────────────────────────┘   │
│                                         │
│  [ Edit manually ]                      │
└─────────────────────────────────────────┘
```

---

## Migration File

```sql
-- db/migrations/YYYYMMDDHHMMSS_create_savings_goals.up.sql

CREATE TABLE savings_goals (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    target_amount INT NOT NULL,
    current_amount INT DEFAULT 0,
    deadline DATE NULL,
    priority ENUM('low', 'medium', 'high', 'critical') DEFAULT 'medium',
    goal_type ENUM('general', 'emergency_fund', 'vacation', 'purchase', 'investment', 'education', 'retirement', 'other') DEFAULT 'general',
    icon VARCHAR(50) NULL,
    color VARCHAR(7) NULL,
    linked_asset_id BIGINT UNSIGNED NULL,
    is_virtual BOOLEAN DEFAULT true,
    auto_save_enabled BOOLEAN DEFAULT false,
    auto_save_amount INT NULL,
    auto_save_frequency ENUM('weekly', 'biweekly', 'monthly') NULL,
    auto_save_asset_id BIGINT UNSIGNED NULL,
    status ENUM('active', 'paused', 'completed', 'cancelled') DEFAULT 'active',
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_user_status (user_id, status),
    INDEX idx_deadline (deadline),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (linked_asset_id) REFERENCES assets(id) ON DELETE SET NULL,
    FOREIGN KEY (auto_save_asset_id) REFERENCES assets(id) ON DELETE SET NULL
);

CREATE TABLE savings_contributions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    savings_goal_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    amount INT NOT NULL,
    contribution_type ENUM('manual', 'auto', 'transfer_in', 'transfer_out', 'adjustment') NOT NULL,
    transaction_id BIGINT UNSIGNED NULL,
    source_asset_id BIGINT UNSIGNED NULL,
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_goal_date (savings_goal_id, created_at),
    FOREIGN KEY (savings_goal_id) REFERENCES savings_goals(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_id) REFERENCES transactions_v2(id) ON DELETE SET NULL,
    FOREIGN KEY (source_asset_id) REFERENCES assets(id) ON DELETE SET NULL
);
```

---

## Typical Use Cases

### Use Case 1: Building Emergency Fund
1. User creates "Emergency Fund" goal with target = 3 months expenses
2. Sets as "critical" priority
3. Enables auto-save of Rp 1,000,000/month from main wallet
4. System automatically transfers monthly
5. User can add extra contributions when possible
6. Dashboard shows progress and time-to-goal

### Use Case 2: Saving for Vacation
1. User creates "Bali Trip" with target Rp 10,000,000
2. Sets deadline 6 months from now
3. System calculates needed: ~Rp 1,700,000/month
4. User can see if they're on track
5. Can adjust contributions based on monthly income

### Use Case 3: Multiple Competing Goals
1. User has 3 active goals
2. Gets paid and wants to save Rp 5,000,000
3. Uses "Smart Allocation" feature
4. System suggests split based on priority/urgency
5. User reviews and applies allocation in one click
