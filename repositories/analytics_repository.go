package repositories

import (
	"gorm.io/gorm"
	"my-api/models"
	"time"
)

type AnalyticsRepository interface {
	GetTransactionsByDateRange(userID uint, startDate, endDate time.Time, assetID *uint64) ([]models.TransactionV2, error)
	GetSpendingByCategory(userID uint, startDate, endDate time.Time, transactionType int, assetID *uint64) ([]map[string]interface{}, error)
	GetSpendingByBank(userID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error)
	GetSpendingByAsset(userID uint, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetIncomeVsExpense(userID uint, startDate, endDate time.Time, assetID *uint64) (map[string]interface{}, error)
	GetMonthlyTrend(userID uint, startDate, endDate time.Time, assetID *uint64) ([]map[string]interface{}, error)
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
		Select("COALESCE(banks.id, 0) as bank_id, COALESCE(banks.bank_name, 'No Bank') as bank_name, SUM(transactions.amount) as total_amount, COUNT(*) as count").
		Joins("LEFT JOIN banks ON transactions.bank_id = banks.id").
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
// GetSpendingByAsset returns spending grouped by asset/wallet
func (r *analyticsRepository) GetSpendingByAsset(userID uint, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.Table("transactions").
		Select(`
			assets.id as asset_id, 
			assets.name as asset_name, 
			assets.type as asset_type,
			assets.currency as asset_currency,
			SUM(CASE WHEN transactions.transaction_type = 1 THEN transactions.amount ELSE 0 END) as total_income,
			SUM(CASE WHEN transactions.transaction_type = 2 THEN transactions.amount ELSE 0 END) as total_expense,
			COUNT(*) as transaction_count
		`).
		Joins("INNER JOIN assets ON transactions.asset_id = assets.id").
		Where("transactions.user_id = ? AND transactions.date BETWEEN ? AND ?", userID, startDate, endDate).
		Group("assets.id, assets.name, assets.type, assets.currency").
		Order("total_expense DESC").
		Scan(&results).Error

	return results, err
}
func (r *analyticsRepository) GetIncomeVsExpense(userID uint, startDate, endDate time.Time, assetID *uint64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Optimize with a single query instead of 4 separate queries
	type QueryResult struct {
		TotalIncome   int64 `gorm:"column:total_income"`
		TotalExpense  int64 `gorm:"column:total_expense"`
		IncomeCount   int64 `gorm:"column:income_count"`
		ExpenseCount  int64 `gorm:"column:expense_count"`
	}

	var queryResult QueryResult
	query := r.db.Model(&models.TransactionV2{}).
		Select(`
			COALESCE(SUM(CASE WHEN transaction_type = 1 THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN transaction_type = 2 THEN amount ELSE 0 END), 0) as total_expense,
			COUNT(CASE WHEN transaction_type = 1 THEN 1 END) as income_count,
			COUNT(CASE WHEN transaction_type = 2 THEN 1 END) as expense_count
		`).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate)

	if assetID != nil {
		query = query.Where("asset_id = ?", *assetID)
	}

	err := query.Scan(&queryResult).Error
	if err != nil {
		return nil, err
	}

	result["total_income"] = queryResult.TotalIncome
	result["total_expense"] = queryResult.TotalExpense
	result["income_count"] = queryResult.IncomeCount
	result["expense_count"] = queryResult.ExpenseCount
	result["net_amount"] = queryResult.TotalIncome - queryResult.TotalExpense

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
