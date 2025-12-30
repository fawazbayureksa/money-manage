package services

import (
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"time"
)

type TransactionService interface {
	GetTransactions(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID *uint) ([]dto.TransactionResponse, *dto.PaginationResponse, error)
	GetTransactionByID(id, userID uint) (*dto.TransactionResponse, error)
	CreateTransaction(transaction *models.Transaction) error
	UpdateTransaction(transaction *models.Transaction) error
	DeleteTransaction(id, userID uint) error
}

type transactionService struct {
	transactionRepo repositories.TransactionRepository
}

func NewTransactionService(transactionRepo repositories.TransactionRepository) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
	}
}

func (s *transactionService) GetTransactions(userID uint, page, limit int, startDate, endDate *time.Time, transactionType *int, categoryID, bankID *uint) ([]dto.TransactionResponse, *dto.PaginationResponse, error) {
	transactions, total, err := s.transactionRepo.GetAll(userID, page, limit, startDate, endDate, transactionType, categoryID, bankID)
	if err != nil {
		return nil, nil, err
	}

	// Convert to DTO
	transactionResponses := make([]dto.TransactionResponse, len(transactions))
	for i, t := range transactions {
		transactionResponses[i] = dto.TransactionResponse{
			ID:              t.ID,
			Description:     t.Description,
			Amount:          t.Amount,
			TransactionType: t.TransactionType,
			Date:            t.Date,
			CategoryName:    t.Category.CategoryName,
			BankName:        t.Bank.BankName,
		}
	}

	// Create pagination response
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

func (s *transactionService) GetTransactionByID(id, userID uint) (*dto.TransactionResponse, error) {
	transaction, err := s.transactionRepo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	response := &dto.TransactionResponse{
		ID:              transaction.ID,
		Description:     transaction.Description,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		Date:            transaction.Date,
		CategoryName:    transaction.Category.CategoryName,
		BankName:        transaction.Bank.BankName,
	}

	return response, nil
}

func (s *transactionService) CreateTransaction(transaction *models.Transaction) error {
	return s.transactionRepo.Create(transaction)
}

func (s *transactionService) UpdateTransaction(transaction *models.Transaction) error {
	return s.transactionRepo.Update(transaction)
}

func (s *transactionService) DeleteTransaction(id, userID uint) error {
	return s.transactionRepo.Delete(id, userID)
}
