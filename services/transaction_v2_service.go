package services

import (
	"errors"
	"log"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"time"
)

type TransactionV2Service interface {
	GetTransactions(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID *uint, assetID *uint64) ([]dto.TransactionV2Response, *dto.PaginationResponse, error)
	GetTransactionByID(id, userID uint) (*dto.TransactionV2Response, error)
	CreateTransaction(transaction *models.TransactionV2, splits []dto.SplitItem) error
	UpdateTransaction(transaction *models.TransactionV2, oldAmount int, oldType int, splits *[]dto.SplitItem) error
	DeleteTransaction(id, userID uint) error
	GetAssetTransactions(assetID uint64, userID uint, page, limit int) (*dto.AssetTransactionsResponse, error)
	AddTagsToTransaction(transactionID, userID uint, tagIDs []uint) error
	RemoveTagFromTransaction(transactionID, userID, tagID uint) error
	ReplaceTagsOnTransaction(transactionID, userID uint, tagIDs []uint) error
}

type transactionV2Service struct {
	transactionRepo repositories.TransactionV2Repository
	assetRepo       *repositories.AssetRepository
	tagRepo         repositories.TagRepository
}

func NewTransactionV2Service(transactionRepo repositories.TransactionV2Repository, assetRepo *repositories.AssetRepository, tagRepo repositories.TagRepository) TransactionV2Service {
	return &transactionV2Service{
		transactionRepo: transactionRepo,
		assetRepo:       assetRepo,
		tagRepo:         tagRepo,
	}
}

