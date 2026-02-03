package dto

import "my-api/models"

type UserSettingsResponse struct {
	ID               uint64                `json:"id"`
	UserID           uint                  `json:"user_id"`
	PayCycleType     models.PayCycleType   `json:"pay_cycle_type"`
	PayDay           *int                  `json:"pay_day"`
	CycleStartOffset int                   `json:"cycle_start_offset"`
	CreatedAt        string                `json:"created_at"`
	UpdatedAt        *string               `json:"updated_at"`
}

type CreateUserSettingsRequest struct {
	PayCycleType     models.PayCycleType `json:"pay_cycle_type" binding:"required,oneof=calendar last_weekday custom_day bi_weekly"`
	PayDay           *int                `json:"pay_day"`
	CycleStartOffset int                 `json:"cycle_start_offset" binding:"min=0,max=31"`
}

type UpdateUserSettingsRequest struct {
	PayCycleType     models.PayCycleType `json:"pay_cycle_type" binding:"omitempty,oneof=calendar last_weekday custom_day bi_weekly"`
	PayDay           *int                `json:"pay_day"`
	CycleStartOffset *int                `json:"cycle_start_offset" binding:"omitempty,min=0,max=31"`
}

// ValidatePayCycleSettings validates the consistency between pay_cycle_type and pay_day
func (req *CreateUserSettingsRequest) Validate() error {
	switch req.PayCycleType {
	case models.PayCycleCustomDay:
		if req.PayDay == nil {
			return &ValidationError{Field: "pay_day", Message: "pay_day is required for custom_day pay cycle type"}
		}
		if *req.PayDay < 1 || *req.PayDay > 31 {
			return &ValidationError{Field: "pay_day", Message: "pay_day must be between 1 and 31 for custom_day"}
		}
	case models.PayCycleBiWeekly:
		if req.PayDay == nil {
			return &ValidationError{Field: "pay_day", Message: "pay_day is required for bi_weekly pay cycle type"}
		}
		if *req.PayDay < 0 || *req.PayDay > 6 {
			return &ValidationError{Field: "pay_day", Message: "pay_day must be between 0 and 6 (day of week) for bi_weekly"}
		}
	case models.PayCycleLastWeekday, models.PayCycleCalendar:
		// pay_day is not needed for these types
		req.PayDay = nil
	}
	return nil
}

func (req *UpdateUserSettingsRequest) Validate() error {
	if req.PayCycleType != "" {
		switch req.PayCycleType {
		case models.PayCycleCustomDay:
			if req.PayDay == nil {
				return &ValidationError{Field: "pay_day", Message: "pay_day is required for custom_day pay cycle type"}
			}
			if *req.PayDay < 1 || *req.PayDay > 31 {
				return &ValidationError{Field: "pay_day", Message: "pay_day must be between 1 and 31 for custom_day"}
			}
		case models.PayCycleBiWeekly:
			if req.PayDay == nil {
				return &ValidationError{Field: "pay_day", Message: "pay_day is required for bi_weekly pay cycle type"}
			}
			if *req.PayDay < 0 || *req.PayDay > 6 {
				return &ValidationError{Field: "pay_day", Message: "pay_day must be between 0 and 6 (day of week) for bi_weekly"}
			}
		case models.PayCycleLastWeekday, models.PayCycleCalendar:
			// pay_day is not needed for these types
			req.PayDay = nil
		}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
