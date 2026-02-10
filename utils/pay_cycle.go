package utils

import (
	"time"
)

// PayCycleType represents the type of pay cycle
type PayCycleType string

const (
	PayCycleCalendar    PayCycleType = "calendar"
	PayCycleLastWeekday PayCycleType = "last_weekday"
	PayCycleCustomDay   PayCycleType = "custom_day"
	PayCycleBiWeekly    PayCycleType = "bi_weekly"
)

// UserSettingsInterface defines the interface for user settings
type UserSettingsInterface interface {
	GetPayCycleType() PayCycleType
	GetPayDay() *int
	GetCycleStartOffset() int
}

// FinancialPeriod represents a financial period with start and end dates
type FinancialPeriod struct {
	PeriodLabel string    `json:"period_label"` // e.g., "2026-01", "Week 1", etc.
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

// GetLastWeekdayOfMonth returns the last weekday (Mon-Fri) of the given month
func GetLastWeekdayOfMonth(year int, month time.Month) time.Time {
	// Get last day of month
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)

	// Move back to Friday if it's Saturday or Sunday
	for lastDay.Weekday() == time.Saturday || lastDay.Weekday() == time.Sunday {
		lastDay = lastDay.AddDate(0, 0, -1)
	}
	return lastDay
}

// GetFinancialPeriodForDate returns the financial period for a specific date
// based on user's pay cycle settings
func GetFinancialPeriodForDate(settings UserSettingsInterface, targetDate time.Time) FinancialPeriod {
	if settings == nil || settings.GetPayCycleType() == PayCycleCalendar {
		return getCalendarPeriod(targetDate)
	}

	switch settings.GetPayCycleType() {
	case PayCycleLastWeekday:
		return getLastWeekdayPeriod(targetDate, settings.GetCycleStartOffset())
	case PayCycleCustomDay:
		if settings.GetPayDay() != nil {
			return getCustomDayPeriod(targetDate, *settings.GetPayDay(), settings.GetCycleStartOffset())
		}
		return getCalendarPeriod(targetDate)
	case PayCycleBiWeekly:
		if settings.GetPayDay() != nil {
			return getBiWeeklyPeriod(targetDate, *settings.GetPayDay(), settings.GetCycleStartOffset())
		}
		return getCalendarPeriod(targetDate)
	default:
		return getCalendarPeriod(targetDate)
	}
}

// GetFinancialPeriods returns all financial periods between start and end dates
func GetFinancialPeriods(settings UserSettingsInterface, startDate, endDate time.Time) []FinancialPeriod {
	var periods []FinancialPeriod

	if settings == nil || settings.GetPayCycleType() == PayCycleCalendar {
		return getCalendarPeriods(startDate, endDate)
	}

	// Start with the period containing startDate
	currentPeriod := GetFinancialPeriodForDate(settings, startDate)
	
	for !currentPeriod.StartDate.After(endDate) {
		// Only add if the period overlaps with our date range
		if !currentPeriod.EndDate.Before(startDate) {
			periods = append(periods, currentPeriod)
		}

		// Move to next period
		nextDate := currentPeriod.EndDate.AddDate(0, 0, 1)
		if nextDate.After(endDate) {
			break
		}
		currentPeriod = GetFinancialPeriodForDate(settings, nextDate)
	}

	return periods
}

// getCalendarPeriod returns a standard calendar month period
func getCalendarPeriod(targetDate time.Time) FinancialPeriod {
	year, month, _ := targetDate.Date()
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, month+1, 0, 23, 59, 59, 0, time.UTC)

	return FinancialPeriod{
		PeriodLabel: startDate.Format("2006-01"),
		StartDate:   startDate,
		EndDate:     endDate,
	}
}

// getCalendarPeriods returns all calendar month periods between dates
func getCalendarPeriods(startDate, endDate time.Time) []FinancialPeriod {
	var periods []FinancialPeriod

	current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !current.After(end) {
		period := getCalendarPeriod(current)
		periods = append(periods, period)
		current = current.AddDate(0, 1, 0)
	}

	return periods
}

