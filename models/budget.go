package models

import (
	"my-api/utils"
)

type Budget struct {
	ID          uint              `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	UserID      uint              `gorm:"not null;index;type:int unsigned" json:"user_id"`
	CategoryID  uint              `gorm:"not null;index;type:int unsigned" json:"category_id"`
	Amount      int               `gorm:"not null" json:"amount"`
	Period      string            `gorm:"size:20;not null" json:"period"` // monthly, yearly
	StartDate   utils.CustomTime  `gorm:"not null;type:datetime" json:"start_date"`
	EndDate     utils.CustomTime  `gorm:"not null;type:datetime" json:"end_date"`
	IsActive    bool              `gorm:"default:true" json:"is_active"`
	AlertAt     int               `gorm:"default:80" json:"alert_at"` // Alert at 80% of budget
	Description string            `gorm:"size:500" json:"description"`
	CreatedAt   utils.CustomTime  `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt   utils.CustomTime  `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`

	// Relations
	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

type BudgetAlert struct {
	ID          uint              `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	BudgetID    uint              `gorm:"not null;index;type:int unsigned" json:"budget_id"`
	UserID      uint              `gorm:"not null;index;type:int unsigned" json:"user_id"`
	Percentage  int               `gorm:"not null" json:"percentage"`
	SpentAmount int               `gorm:"not null" json:"spent_amount"`
	Message     string            `gorm:"size:500" json:"message"`
	IsRead      bool              `gorm:"default:false" json:"is_read"`
	CreatedAt   utils.CustomTime  `gorm:"autoCreateTime;type:datetime" json:"created_at"`

	Budget Budget `gorm:"foreignKey:BudgetID" json:"budget,omitempty"`
}
