package dto

import (
	"my-api/utils"
	"time"
)

// ─── Requests ────────────────────────────────────────────────────────────────

type CreateDebtRequest struct {
	Name           string           `json:"name" binding:"required,max=100"`
	Description    string           `json:"description" binding:"omitempty,max=500"`
	DebtType       string           `json:"debt_type" binding:"required,oneof=credit_card personal_loan mortgage car_loan student_loan other"`
	CreditorName   string           `json:"creditor_name" binding:"omitempty,max=100"`
	OriginalAmount int              `json:"original_amount" binding:"required,min=1"`
	CurrentBalance int              `json:"current_balance" binding:"required,min=0"`
	InterestRate   float64          `json:"interest_rate" binding:"required,min=0,max=100"`
	InterestType   string           `json:"interest_type" binding:"omitempty,oneof=fixed variable"`
	MinimumPayment int              `json:"minimum_payment" binding:"required,min=1"`
	PaymentDueDay  int              `json:"payment_due_day" binding:"required,min=1,max=31"`
	StartDate      utils.CustomTime `json:"start_date" binding:"required"`
	PaymentAssetID *uint64          `json:"payment_asset_id"`
	Notes          string           `json:"notes" binding:"omitempty"`
}

type UpdateDebtRequest struct {
	Name             string           `json:"name" binding:"omitempty,max=100"`
	Description      string           `json:"description" binding:"omitempty,max=500"`
	CreditorName     string           `json:"creditor_name" binding:"omitempty,max=100"`
	InterestRate     *float64         `json:"interest_rate" binding:"omitempty,min=0,max=100"`
	InterestType     string           `json:"interest_type" binding:"omitempty,oneof=fixed variable"`
	MinimumPayment   *int             `json:"minimum_payment" binding:"omitempty,min=1"`
	PaymentDueDay    *int             `json:"payment_due_day" binding:"omitempty,min=1,max=31"`
	PaymentAssetID   *uint64          `json:"payment_asset_id"`
	IncludeInNetWorth *bool           `json:"include_in_net_worth"`
	AutoTrackInterest *bool           `json:"auto_track_interest"`
	Status           string           `json:"status" binding:"omitempty,oneof=active paid_off defaulted settled"`
	Notes            string           `json:"notes" binding:"omitempty"`
	ExpectedPayoffDate *utils.CustomTime `json:"expected_payoff_date"`
}

type RecordDebtPaymentRequest struct {
	Amount          int              `json:"amount" binding:"required,min=1"`
	PaymentType     string           `json:"payment_type" binding:"required,oneof=regular extra interest_only payoff adjustment"`
	PrincipalAmount int              `json:"principal_amount" binding:"omitempty,min=0"`
	InterestAmount  int              `json:"interest_amount" binding:"omitempty,min=0"`
	FeesAmount      int              `json:"fees_amount" binding:"omitempty,min=0"`
	PaymentDate     utils.CustomTime `json:"payment_date" binding:"required"`
	SourceAssetID   *uint            `json:"source_asset_id"`
	Notes           string           `json:"notes" binding:"omitempty"`
}

type UpdateDebtBalanceRequest struct {
	NewBalance int              `json:"new_balance" binding:"required,min=0"`
	AsOfDate   utils.CustomTime `json:"as_of_date" binding:"required"`
}

// ─── Responses ───────────────────────────────────────────────────────────────

type DebtProjections struct {
	PayoffDateMinimum    string `json:"payoff_date_minimum"`
	TotalInterestMinimum int    `json:"total_interest_minimum"`
	MonthsRemainingMin   int    `json:"months_remaining_minimum"`
}

type DebtResponse struct {
	ID               uint             `json:"id"`
	Name             string           `json:"name"`
	DebtType         string           `json:"debt_type"`
	CreditorName     string           `json:"creditor_name,omitempty"`
	OriginalAmount   int              `json:"original_amount"`
	CurrentBalance   int              `json:"current_balance"`
	PaidOffAmount    int              `json:"paid_off_amount"`
	PaidOffPercentage float64         `json:"paid_off_percentage"`
	InterestRate     float64          `json:"interest_rate"`
	InterestType     string           `json:"interest_type"`
	MinimumPayment   int              `json:"minimum_payment"`
	PaymentDueDay    int              `json:"payment_due_day"`
	DaysUntilDue     int              `json:"days_until_due"`
	CurrentMonthPaid bool             `json:"current_month_paid"`
	Status           string           `json:"status"`
	StartDate        utils.CustomTime `json:"start_date"`
	Notes            string           `json:"notes,omitempty"`
	CreatedAt        utils.CustomTime `json:"created_at"`
	Projections      *DebtProjections `json:"projections,omitempty"`
}

