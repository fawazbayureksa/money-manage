package services

import (
	"errors"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"my-api/utils"
	"time"
)

type RecurringTransactionService interface {
	Create(userID uint, req *dto.CreateRecurringTransactionRequest) (*dto.RecurringTransactionResponse, error)
	GetByID(id uint64, userID uint) (*dto.RecurringTransactionResponse, error)
	GetAll(userID uint, filter *dto.RecurringTransactionFilterRequest) (*dto.PaginationResponse, error)
	Update(id uint64, userID uint, req *dto.UpdateRecurringTransactionRequest) (*dto.RecurringTransactionResponse, error)
	Delete(id uint64, userID uint) error
}

type recurringTransactionService struct {
	repo repositories.RecurringTransactionRepository
}

func NewRecurringTransactionService(repo repositories.RecurringTransactionRepository) RecurringTransactionService {
	return &recurringTransactionService{repo: repo}
}

func (s *recurringTransactionService) Create(userID uint, req *dto.CreateRecurringTransactionRequest) (*dto.RecurringTransactionResponse, error) {
	if err := validateFrequencyFields(req.Frequency, req.DayOfMonth, req.DayOfWeek); err != nil {
		return nil, err
	}

	next := computeNextOccurrence(req.Frequency, req.DayOfMonth, req.DayOfWeek, req.StartDate.Time)

	rt := &models.RecurringTransaction{
		UserID:          userID,
		AssetID:         req.AssetID,
		CategoryID:      req.CategoryID,
		Description:     req.Description,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		Frequency:       req.Frequency,
		DayOfMonth:      req.DayOfMonth,
		DayOfWeek:       req.DayOfWeek,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		NextOccurrence:  utils.CustomTime{Time: next},
		IsActive:        true,
	}

	if err := s.repo.Create(rt); err != nil {
		return nil, err
	}

	return s.FindByIDAndMap(rt.ID, userID)
}

func (s *recurringTransactionService) GetByID(id uint64, userID uint) (*dto.RecurringTransactionResponse, error) {
	return s.FindByIDAndMap(id, userID)
}

func (s *recurringTransactionService) GetAll(userID uint, filter *dto.RecurringTransactionFilterRequest) (*dto.PaginationResponse, error) {
	filter.SetDefaults()

	records, total, err := s.repo.FindAll(userID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.RecurringTransactionResponse, len(records))
	for i, r := range records {
		responses[i] = toRecurringTransactionResponse(r)
	}

	return dto.NewPaginationResponse(responses, filter.Page, filter.PageSize, total), nil
}

func (s *recurringTransactionService) Update(id uint64, userID uint, req *dto.UpdateRecurringTransactionRequest) (*dto.RecurringTransactionResponse, error) {
	rt, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, errors.New("recurring transaction not found")
	}

	if req.AssetID != nil {
		rt.AssetID = *req.AssetID
	}
	if req.CategoryID != nil {
		rt.CategoryID = req.CategoryID
	}
	if req.Description != "" {
		rt.Description = req.Description
	}
	if req.Amount != nil {
		rt.Amount = *req.Amount
	}
	if req.TransactionType != nil {
		rt.TransactionType = *req.TransactionType
	}
	if req.IsActive != nil {
		rt.IsActive = *req.IsActive
	}
	if req.StartDate != nil {
		rt.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		rt.EndDate = req.EndDate
	}

	// Update frequency-related fields only when frequency changes
	frequencyChanged := req.Frequency != "" && req.Frequency != rt.Frequency
	if req.Frequency != "" {
		rt.Frequency = req.Frequency
	}
	if req.DayOfMonth != nil {
		rt.DayOfMonth = req.DayOfMonth
	}
	if req.DayOfWeek != nil {
		rt.DayOfWeek = req.DayOfWeek
	}

	if frequencyChanged || req.DayOfMonth != nil || req.DayOfWeek != nil || req.StartDate != nil {
		if err := validateFrequencyFields(rt.Frequency, rt.DayOfMonth, rt.DayOfWeek); err != nil {
			return nil, err
		}
		next := computeNextOccurrence(rt.Frequency, rt.DayOfMonth, rt.DayOfWeek, rt.StartDate.Time)
		rt.NextOccurrence = utils.CustomTime{Time: next}
	}

	if err := s.repo.Update(rt); err != nil {
		return nil, err
	}

	return s.FindByIDAndMap(rt.ID, userID)
}

