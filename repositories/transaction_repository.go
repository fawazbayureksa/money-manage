package repositories

import (
	"my-api/models"
	"time"
	"gorm.io/gorm"
	"fmt"
)

type TransactionRepository interface {
	GetAll(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID *uint) ([]models.Transaction, int64, error)
	GetByID(id, userID uint) (*models.Transaction, error)
	Create(transaction *models.Transaction) error
	Update(transaction *models.Transaction) error
	Delete(id, userID uint) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) GetAll(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID *uint) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.db.Model(&models.Transaction{}).Where("user_id = ?", userID)

	// Apply filters
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}
	if transactionType != nil {
		fmt.Println("Filtering by transaction type:", *transactionType)
		query = query.Where("transaction_type = ?", *transactionType)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if bankID != nil {
		query = query.Where("bank_id = ?", *bankID)
	}

	// Count total
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	err := query.
		Preload("Category").
		Preload("Bank").
		Order("date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *transactionRepository) GetByID(id, userID uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.
		Preload("Category").
		Preload("Bank").
		Where("id = ? AND user_id = ?", id, userID).
		First(&transaction).Error
	
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) Create(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) Update(transaction *models.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Transaction{}).Error
}
