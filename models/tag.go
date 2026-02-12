package models

import (
	"my-api/utils"
	"gorm.io/gorm"
)

// Tag represents a user-defined tag that can be attached to transactions
type Tag struct {
	ID         uint             `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	UserID     uint             `gorm:"not null;index;type:bigint unsigned" json:"user_id"`
	Name       string           `gorm:"size:50;not null" json:"name"`
	Color      string           `gorm:"size:7;default:'#6366F1'" json:"color"`
	Icon       string           `gorm:"size:10" json:"icon,omitempty"`
	UsageCount int              `gorm:"default:0" json:"usage_count"`
	CreatedAt  utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt  utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TransactionTag represents the many-to-many relationship between transactions and tags
type TransactionTag struct {
	TransactionID uint             `gorm:"primaryKey;type:bigint unsigned" json:"transaction_id"`
	TagID         uint             `gorm:"primaryKey;type:bigint unsigned" json:"tag_id"`
	CreatedAt     utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`

	// Relations
	Tag Tag `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

func (Tag) TableName() string {
	return "tags"
}

func (TransactionTag) TableName() string {
	return "transaction_tags"
}
