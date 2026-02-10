package services

import (
	"errors"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"time"
)

type TransactionV2Service interface {
	GetTransactions(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID *uint, assetID *uint64) ([]dto.TransactionV2Response, *dto.PaginationResponse, error)
	GetTransactionByID(id, userID uint) (*dto.TransactionV2Response, error)
	CreateTransaction(transaction *models.TransactionV2) error
	UpdateTransaction(transaction *models.TransactionV2, oldAmount int, oldType int) error
	DeleteTransaction(id, userID uint) error
	GetAssetTransactions(assetID uint64, userID uint, page, limit int) (*dto.AssetTransactionsResponse, error)
}

type transactionV2Service struct {
	transactionRepo repositories.TransactionV2Repository
	assetRepo       *repositories.AssetRepository
}

func NewTransactionV2Service(transactionRepo repositories.TransactionV2Repository, assetRepo *repositories.AssetRepository) TransactionV2Service {
	return &transactionV2Service{
		transactionRepo: transactionRepo,
		assetRepo:       assetRepo,
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

		transactionResponses[i] = dto.TransactionV2Response{
			ID:              t.ID,
			Description:     t.Description,
			Amount:          t.Amount,
			TransactionType: t.TransactionType,
			Date:            t.Date,
			CategoryName:    t.Category.CategoryName,
			BankName:        t.Bank.BankName,
			AssetID:         t.AssetID,
			AssetName:       assetName,
			AssetType:       assetType,
			AssetBalance:    assetBalance,
			AssetCurrency:   assetCurrency,
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

	response := &dto.TransactionV2Response{
		ID:              transaction.ID,
		Description:     transaction.Description,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		Date:            transaction.Date,
		CategoryName:    transaction.Category.CategoryName,
		BankName:        transaction.Bank.BankName,
		AssetID:         transaction.AssetID,
		AssetName:       assetName,
		AssetType:       assetType,
		AssetBalance:    assetBalance,
		AssetCurrency:   assetCurrency,
	}

	return response, nil
}

func (s *transactionV2Service) CreateTransaction(transaction *models.TransactionV2) error {
	return s.transactionRepo.CreateWithBalanceUpdate(transaction)
}

func (s *transactionV2Service) UpdateTransaction(transaction *models.TransactionV2, oldAmount int, oldType int) error {
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

		transactionResponses[i] = dto.TransactionV2Response{
			ID:              t.ID,
			Description:     t.Description,
			Amount:          t.Amount,
			TransactionType: t.TransactionType,
			Date:            t.Date,
			CategoryName:    t.Category.CategoryName,
			BankName:        t.Bank.BankName,
			AssetID:         t.AssetID,
			AssetName:       asset.Name,
			AssetType:       asset.Type,
			AssetBalance:    asset.Balance,
			AssetCurrency:   asset.Currency,
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