// getLastWeekdayPeriod calculates financial period based on last weekday of month
func getLastWeekdayPeriod(targetDate time.Time, offset int) FinancialPeriod {
	year, month, _ := targetDate.Date()

	// Get last weekday of previous month
	prevMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
	lastWeekdayPrev := GetLastWeekdayOfMonth(prevMonth.Year(), prevMonth.Month())
	periodStart := lastWeekdayPrev.AddDate(0, 0, offset)

	// Get last weekday of current month
	lastWeekdayCurrent := GetLastWeekdayOfMonth(year, month)
	periodEnd := lastWeekdayCurrent.AddDate(0, 0, offset-1)

	// If target date is before current period start, move back one month
	if targetDate.Before(periodStart) {
		lastWeekdayPrevPrev := GetLastWeekdayOfMonth(prevMonth.Year(), prevMonth.Month()-1)
		periodStart = lastWeekdayPrevPrev.AddDate(0, 0, offset)
		periodEnd = lastWeekdayPrev.AddDate(0, 0, offset-1)
		month = prevMonth.Month()
		year = prevMonth.Year()
	} else if targetDate.After(periodEnd) {
		// If target date is after current period end, move forward one month
		nextMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		lastWeekdayNext := GetLastWeekdayOfMonth(nextMonth.Year(), nextMonth.Month())
		periodStart = lastWeekdayCurrent.AddDate(0, 0, offset)
		periodEnd = lastWeekdayNext.AddDate(0, 0, offset-1)
		month = nextMonth.Month()
		year = nextMonth.Year()
	}

	// Set time to start/end of day
	periodStart = time.Date(periodStart.Year(), periodStart.Month(), periodStart.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd = time.Date(periodEnd.Year(), periodEnd.Month(), periodEnd.Day(), 23, 59, 59, 0, time.UTC)

	return FinancialPeriod{
		PeriodLabel: time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01"),
		StartDate:   periodStart,
		EndDate:     periodEnd,
	}
}

// getCustomDayPeriod calculates financial period based on custom day of month
func getCustomDayPeriod(targetDate time.Time, payDay int, offset int) FinancialPeriod {
	year, month, _ := targetDate.Date()

	// Ensure payDay is valid (1-31)
	if payDay < 1 {
		payDay = 1
	}
	if payDay > 31 {
		payDay = 31
	}

	// Get the pay day for this month (handle months with fewer days)
	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	actualPayDay := payDay
	if payDay > lastDayOfMonth {
		actualPayDay = lastDayOfMonth
	}

	payDate := time.Date(year, month, actualPayDay, 0, 0, 0, 0, time.UTC)
	periodStart := payDate.AddDate(0, 0, offset)

	// If target date is before current period start, use previous month
	if targetDate.Before(periodStart) {
		prevMonth := time.Date(year, month-1, 1, 0, 0, 0, 0, time.UTC)
		lastDayPrevMonth := time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		actualPayDayPrev := payDay
		if payDay > lastDayPrevMonth {
			actualPayDayPrev = lastDayPrevMonth
		}
		payDate = time.Date(prevMonth.Year(), prevMonth.Month(), actualPayDayPrev, 0, 0, 0, 0, time.UTC)
		periodStart = payDate.AddDate(0, 0, offset)
		month = prevMonth.Month()
		year = prevMonth.Year()
	}

	// Calculate period end (day before next period start)
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	lastDayNextMonth := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	actualPayDayNext := payDay
	if payDay > lastDayNextMonth {
		actualPayDayNext = lastDayNextMonth
	}
	nextPayDate := time.Date(nextMonth.Year(), nextMonth.Month(), actualPayDayNext, 0, 0, 0, 0, time.UTC)
	periodEnd := nextPayDate.AddDate(0, 0, offset-1)

	periodEnd = time.Date(periodEnd.Year(), periodEnd.Month(), periodEnd.Day(), 23, 59, 59, 0, time.UTC)

	return FinancialPeriod{
		PeriodLabel: time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01"),
		StartDate:   periodStart,
		EndDate:     periodEnd,
	}
}

// getBiWeeklyPeriod calculates financial period based on bi-weekly pay schedule
func getBiWeeklyPeriod(targetDate time.Time, startDayOfWeek int, offset int) FinancialPeriod {
	// For bi-weekly, we need a reference start date
	// Let's use the first occurrence of the pay day in the current year
	year := targetDate.Year()
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// Find first occurrence of the pay day
	daysUntilPayDay := (startDayOfWeek - int(jan1.Weekday()) + 7) % 7
	firstPayDay := jan1.AddDate(0, 0, daysUntilPayDay)

	// Calculate weeks since first pay day
	daysSinceFirst := int(targetDate.Sub(firstPayDay).Hours() / 24)
	periodNumber := daysSinceFirst / 14

	// Calculate period start (every 14 days)
	periodPayDay := firstPayDay.AddDate(0, 0, periodNumber*14)
	periodStart := periodPayDay.AddDate(0, 0, offset)
	periodEnd := periodStart.AddDate(0, 0, 13)

	periodEnd = time.Date(periodEnd.Year(), periodEnd.Month(), periodEnd.Day(), 23, 59, 59, 0, time.UTC)

	return FinancialPeriod{
		PeriodLabel: periodStart.Format("2006-01-02") + " to " + periodEnd.Format("01-02"),
		StartDate:   periodStart,
		EndDate:     periodEnd,
	}
}

// AdjustDateRangeForPayCycle adjusts start and end dates to align with financial periods
func AdjustDateRangeForPayCycle(settings UserSettingsInterface, startDate, endDate time.Time) (time.Time, time.Time) {
	if settings == nil || settings.GetPayCycleType() == PayCycleCalendar {
		return startDate, endDate
	}

	startPeriod := GetFinancialPeriodForDate(settings, startDate)
	endPeriod := GetFinancialPeriodForDate(settings, endDate)

	return startPeriod.StartDate, endPeriod.EndDate
}
