package utils

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCustomTimeMarshalJSON(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	ct := CustomTime{Time: t1}

	data, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("Failed to marshal CustomTime: %v", err)
	}

	expected := `"2024-01-15 10:30:00"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestCustomTimeUnmarshalJSONDateTimeFormat(t *testing.T) {
	var ct CustomTime
	err := json.Unmarshal([]byte(`"2024-01-15 10:30:00"`), &ct)
	if err != nil {
		t.Fatalf("Failed to unmarshal CustomTime with datetime format: %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date: %v", ct.Time)
	}
	if ct.Hour() != 10 || ct.Minute() != 30 {
		t.Errorf("Unexpected time: %v", ct.Time)
	}
}

func TestCustomTimeUnmarshalJSONRFC3339(t *testing.T) {
	var ct CustomTime
	err := json.Unmarshal([]byte(`"2024-01-15T10:30:00Z"`), &ct)
	if err != nil {
		t.Fatalf("Failed to unmarshal RFC3339 CustomTime: %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date: %v", ct.Time)
	}
}

func TestCustomTimeUnmarshalJSONDateOnly(t *testing.T) {
	var ct CustomTime
	err := json.Unmarshal([]byte(`"2024-01-15"`), &ct)
	if err != nil {
		t.Fatalf("Failed to unmarshal date-only CustomTime: %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date: %v", ct.Time)
	}
}

func TestCustomTimeUnmarshalJSONInvalidFormat(t *testing.T) {
	var ct CustomTime
	err := json.Unmarshal([]byte(`"not-a-date"`), &ct)
	if err == nil {
		t.Error("Expected error for invalid date format, got nil")
	}
}

func TestCustomTimeRoundTrip(t *testing.T) {
	original := time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC)
	ct := CustomTime{Time: original}

	data, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded CustomTime
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !decoded.Time.Equal(original) {
		t.Errorf("Round-trip failed: expected %v, got %v", original, decoded.Time)
	}
}

func TestCustomTimeValue(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	ct := CustomTime{Time: t1}

	val, err := ct.Value()
	if err != nil {
		t.Fatalf("Failed to get Value: %v", err)
	}

	v, ok := val.(time.Time)
	if !ok {
		t.Fatalf("Expected time.Time from Value, got %T", val)
	}
	if !v.Equal(t1) {
		t.Errorf("Expected %v, got %v", t1, v)
	}
}

func TestCustomTimeScanTimeTime(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	ct := &CustomTime{}

	err := ct.Scan(t1)
	if err != nil {
		t.Fatalf("Failed to scan time.Time: %v", err)
	}
	if !ct.Time.Equal(t1) {
		t.Errorf("Expected %v, got %v", t1, ct.Time)
	}
}

func TestCustomTimeScanNil(t *testing.T) {
	ct := &CustomTime{}
	err := ct.Scan(nil)
	if err != nil {
		t.Fatalf("Failed to scan nil: %v", err)
	}
	if !ct.Time.IsZero() {
		t.Error("Expected zero time for nil input")
	}
}

func TestCustomTimeScanString(t *testing.T) {
	ct := &CustomTime{}
	err := ct.Scan("2024-01-15 10:30:00")
	if err != nil {
		t.Fatalf("Failed to scan string: %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date after scan: %v", ct.Time)
	}
}

func TestCustomTimeScanBytes(t *testing.T) {
	ct := &CustomTime{}
	err := ct.Scan([]byte("2024-01-15"))
	if err != nil {
		t.Fatalf("Failed to scan bytes: %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date after scan: %v", ct.Time)
	}
}

func TestCustomTimeScanInvalidType(t *testing.T) {
	ct := &CustomTime{}
	err := ct.Scan(12345)
	if err == nil {
		t.Error("Expected error for unsupported scan type")
	}
}

func TestCustomTimeUnmarshalTextDateOnly(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalText([]byte("2024-01-15"))
	if err != nil {
		t.Fatalf("Failed to unmarshal text (date): %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date: %v", ct.Time)
	}
}

func TestCustomTimeUnmarshalTextEmpty(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalText([]byte(""))
	if err != nil {
		t.Fatalf("Failed to unmarshal empty text: %v", err)
	}
	if !ct.Time.IsZero() {
		t.Error("Expected zero time for empty text input")
	}
}

func TestCustomTimeUnmarshalTextDateTime(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalText([]byte("2024-01-15 10:30:00"))
	if err != nil {
		t.Fatalf("Failed to unmarshal text (datetime): %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date: %v", ct.Time)
	}
}

func TestCustomTimeUnmarshalTextRFC3339(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalText([]byte("2024-01-15T10:30:00Z"))
	if err != nil {
		t.Fatalf("Failed to unmarshal RFC3339 text: %v", err)
	}
	if ct.Year() != 2024 || ct.Month() != 1 || ct.Day() != 15 {
		t.Errorf("Unexpected date: %v", ct.Time)
	}
}

func TestCustomTimeUnmarshalTextInvalid(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalText([]byte("not-a-date"))
	if err == nil {
		t.Error("Expected error for invalid text format")
	}
}
