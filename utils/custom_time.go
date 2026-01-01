package utils

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const DateTimeFormat = "2006-01-02 15:04:05"

type CustomTime struct {
	time.Time
}

// MarshalJSON implements json.Marshaler interface
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", ct.Time.Format(DateTimeFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	str := string(data)
	if len(str) > 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// Try to parse with datetime format first
	parsed, err := time.Parse(DateTimeFormat, str)
	if err != nil {
		// Try ISO 8601 format (RFC3339)
		parsed, err = time.Parse(time.RFC3339, str)
		if err != nil {
			// Try date only format
			parsed, err = time.Parse("2006-01-02", str)
			if err != nil {
				return fmt.Errorf("invalid time format: %s", str)
			}
		}
	}

	ct.Time = parsed
	return nil
}

// Value implements driver.Valuer interface for database storage
func (ct CustomTime) Value() (driver.Value, error) {
	return ct.Time, nil
}

// Scan implements sql.Scanner interface for reading from database
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ct.Time = v
		return nil
	case []byte:
		return ct.UnmarshalJSON(v)
	case string:
		return ct.UnmarshalJSON([]byte(v))
	default:
		return fmt.Errorf("cannot scan type %T into CustomTime", value)
	}
}
