package models

import (
	"my-api/utils"
	"time"
)

type PayCycleType string

const (
	PayCycleCalendar    PayCycleType = "calendar"     // Standard calendar month
	PayCycleLastWeekday PayCycleType = "last_weekday" // Last weekday of month (Mon-Fri)
	PayCycleCustomDay   PayCycleType = "custom_day"   // Specific day of month (e.g., 25th)
	PayCycleBiWeekly    PayCycleType = "bi_weekly"    // Every 2 weeks
)

type UserSettings struct {
	ID               uint64       `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	UserID           uint         `json:"user_id" gorm:"type:int unsigned;uniqueIndex;not null"`
	PayCycleType     PayCycleType `json:"pay_cycle_type" gorm:"type:enum('calendar','last_weekday','custom_day','bi_weekly');default:'calendar';not null"`
	PayDay           *int         `json:"pay_day" gorm:"type:int;default:null"` // Day of month (1-31) or day of week (0-6)
	CycleStartOffset int          `json:"cycle_start_offset" gorm:"type:int;default:1;not null"`
	CreatedAt        time.Time    `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not null"`
	UpdatedAt        *time.Time   `json:"updated_at" gorm:"type:timestamp;default:null;onUpdate:CURRENT_TIMESTAMP"`
	User             User         `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// Implement UserSettingsInterface from utils package
func (u *UserSettings) GetPayCycleType() utils.PayCycleType {
	return utils.PayCycleType(u.PayCycleType)
}

func (u *UserSettings) GetPayDay() *int {
	return u.PayDay
}

func (u *UserSettings) GetCycleStartOffset() int {
	return u.CycleStartOffset
}
