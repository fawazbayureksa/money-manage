package repositories

import (
	"gorm.io/gorm"
	"my-api/models"
	"time"
)

type EmailSyncRepository interface {
	// OAuth state helpers
	SaveState(state *models.EmailOAuthState) error
	FindAndDeleteState(state string) (*models.EmailOAuthState, error)
	DeleteExpiredStates() error

	// Token helpers
	GetToken(userID uint) (*models.EmailSyncToken, error)
	SaveToken(token *models.EmailSyncToken) error
	DeleteToken(userID uint) error

	// Sync log helpers
	IsMessageSynced(userID uint, gmailMessageID string) bool
	SaveLog(log *models.EmailSyncLog) error
	GetLogs(userID uint, limit, offset int) ([]models.EmailSyncLog, int64, error)
}

type emailSyncRepository struct {
	db *gorm.DB
}

func NewEmailSyncRepository(db *gorm.DB) EmailSyncRepository {
	return &emailSyncRepository{db: db}
}

// ----- OAuth state -----

func (r *emailSyncRepository) SaveState(state *models.EmailOAuthState) error {
	return r.db.Create(state).Error
}

func (r *emailSyncRepository) FindAndDeleteState(state string) (*models.EmailOAuthState, error) {
	var record models.EmailOAuthState
	err := r.db.Where("state = ? AND expires_at > ?", state, time.Now()).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	r.db.Delete(&record)
	return &record, nil
}

func (r *emailSyncRepository) DeleteExpiredStates() error {
	return r.db.Where("expires_at <= ?", time.Now()).
		Delete(&models.EmailOAuthState{}).Error
}

// ----- Token -----

func (r *emailSyncRepository) GetToken(userID uint) (*models.EmailSyncToken, error) {
	var token models.EmailSyncToken
	err := r.db.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *emailSyncRepository) SaveToken(token *models.EmailSyncToken) error {
	return r.db.Save(token).Error
}

func (r *emailSyncRepository) DeleteToken(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.EmailSyncToken{}).Error
}

// ----- Sync logs -----

func (r *emailSyncRepository) IsMessageSynced(userID uint, gmailMessageID string) bool {
	var count int64
	r.db.Model(&models.EmailSyncLog{}).
		Where("user_id = ? AND gmail_message_id = ?", userID, gmailMessageID).
		Count(&count)
	return count > 0
}

func (r *emailSyncRepository) SaveLog(log *models.EmailSyncLog) error {
	return r.db.Create(log).Error
}

func (r *emailSyncRepository) GetLogs(userID uint, limit, offset int) ([]models.EmailSyncLog, int64, error) {
	var logs []models.EmailSyncLog
	var total int64

	query := r.db.Model(&models.EmailSyncLog{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}
