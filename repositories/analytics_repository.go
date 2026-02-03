package repositories

import (
	"gorm.io/gorm"
	"my-api/models"
	"my-api/utils"
	"time"
)

type AnalyticsRepository interface {
	GetTransactionsByDateRange(userID uint, startDate, endDate time.Time, assetID *uint64) ([]models.TransactionV2, error)
	GetSpendingByCategory(userID uint, startDate, endDate time.Time, transactionType int, assetID *uint64) ([]map[string]interface{}, error)
	GetSpendingByBank(userID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error)
	GetIncomeVsExpense(userID uint, startDate, endDate time.Time, assetID *uint64) (map[string]interface{}, error)
	GetMonthlyTrend(userID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error)
	GetMonthlyTrendByPayCycle(userID uint, startDate, endDate time.Time, assetID *uint64, settings *models.UserSettings) ([]map[string]interface{}, error)
	GetRecentTransactions(userID uint, limit int, assetID *uint64) ([]models.TransactionV2, error)
	GetCategoryTrend(userID uint, categoryID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetTransactionsByDateRange(userID uint, startDate, endDate time.Time, assetID *uint64) ([]models.TransactionV2, error) {
	var transactions []models.TransactionV2
	query := r.db.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate)
	if assetID != nil {
		query = query.Where("asset_id = ?", *assetID)
	}
	err := query.Order("date DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *analyticsRepository) GetSpendingByCategory(userID uint, startDate, endDate time.Time, transactionType int, assetID *uint64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.Table("transactions").
		Select("categories.id as category_id, categories.category_name, SUM(transactions.amount) as total_amount, COUNT(*) as count").
		Joins("JOIN categories ON transactions.category_id = categories.id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.date BETWEEN ? AND ?",
			userID, transactionType, startDate, endDate)
	if assetID != nil {
		query = query.Where("transactions.asset_id = ?", *assetID)
	}
	err := query.Group("categories.id, categories.category_name").
		Order("total_amount DESC").
		Scan(&results).Error

	return results, err
}

func (r *analyticsRepository) GetSpendingByBank(userID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.Table("transactions").
		Select("banks.id as bank_id, banks.bank_name, SUM(transactions.amount) as total_amount, COUNT(*) as count").
		Joins("JOIN banks ON transactions.bank_id = banks.id").
		Where("transactions.user_id = ? AND transactions.date BETWEEN ? AND ?",
			userID, startDate, endDate)
	if assetID != nil {
		query = query.Where("transactions.asset_id = ?", *assetID)
	}
	err := query.Group("banks.id, banks.bank_name").
		Order("total_amount DESC").
		Scan(&results).Error

	return results, err
}

func (r *analyticsRepository) GetIncomeVsExpense(userID uint, startDate, endDate time.Time, assetID *uint64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var income, expense int64
	var incomeCount, expenseCount int64

	queryIncome := r.db.Model(&models.TransactionV2{}).
		Where("user_id = ? AND transaction_type = ? AND DATE(date) BETWEEN ? AND ?", userID, 1, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if assetID != nil {
		queryIncome = queryIncome.Where("asset_id = ?", *assetID)
	}
	queryIncome.Select("COALESCE(SUM(amount), 0)").Scan(&income)

	queryIncomeCount := r.db.Model(&models.TransactionV2{}).
		Where("user_id = ? AND transaction_type = ? AND DATE(date) BETWEEN ? AND ?", userID, 1, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if assetID != nil {
		queryIncomeCount = queryIncomeCount.Where("asset_id = ?", *assetID)
	}
	queryIncomeCount.Count(&incomeCount)

	queryExpense := r.db.Model(&models.TransactionV2{}).
		Where("user_id = ? AND transaction_type = ? AND DATE(date) BETWEEN ? AND ?", userID, 2, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if assetID != nil {
		queryExpense = queryExpense.Where("asset_id = ?", *assetID)
	}
	queryExpense.Select("COALESCE(SUM(amount), 0)").Scan(&expense)

	queryExpenseCount := r.db.Model(&models.TransactionV2{}).
		Where("user_id = ? AND transaction_type = ? AND DATE(date) BETWEEN ? AND ?", userID, 2, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if assetID != nil {
		queryExpenseCount = queryExpenseCount.Where("asset_id = ?", *assetID)
	}
	queryExpenseCount.Count(&expenseCount)

	result["total_income"] = income
	result["total_expense"] = expense
	result["income_count"] = incomeCount
	result["expense_count"] = expenseCount
	result["net_amount"] = income - expense

	return result, nil
}

func (r *analyticsRepository) GetMonthlyTrend(userID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `SELECT 
		DATE_FORMAT(date, '%Y-%m') as month,
		SUM(CASE WHEN transaction_type = 1 THEN amount ELSE 0 END) as income,
		SUM(CASE WHEN transaction_type = 2 THEN amount ELSE 0 END) as expense
	FROM transactions
	WHERE user_id = ? AND date BETWEEN ? AND ? AND category_id != ?`

	var args []interface{}
	args = append(args, userID, startDate, endDate,18)

	if assetID != nil {
		query += ` AND asset_id = ?`
		args = append(args, *assetID)
	}

	query += ` GROUP BY DATE_FORMAT(date, '%Y-%m') ORDER BY month ASC`

	err := r.db.Raw(query, args...).Scan(&results).Error

	return results, err
}

func (r *analyticsRepository) GetRecentTransactions(userID uint, limit int, assetID *uint64) ([]models.TransactionV2, error) {
	var transactions []models.TransactionV2
	query := r.db.Preload("Category").Preload("Bank").Preload("Asset").
		Where("user_id = ?", userID)
	if assetID != nil {
		query = query.Where("asset_id = ?", *assetID)
	}
	err := query.Order("date DESC, created_at DESC").
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}

func (r *analyticsRepository) GetCategoryTrend(userID uint, categoryID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `SELECT 
		DATE(date) as date,
		SUM(amount) as amount
	FROM transactions
	WHERE user_id = ? AND category_id = ? AND date BETWEEN ? AND ?`

	var args []interface{}
	args = append(args, userID, categoryID, startDate, endDate)

	if assetID != nil {
		query += ` AND asset_id = ?`
		args = append(args, *assetID)
	}

	query += ` GROUP BY DATE(date) ORDER BY date ASC`

	err := r.db.Raw(query, args...).Scan(&results).Error

	return results, err
}

// GetMonthlyTrendByPayCycle returns monthly trends based on user's financial periods
func (r *analyticsRepository) GetMonthlyTrendByPayCycle(userID uint, startDate, endDate time.Time, assetID *uint64, settings *models.UserSettings) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// Get all financial periods in the date range
	periods := utils.GetFinancialPeriods(settings, startDate, endDate)

	// Query transactions for each period
	for _, period := range periods {
		var income, expense int64

		// Get income for this period
		queryIncome := r.db.Model(&models.TransactionV2{}).
			Where("user_id = ? AND transaction_type = ? AND date BETWEEN ? AND ? AND category_id != ?",
				userID, 1, period.StartDate, period.EndDate, 18)
		if assetID != nil {
			queryIncome = queryIncome.Where("asset_id = ?", *assetID)
		}
		queryIncome.Select("COALESCE(SUM(amount), 0)").Scan(&income)

		// Get expense for this period
		queryExpense := r.db.Model(&models.TransactionV2{}).
			Where("user_id = ? AND transaction_type = ? AND date BETWEEN ? AND ? AND category_id != ?",
				userID, 2, period.StartDate, period.EndDate, 18)
		if assetID != nil {
			queryExpense = queryExpense.Where("asset_id = ?", *assetID)
		}
		queryExpense.Select("COALESCE(SUM(amount), 0)").Scan(&expense)

		results = append(results, map[string]interface{}{
			"period":       period.PeriodLabel,
			"period_start": period.StartDate.Format("2006-01-02"),
			"period_end":   period.EndDate.Format("2006-01-02"),
			"income":       income,
			"expense":      expense,
		})
	}

	return results, nil
}