func (s *recurringTransactionService) Delete(id uint64, userID uint) error {
	if _, err := s.repo.FindByID(id, userID); err != nil {
		return errors.New("recurring transaction not found")
	}
	return s.repo.Delete(id, userID)
}

// FindByIDAndMap is a helper used internally.
func (s *recurringTransactionService) FindByIDAndMap(id uint64, userID uint) (*dto.RecurringTransactionResponse, error) {
	rt, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, errors.New("recurring transaction not found")
	}
	resp := toRecurringTransactionResponse(*rt)
	return &resp, nil
}

// ---- Helpers ----

func validateFrequencyFields(frequency string, dayOfMonth, dayOfWeek *uint) error {
	switch frequency {
	case "monthly", "yearly":
		if dayOfMonth == nil {
			return errors.New("day_of_month is required for monthly/yearly frequency")
		}
	case "weekly", "bi_weekly":
		if dayOfWeek == nil {
			return errors.New("day_of_week is required for weekly/bi_weekly frequency")
		}
	}
	return nil
}

// computeNextOccurrence returns the next occurrence date on or after today.
func computeNextOccurrence(frequency string, dayOfMonth, dayOfWeek *uint, startDate time.Time) time.Time {
	now := time.Now()
	base := startDate
	if base.Before(now) {
		base = now
	}
	// Normalize to date only (midnight)
	base = time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, base.Location())

	switch frequency {
	case "daily":
		return base

	case "weekly", "bi_weekly":
		targetDay := time.Weekday(0)
		if dayOfWeek != nil {
			targetDay = time.Weekday(*dayOfWeek)
		}
		d := base
		for d.Weekday() != targetDay {
			d = d.AddDate(0, 0, 1)
		}
		return d

	case "monthly":
		dom := 1
		if dayOfMonth != nil {
			dom = int(*dayOfMonth)
		}
		d := time.Date(base.Year(), base.Month(), dom, 0, 0, 0, 0, base.Location())
		if d.Before(base) {
			d = d.AddDate(0, 1, 0)
		}
		return d

	case "yearly":
		dom := 1
		if dayOfMonth != nil {
			dom = int(*dayOfMonth)
		}
		d := time.Date(base.Year(), base.Month(), dom, 0, 0, 0, 0, base.Location())
		if d.Before(base) {
			d = d.AddDate(1, 0, 0)
		}
		return d
	}

	return base
}

// advanceOccurrence returns the occurrence after the given date.
func advanceOccurrence(frequency string, dayOfMonth, dayOfWeek *uint, current time.Time) time.Time {
	switch frequency {
	case "daily":
		return current.AddDate(0, 0, 1)
	case "weekly":
		return current.AddDate(0, 0, 7)
	case "bi_weekly":
		return current.AddDate(0, 0, 14)
	case "monthly":
		next := current.AddDate(0, 1, 0)
		if dayOfMonth != nil {
			next = time.Date(next.Year(), next.Month(), int(*dayOfMonth), 0, 0, 0, 0, next.Location())
		}
		return next
	case "yearly":
		next := current.AddDate(1, 0, 0)
		if dayOfMonth != nil {
			next = time.Date(next.Year(), next.Month(), int(*dayOfMonth), 0, 0, 0, 0, next.Location())
		}
		return next
	}
	return current.AddDate(0, 0, 1)
}

func toRecurringTransactionResponse(r models.RecurringTransaction) dto.RecurringTransactionResponse {
	resp := dto.RecurringTransactionResponse{
		ID:              r.ID,
		AssetID:         r.AssetID,
		CategoryID:      r.CategoryID,
		Description:     r.Description,
		Amount:          r.Amount,
		TransactionType: r.TransactionType,
		Frequency:       r.Frequency,
		DayOfMonth:      r.DayOfMonth,
		DayOfWeek:       r.DayOfWeek,
		StartDate:       r.StartDate,
		EndDate:         r.EndDate,
		NextOccurrence:  r.NextOccurrence,
		IsActive:        r.IsActive,
		CreatedAt:       r.CreatedAt,
	}
	if r.Asset.ID > 0 {
		resp.AssetName = r.Asset.Name
	}
	if r.Category != nil && r.Category.ID > 0 {
		resp.CategoryName = r.Category.CategoryName
	}
	return resp
}
