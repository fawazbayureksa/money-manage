package repositories

import (
	"my-api/dto"
	"my-api/models"
	"gorm.io/gorm"
)

type BankRepository interface {
	FindAll(filter *dto.BankFilterRequest) ([]models.Bank, int64, error)
	FindByID(id uint) (*models.Bank, error)
	Create(bank *models.Bank) error
	Delete(id uint) error
}

type bankRepository struct {
	db *gorm.DB
}

func NewBankRepository(db *gorm.DB) BankRepository {
	return &bankRepository{db: db}
}

func (r *bankRepository) FindAll(filter *dto.BankFilterRequest) ([]models.Bank, int64, error) {
	var banks []models.Bank
	var total int64

	query := r.db.Model(&models.Bank{})

	// Apply filters
	if filter.BankName != "" {
		query = query.Where("bank_name LIKE ?", "%"+filter.BankName+"%")
	}
	if filter.Color != "" {
		query = query.Where("color = ?", filter.Color)
	}
	if filter.Search != "" {
		query = query.Where("bank_name LIKE ?", "%"+filter.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "id"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	query = query.Order(sortBy + " " + filter.SortDir)

	// Apply pagination
	query = query.Offset(filter.GetOffset()).Limit(filter.PageSize)

	if err := query.Find(&banks).Error; err != nil {
		return nil, 0, err
	}

	return banks, total, nil
}

func (r *bankRepository) FindByID(id uint) (*models.Bank, error) {
	var bank models.Bank
	if err := r.db.First(&bank, id).Error; err != nil {
		return nil, err
	}
	return &bank, nil
}

func (r *bankRepository) Create(bank *models.Bank) error {
	return r.db.Create(bank).Error
}

func (r *bankRepository) Delete(id uint) error {
	return r.db.Delete(&models.Bank{}, id).Error
}
