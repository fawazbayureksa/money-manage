package dto

import "time"

type AnalyticsRequest struct {
	StartDate time.Time `form:"start_date" binding:"required"`
	EndDate   time.Time `form:"end_date" binding:"required"`
	GroupBy   string    `form:"group_by" binding:"omitempty,oneof=day week month year"`
}

type SpendingByCategoryResponse struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  int     `json:"total_amount"`
	Percentage   float64 `json:"percentage"`
	Count        int     `json:"count"`
}

type IncomeVsExpenseResponse struct {
	TotalIncome   int     `json:"total_income"`
	TotalExpense  int     `json:"total_expense"`
	NetAmount     int     `json:"net_amount"`
	IncomeCount   int     `json:"income_count"`
	ExpenseCount  int     `json:"expense_count"`
	SavingsRate   float64 `json:"savings_rate"`
}

type TrendDataPoint struct {
	Date    string `json:"date"`
	Income  int    `json:"income"`
	Expense int    `json:"expense"`
	Net     int    `json:"net"`
}

type TrendAnalysisResponse struct {
	Period     string          `json:"period"`
	DataPoints []TrendDataPoint `json:"data_points"`
	Summary    IncomeVsExpenseResponse `json:"summary"`
}

type SpendingByBankResponse struct {
	BankID      uint    `json:"bank_id"`
	BankName    string  `json:"bank_name"`
	TotalAmount int     `json:"total_amount"`
	Percentage  float64 `json:"percentage"`
	Count       int     `json:"count"`
}

type MonthlyComparisonResponse struct {
	Month        string `json:"month"`
	Income       int    `json:"income"`
	Expense      int    `json:"expense"`
	Net          int    `json:"net"`
	IncomeChange float64 `json:"income_change"` // % change from previous month
	ExpenseChange float64 `json:"expense_change"`
}

type DashboardSummaryResponse struct {
	CurrentMonth    IncomeVsExpenseResponse     `json:"current_month"`
	LastMonth       IncomeVsExpenseResponse     `json:"last_month"`
	TopCategories   []SpendingByCategoryResponse `json:"top_categories"`
	RecentTransactions []TransactionResponse    `json:"recent_transactions"`
	BudgetSummary   BudgetSummaryResponse       `json:"budget_summary"`
}

type BudgetSummaryResponse struct {
	TotalBudgets    int     `json:"total_budgets"`
	ActiveBudgets   int     `json:"active_budgets"`
	ExceededBudgets int     `json:"exceeded_budgets"`
	WarningBudgets  int     `json:"warning_budgets"`
	TotalBudgeted   int     `json:"total_budgeted"`
	TotalSpent      int     `json:"total_spent"`
	AverageUtilization float64 `json:"average_utilization"`
}

type TransactionResponse struct {
	ID              uint      `json:"id"`
	Description     string    `json:"description"`
	Amount          int       `json:"amount"`
	TransactionType int       `json:"transaction_type"`
	Date            time.Time `json:"date"`
	CategoryName    string    `json:"category_name"`
	BankName        string    `json:"bank_name"`
}

type YearlyReportResponse struct {
	Year             int                         `json:"year"`
	TotalIncome      int                         `json:"total_income"`
	TotalExpense     int                         `json:"total_expense"`
	NetSavings       int                         `json:"net_savings"`
	MonthlyBreakdown []MonthlyComparisonResponse `json:"monthly_breakdown"`
	TopExpenseCategories []SpendingByCategoryResponse `json:"top_expense_categories"`
	TopIncomeCategories  []SpendingByCategoryResponse `json:"top_income_categories"`
}

type CategoryTrendResponse struct {
	CategoryID   uint              `json:"category_id"`
	CategoryName string            `json:"category_name"`
	DataPoints   []TrendDataPoint  `json:"data_points"`
	TotalAmount  int               `json:"total_amount"`
	AverageAmount float64          `json:"average_amount"`
}
