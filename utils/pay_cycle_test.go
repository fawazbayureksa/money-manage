package utils

import (
	"testing"
	"time"
)

// mockSettings implements UserSettingsInterface for testing.
type mockSettings struct {
	payCycleType     PayCycleType
	payDay           *int
	cycleStartOffset int
}

func (m *mockSettings) GetPayCycleType() PayCycleType { return m.payCycleType }
func (m *mockSettings) GetPayDay() *int               { return m.payDay }
func (m *mockSettings) GetCycleStartOffset() int      { return m.cycleStartOffset }

func intPtr(i int) *int { return &i }

// TestGetLastWeekdayOfMonthNeverWeekend verifies every returned date is Mon-Fri.
func TestGetLastWeekdayOfMonthNeverWeekend(t *testing.T) {
	for month := time.January; month <= time.December; month++ {
		result := GetLastWeekdayOfMonth(2024, month)
		wd := result.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			t.Errorf("Month %v: last weekday is %v (weekend)", month, wd)
		}
	}
}

// TestGetLastWeekdayOfMonthIsInMonth verifies the returned date is in the same month.
func TestGetLastWeekdayOfMonthIsInMonth(t *testing.T) {
	for month := time.January; month <= time.December; month++ {
		result := GetLastWeekdayOfMonth(2024, month)
		if result.Month() != month {
			t.Errorf("Month %v: last weekday falls in %v", month, result.Month())
		}
	}
}

func TestGetFinancialPeriodForDateNilIsCalendar(t *testing.T) {
	targetDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	period := GetFinancialPeriodForDate(nil, targetDate)

	if period.PeriodLabel != "2024-03" {
		t.Errorf("Expected '2024-03', got '%s'", period.PeriodLabel)
	}
	if period.StartDate.Day() != 1 || period.StartDate.Month() != time.March {
		t.Errorf("Expected start 2024-03-01, got %v", period.StartDate)
	}
	if period.EndDate.Day() != 31 || period.EndDate.Month() != time.March {
		t.Errorf("Expected end 2024-03-31, got %v", period.EndDate)
	}
}

func TestGetFinancialPeriodForDateExplicitCalendar(t *testing.T) {
	targetDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	settings := &mockSettings{payCycleType: PayCycleCalendar}

	period := GetFinancialPeriodForDate(settings, targetDate)

	if period.PeriodLabel != "2024-03" {
		t.Errorf("Expected '2024-03', got '%s'", period.PeriodLabel)
	}
}

func TestGetFinancialPeriodForDateCustomDay(t *testing.T) {
	// Pay day is the 15th; target is the 20th — should be in Mar 15-Apr 14 range.
	targetDate := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	settings := &mockSettings{
		payCycleType:     PayCycleCustomDay,
		payDay:           intPtr(15),
		cycleStartOffset: 0,
	}

	period := GetFinancialPeriodForDate(settings, targetDate)

	if period.StartDate.Day() != 15 || period.StartDate.Month() != time.March {
		t.Errorf("Expected period start 2024-03-15, got %v", period.StartDate)
	}
}

func TestGetFinancialPeriodCustomDayNilPayDay(t *testing.T) {
	targetDate := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	settings := &mockSettings{
		payCycleType: PayCycleCustomDay,
		payDay:       nil,
	}

	// Nil pay day falls back to calendar period.
	period := GetFinancialPeriodForDate(settings, targetDate)

	if period.PeriodLabel != "2024-03" {
		t.Errorf("Expected calendar fallback '2024-03', got '%s'", period.PeriodLabel)
	}
}

func TestGetFinancialPeriodBiWeeklyNilPayDay(t *testing.T) {
	targetDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	settings := &mockSettings{
		payCycleType: PayCycleBiWeekly,
		payDay:       nil,
	}

	// Nil pay day falls back to calendar.
	period := GetFinancialPeriodForDate(settings, targetDate)

	if period.PeriodLabel != "2024-03" {
		t.Errorf("Expected calendar fallback '2024-03', got '%s'", period.PeriodLabel)
	}
}

func TestGetFinancialPeriodBiWeeklyDuration(t *testing.T) {
	targetDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	settings := &mockSettings{
		payCycleType:     PayCycleBiWeekly,
		payDay:           intPtr(1), // Monday
		cycleStartOffset: 0,
	}

	period := GetFinancialPeriodForDate(settings, targetDate)

	// Bi-weekly periods span 14 days (end is at 23:59:59 so ~13 whole days + partial).
	duration := period.EndDate.Sub(period.StartDate)
	if int(duration.Hours()/24) < 13 {
		t.Errorf("Bi-weekly period should be ~14 days, got %v", duration)
	}
}

func TestGetFinancialPeriodLastWeekday(t *testing.T) {
	targetDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	settings := &mockSettings{
		payCycleType:     PayCycleLastWeekday,
		cycleStartOffset: 0,
	}

	period := GetFinancialPeriodForDate(settings, targetDate)

	// Start date should be a weekday.
	if period.StartDate.Weekday() == time.Saturday || period.StartDate.Weekday() == time.Sunday {
		t.Errorf("Period start should be a weekday, got %v", period.StartDate.Weekday())
	}
}

func TestGetFinancialPeriodsCalendar(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

	periods := GetFinancialPeriods(nil, start, end)

	if len(periods) != 3 {
		t.Errorf("Expected 3 periods (Jan, Feb, Mar), got %d", len(periods))
	}
}

func TestGetFinancialPeriodsSingleMonth(t *testing.T) {
	start := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC)

	periods := GetFinancialPeriods(nil, start, end)

	if len(periods) != 1 {
		t.Errorf("Expected 1 period, got %d", len(periods))
	}
	if periods[0].PeriodLabel != "2024-05" {
		t.Errorf("Expected '2024-05', got '%s'", periods[0].PeriodLabel)
	}
}

func TestGetFinancialPeriodsSixMonths(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

	periods := GetFinancialPeriods(nil, start, end)

	if len(periods) != 6 {
		t.Errorf("Expected 6 periods, got %d", len(periods))
	}
}

func TestAdjustDateRangeForPayCycleNilSettings(t *testing.T) {
	start := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)

	newStart, newEnd := AdjustDateRangeForPayCycle(nil, start, end)

	if !newStart.Equal(start) || !newEnd.Equal(end) {
		t.Error("Calendar mode should not adjust dates")
	}
}

func TestAdjustDateRangeForPayCycleCalendar(t *testing.T) {
	settings := &mockSettings{payCycleType: PayCycleCalendar}
	start := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)

	newStart, newEnd := AdjustDateRangeForPayCycle(settings, start, end)

	if !newStart.Equal(start) || !newEnd.Equal(end) {
		t.Error("Calendar mode should not adjust dates")
	}
}

func TestAdjustDateRangeForPayCycleCustomDay(t *testing.T) {
	settings := &mockSettings{
		payCycleType:     PayCycleCustomDay,
		payDay:           intPtr(15),
		cycleStartOffset: 0,
	}
	start := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC)

	newStart, newEnd := AdjustDateRangeForPayCycle(settings, start, end)

	// newStart should be on or before start.
	if newStart.After(start) {
		t.Errorf("Adjusted start %v should be <= original start %v", newStart, start)
	}
	// newEnd should be on or after end.
	if newEnd.Before(end) {
		t.Errorf("Adjusted end %v should be >= original end %v", newEnd, end)
	}
}
