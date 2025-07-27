package models

import (
    "time"
)

type Transaction struct {
    ID             uint      `gorm:"primaryKey;autoIncrement"`
    Description    string    `gorm:"size:200;not null"`
    UserID         uint      `gorm:"not null"`
    CategoryID     uint      `gorm:"not null"`
    BankID         uint      `gorm:"not null"`
    Amount         int       `gorm:"not null"`
    TransactionType int      `gorm:"not null"` // e.g., "income", "expense"
    Date           time.Time `gorm:"not null"`
    CreatedAt      time.Time `gorm:"autoCreateTime"`
    UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}