type DebtSummary struct {
	TotalDebt             int     `json:"total_debt"`
	TotalMinimumPayment   int     `json:"total_minimum_payment"`
	TotalPaidThisMonth    int     `json:"total_paid_this_month"`
	DebtsPaidThisMonth    int     `json:"debts_paid_this_month"`
	DebtsUnpaidThisMonth  int     `json:"debts_unpaid_this_month"`
	AvgInterestRate       float64 `json:"avg_interest_rate"`
	ProjectedPayoffDate   string  `json:"projected_payoff_date"`
	TotalInterestMinimum  int     `json:"total_interest_if_minimum"`
}

type DebtsListResponse struct {
	Summary DebtSummary    `json:"summary"`
	Debts   []DebtResponse `json:"debts"`
}

type DebtDetailResponse struct {
	DebtResponse
	Payments   []DebtPaymentHistoryItem  `json:"payments,omitempty"`
	Milestones []DebtMilestoneItem       `json:"milestones,omitempty"`
}

type DebtPaymentHistoryItem struct {
	ID              uint             `json:"id"`
	Amount          int              `json:"amount"`
	PaymentType     string           `json:"payment_type"`
	PrincipalAmount int              `json:"principal_amount"`
	InterestAmount  int              `json:"interest_amount"`
	FeesAmount      int              `json:"fees_amount"`
	BalanceBefore   int              `json:"balance_before"`
	BalanceAfter    int              `json:"balance_after"`
	PaymentDate     utils.CustomTime `json:"payment_date"`
	Notes           string           `json:"notes,omitempty"`
	CreatedAt       utils.CustomTime `json:"created_at"`
}

type DebtMilestoneItem struct {
	ID            uint       `json:"id"`
	MilestoneType string     `json:"milestone_type"`
	Description   string     `json:"description,omitempty"`
	ReachedAt     *time.Time `json:"reached_at,omitempty"`
	IsCelebrated  bool       `json:"is_celebrated"`
}

type RecordDebtPaymentResponse struct {
	PaymentID     uint `json:"payment_id"`
	Amount        int  `json:"amount"`
	BalanceBefore int  `json:"balance_before"`
	BalanceAfter  int  `json:"balance_after"`
	IsPaidOff     bool `json:"is_paid_off"`
}

// ─── Strategy / Payoff Calculation ───────────────────────────────────────────

type DebtOrderItem struct {
	DebtID       uint    `json:"debt_id"`
	Name         string  `json:"name"`
	Balance      int     `json:"balance,omitempty"`
	InterestRate float64 `json:"interest_rate,omitempty"`
}

type PayoffMilestoneItem struct {
	Date  string `json:"date"`
	Event string `json:"event"`
}

type PayoffStrategyItem struct {
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	PayoffDate    string                `json:"payoff_date"`
	TotalInterest int                   `json:"total_interest"`
	InterestSaved int                   `json:"interest_saved"`
	MonthsSaved   int                   `json:"months_saved"`
	Order         []DebtOrderItem       `json:"order"`
	Milestones    []PayoffMilestoneItem `json:"milestones,omitempty"`
}

type StrategyRecommendation struct {
	Strategy string `json:"strategy"`
	Reason   string `json:"reason"`
}

type StrategiesResponse struct {
	CurrentMonthlyPayment int                    `json:"current_monthly_payment"`
	ExtraAvailable        int                    `json:"extra_available"`
	Strategies            []PayoffStrategyItem   `json:"strategies"`
	Recommendation        StrategyRecommendation `json:"recommendation"`
}

type TimelineMonth struct {
	Month          string `json:"month"`
	Payment        int    `json:"payment"`
	PrincipalPaid  int    `json:"principal_paid"`
	InterestPaid   int    `json:"interest_paid"`
	RemainingBalance int  `json:"remaining_balance"`
}

type TimelineResponse struct {
	DebtID       uint            `json:"debt_id"`
	Strategy     string          `json:"strategy"`
	PayoffDate   string          `json:"payoff_date"`
	TotalInterest int            `json:"total_interest"`
	Timeline     []TimelineMonth `json:"timeline"`
}
