package models

import "my-api/utils"

// TransactionSplit represents a single category-level split of a parent transaction.
type TransactionSplit struct {
	ID            uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	TransactionID uint             `gorm:"not null;index;type:int unsigned" json:"transaction_id"`
	CategoryID    *uint            `gorm:"index;type:int unsigned" json:"category_id"`
	Amount        int              `gorm:"not null" json:"amount"`
	Description   string           `gorm:"size:200" json:"description"`
	CreatedAt     utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt     utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`

	// Relations
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (TransactionSplit) TableName() string {
	return "transaction_splits"
}
