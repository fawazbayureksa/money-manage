package models

import (
    "time"
)

type Transaction struct {
    ID              uint      `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
    Description     string    `gorm:"size:200;not null" json:"description"`
    UserID          uint      `gorm:"not null;index;type:int unsigned" json:"user_id"`
    CategoryID      uint      `gorm:"not null;index;type:int unsigned" json:"category_id"`
    BankID          uint      `gorm:"not null;index;type:int unsigned" json:"bank_id"`
    Amount          int       `gorm:"not null" json:"amount"`
    TransactionType int       `gorm:"not null" json:"transaction_type"` // 1=income, 2=expense
    Date            time.Time `gorm:"not null;index" json:"date"`
    CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    // Relations
    User     User     `gorm:"foreignKey:UserID" json:"-"`
    Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Bank     Bank     `gorm:"foreignKey:BankID" json:"bank,omitempty"`
}