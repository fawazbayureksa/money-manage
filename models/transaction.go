package models

import (
    "time"
)
type Transaction struct {
    ID           uint      `gorm:"primaryKey;autoIncrement"`
    Description  string    `gorm:"size:200;not null"`
    UserID       uint      `gorm:"not null"`
	CategoryID    uint     `gorm:"not null;foreignKey:UserID"`
	BankID       uint      `gorm:"not null;foreignKey:UserID"`
	Amount       float64    `gorm:"not null"`
    CreatedAt    time.Time `gorm:"autoCreateTime"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}