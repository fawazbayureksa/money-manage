package repositories

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"my-api/models"
	"time"
)

type TransactionV2Repository interface {
	GetAll(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID *uint, assetID *uint64) ([]models.TransactionV2, int64, error)
	GetByID(id, userID uint) (*models.TransactionV2, error)
	GetByIDWithAsset(id, userID uint) (*models.TransactionV2, error)
	CreateWithBalanceUpdate(transaction *models.TransactionV2) error
	UpdateWithBalanceUpdate(transaction *models.TransactionV2, oldAmount int, oldType int) error
	DeleteWithBalanceRollback(id, userID uint) error
	GetByAssetID(assetID uint64, userID uint, page, limit int) ([]models.TransactionV2, int64, error)
}

type transactionV2Repository struct {
	db *gorm.DB
}

func NewTransactionV2Repository(db *gorm.DB) TransactionV2Repository {
	return &transactionV2Repository{db: db}
}

func (r *transactionV2Repository) GetAll(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID *uint, assetID *uint64) ([]models.TransactionV2, int64, error) {
	var transactions []models.TransactionV2
	var total int64

	query := r.db.Model(&models.TransactionV2{}).Where("user_id = ?", userID)

	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}
	if transactionType != nil {
		query = query.Where("transaction_type = ?", *transactionType)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if assetID != nil {
		query = query.Where("asset_id = ?", *assetID)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("Category").
		Preload("Bank").
		Preload("Asset").
		Order("date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *transactionV2Repository) GetByID(id, userID uint) (*models.TransactionV2, error) {
	var transaction models.TransactionV2
	err := r.db.
		Preload("Category").
		Preload("Bank").
		Preload("Asset").
		Where("id = ? AND user_id = ?", id, userID).
		First(&transaction).Error

	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionV2Repository) GetByIDWithAsset(id, userID uint) (*models.TransactionV2, error) {
	return r.GetByID(id, userID)
}

func (r *transactionV2Repository) CreateWithBalanceUpdate(transaction *models.TransactionV2) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var asset models.Asset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&asset, transaction.AssetID).Error; err != nil {
			return errors.New("asset not found")
		}

		if asset.UserID != uint64(transaction.UserID) {
			return errors.New("unauthorized: asset does not belong to user")
		}

		if transaction.TransactionType == 2 && asset.Balance < float64(transaction.Amount) {
			return errors.New("insufficient balance")
		}

		if transaction.TransactionType == 1 {
			asset.Balance += float64(transaction.Amount)
		} else {
			asset.Balance -= float64(transaction.Amount)
		}

		if err := tx.Save(&asset).Error; err != nil {
			return err
		}

		return tx.Create(transaction).Error
	})
}

func (r *transactionV2Repository) UpdateWithBalanceUpdate(transaction *models.TransactionV2, oldAmount int, oldType int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var asset models.Asset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&asset, transaction.AssetID).Error; err != nil {
			return errors.New("asset not found")
		}

		if asset.UserID != uint64(transaction.UserID) {
			return errors.New("unauthorized: asset does not belong to user")
		}

		if oldType == 1 {
			asset.Balance -= float64(oldAmount)
		} else {
			asset.Balance += float64(oldAmount)
		}

		if transaction.TransactionType == 2 && asset.Balance < float64(transaction.Amount) {
			return errors.New("insufficient balance")
		}

		if transaction.TransactionType == 1 {
			asset.Balance += float64(transaction.Amount)
		} else {
			asset.Balance -= float64(transaction.Amount)
		}

		if err := tx.Save(&asset).Error; err != nil {
			return err
		}

		return tx.Save(transaction).Error
	})
}

func (r *transactionV2Repository) DeleteWithBalanceRollback(id, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var transaction models.TransactionV2
		if err := tx.Where("id = ? AND user_id = ?", id, userID).
			First(&transaction).Error; err != nil {
			return err
		}

		var asset models.Asset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&asset, transaction.AssetID).Error; err != nil {
			return err
		}

		if transaction.TransactionType == 1 {
			asset.Balance -= float64(transaction.Amount)
		} else {
			asset.Balance += float64(transaction.Amount)
		}

		if err := tx.Save(&asset).Error; err != nil {
			return err
		}

		return tx.Delete(&transaction).Error
	})
}

func (r *transactionV2Repository) GetByAssetID(assetID uint64, userID uint, page, limit int) ([]models.TransactionV2, int64, error) {
	var transactions []models.TransactionV2
	var total int64

	query := r.db.Model(&models.TransactionV2{}).
		Where("asset_id = ? AND user_id = ?", assetID, userID)

	query.Count(&total)

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
