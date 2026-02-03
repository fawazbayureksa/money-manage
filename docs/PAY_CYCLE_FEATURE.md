# Pay Cycle / Financial Period Feature - Implementation Summary

## Overview
This feature allows users to configure custom pay cycles so their financial analytics align with their actual income schedule rather than calendar months. This is particularly useful for users who get paid on specific days (like the last weekday of the month) and want to track expenses from that point forward.

## Problem Solved
**Before**: User gets paid on January 30 (last weekday), but January 31 transactions are counted in "January" analytics, which doesn't match their actual financial cycle.

**After**: User can configure their pay cycle (e.g., "last weekday"), and the system will group January 31 transactions as part of their "February" financial period that started on January 31.

---

## What Was Implemented

### 1. Database Schema
**File**: `db/migrations/20260201100000_create_user_settings_table.up.sql`

Created `user_settings` table with:
- `pay_cycle_type`: ENUM ('calendar', 'last_weekday', 'custom_day', 'bi_weekly')
- `pay_day`: INT (day of month for custom_day, day of week for bi_weekly)
- `cycle_start_offset`: INT (days after payday to start counting, default: 1)

### 2. Models
**File**: `models/user_settings.go`

- `UserSettings` struct with GORM annotations
- `PayCycleType` constants
- Implements `utils.UserSettingsInterface` for dependency inversion

### 3. Utilities
**File**: `utils/pay_cycle.go`

Core date calculation functions:
- `GetLastWeekdayOfMonth()`: Finds last Mon-Fri of month
- `GetFinancialPeriodForDate()`: Returns financial period for any date
- `GetFinancialPeriods()`: Returns all periods in a date range
- Support for 4 pay cycle types:
  - **calendar**: Standard calendar month
  - **last_weekday**: Last working day of month (your use case)
  - **custom_day**: Specific day of month (e.g., 25th)
  - **bi_weekly**: Every 2 weeks

### 4. Repository Layer
**File**: `repositories/user_settings_repository.go`

CRUD operations for user settings:
- `FindByUserID()`: Get settings for a user
- `Create()`, `Update()`, `Delete()`
- `Upsert()`: Create or update in one operation

### 5. Service Layer
**File**: `services/user_settings_service.go`

Business logic:
- Validation of pay cycle configurations
- Returns default settings if none exist
- Converts models to DTOs

### 6. DTOs
**File**: `dto/user_settings_dto.go`

- `CreateUserSettingsRequest` with validation
- `UpdateUserSettingsRequest`
- `UserSettingsResponse`
- Custom `Validate()` methods to ensure consistency

### 7. Controller
**File**: `controllers/user_settings_controller.go`

REST API endpoints:
- `GET /api/user/settings` - Get current settings
- `POST /api/user/settings` - Create settings
- `PUT /api/user/settings` - Update settings
- `DELETE /api/user/settings` - Delete settings (reset to defaults)

### 8. Analytics Integration
**Files**: 
- `repositories/analytics_repository.go`
- `services/analytics_service.go`
- `controllers/analytics_controller.go`

Enhanced analytics to support pay cycles:
- New method: `GetMonthlyTrendByPayCycle()`
- New service method: `GetTrendAnalysisWithPayCycle()`
- Analytics controller now accepts `use_pay_cycle=true` query parameter

### 9. Routes
**File**: `routes/routes.go`

Registered new endpoints and wired dependencies

---

## How to Use

### Step 1: Configure Your Pay Cycle

**For your specific case (last weekday of month):**

```bash
POST /api/user/settings
Authorization: Bearer <your_token>
Content-Type: application/json

{
  "pay_cycle_type": "last_weekday",
  "cycle_start_offset": 1
}
```

**Explanation:**
- `pay_cycle_type: "last_weekday"` - System will use last working day (Mon-Fri) of each month
- `cycle_start_offset: 1` - Your financial period starts 1 day after payday (so if paid on Jan 30, period starts Jan 31)

**Other Examples:**

```json
// Get paid on 25th of every month
{
  "pay_cycle_type": "custom_day",
  "pay_day": 25,
  "cycle_start_offset": 0
}

// Bi-weekly pay on Fridays
{
  "pay_cycle_type": "bi_weekly",
  "pay_day": 5,  // 5 = Friday (0=Sunday, 6=Saturday)
  "cycle_start_offset": 0
}
```

### Step 2: Use Pay Cycle in Analytics

**Standard analytics (calendar month):**
```bash
GET /api/analytics/trend?start_date=2026-01-01&end_date=2026-03-31
```

**Analytics with YOUR pay cycle:**
```bash
GET /api/analytics/trend?start_date=2026-01-01&end_date=2026-03-31&use_pay_cycle=true
```

### Step 3: View Your Settings

```bash
GET /api/user/settings
Authorization: Bearer <your_token>
```

