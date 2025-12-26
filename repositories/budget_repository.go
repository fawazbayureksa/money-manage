package repositories

import (
	"my-api/dto"
	"my-api/models"
	"time"
	"gorm.io/gorm"
)

type BudgetRepository interface {
	Create(budget *models.Budget) error
	FindByID(id uint, userID uint) (*models.Budget, error)
	FindAll(userID uint, filter *dto.BudgetFilterRequest) ([]models.Budget, int64, error)
	Update(budget *models.Budget) error
	Delete(id uint, userID uint) error
	FindActiveBudgets(userID uint) ([]models.Budget, error)
	GetSpentAmount(budgetID uint, startDate, endDate time.Time) (int, error)
	FindBudgetByCategory(userID, categoryID uint, startDate, endDate time.Time) (*models.Budget, error)
	
	// Budget Alerts
	CreateAlert(alert *models.BudgetAlert) error
	GetUserAlerts(userID uint, unreadOnly bool) ([]models.BudgetAlert, error)
	MarkAlertAsRead(alertID uint, userID uint) error
}

type budgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Create(budget *models.Budget) error {
	return r.db.Create(budget).Error
}

func (r *budgetRepository) FindByID(id uint, userID uint) (*models.Budget, error) {
	var budget models.Budget
	err := r.db.Preload("Category").
		Where("id = ? AND user_id = ?", id, userID).
		First(&budget).Error
	return &budget, err
}

func (r *budgetRepository) FindAll(userID uint, filter *dto.BudgetFilterRequest) ([]models.Budget, int64, error) {
	var budgets []models.Budget
	var total int64

	query := r.db.Model(&models.Budget{}).Where("user_id = ?", userID)

	if filter.CategoryID != 0 {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.Period != "" {
		query = query.Where("period = ?", filter.Period)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		query = query.Where("description LIKE ?", "%"+filter.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	query = query.Order(sortBy + " " + filter.SortDir)
	query = query.Offset(filter.GetOffset()).Limit(filter.PageSize)

	err := query.Preload("Category").Find(&budgets).Error
	return budgets, total, err
}

func (r *budgetRepository) Update(budget *models.Budget) error {
	return r.db.Save(budget).Error
}

func (r *budgetRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Budget{}).Error
}

func (r *budgetRepository) FindActiveBudgets(userID uint) ([]models.Budget, error) {
	var budgets []models.Budget
	now := time.Now()
	err := r.db.Preload("Category").
		Where("user_id = ? AND is_active = ? AND start_date <= ? AND end_date >= ?", 
			userID, true, now, now).
		Find(&budgets).Error
	return budgets, err
}

func (r *budgetRepository) GetSpentAmount(budgetID uint, startDate, endDate time.Time) (int, error) {
	var budget models.Budget
	if err := r.db.First(&budget, budgetID).Error; err != nil {
		return 0, err
	}

	var total int64
	err := r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND category_id = ? AND transaction_type = ? AND date BETWEEN ? AND ?",
			budget.UserID, budget.CategoryID, 2, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error

	return int(total), err
}

func (r *budgetRepository) FindBudgetByCategory(userID, categoryID uint, startDate, endDate time.Time) (*models.Budget, error) {
	var budget models.Budget
	err := r.db.Where("user_id = ? AND category_id = ? AND start_date <= ? AND end_date >= ?",
		userID, categoryID, endDate, startDate).
		First(&budget).Error
	return &budget, err
}

func (r *budgetRepository) CreateAlert(alert *models.BudgetAlert) error {
	return r.db.Create(alert).Error
}

func (r *budgetRepository) GetUserAlerts(userID uint, unreadOnly bool) ([]models.BudgetAlert, error) {
	var alerts []models.BudgetAlert
	query := r.db.Preload("Budget.Category").Where("user_id = ?", userID)
	
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	
	err := query.Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *budgetRepository) MarkAlertAsRead(alertID uint, userID uint) error {
	return r.db.Model(&models.BudgetAlert{}).
		Where("id = ? AND user_id = ?", alertID, userID).
		Update("is_read", true).Error
}
