package models

import "my-api/utils"

type SavingsGoal struct {
	ID            uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	UserID        uint             `gorm:"not null;index;type:int unsigned" json:"user_id"`
	AssetID       *uint64          `gorm:"index;type:bigint unsigned" json:"asset_id"`
	Name          string           `gorm:"size:255;not null" json:"name"`
	Description   string           `gorm:"size:500" json:"description"`
	TargetAmount  int              `gorm:"not null" json:"target_amount"`
	CurrentAmount int              `gorm:"not null;default:0" json:"current_amount"`
	Currency      string           `gorm:"size:10;not null;default:'IDR'" json:"currency"`
	Deadline      utils.CustomTime `gorm:"type:datetime" json:"deadline"`
	Status        string           `gorm:"size:20;not null;default:'active'" json:"status"` // active, completed, cancelled
	Icon          string           `gorm:"size:50" json:"icon"`
	Color         string           `gorm:"size:20" json:"color"`
	CreatedAt     utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt     utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`

	// Relations
	User          User              `gorm:"foreignKey:UserID" json:"-"`
	Asset         *Asset            `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	Contributions []SavingsContribution `gorm:"foreignKey:GoalID" json:"contributions,omitempty"`
}

type SavingsContribution struct {
	ID        uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	GoalID    uint             `gorm:"not null;index;type:int unsigned" json:"goal_id"`
	UserID    uint             `gorm:"not null;index;type:int unsigned" json:"user_id"`
	Amount    int              `gorm:"not null" json:"amount"`
	Note      string           `gorm:"size:500" json:"note"`
	Date      utils.CustomTime `gorm:"not null;type:datetime" json:"date"`
	CreatedAt utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`

	Goal SavingsGoal `gorm:"foreignKey:GoalID" json:"goal,omitempty"`
}