func (s *transactionV2Service) GetTransactions(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID *uint, assetID *uint64) ([]dto.TransactionV2Response, *dto.PaginationResponse, error) {
	transactions, total, err := s.transactionRepo.GetAll(userID, page, limit, startDate, endDate, transactionType, categoryID, assetID)
	if err != nil {
		return nil, nil, err
	}

	transactionResponses := make([]dto.TransactionV2Response, len(transactions))
	for i, t := range transactions {
		assetName := ""
		assetType := ""
		assetBalance := 0.0
		assetCurrency := ""

		if t.Asset.ID != 0 {
			assetName = t.Asset.Name
			assetType = t.Asset.Type
			assetBalance = t.Asset.Balance
			assetCurrency = t.Asset.Currency
		}

		categoryName := ""
		if t.Category != nil {
			categoryName = t.Category.CategoryName
		}
		bankName := ""
		if t.Bank != nil {
			bankName = t.Bank.BankName
		}

		transactionResponses[i] = dto.TransactionV2Response{
			ID:              t.ID,
			Description:     t.Description,
			Amount:          t.Amount,
			TransactionType: t.TransactionType,
			Date:            t.Date,
			CategoryName:    categoryName,
			BankName:        bankName,
			AssetID:         t.AssetID,
			AssetName:       assetName,
			AssetType:       assetType,
			AssetBalance:    assetBalance,
			AssetCurrency:   assetCurrency,
			Tags:            t.Tags,
			Splits:          mapSplits(t.Splits),
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	pagination := &dto.PaginationResponse{
		Page:       page,
		PageSize:   limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	return transactionResponses, pagination, nil
}

func (s *transactionV2Service) GetTransactionByID(id, userID uint) (*dto.TransactionV2Response, error) {
	transaction, err := s.transactionRepo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	assetName := ""
	assetType := ""
	assetBalance := 0.0
	assetCurrency := ""

	if transaction.Asset.ID != 0 {
		assetName = transaction.Asset.Name
		assetType = transaction.Asset.Type
		assetBalance = transaction.Asset.Balance
		assetCurrency = transaction.Asset.Currency
	}

	catName := ""
	if transaction.Category != nil {
		catName = transaction.Category.CategoryName
	}
	bkName := ""
	if transaction.Bank != nil {
		bkName = transaction.Bank.BankName
	}

	response := &dto.TransactionV2Response{
		ID:              transaction.ID,
		Description:     transaction.Description,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		Date:            transaction.Date,
		CategoryName:    catName,
		BankName:        bkName,
		AssetID:         transaction.AssetID,
		AssetName:       assetName,
		AssetType:       assetType,
		AssetBalance:    assetBalance,
		AssetCurrency:   assetCurrency,
		Tags:            transaction.Tags,
		Splits:          mapSplits(transaction.Splits),
	}

	return response, nil
}

// mapSplits converts model splits to DTO split responses.
func mapSplits(splits []models.TransactionSplit) []dto.SplitItemResponse {
	if len(splits) == 0 {
		return nil
	}
	out := make([]dto.SplitItemResponse, len(splits))
	for i, s := range splits {
		catName := ""
		if s.Category != nil {
			catName = s.Category.CategoryName
		}
		out[i] = dto.SplitItemResponse{
			ID:           s.ID,
			CategoryID:   s.CategoryID,
			CategoryName: catName,
			Amount:       s.Amount,
			Description:  s.Description,
		}
	}
	return out
}

// buildModelSplits converts DTO split items to model splits for a given transaction.
func buildModelSplits(splits []dto.SplitItem, transactionID uint) []models.TransactionSplit {
	out := make([]models.TransactionSplit, len(splits))
	for i, s := range splits {
		out[i] = models.TransactionSplit{
			TransactionID: transactionID,
			CategoryID:    s.CategoryID,
			Amount:        s.Amount,
			Description:   s.Description,
		}
	}
	return out
}

// validateSplits checks that splits are non-empty and their amounts sum to totalAmount.
func validateSplits(splits []dto.SplitItem, totalAmount int) error {
	sum := 0
	for _, s := range splits {
		sum += s.Amount
	}
	if sum != totalAmount {
		return errors.New("split amounts must sum to the transaction total amount")
	}
	return nil
}

func (s *transactionV2Service) CreateTransaction(transaction *models.TransactionV2, splits []dto.SplitItem) error {
	if len(splits) > 0 {
		if err := validateSplits(splits, transaction.Amount); err != nil {
			return err
		}
		transaction.Splits = buildModelSplits(splits, 0)
	}
	return s.transactionRepo.CreateWithBalanceUpdate(transaction)
}

func (s *transactionV2Service) UpdateTransaction(transaction *models.TransactionV2, oldAmount int, oldType int, splits *[]dto.SplitItem) error {
	if splits != nil {
		if err := validateSplits(*splits, transaction.Amount); err != nil {
			return err
		}
		transaction.Splits = buildModelSplits(*splits, transaction.ID)
	}
	return s.transactionRepo.UpdateWithBalanceUpdate(transaction, oldAmount, oldType)
}

func (s *transactionV2Service) DeleteTransaction(id, userID uint) error {
	return s.transactionRepo.DeleteWithBalanceRollback(id, userID)
}

func (s *transactionV2Service) GetAssetTransactions(assetID uint64, userID uint, page, limit int) (*dto.AssetTransactionsResponse, error) {
	asset, err := s.assetRepo.GetAssetByID(assetID)
	if err != nil {
		return nil, err
	}

	if asset.UserID != uint64(userID) {
		return nil, errors.New("unauthorized")
	}

	transactions, total, err := s.transactionRepo.GetByAssetID(assetID, userID, page, limit)
	if err != nil {
		return nil, err
	}

	transactionResponses := make([]dto.TransactionV2Response, len(transactions))
	totalIncome := 0.0
	totalExpense := 0.0

	for i, t := range transactions {
		if t.TransactionType == 1 {
			totalIncome += float64(t.Amount)
		} else {
			totalExpense += float64(t.Amount)
		}

		assetCatName := ""
		if t.Category != nil {
			assetCatName = t.Category.CategoryName
		}
		assetBkName := ""
		if t.Bank != nil {
			assetBkName = t.Bank.BankName
		}

		transactionResponses[i] = dto.TransactionV2Response{
			ID:              t.ID,
			Description:     t.Description,
			Amount:          t.Amount,
			TransactionType: t.TransactionType,
			Date:            t.Date,
			CategoryName:    assetCatName,
			BankName:        assetBkName,
			AssetID:         t.AssetID,
			AssetName:       asset.Name,
			AssetType:       asset.Type,
			AssetBalance:    asset.Balance,
			AssetCurrency:   asset.Currency,
			Tags:            t.Tags,
			Splits:          mapSplits(t.Splits),
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &dto.AssetTransactionsResponse{
		AssetID:        asset.ID,
		AssetName:      asset.Name,
		AssetType:      asset.Type,
		CurrentBalance: asset.Balance,
		Currency:       asset.Currency,
		Transactions:   transactionResponses,
		TotalIncome:    totalIncome,
		TotalExpense:   totalExpense,
	}, nil
}

func (s *transactionV2Service) AddTagsToTransaction(transactionID, userID uint, tagIDs []uint) error {
	// Verify transaction belongs to user
	_, err := s.transactionRepo.GetByID(transactionID, userID)
	if err != nil {
		return err
	}

	// Verify all tags belong to user
	for _, tagID := range tagIDs {
		_, err := s.tagRepo.FindByID(tagID, userID)
		if err != nil {
			return errors.New("one or more tags not found or do not belong to you")
		}
	}

	// Add tags to transaction
	err = s.transactionRepo.AddTagsToTransaction(transactionID, tagIDs)
	if err != nil {
		return err
	}

	// Increment usage count for all tags
	for _, tagID := range tagIDs {
		if err := s.tagRepo.IncrementUsage(tagID); err != nil {
			// Log the error but don't fail the operation
			// as the tags are already added successfully
			log.Printf("Failed to increment usage count for tag %d: %v", tagID, err)
		}
	}

	return nil
}

func (s *transactionV2Service) RemoveTagFromTransaction(transactionID, userID, tagID uint) error {
	// Verify transaction belongs to user
	_, err := s.transactionRepo.GetByID(transactionID, userID)
	if err != nil {
		return err
	}

	return s.transactionRepo.RemoveTagFromTransaction(transactionID, tagID)
}

func (s *transactionV2Service) ReplaceTagsOnTransaction(transactionID, userID uint, tagIDs []uint) error {
	_, err := s.transactionRepo.GetByID(transactionID, userID)
	if err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if _, err := s.tagRepo.FindByID(tagID, userID); err != nil {
			return errors.New("one or more tags not found or do not belong to you")
		}
	}
	return s.transactionRepo.ReplaceTagsOnTransaction(transactionID, tagIDs)
}
