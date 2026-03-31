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
	AddTagsToTransaction(transactionID uint, tagIDs []uint) error
	RemoveTagFromTransaction(transactionID uint, tagID uint) error
	ReplaceTagsOnTransaction(transactionID uint, tagIDs []uint) error
	CreateSplits(tx interface{}, transactionID uint, splits []models.TransactionSplit) error
	DeleteSplitsByTransactionID(transactionID uint) error
	GetSplitsByTransactionID(transactionID uint) ([]models.TransactionSplit, error)
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
		Preload("Tags").
		Preload("Splits.Category").
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
		Preload("Tags").
		Preload("Splits.Category").
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

		// Omit Splits from the parent Create so GORM doesn't try to insert them here.
		splits := transaction.Splits
		transaction.Splits = nil
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
		transaction.Splits = splits

		// Persist splits within the same transaction.
		if len(splits) > 0 {
			for i := range splits {
				splits[i].TransactionID = transaction.ID
			}
			if err := tx.Create(&splits).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *transactionV2Repository) UpdateWithBalanceUpdate(transaction *models.TransactionV2, oldAmount int, oldType int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// No asset linked — just save without balance adjustment
		if transaction.AssetID == 0 {
			if err := tx.Omit("created_at").Save(transaction).Error; err != nil {
				return err
			}
			return r.replaceSplitsInTx(tx, transaction.ID, transaction.Splits)
		}

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

		splits := transaction.Splits
		transaction.Splits = nil
		if err := tx.Omit("created_at").Save(transaction).Error; err != nil {
			return err
		}
		transaction.Splits = splits

		return r.replaceSplitsInTx(tx, transaction.ID, splits)
	})
}

// replaceSplitsInTx deletes existing splits and inserts the new set within tx.
// A nil splits slice means no replacement is performed (leave existing as-is).
func (r *transactionV2Repository) replaceSplitsInTx(tx *gorm.DB, transactionID uint, splits []models.TransactionSplit) error {
	if splits == nil {
		return nil
	}
	if err := tx.Where("transaction_id = ?", transactionID).Delete(&models.TransactionSplit{}).Error; err != nil {
		return err
	}
	if len(splits) > 0 {
		for i := range splits {
			splits[i].ID = 0
			splits[i].TransactionID = transactionID
		}
		if err := tx.Create(&splits).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *transactionV2Repository) DeleteWithBalanceRollback(id, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var transaction models.TransactionV2
		if err := tx.Where("id = ? AND user_id = ?", id, userID).
			First(&transaction).Error; err != nil {
			return err
		}

		// No asset linked — just delete without balance rollback
		if transaction.AssetID == 0 {
			return tx.Delete(&transaction).Error
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
		Preload("Tags").
		Preload("Splits.Category").
		Order("date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *transactionV2Repository) AddTagsToTransaction(transactionID uint, tagIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var transaction models.TransactionV2
		if err := tx.First(&transaction, transactionID).Error; err != nil {
			return err
		}

		// Get the tags
		var tags []models.Tag
		if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
			return err
		}

		// Associate tags with transaction
		if err := tx.Model(&transaction).Association("Tags").Append(&tags); err != nil {
			return err
		}

		return nil
	})
}

func (r *transactionV2Repository) ReplaceTagsOnTransaction(transactionID uint, tagIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var transaction models.TransactionV2
		if err := tx.First(&transaction, transactionID).Error; err != nil {
			return err
		}
		var tags []models.Tag
		if len(tagIDs) > 0 {
			if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
				return err
			}
		}
		return tx.Model(&transaction).Association("Tags").Replace(&tags)
	})
}

func (r *transactionV2Repository) RemoveTagFromTransaction(transactionID uint, tagID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var transaction models.TransactionV2
		if err := tx.First(&transaction, transactionID).Error; err != nil {
			return err
		}

		var tag models.Tag
		if err := tx.First(&tag, tagID).Error; err != nil {
			return err
		}

		// Remove tag from transaction
		if err := tx.Model(&transaction).Association("Tags").Delete(&tag); err != nil {
			return err
		}

		return nil
	})
}

func (r *transactionV2Repository) CreateSplits(_ interface{}, transactionID uint, splits []models.TransactionSplit) error {
	for i := range splits {
		splits[i].TransactionID = transactionID
	}
	return r.db.Create(&splits).Error
}

func (r *transactionV2Repository) DeleteSplitsByTransactionID(transactionID uint) error {
	return r.db.Where("transaction_id = ?", transactionID).Delete(&models.TransactionSplit{}).Error
}

func (r *transactionV2Repository) GetSplitsByTransactionID(transactionID uint) ([]models.TransactionSplit, error) {
	var splits []models.TransactionSplit
	err := r.db.Preload("Category").
		Where("transaction_id = ?", transactionID).
		Find(&splits).Error
	return splits, err
}
