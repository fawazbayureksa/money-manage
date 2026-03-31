package dto

import (
	"my-api/utils"
	"time"
)

// ---- Recurring Transactions ----

type CreateRecurringTransactionRequest struct {
	AssetID         uint64           `json:"asset_id" binding:"required"`
	CategoryID      *uint            `json:"category_id"`
	Description     string           `json:"description" binding:"required,max=200"`
	Amount          int              `json:"amount" binding:"required,min=1"`
	TransactionType int              `json:"transaction_type" binding:"required,oneof=1 2"`
	Frequency       string           `json:"frequency" binding:"required,oneof=daily weekly bi_weekly monthly yearly"`
	DayOfMonth      *uint            `json:"day_of_month" binding:"omitempty,min=1,max=31"`
	DayOfWeek       *uint            `json:"day_of_week" binding:"omitempty,min=0,max=6"`
	StartDate       utils.CustomTime `json:"start_date" binding:"required"`
	EndDate         *utils.CustomTime `json:"end_date"`
}

type UpdateRecurringTransactionRequest struct {
	AssetID         *uint64           `json:"asset_id"`
	CategoryID      *uint             `json:"category_id"`
	Description     string            `json:"description" binding:"omitempty,max=200"`
	Amount          *int              `json:"amount" binding:"omitempty,min=1"`
	TransactionType *int              `json:"transaction_type" binding:"omitempty,oneof=1 2"`
	Frequency       string            `json:"frequency" binding:"omitempty,oneof=daily weekly bi_weekly monthly yearly"`
	DayOfMonth      *uint             `json:"day_of_month" binding:"omitempty,min=1,max=31"`
	DayOfWeek       *uint             `json:"day_of_week" binding:"omitempty,min=0,max=6"`
	StartDate       *utils.CustomTime `json:"start_date"`
	EndDate         *utils.CustomTime `json:"end_date"`
	IsActive        *bool             `json:"is_active"`
}

type RecurringTransactionResponse struct {
	ID              uint64            `json:"id"`
	AssetID         uint64            `json:"asset_id"`
	AssetName       string            `json:"asset_name,omitempty"`
	CategoryID      *uint             `json:"category_id"`
	CategoryName    string            `json:"category_name,omitempty"`
	Description     string            `json:"description"`
	Amount          int               `json:"amount"`
	TransactionType int               `json:"transaction_type"`
	Frequency       string            `json:"frequency"`
	DayOfMonth      *uint             `json:"day_of_month,omitempty"`
	DayOfWeek       *uint             `json:"day_of_week,omitempty"`
	StartDate       utils.CustomTime  `json:"start_date"`
	EndDate         *utils.CustomTime `json:"end_date,omitempty"`
	NextOccurrence  utils.CustomTime  `json:"next_occurrence"`
	IsActive        bool              `json:"is_active"`
	CreatedAt       utils.CustomTime  `json:"created_at"`
}

type RecurringTransactionFilterRequest struct {
	PaginationRequest
	IsActive  *bool  `form:"is_active"`
	Frequency string `form:"frequency"`
	AssetID   *uint64 `form:"asset_id"`
}

// ---- Cash Flow Forecast ----

type CashFlowForecastRequest struct {
	Days    int     `form:"days" binding:"omitempty,min=1,max=365"`
	AssetID *uint64 `form:"asset_id"`
}

// ForecastEvent represents a single income or expense event on a forecast day.
type ForecastEvent struct {
	Description     string `json:"description"`
	Amount          int    `json:"amount"`
	TransactionType int    `json:"transaction_type"` // 1=income, 2=expense
	Source          string `json:"source"`           // "recurring" or "estimated"
	RecurringID     *uint64 `json:"recurring_id,omitempty"`
}

// ForecastDay represents the projected financial state for a single day.
type ForecastDay struct {
	Date             string          `json:"date"`
	Events           []ForecastEvent `json:"events"`
	DayIncome        int             `json:"day_income"`
	DayExpense       int             `json:"day_expense"`
	ProjectedBalance float64         `json:"projected_balance"`
	IsNegative       bool            `json:"is_negative"`
}

// NegativeBalanceAlert is generated when the projected balance goes negative.
type NegativeBalanceAlert struct {
	Date             string  `json:"date"`
	ProjectedBalance float64 `json:"projected_balance"`
	TriggerEvent     string  `json:"trigger_event"`
}

// CashFlowForecastResponse is the top-level response for the forecast endpoint.
type CashFlowForecastResponse struct {
	AssetID          *uint64                `json:"asset_id,omitempty"`
	CurrentBalance   float64                `json:"current_balance"`
	ForecastDays     int                    `json:"forecast_days"`
	StartDate        string                 `json:"start_date"`
	EndDate          string                 `json:"end_date"`
	MinBalance       float64                `json:"min_balance"`
	MinBalanceDate   string                 `json:"min_balance_date"`
	MaxBalance       float64                `json:"max_balance"`
	MaxBalanceDate   string                 `json:"max_balance_date"`
	NegativeAlerts   []NegativeBalanceAlert `json:"negative_alerts"`
	DailyProjections []ForecastDay          `json:"daily_projections"`
	AvgDailyExpense  float64                `json:"avg_daily_expense"`
}

// ScenarioTransaction is a one-off transaction used in what-if scenario planning.
type ScenarioTransaction struct {
	Description     string    `json:"description" binding:"required,max=200"`
	Amount          int       `json:"amount" binding:"required,min=1"`
	TransactionType int       `json:"transaction_type" binding:"required,oneof=1 2"`
	Date            time.Time `json:"date" binding:"required"`
}

type CashFlowScenarioRequest struct {
	Days         int                   `json:"days" binding:"omitempty,min=1,max=365"`
	AssetID      *uint64               `json:"asset_id"`
	Transactions []ScenarioTransaction `json:"transactions" binding:"required,min=1"`
}
