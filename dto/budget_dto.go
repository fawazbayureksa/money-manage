package dto

import "time"

type CreateBudgetRequest struct {
	CategoryID  uint      `json:"category_id" binding:"required"`
	Amount      int       `json:"amount" binding:"required,min=1"`
	Period      string    `json:"period" binding:"required,oneof=monthly yearly"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	AlertAt     int       `json:"alert_at" binding:"omitempty,min=1,max=100"`
	Description string    `json:"description" binding:"omitempty,max=500"`
}

type UpdateBudgetRequest struct {
	Amount      int    `json:"amount" binding:"omitempty,min=1"`
	AlertAt     int    `json:"alert_at" binding:"omitempty,min=1,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool  `json:"is_active"`
}

type BudgetResponse struct {
	ID           uint      `json:"id"`
	CategoryID   uint      `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       int       `json:"amount"`
	Period       string    `json:"period"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	IsActive     bool      `json:"is_active"`
	AlertAt      int       `json:"alert_at"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

type BudgetWithSpendingResponse struct {
	BudgetResponse
	SpentAmount    int     `json:"spent_amount"`
	RemainingAmount int    `json:"remaining_amount"`
	PercentageUsed float64 `json:"percentage_used"`
	Status         string  `json:"status"` // safe, warning, exceeded
	DaysRemaining  int     `json:"days_remaining"`
}

type BudgetFilterRequest struct {
	PaginationRequest
	CategoryID uint   `form:"category_id"`
	Period     string `form:"period"`
	IsActive   *bool  `form:"is_active"`
	Status     string `form:"status"` // all, active, exceeded, warning
}

type BudgetAlertResponse struct {
	ID          uint      `json:"id"`
	BudgetID    uint      `json:"budget_id"`
	Percentage  int       `json:"percentage"`
	SpentAmount int       `json:"spent_amount"`
	Message     string    `json:"message"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}
