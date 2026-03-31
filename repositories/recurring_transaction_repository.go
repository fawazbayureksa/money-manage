package repositories

import (
	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
	"time"
)

type RecurringTransactionRepository interface {
	Create(rt *models.RecurringTransaction) error
	FindByID(id uint64, userID uint) (*models.RecurringTransaction, error)
	FindAll(userID uint, filter *dto.RecurringTransactionFilterRequest) ([]models.RecurringTransaction, int64, error)
	FindActiveByUserID(userID uint) ([]models.RecurringTransaction, error)
	FindUpcoming(userID uint, from, to time.Time) ([]models.RecurringTransaction, error)
	Update(rt *models.RecurringTransaction) error
	Delete(id uint64, userID uint) error
}

type recurringTransactionRepository struct {
	db *gorm.DB
}

func NewRecurringTransactionRepository(db *gorm.DB) RecurringTransactionRepository {
	return &recurringTransactionRepository{db: db}
}

func (r *recurringTransactionRepository) Create(rt *models.RecurringTransaction) error {
	return r.db.Create(rt).Error
}

func (r *recurringTransactionRepository) FindByID(id uint64, userID uint) (*models.RecurringTransaction, error) {
	var rt models.RecurringTransaction
	err := r.db.Preload("Asset").Preload("Category").
		Where("id = ? AND user_id = ?", id, userID).
		First(&rt).Error
	return &rt, err
}

func (r *recurringTransactionRepository) FindAll(userID uint, filter *dto.RecurringTransactionFilterRequest) ([]models.RecurringTransaction, int64, error) {
	var records []models.RecurringTransaction
	var total int64

	query := r.db.Model(&models.RecurringTransaction{}).Where("user_id = ?", userID)

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Frequency != "" {
		query = query.Where("frequency = ?", filter.Frequency)
	}
	if filter.AssetID != nil {
		query = query.Where("asset_id = ?", *filter.AssetID)
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

	err := query.Preload("Asset").Preload("Category").Find(&records).Error
	return records, total, err
}

func (r *recurringTransactionRepository) FindActiveByUserID(userID uint) ([]models.RecurringTransaction, error) {
	var records []models.RecurringTransaction
	err := r.db.Preload("Asset").Preload("Category").
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&records).Error
	return records, err
}

func (r *recurringTransactionRepository) FindUpcoming(userID uint, from, to time.Time) ([]models.RecurringTransaction, error) {
	var records []models.RecurringTransaction
	err := r.db.Preload("Asset").Preload("Category").
		Where("user_id = ? AND is_active = ? AND next_occurrence BETWEEN ? AND ?",
			userID, true, from, to).
		Order("next_occurrence ASC").
		Find(&records).Error
	return records, err
}

func (r *recurringTransactionRepository) Update(rt *models.RecurringTransaction) error {
	return r.db.Save(rt).Error
}

func (r *recurringTransactionRepository) Delete(id uint64, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.RecurringTransaction{}).Error
}
