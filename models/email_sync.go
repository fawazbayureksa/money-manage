package models

import "time"

// EmailOAuthState is a short-lived record used to correlate an OAuth2 callback with a user.
type EmailOAuthState struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	State     string    `gorm:"size:255;not null;uniqueIndex" json:"state"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (EmailOAuthState) TableName() string { return "email_oauth_states" }

// EmailSyncToken stores the Gmail OAuth2 access/refresh token for a user.
type EmailSyncToken struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint       `gorm:"not null;uniqueIndex" json:"user_id"`
	AccessToken  string     `gorm:"type:text;not null" json:"access_token"`
	RefreshToken string     `gorm:"type:text;not null" json:"refresh_token"`
	TokenType    string     `gorm:"size:50;not null;default:Bearer" json:"token_type"`
	Expiry       *time.Time `json:"expiry"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (EmailSyncToken) TableName() string { return "email_sync_tokens" }

// EmailSyncLog records each Gmail message that was processed during a sync.
type EmailSyncLog struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint       `gorm:"not null;index" json:"user_id"`
	GmailMessageID  string     `gorm:"size:255;not null" json:"gmail_message_id"`
	Subject         string     `gorm:"size:500" json:"subject"`
	FromEmail       string     `gorm:"size:255" json:"from_email"`
	BankName        string     `gorm:"size:100" json:"bank_name"`
	Amount          int        `gorm:"not null;default:0" json:"amount"`
	AssetID         uint64     `gorm:"not null;default:0" json:"asset_id"`
	TransactionID   uint       `gorm:"not null;default:0" json:"transaction_id"`
	Status          string     `gorm:"size:50;not null;default:imported" json:"status"`
	ErrorMessage    string     `gorm:"type:text" json:"error_message"`
	EmailDate       *time.Time `json:"email_date"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (EmailSyncLog) TableName() string { return "email_sync_logs" }
