# Pay Cycle Quick Start Guide

## Your Use Case: Last Weekday Pay Schedule

### Step 1: Run Database Migration
```bash
cd /Users/fawwazbayureksa/Documents/project/go-projects/my-api
# Apply the migration manually or fix the existing migration issues first
```

### Step 2: Configure Your Pay Cycle
```bash
curl -X POST http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pay_cycle_type": "last_weekday",
    "cycle_start_offset": 1
  }'
```

**What this does:**
- Finds the last weekday (Mon-Fri) of each month as your payday
- Starts your expense tracking period 1 day after payday
- Example: Paid on Jan 30 (Fri) → Period starts Jan 31 (Sat)

### Step 3: Use Analytics with Your Pay Cycle

**Without pay cycle (traditional calendar):**
```bash
curl -X GET "http://localhost:8080/api/analytics/trend?start_date=2026-01-01&end_date=2026-03-31" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**With YOUR pay cycle:**
```bash
curl -X GET "http://localhost:8080/api/analytics/trend?start_date=2026-01-01&end_date=2026-03-31&use_pay_cycle=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/user/settings` | Get your pay cycle settings |
| POST | `/api/user/settings` | Create pay cycle settings |
| PUT | `/api/user/settings` | Update pay cycle settings |
| DELETE | `/api/user/settings` | Reset to calendar month |
| GET | `/api/analytics/trend?use_pay_cycle=true` | Get analytics using your pay cycle |

## Pay Cycle Types

### 1. Calendar (Default)
Standard calendar months (Jan 1-31, Feb 1-28, etc.)
```json
{"pay_cycle_type": "calendar"}
```

### 2. Last Weekday (Your Case)
Last working day of month
```json
{
  "pay_cycle_type": "last_weekday",
  "cycle_start_offset": 1
}
```

### 3. Custom Day
Specific day of month (e.g., 25th)
```json
{
  "pay_cycle_type": "custom_day",
  "pay_day": 25,
  "cycle_start_offset": 0
}
```

### 4. Bi-Weekly
Every 2 weeks on specific day
```json
{
  "pay_cycle_type": "bi_weekly",
  "pay_day": 5,
  "cycle_start_offset": 0
}
```
*Note: pay_day is day of week (0=Sunday, 5=Friday)*

## Example Response

```json
{
  "success": true,
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

## Troubleshooting

### Settings Not Found
If you get "user settings not found", create them first:
```bash
POST /api/user/settings
```

### Migration Failed
Check database connection and run migration manually:
```sql
-- See db/migrations/20260201100000_create_user_settings_table.up.sql
```

### Analytics Not Using Pay Cycle
Make sure you add `use_pay_cycle=true` to the query parameters.

## Code Structure

```
my-api/
├── models/user_settings.go          # Data model
├── utils/pay_cycle.go               # Date calculation logic
├── repositories/user_settings_repository.go  # Database operations
├── services/user_settings_service.go         # Business logic
├── controllers/user_settings_controller.go   # API handlers
├── dto/user_settings_dto.go         # Request/Response structures
└── db/migrations/
    └── 20260201100000_create_user_settings_table.up.sql
```

## Complete Example: January 2026

**Scenario:**
- Today: January 31, 2026 (Saturday)
- Last payday: January 30, 2026 (Friday)
- Next payday: February 27, 2026 (Friday)

**Your Financial Period "February 2026":**
- Starts: January 31, 2026 (day after payday)
- Ends: February 26, 2026 (day before next payday)
- Any transaction on Jan 31 - Feb 26 counts as "February"

**Calendar Month "February 2026":**
- Starts: February 1, 2026
- Ends: February 28, 2026
- Transaction on Jan 31 would count as "January"

## Testing with Postman

1. Import the following collection:
```json
{
  "info": {"name": "Pay Cycle API"},
  "item": [
    {
      "name": "Create Pay Cycle Settings",
      "request": {
        "method": "POST",
        "url": "{{base_url}}/api/user/settings",
        "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
        "body": {
          "mode": "raw",
          "raw": "{\"pay_cycle_type\":\"last_weekday\",\"cycle_start_offset\":1}"
        }
      }
    },
    {
      "name": "Get Analytics with Pay Cycle",
      "request": {
        "method": "GET",
        "url": "{{base_url}}/api/analytics/trend?start_date=2026-01-01&end_date=2026-03-31&use_pay_cycle=true",
        "header": [{"key": "Authorization", "value": "Bearer {{token}}"}]
      }
    }
  ]
}
```

2. Set variables:
   - `base_url`: http://localhost:8080
   - `token`: Your JWT token

---

For detailed documentation, see [PAY_CYCLE_FEATURE.md](./PAY_CYCLE_FEATURE.md)
