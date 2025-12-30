package services

import (
	"errors"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"gorm.io/gorm"
)

type BankService interface {
	GetAllBanks(filter *dto.BankFilterRequest) (*dto.PaginationResponse, error)
	GetBankByID(id uint) (*dto.BankResponse, error)
	CreateBank(req *dto.CreateBankRequest) (*dto.BankResponse, error)
	DeleteBank(id uint) error
}

type bankService struct {
	repo repositories.BankRepository
}

func NewBankService(repo repositories.BankRepository) BankService {
	return &bankService{repo: repo}
}

func (s *bankService) GetAllBanks(filter *dto.BankFilterRequest) (*dto.PaginationResponse, error) {
	filter.SetDefaults()

	banks, total, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	bankResponses := make([]dto.BankResponse, len(banks))
	for i, bank := range banks {
		bankResponses[i] = s.toBankResponse(&bank)
	}

	return dto.NewPaginationResponse(bankResponses, filter.Page, filter.PageSize, total), nil
}

func (s *bankService) GetBankByID(id uint) (*dto.BankResponse, error) {
	bank, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bank not found")
		}
		return nil, err
	}

	response := s.toBankResponse(bank)
	return &response, nil
}

func (s *bankService) CreateBank(req *dto.CreateBankRequest) (*dto.BankResponse, error) {
	bank := &models.Bank{
		BankName: req.BankName,
		Color:    req.Color,
		Image:    req.Image,
	}

	if err := s.repo.Create(bank); err != nil {
		return nil, err
	}

	response := s.toBankResponse(bank)
	return &response, nil
}

func (s *bankService) DeleteBank(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bank not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

// Helper function to convert model to response DTO
func (s *bankService) toBankResponse(bank *models.Bank) dto.BankResponse {
	return dto.BankResponse{
		ID:       bank.ID,
		BankName: bank.BankName,
		Color:    bank.Color,
		Image:    bank.Image,
	}
}