Response:
```json
{
  "success": true,
  "message": "User settings retrieved successfully",
  "data": {
    "id": 1,
    "user_id": 5,
    "pay_cycle_type": "last_weekday",
    "pay_day": null,
    "cycle_start_offset": 1,
    "created_at": "2026-02-01 13:00:00",
    "updated_at": null
  }
}
```

---

## Example: How It Works for Your Case

### Scenario:
- You get paid on **January 30, 2026 (Friday)**
- January 31 is Saturday
- You want to track expenses starting from Jan 31

### Traditional Calendar View:
```
January 2026:
  - Jan 1-31: All transactions grouped as "January"
  
February 2026:
  - Feb 1-28: All transactions grouped as "February"
```

### With Pay Cycle (`last_weekday` + `offset: 1`):
```
Financial Period "January 2026":
  - Dec 31, 2025 to Jan 29, 2026
  - (Previous month's last weekday + 1 day to current month's last weekday)

Financial Period "February 2026":
  - Jan 30, 2026 to Feb 26, 2026
  - (Jan 30 was last weekday, so period starts Jan 31)
```

### API Response with `use_pay_cycle=true`:
```json
{
  "data": [
    {
      "period": "2026-01",
      "period_start": "2025-12-31",
      "period_end": "2026-01-29",
      "income": 5000000,
      "expense": 3200000
    },
    {
      "period": "2026-02",
      "period_start": "2026-01-30",
      "period_end": "2026-02-26",
      "income": 5000000,
      "expense": 3800000
    }
  ]
}
```

---

## Database Migration

To apply the changes to your database:

```bash
cd /Users/fawwazbayureksa/Documents/project/go-projects/my-api
go run cmd/migrate/main.go
```

Or manually run the SQL:
```sql
CREATE TABLE user_settings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL UNIQUE,
    pay_cycle_type ENUM('calendar', 'last_weekday', 'custom_day', 'bi_weekly') NOT NULL DEFAULT 'calendar',
    pay_day INT DEFAULT NULL,
    cycle_start_offset INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## Testing

### Test the Settings Endpoint:
```bash
# Create settings
curl -X POST http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"pay_cycle_type":"last_weekday","cycle_start_offset":1}'

# Get settings
curl -X GET http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer YOUR_TOKEN"

# Update settings
curl -X PUT http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"cycle_start_offset":0}'
```

### Test Analytics with Pay Cycle:
```bash
# Without pay cycle (calendar month)
curl -X GET "http://localhost:8080/api/analytics/trend?start_date=2026-01-01&end_date=2026-03-31" \
  -H "Authorization: Bearer YOUR_TOKEN"

# With pay cycle (your financial periods)
curl -X GET "http://localhost:8080/api/analytics/trend?start_date=2026-01-01&end_date=2026-03-31&use_pay_cycle=true" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Files Created/Modified

### New Files:
1. `db/migrations/20260201100000_create_user_settings_table.up.sql`
2. `db/migrations/20260201100000_create_user_settings_table.down.sql`
3. `models/user_settings.go`
4. `utils/pay_cycle.go`
5. `repositories/user_settings_repository.go`
6. `services/user_settings_service.go`
7. `dto/user_settings_dto.go`
8. `controllers/user_settings_controller.go`

### Modified Files:
1. `repositories/analytics_repository.go` - Added `GetMonthlyTrendByPayCycle()`
2. `services/analytics_service.go` - Added `GetTrendAnalysisWithPayCycle()`
3. `controllers/analytics_controller.go` - Added pay cycle support
4. `routes/routes.go` - Registered new endpoints

---

## Architecture Decisions

### 1. Interface-Based Design
Used `utils.UserSettingsInterface` to avoid circular dependencies between `utils` and `models` packages.

### 2. Offset Flexibility
`cycle_start_offset` allows users to customize when their expense tracking starts relative to payday:
- `0`: Same day as payday
- `1`: Day after payday (your preference)
- `2+`: Multiple days after

### 3. Backward Compatibility
- Default behavior unchanged (calendar months)
- Pay cycle is opt-in via `use_pay_cycle=true` parameter
- Existing users continue working without changes

### 4. Validation
DTOs validate that `pay_day` is provided when required for specific pay cycle types.

---

## Next Steps / Future Enhancements

1. **Frontend Integration**: Update UI to configure pay cycle settings
2. **Dashboard Widget**: Show "current financial period" on dashboard
3. **Budget Integration**: Align budgets with financial periods instead of calendar months
4. **Reports**: Generate reports by financial period instead of month
5. **Multiple Pay Cycles**: Support users with multiple income sources
6. **Calendar View**: Visual calendar showing financial period boundaries

---

## Support

If you encounter issues:
1. Check database migration was successful
2. Verify user_settings table exists
3. Ensure authentication token is valid
4. Check server logs for detailed error messages

For questions about the implementation, refer to the inline code comments or the API documentation.
