package dto

import "my-api/utils"

// CreateSavingsGoalRequest represents a request to create a savings goal
type CreateSavingsGoalRequest struct {
	Name         string           `json:"name" binding:"required,max=255"`
	Description  string           `json:"description" binding:"omitempty,max=500"`
	TargetAmount int              `json:"target_amount" binding:"required,min=1"`
	Currency     string           `json:"currency" binding:"omitempty,max=10"`
	Deadline     utils.CustomTime `json:"deadline"`
	AssetID      *uint64          `json:"asset_id"`
	Icon         string           `json:"icon" binding:"omitempty,max=50"`
	Color        string           `json:"color" binding:"omitempty,max=20"`
}

// UpdateSavingsGoalRequest represents a request to update a savings goal
type UpdateSavingsGoalRequest struct {
	Name         string           `json:"name" binding:"omitempty,max=255"`
	Description  string           `json:"description" binding:"omitempty,max=500"`
	TargetAmount int              `json:"target_amount" binding:"omitempty,min=1"`
	Currency     string           `json:"currency" binding:"omitempty,max=10"`
	Deadline     *utils.CustomTime `json:"deadline"`
	AssetID      *uint64          `json:"asset_id"`
	Icon         string           `json:"icon" binding:"omitempty,max=50"`
	Color        string           `json:"color" binding:"omitempty,max=20"`
	Status       string           `json:"status" binding:"omitempty,oneof=active completed cancelled"`
}

// SavingsGoalResponse represents a savings goal in API responses
type SavingsGoalResponse struct {
	ID                   uint             `json:"id"`
	Name                 string           `json:"name"`
	Description          string           `json:"description"`
	TargetAmount         int              `json:"target_amount"`
	CurrentAmount        int              `json:"current_amount"`
	RemainingAmount      int              `json:"remaining_amount"`
	Currency             string           `json:"currency"`
	Deadline             utils.CustomTime `json:"deadline"`
	Status               string           `json:"status"`
	Icon                 string           `json:"icon"`
	Color                string           `json:"color"`
	ProgressPercentage   float64          `json:"progress_percentage"`
	MonthlyTargetAmount  int              `json:"monthly_target_amount"`
	MonthsRemaining      int              `json:"months_remaining"`
	DaysRemaining        int              `json:"days_remaining"`
	AssetID              *uint64          `json:"asset_id"`
	AssetName            string           `json:"asset_name,omitempty"`
	CreatedAt            utils.CustomTime `json:"created_at"`
}

// SavingsGoalFilterRequest contains query params for listing goals
type SavingsGoalFilterRequest struct {
	PaginationRequest
	Status string `form:"status" binding:"omitempty,oneof=active completed cancelled"`
}

// AddContributionRequest represents a request to add a contribution to a goal
type AddContributionRequest struct {
	Amount int              `json:"amount" binding:"required,min=1"`
	Note   string           `json:"note" binding:"omitempty,max=500"`
	Date   utils.CustomTime `json:"date"`
}

// SavingsContributionResponse represents a contribution in API responses
type SavingsContributionResponse struct {
	ID        uint             `json:"id"`
	GoalID    uint             `json:"goal_id"`
	Amount    int              `json:"amount"`
	Note      string           `json:"note"`
	Date      utils.CustomTime `json:"date"`
	CreatedAt utils.CustomTime `json:"created_at"`
}

// ContributionFilterRequest contains query params for listing contributions
type ContributionFilterRequest struct {
	PaginationRequest
}
