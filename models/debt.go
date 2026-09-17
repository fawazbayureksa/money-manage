package models

import (
	"my-api/utils"
	"time"

	"gorm.io/gorm"
)

type Debt struct {
	ID           uint   `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	UserID       uint   `gorm:"not null;index;type:int unsigned" json:"user_id"`
	Name         string `gorm:"size:100;not null" json:"name"`
	Description  string `gorm:"type:text" json:"description,omitempty"`
	DebtType     string `gorm:"type:enum('credit_card','personal_loan','mortgage','car_loan','student_loan','other');not null" json:"debt_type"`
	CreditorName string `gorm:"size:100" json:"creditor_name,omitempty"`

	OriginalAmount int     `gorm:"not null" json:"original_amount"`
	CurrentBalance int     `gorm:"not null" json:"current_balance"`
	InterestRate   float64 `gorm:"type:decimal(5,2);not null" json:"interest_rate"`
	InterestType   string  `gorm:"type:enum('fixed','variable');default:'fixed'" json:"interest_type"`

	MinimumPayment   int  `gorm:"not null" json:"minimum_payment"`
	PaymentDueDay    int  `gorm:"not null" json:"payment_due_day"`
	CurrentMonthPaid bool `gorm:"default:false" json:"current_month_paid"`

	StartDate          utils.CustomTime  `gorm:"type:date;not null" json:"start_date"`
	ExpectedPayoffDate *utils.CustomTime `gorm:"type:date" json:"expected_payoff_date,omitempty"`
	ActualPayoffDate   *utils.CustomTime `gorm:"type:date" json:"actual_payoff_date,omitempty"`

	Status         string  `gorm:"type:enum('active','paid_off','defaulted','settled');default:'active'" json:"status"`
	PaymentAssetID *uint64 `gorm:"type:bigint unsigned" json:"payment_asset_id,omitempty"`

	IncludeInNetWorth bool `gorm:"default:true" json:"include_in_net_worth"`
	AutoTrackInterest bool `gorm:"default:true" json:"auto_track_interest"`

	Notes     string           `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
	UpdatedAt utils.CustomTime `gorm:"autoUpdateTime;type:datetime" json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relations
	Payments   []DebtPayment   `gorm:"foreignKey:DebtID" json:"payments,omitempty"`
	Milestones []DebtMilestone `gorm:"foreignKey:DebtID" json:"milestones,omitempty"`
}

// PaidOffAmount returns the amount already paid off from the original balance.
func (d *Debt) PaidOffAmount() int {
	paid := d.OriginalAmount - d.CurrentBalance
	if paid < 0 {
		return 0
	}
	return paid
}

// PaidOffPercentage returns what percentage of the debt has been paid off.
func (d *Debt) PaidOffPercentage() float64 {
	if d.OriginalAmount == 0 {
		return 0
	}
	return float64(d.PaidOffAmount()) / float64(d.OriginalAmount) * 100
}

// DaysUntilDue returns the number of days until the next payment is due.
func (d *Debt) DaysUntilDue() int {
	now := time.Now()
	dueDate := time.Date(now.Year(), now.Month(), d.PaymentDueDay, 0, 0, 0, 0, now.Location())

	// If due date has passed this month, calculate for next month
	if now.After(dueDate) {
		dueDate = dueDate.AddDate(0, 1, 0)
	}

	days := int(dueDate.Sub(now).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// MonthlyInterest returns the estimated interest accrued this month.
func (d *Debt) MonthlyInterest() int {
	return int(float64(d.CurrentBalance) * (d.InterestRate / 100 / 12))
}

// ---

type DebtPayment struct {
	ID              uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	DebtID          uint             `gorm:"not null;index;type:int unsigned" json:"debt_id"`
	UserID          uint             `gorm:"not null;type:int unsigned" json:"user_id"`
	Amount          int              `gorm:"not null" json:"amount"`
	PaymentType     string           `gorm:"type:enum('regular','extra','interest_only','payoff','adjustment');not null" json:"payment_type"`
	PrincipalAmount int              `gorm:"default:0" json:"principal_amount"`
	InterestAmount  int              `gorm:"default:0" json:"interest_amount"`
	FeesAmount      int              `gorm:"default:0" json:"fees_amount"`
	BalanceBefore   int              `gorm:"not null" json:"balance_before"`
	BalanceAfter    int              `gorm:"not null" json:"balance_after"`
	TransactionID   *uint            `gorm:"type:bigint unsigned" json:"transaction_id,omitempty"`
	PaymentDate     utils.CustomTime `gorm:"type:date;not null" json:"payment_date"`
	Notes           string           `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
}

// ---

type DebtMilestone struct {
	ID            uint             `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	DebtID        uint             `gorm:"not null;index;type:int unsigned" json:"debt_id"`
	UserID        uint             `gorm:"not null;type:int unsigned" json:"user_id"`
	MilestoneType string           `gorm:"type:enum('25_percent','50_percent','75_percent','paid_off','custom');not null" json:"milestone_type"`
	Description   string           `gorm:"size:255" json:"description,omitempty"`
	ReachedAt     *time.Time       `json:"reached_at,omitempty"`
	IsCelebrated  bool             `gorm:"default:false" json:"is_celebrated"`
	CreatedAt     utils.CustomTime `gorm:"autoCreateTime;type:datetime" json:"created_at"`
}
