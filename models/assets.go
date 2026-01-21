package models

import "time"

// Asset represents a wallet/asset belonging to a user.
type Asset struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    uint64    `gorm:"not null;index" json:"user_id"`
    Name      string    `gorm:"size:255;not null" json:"name"`
    Type      string    `gorm:"size:100" json:"type"`
    Balance   float64   `gorm:"type:decimal(20,8);not null;default:0" json:"balance"`
    Currency  string    `gorm:"size:10;not null" json:"currency"`
    BankName  string    `gorm:"size:255" json:"bank_name"`
    AccountNo string    `gorm:"size:100" json:"account_no"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
