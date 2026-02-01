package services

import (
	"errors"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"time"

	"gorm.io/gorm"
)

type UserSettingsService interface {
	GetUserSettings(userID uint) (*dto.UserSettingsResponse, error)
	CreateUserSettings(userID uint, req *dto.CreateUserSettingsRequest) (*dto.UserSettingsResponse, error)
	UpdateUserSettings(userID uint, req *dto.UpdateUserSettingsRequest) (*dto.UserSettingsResponse, error)
	DeleteUserSettings(userID uint) error
}

type userSettingsService struct {
	repo repositories.UserSettingsRepository
}

func NewUserSettingsService(repo repositories.UserSettingsRepository) UserSettingsService {
	return &userSettingsService{repo: repo}
}

func (s *userSettingsService) GetUserSettings(userID uint) (*dto.UserSettingsResponse, error) {
	settings, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default settings if not found
			return &dto.UserSettingsResponse{
				UserID:           userID,
				PayCycleType:     models.PayCycleCalendar,
				PayDay:           nil,
				CycleStartOffset: 1,
			}, nil
		}
		return nil, err
	}

	return s.toUserSettingsResponse(settings), nil
}

func (s *userSettingsService) CreateUserSettings(userID uint, req *dto.CreateUserSettingsRequest) (*dto.UserSettingsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if settings already exist
	existing, err := s.repo.FindByUserID(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user settings already exist, use update instead")
	}

	// Create settings
	settings := &models.UserSettings{
		UserID:           userID,
		PayCycleType:     req.PayCycleType,
		PayDay:           req.PayDay,
		CycleStartOffset: req.CycleStartOffset,
		CreatedAt:        time.Now(),
	}

	err = s.repo.Create(settings)
	if err != nil {
		return nil, err
	}

	return s.toUserSettingsResponse(settings), nil
}

func (s *userSettingsService) UpdateUserSettings(userID uint, req *dto.UpdateUserSettingsRequest) (*dto.UserSettingsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get existing settings
	settings, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user settings not found, please create first")
		}
		return nil, err
	}

	// Update fields
	if req.PayCycleType != "" {
		settings.PayCycleType = req.PayCycleType
	}
	if req.PayDay != nil {
		settings.PayDay = req.PayDay
	}
	if req.CycleStartOffset != nil {
		settings.CycleStartOffset = *req.CycleStartOffset
	}

	err = s.repo.Update(settings)
	if err != nil {
		return nil, err
	}

	return s.toUserSettingsResponse(settings), nil
}

func (s *userSettingsService) DeleteUserSettings(userID uint) error {
	return s.repo.Delete(userID)
}

func (s *userSettingsService) toUserSettingsResponse(settings *models.UserSettings) *dto.UserSettingsResponse {
	var updatedAt *string
	if settings.UpdatedAt != nil {
		updated := settings.UpdatedAt.Format("2006-01-02 15:04:05")
		updatedAt = &updated
	}

	return &dto.UserSettingsResponse{
		ID:               settings.ID,
		UserID:           settings.UserID,
		PayCycleType:     settings.PayCycleType,
		PayDay:           settings.PayDay,
		CycleStartOffset: settings.CycleStartOffset,
		CreatedAt:        settings.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        updatedAt,
	}
}
