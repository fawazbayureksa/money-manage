package services

import (
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"strconv"
	"time"
)

type AnalyticsService interface {
	GetSpendingByCategory(userID uint, req *dto.AnalyticsRequest) ([]dto.SpendingByCategoryResponse, error)
	GetIncomeVsExpense(userID uint, req *dto.AnalyticsRequest) (*dto.IncomeVsExpenseResponse, error)
	GetTrendAnalysis(userID uint, req *dto.AnalyticsRequest) (*dto.TrendAnalysisResponse, error)
	GetSpendingByBank(userID uint, req *dto.AnalyticsRequest) ([]dto.SpendingByBankResponse, error)
	GetMonthlyComparison(userID uint, months int) ([]dto.MonthlyComparisonResponse, error)
	GetDashboardSummary(userID uint, startDate, endDate *time.Time) (*dto.DashboardSummaryResponse, error)
	GetYearlyReport(userID uint, year int) (*dto.YearlyReportResponse, error)
	GetCategoryTrend(userID uint, categoryID uint, req *dto.AnalyticsRequest) (*dto.CategoryTrendResponse, error)
}

type analyticsService struct {
	analyticsRepo repositories.AnalyticsRepository
	budgetRepo    repositories.BudgetRepository
}

func NewAnalyticsService(analyticsRepo repositories.AnalyticsRepository, budgetRepo repositories.BudgetRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		budgetRepo:    budgetRepo,
	}
}

// Helper function to safely convert interface{} to int
func toInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		// Try to parse string to int
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		return 0
	default:
		return 0
	}
}

// Helper function to safely convert interface{} to uint
func toUint(val interface{}) uint {
	switch v := val.(type) {
	case uint:
		return v
	case uint64:
		return uint(v)
	case int:
		return uint(v)
	case int64:
		return uint(v)
	case float64:
		return uint(v)
	default:
		return 0
	}
}

func (s *analyticsService) GetSpendingByCategory(userID uint, req *dto.AnalyticsRequest) ([]dto.SpendingByCategoryResponse, error) {
	results, err := s.analyticsRepo.GetSpendingByCategory(userID, req.StartDate, req.EndDate, 2) // 2 = expense
	if err != nil {
		return nil, err
	}

	var totalAmount int64
	for _, result := range results {
		totalAmount += int64(toInt(result["total_amount"]))
	}

	responses := make([]dto.SpendingByCategoryResponse, len(results))
	for i, result := range results {
		amount := toInt(result["total_amount"])
		percentage := float64(0)
		if totalAmount > 0 {
			percentage = float64(amount) / float64(totalAmount) * 100
		}

		responses[i] = dto.SpendingByCategoryResponse{
			CategoryID:   toUint(result["category_id"]),
			CategoryName: result["category_name"].(string),
			TotalAmount:  amount,
			Percentage:   percentage,
			Count:        toInt(result["count"]),
		}
	}

	return responses, nil
}

