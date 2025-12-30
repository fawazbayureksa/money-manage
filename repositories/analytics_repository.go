package repositories

import (
	"my-api/models"
	"time"
	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	GetTransactionsByDateRange(userID uint, startDate, endDate time.Time) ([]models.Transaction, error)
	GetSpendingByCategory(userID uint, startDate, endDate time.Time, transactionType int) ([]map[string]interface{}, error)
	GetSpendingByBank(userID uint, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetIncomeVsExpense(userID uint, startDate, endDate time.Time) (map[string]interface{}, error)
	GetMonthlyTrend(userID uint, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetRecentTransactions(userID uint, limit int) ([]models.Transaction, error)
	GetCategoryTrend(userID uint, categoryID uint, startDate, endDate time.Time) ([]map[string]interface{}, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetTransactionsByDateRange(userID uint, startDate, endDate time.Time) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("date DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *analyticsRepository) GetSpendingByCategory(userID uint, startDate, endDate time.Time, transactionType int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	err := r.db.Table("transactions").
		Select("categories.id as category_id, categories.category_name, SUM(transactions.amount) as total_amount, COUNT(*) as count").
		Joins("JOIN categories ON transactions.category_id = categories.id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.date BETWEEN ? AND ?",
			userID, transactionType, startDate, endDate).
		Group("categories.id, categories.category_name").
		Order("total_amount DESC").
		Scan(&results).Error

	return results, err
}

func (r *analyticsRepository) GetSpendingByBank(userID uint, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	err := r.db.Table("transactions").
		Select("banks.id as bank_id, banks.bank_name, SUM(transactions.amount) as total_amount, COUNT(*) as count").
		Joins("JOIN banks ON transactions.bank_id = banks.id").
		Where("transactions.user_id = ? AND transactions.date BETWEEN ? AND ?",
			userID, startDate, endDate).
		Group("banks.id, banks.bank_name").
		Order("total_amount DESC").
		Scan(&results).Error

	return results, err
}

func (r *analyticsRepository) GetIncomeVsExpense(userID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	var income, expense int64
	var incomeCount, expenseCount int64

	r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND date BETWEEN ? AND ?", userID, 1, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&income)

	r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND date BETWEEN ? AND ?", userID, 1, startDate, endDate).
		Count(&incomeCount)

	r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND date BETWEEN ? AND ?", userID, 2, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&expense)

	r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND date BETWEEN ? AND ?", userID, 2, startDate, endDate).
		Count(&expenseCount)

	result["total_income"] = income
	result["total_expense"] = expense
	result["income_count"] = incomeCount
	result["expense_count"] = expenseCount
	result["net_amount"] = income - expense

	return result, nil
}

func (r *analyticsRepository) GetMonthlyTrend(userID uint, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	err := r.db.Raw(`
		SELECT 
			DATE_FORMAT(date, '%Y-%m') as month,
			SUM(CASE WHEN transaction_type = 1 THEN amount ELSE 0 END) as income,
			SUM(CASE WHEN transaction_type = 2 THEN amount ELSE 0 END) as expense
		FROM transactions
		WHERE user_id = ? AND date BETWEEN ? AND ?
		GROUP BY DATE_FORMAT(date, '%Y-%m')
		ORDER BY month ASC
	`, userID, startDate, endDate).Scan(&results).Error

	return results, err
}

func (r *analyticsRepository) GetRecentTransactions(userID uint, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Preload("Category").Preload("Bank").
		Where("user_id = ?", userID).
		Order("date DESC, created_at DESC").
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}

func (r *analyticsRepository) GetCategoryTrend(userID uint, categoryID uint, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	err := r.db.Raw(`
		SELECT 
			DATE(date) as date,
			SUM(amount) as amount
		FROM transactions
		WHERE user_id = ? AND category_id = ? AND date BETWEEN ? AND ?
		GROUP BY DATE(date)
		ORDER BY date ASC
	`, userID, categoryID, startDate, endDate).Scan(&results).Error

	return results, err
}
