package models

import "my-api/utils"

// RecurringTransaction represents a scheduled recurring transaction template.
// Frequencies: daily, weekly, bi_weekly, monthly, yearly
type RecurringTransaction struct {
	ID              uint64           `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	UserID          uint             `gorm:"not null;index;type:int unsigned" json:"user_id"`
	AssetID         uint64           `gorm:"not null;index;type:bigint unsigned" json:"asset_id"`
	CategoryID      *uint            `gorm:"index;type:int unsigned" json:"category_id"`
	Description     string           `gorm:"size:200;not null" json:"description"`
	Amount          int              `gorm:"not null" json:"amount"`
	TransactionType int              `gorm:"not null" json:"transaction_type"` // 1=income, 2=expense
	Frequency       string           `gorm:"size:20;not null" json:"frequency"` // daily, weekly, bi_weekly, monthly, yearly
	DayOfMonth      *uint            `gorm:"type:tinyint unsigned" json:"day_of_month"` // 1-31 for monthly
	DayOfWeek       *uint            `gorm:"type:tinyint unsigned" json:"day_of_week"`  // 0=Sunday for weekly/bi_weekly
	StartDate       utils.CustomTime `gorm:"not null;type:date" json:"start_date"`
	EndDate         *utils.CustomTime `gorm:"type:date" json:"end_date"`
	NextOccurrence  utils.CustomTime `gorm:"not null;index;type:date" json:"next_occurrence"`
	IsActive        bool             `gorm:"default:true" json:"is_active"`
	CreatedAt       utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt       utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`

	// Relations
	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Asset    Asset     `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}
