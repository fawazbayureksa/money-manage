package models

import (
	"my-api/utils"
)

type Debt struct {
	ID              uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	UserID          uint             `gorm:"not null;index;type:int unsigned" json:"user_id"`
	Name            string           `gorm:"size:200;not null" json:"name"`
	DebtType        string           `gorm:"size:50;not null" json:"debt_type"` // credit_card, loan, mortgage, student_loan, personal, other
	OriginalAmount  int              `gorm:"not null" json:"original_amount"`
	CurrentBalance  int              `gorm:"not null" json:"current_balance"`
	InterestRate    float64          `gorm:"not null;default:0" json:"interest_rate"` // Annual percentage rate
	MinimumPayment  int              `gorm:"not null;default:0" json:"minimum_payment"`
	DueDay          int              `gorm:"not null;default:1" json:"due_day"` // Day of month (1-31)
	StartDate       utils.CustomTime `gorm:"not null;type:datetime" json:"start_date"`
	LenderName      string           `gorm:"size:200" json:"lender_name"`
	Notes           string           `gorm:"size:500" json:"notes"`
	IsActive        bool             `gorm:"default:true" json:"is_active"`
	CreatedAt       utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt       utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`

	// Relations
	User     User          `gorm:"foreignKey:UserID" json:"-"`
	Payments []DebtPayment `gorm:"foreignKey:DebtID" json:"payments,omitempty"`
}

type DebtPayment struct {
	ID          uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	DebtID      uint             `gorm:"not null;index;type:int unsigned" json:"debt_id"`
	UserID      uint             `gorm:"not null;index;type:int unsigned" json:"user_id"`
	Amount      int              `gorm:"not null" json:"amount"`
	PaymentDate utils.CustomTime `gorm:"not null;type:datetime" json:"payment_date"`
	Notes       string           `gorm:"size:500" json:"notes"`
	CreatedAt   utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`

	// Relations
	Debt Debt `gorm:"foreignKey:DebtID" json:"debt,omitempty"`
}

type DebtMilestone struct {
	ID            uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	DebtID        uint             `gorm:"not null;index;type:int unsigned" json:"debt_id"`
	UserID        uint             `gorm:"not null;index;type:int unsigned" json:"user_id"`
	MilestoneType string           `gorm:"size:50;not null" json:"milestone_type"` // percentage_paid, debt_free
	TargetValue   int              `gorm:"not null" json:"target_value"`            // e.g., 25, 50, 75, 100 for percentage
	Message       string           `gorm:"size:500" json:"message"`
	IsRead        bool             `gorm:"default:false" json:"is_read"`
	CreatedAt     utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`

	// Relations
	Debt Debt `gorm:"foreignKey:DebtID" json:"debt,omitempty"`
}
