package models

import (
	"my-api/utils"
)

func (TransactionV2) TableName() string {
	return "transactions"
}

// TransactionV2 represents the transaction model with asset support
type TransactionV2 struct {
	ID              uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	Description     string           `gorm:"size:200;not null" json:"description"`
	UserID          uint             `gorm:"not null;index;type:int unsigned" json:"user_id"`
	CategoryID      uint             `gorm:"not null;index;type:int unsigned" json:"category_id"`
	BankID          uint             `gorm:"index;type:int unsigned" json:"bank_id"`
	AssetID         uint64           `gorm:"not null;index;type:bigint unsigned" json:"asset_id"`
	Amount          int              `gorm:"not null" json:"amount"`
	TransactionType int              `gorm:"not null" json:"transaction_type"` // 1=income, 2=expense
	Date            utils.CustomTime `gorm:"not null;index;type:datetime" json:"date"`
	CreatedAt       utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt       utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`

	// Relations
	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Bank     Bank     `gorm:"foreignKey:BankID" json:"bank,omitempty"`
	Asset    Asset    `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}