func (s *analyticsService) GetIncomeVsExpense(userID uint, req *dto.AnalyticsRequest) (*dto.IncomeVsExpenseResponse, error) {
	result, err := s.analyticsRepo.GetIncomeVsExpense(userID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	income := toInt(result["total_income"])
	expense := toInt(result["total_expense"])
	net := toInt(result["net_amount"])

	savingsRate := float64(0)
	if income > 0 {
		savingsRate = float64(net) / float64(income) * 100
	}

	return &dto.IncomeVsExpenseResponse{
		TotalIncome:  income,
		TotalExpense: expense,
		NetAmount:    net,
		IncomeCount:  toInt(result["income_count"]),
		ExpenseCount: toInt(result["expense_count"]),
		SavingsRate:  savingsRate,
	}, nil
}

func (s *analyticsService) GetTrendAnalysis(userID uint, req *dto.AnalyticsRequest) (*dto.TrendAnalysisResponse, error) {
	results, err := s.analyticsRepo.GetMonthlyTrend(userID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	dataPoints := make([]dto.TrendDataPoint, len(results))
	for i, result := range results {
		income := toInt(result["income"])
		expense := toInt(result["expense"])
		
		dataPoints[i] = dto.TrendDataPoint{
			Date:    result["month"].(string),
			Income:  income,
			Expense: expense,
			Net:     income - expense,
		}
	}

	summary, _ := s.GetIncomeVsExpense(userID, req)

	return &dto.TrendAnalysisResponse{
		Period:     req.GroupBy,
		DataPoints: dataPoints,
		Summary:    *summary,
	}, nil
}

func (s *analyticsService) GetSpendingByBank(userID uint, req *dto.AnalyticsRequest) ([]dto.SpendingByBankResponse, error) {
	results, err := s.analyticsRepo.GetSpendingByBank(userID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	var totalAmount int64
	for _, result := range results {
		totalAmount += int64(toInt(result["total_amount"]))
	}

	responses := make([]dto.SpendingByBankResponse, len(results))
	for i, result := range results {
		amount := toInt(result["total_amount"])
		percentage := float64(0)
		if totalAmount > 0 {
			percentage = float64(amount) / float64(totalAmount) * 100
		}

		responses[i] = dto.SpendingByBankResponse{
			BankID:      toUint(result["bank_id"]),
			BankName:    result["bank_name"].(string),
			TotalAmount: amount,
			Percentage:  percentage,
			Count:       toInt(result["count"]),
		}
	}

	return responses, nil
}

func (s *analyticsService) GetMonthlyComparison(userID uint, months int) ([]dto.MonthlyComparisonResponse, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	results, err := s.analyticsRepo.GetMonthlyTrend(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.MonthlyComparisonResponse, len(results))
	var prevIncome, prevExpense int

	for i, result := range results {
		income := toInt(result["income"])
		expense := toInt(result["expense"])

		incomeChange := float64(0)
		expenseChange := float64(0)

		if i > 0 && prevIncome > 0 {
			incomeChange = float64(income-prevIncome) / float64(prevIncome) * 100
		}
		if i > 0 && prevExpense > 0 {
			expenseChange = float64(expense-prevExpense) / float64(prevExpense) * 100
		}

		responses[i] = dto.MonthlyComparisonResponse{
			Month:         result["month"].(string),
			Income:        income,
			Expense:       expense,
			Net:           income - expense,
			IncomeChange:  incomeChange,
			ExpenseChange: expenseChange,
		}

		prevIncome = income
		prevExpense = expense
	}

	return responses, nil
}

func (s *analyticsService) GetDashboardSummary(userID uint, startDate, endDate *time.Time) (*dto.DashboardSummaryResponse, error) {
	now := time.Now()
	var currentMonthStart, currentMonthEnd time.Time
	
	// Use provided dates or default to current month
	if startDate != nil && endDate != nil {
		currentMonthStart = *startDate
		currentMonthEnd = *endDate
	} else {
		currentMonthStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		currentMonthEnd = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	}
	
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := currentMonthStart.AddDate(0, 0, -1)

	// Current month
	currentReq := &dto.AnalyticsRequest{
		StartDate: currentMonthStart,
		EndDate:   currentMonthEnd,
	}
	currentMonth, _ := s.GetIncomeVsExpense(userID, currentReq)

	// Last month
	lastReq := &dto.AnalyticsRequest{
		StartDate: lastMonthStart,
		EndDate:   lastMonthEnd,
	}
	lastMonth, _ := s.GetIncomeVsExpense(userID, lastReq)

	// Top categories
	topCategories, _ := s.GetSpendingByCategory(userID, currentReq)
	if len(topCategories) > 5 {
		topCategories = topCategories[:5]
	}

	// Recent transactions
	transactions, _ := s.analyticsRepo.GetRecentTransactions(userID, 10)
	recentTransactions := s.toTransactionResponses(transactions)

	// Budget summary
	budgetSummary := s.getBudgetSummary(userID)

	return &dto.DashboardSummaryResponse{
		CurrentMonth:       *currentMonth,
		LastMonth:          *lastMonth,
		TopCategories:      topCategories,
		RecentTransactions: recentTransactions,
		BudgetSummary:      budgetSummary,
	}, nil
}

func (s *analyticsService) GetYearlyReport(userID uint, year int) (*dto.YearlyReportResponse, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	req := &dto.AnalyticsRequest{
		StartDate: startDate,
		EndDate:   endDate,
	}

	summary, _ := s.GetIncomeVsExpense(userID, req)
	monthlyBreakdown, _ := s.GetMonthlyComparison(userID, 12)
	
	expenseCategories, _ := s.analyticsRepo.GetSpendingByCategory(userID, startDate, endDate, 2)
	incomeCategories, _ := s.analyticsRepo.GetSpendingByCategory(userID, startDate, endDate, 1)

	topExpense := s.toSpendingByCategoryResponses(expenseCategories, 10)
	topIncome := s.toSpendingByCategoryResponses(incomeCategories, 10)

	return &dto.YearlyReportResponse{
		Year:                 year,
		TotalIncome:          summary.TotalIncome,
		TotalExpense:         summary.TotalExpense,
		NetSavings:           summary.NetAmount,
		MonthlyBreakdown:     monthlyBreakdown,
		TopExpenseCategories: topExpense,
		TopIncomeCategories:  topIncome,
	}, nil
}

func (s *analyticsService) GetCategoryTrend(userID uint, categoryID uint, req *dto.AnalyticsRequest) (*dto.CategoryTrendResponse, error) {
	results, err := s.analyticsRepo.GetCategoryTrend(userID, categoryID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	dataPoints := make([]dto.TrendDataPoint, len(results))
	var totalAmount int
	
	for i, result := range results {
		amount := toInt(result["amount"])
		totalAmount += amount
		
		dataPoints[i] = dto.TrendDataPoint{
			Date:    result["date"].(string),
			Expense: amount,
		}
	}

	avgAmount := float64(0)
	if len(dataPoints) > 0 {
		avgAmount = float64(totalAmount) / float64(len(dataPoints))
	}

	return &dto.CategoryTrendResponse{
		CategoryID:    categoryID,
		CategoryName:  "", // Would need to fetch from category repo
		DataPoints:    dataPoints,
		TotalAmount:   totalAmount,
		AverageAmount: avgAmount,
	}, nil
}

// Helper functions
func (s *analyticsService) toTransactionResponses(transactions []models.Transaction) []dto.TransactionResponse {
	responses := make([]dto.TransactionResponse, len(transactions))
	for i, t := range transactions {
		categoryName := ""
		bankName := ""
		if t.Category.ID > 0 {
			categoryName = t.Category.CategoryName
		}
		if t.Bank.ID > 0 {
			bankName = t.Bank.BankName
		}

		responses[i] = dto.TransactionResponse{
			ID:              t.ID,
			Description:     t.Description,
			Amount:          t.Amount,
			TransactionType: t.TransactionType,
			Date:            t.Date,
			CategoryName:    categoryName,
			BankName:        bankName,
		}
	}
	return responses
}

func (s *analyticsService) toSpendingByCategoryResponses(results []map[string]interface{}, limit int) []dto.SpendingByCategoryResponse {
	if len(results) > limit {
		results = results[:limit]
	}

	var totalAmount int64
	for _, result := range results {
		totalAmount += int64(toInt(result["total_amount"]))
	}

	responses := make([]dto.SpendingByCategoryResponse, len(results))
	for i, result := range results {
		amount := toInt(result["total_amount"])
		percentage := float64(0)
		if totalAmount > 0 {
			percentage = float64(amount) / float64(totalAmount) * 100
		}

		responses[i] = dto.SpendingByCategoryResponse{
			CategoryID:   toUint(result["category_id"]),
			CategoryName: result["category_name"].(string),
			TotalAmount:  amount,
			Percentage:   percentage,
			Count:        toInt(result["count"]),
		}
	}

	return responses
}

func (s *analyticsService) getBudgetSummary(userID uint) dto.BudgetSummaryResponse {
	budgets, _ := s.budgetRepo.FindActiveBudgets(userID)
	
	totalBudgeted := 0
	totalSpent := 0
	exceededCount := 0
	warningCount := 0
	activeCount := 0

	for _, budget := range budgets {
		if budget.IsActive {
			activeCount++
		}
		totalBudgeted += budget.Amount
		
		spent, _ := s.budgetRepo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
		totalSpent += spent
		
		percentage := float64(spent) / float64(budget.Amount) * 100
		if percentage >= 100 {
			exceededCount++
		} else if percentage >= float64(budget.AlertAt) {
			warningCount++
		}
	}

	avgUtilization := float64(0)
	if totalBudgeted > 0 {
		avgUtilization = float64(totalSpent) / float64(totalBudgeted) * 100
	}

	return dto.BudgetSummaryResponse{
		TotalBudgets:       len(budgets),
		ActiveBudgets:      activeCount,
		ExceededBudgets:    exceededCount,
		WarningBudgets:     warningCount,
		TotalBudgeted:      totalBudgeted,
		TotalSpent:         totalSpent,
		AverageUtilization: avgUtilization,
	}
}
