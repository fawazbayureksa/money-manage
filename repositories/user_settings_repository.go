package repositories

import (
	"gorm.io/gorm"
	"my-api/models"
)

type UserSettingsRepository interface {
	FindByUserID(userID uint) (*models.UserSettings, error)
	Create(settings *models.UserSettings) error
	Update(settings *models.UserSettings) error
	Delete(userID uint) error
	Upsert(settings *models.UserSettings) error
}

type userSettingsRepository struct {
	db *gorm.DB
}

func NewUserSettingsRepository(db *gorm.DB) UserSettingsRepository {
	return &userSettingsRepository{db: db}
}

func (r *userSettingsRepository) FindByUserID(userID uint) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *userSettingsRepository) Create(settings *models.UserSettings) error {
	return r.db.Create(settings).Error
}

func (r *userSettingsRepository) Update(settings *models.UserSettings) error {
	return r.db.Save(settings).Error
}

func (r *userSettingsRepository) Delete(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.UserSettings{}).Error
}

// Upsert creates or updates user settings
func (r *userSettingsRepository) Upsert(settings *models.UserSettings) error {
	var existing models.UserSettings
	err := r.db.Where("user_id = ?", settings.UserID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new settings
		return r.db.Create(settings).Error
	} else if err != nil {
		return err
	}

	// Update existing settings
	settings.ID = existing.ID
	return r.db.Save(settings).Error
}
