package dto

import (
	"my-api/utils"
	"time"
)

// --- Request DTOs ---

type CreateDebtRequest struct {
	Name           string           `json:"name" binding:"required,max=200"`
	DebtType       string           `json:"debt_type" binding:"required,oneof=credit_card loan mortgage student_loan personal other"`
	OriginalAmount int              `json:"original_amount" binding:"required,min=1"`
	CurrentBalance int              `json:"current_balance" binding:"required,min=0"`
	InterestRate   float64          `json:"interest_rate" binding:"omitempty,min=0,max=100"`
	MinimumPayment int              `json:"minimum_payment" binding:"omitempty,min=0"`
	DueDay         int              `json:"due_day" binding:"omitempty,min=1,max=31"`
	StartDate      utils.CustomTime `json:"start_date" binding:"required"`
	LenderName     string           `json:"lender_name" binding:"omitempty,max=200"`
	Notes          string           `json:"notes" binding:"omitempty,max=500"`
}

type UpdateDebtRequest struct {
	Name           string   `json:"name" binding:"omitempty,max=200"`
	CurrentBalance *int     `json:"current_balance" binding:"omitempty,min=0"`
	InterestRate   *float64 `json:"interest_rate" binding:"omitempty,min=0,max=100"`
	MinimumPayment *int     `json:"minimum_payment" binding:"omitempty,min=0"`
	DueDay         *int     `json:"due_day" binding:"omitempty,min=1,max=31"`
	LenderName     string   `json:"lender_name" binding:"omitempty,max=200"`
	Notes          string   `json:"notes" binding:"omitempty,max=500"`
	IsActive       *bool    `json:"is_active"`
}

type CreateDebtPaymentRequest struct {
	Amount      int              `json:"amount" binding:"required,min=1"`
	PaymentDate utils.CustomTime `json:"payment_date" binding:"required"`
	Notes       string           `json:"notes" binding:"omitempty,max=500"`
}

type DebtFilterRequest struct {
	PaginationRequest
	DebtType string `form:"debt_type"`
	IsActive *bool  `form:"is_active"`
}

type DebtPaymentFilterRequest struct {
	PaginationRequest
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// --- Response DTOs ---

type DebtResponse struct {
	ID              uint             `json:"id"`
	Name            string           `json:"name"`
	DebtType        string           `json:"debt_type"`
	OriginalAmount  int              `json:"original_amount"`
	CurrentBalance  int              `json:"current_balance"`
	InterestRate    float64          `json:"interest_rate"`
	MinimumPayment  int              `json:"minimum_payment"`
	DueDay          int              `json:"due_day"`
	StartDate       utils.CustomTime `json:"start_date"`
	LenderName      string           `json:"lender_name"`
	Notes           string           `json:"notes"`
	IsActive        bool             `json:"is_active"`
	CreatedAt       utils.CustomTime `json:"created_at"`
	TotalPaid       int              `json:"total_paid"`
	PercentagePaid  float64          `json:"percentage_paid"`
	RemainingAmount int              `json:"remaining_amount"`
}

type DebtPaymentResponse struct {
	ID          uint             `json:"id"`
	DebtID      uint             `json:"debt_id"`
	Amount      int              `json:"amount"`
	PaymentDate utils.CustomTime `json:"payment_date"`
	Notes       string           `json:"notes"`
	CreatedAt   utils.CustomTime `json:"created_at"`
	DebtName    string           `json:"debt_name,omitempty"`
}

type DebtMilestoneResponse struct {
	ID            uint      `json:"id"`
	DebtID        uint      `json:"debt_id"`
	MilestoneType string    `json:"milestone_type"`
	TargetValue   int       `json:"target_value"`
	Message       string    `json:"message"`
	IsRead        bool      `json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
	DebtName      string    `json:"debt_name,omitempty"`
}

type DebtSummaryResponse struct {
	TotalDebts        int     `json:"total_debts"`
	ActiveDebts       int     `json:"active_debts"`
	TotalOwed         int     `json:"total_owed"`
	TotalOriginal     int     `json:"total_original"`
	TotalPaid         int     `json:"total_paid"`
	OverallProgress   float64 `json:"overall_progress"`
	TotalMinPayment   int     `json:"total_minimum_payment"`
	HighestInterest   float64 `json:"highest_interest_rate"`
	EstimatedInterest int     `json:"estimated_monthly_interest"`
}

type PayoffStrategyDebt struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	CurrentBalance int     `json:"current_balance"`
	InterestRate   float64 `json:"interest_rate"`
	MinimumPayment int     `json:"minimum_payment"`
	PayoffOrder    int     `json:"payoff_order"`
}

type PayoffStrategyResponse struct {
	Strategy          string               `json:"strategy"` // avalanche, snowball
	Description       string               `json:"description"`
	Debts             []PayoffStrategyDebt `json:"debts"`
	EstimatedMonths   int                  `json:"estimated_months"`
	EstimatedInterest int                  `json:"estimated_total_interest"`
}

type PayoffProjectionMonth struct {
	Month          int `json:"month"`
	Payment        int `json:"payment"`
	Principal      int `json:"principal"`
	Interest       int `json:"interest"`
	RemainingBalance int `json:"remaining_balance"`
}

type PayoffProjectionResponse struct {
	DebtID             uint                     `json:"debt_id"`
	DebtName           string                   `json:"debt_name"`
	CurrentBalance     int                      `json:"current_balance"`
	InterestRate       float64                  `json:"interest_rate"`
	MonthlyPayment     int                      `json:"monthly_payment"`
	EstimatedMonths    int                      `json:"estimated_months"`
	EstimatedPayoffDate string                  `json:"estimated_payoff_date"`
	TotalInterest      int                      `json:"total_interest"`
	TotalPayment       int                      `json:"total_payment"`
	Schedule           []PayoffProjectionMonth  `json:"schedule,omitempty"`
}
