package models

import (
    "time"
)

type Category struct {
    ID           uint      `gorm:"primaryKey;autoIncrement;type:int unsigned"`
    CategoryName string    `gorm:"size:200;not null"`
    Description  string    `gorm:"size:200;not null"`
    UserID       uint      `gorm:"not null;type:int unsigned"`
    CreatedAt    time.Time `gorm:"autoCreateTime"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}