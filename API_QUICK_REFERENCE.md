# Quick API Reference

## Authentication Required Endpoints

All endpoints below require:
```
Authorization: Bearer YOUR_JWT_TOKEN
```

---

## Budget Management

### Create Budget
```
POST /api/budgets
{
  "category_id": 1,
  "amount": 5000000,
  "period": "monthly",
  "start_date": "2025-01-01T00:00:00Z",
  "alert_at": 80
}
```

### Get All Budgets
```
GET /api/budgets?page=1&page_size=10
GET /api/budgets?category_id=1
GET /api/budgets?period=monthly
GET /api/budgets?is_active=true
```

### Get Budget Detail
```
GET /api/budgets/1
```

### Update Budget
```
PUT /api/budgets/1
{
  "amount": 6000000,
  "alert_at": 85
}
```

### Delete Budget
```
DELETE /api/budgets/1
```

### Get Active Budgets Status
```
GET /api/budgets/status
```

### Get Alerts
```
GET /api/budget-alerts
GET /api/budget-alerts?unread_only=true
```

### Mark Alert as Read
```
PUT /api/budget-alerts/1/read
```

---

## Analytics

### Dashboard Summary
```
GET /api/analytics/dashboard

Returns:
- Current month stats
- Last month stats
- Top 5 categories
- Recent 10 transactions
- Budget summary
```

### Spending by Category
```
GET /api/analytics/spending-by-category?start_date=2025-01-01&end_date=2025-01-31

Returns category breakdown with percentages
```

### Spending by Bank
```
GET /api/analytics/spending-by-bank?start_date=2025-01-01&end_date=2025-01-31

Returns bank usage distribution
```

### Income vs Expense
```
GET /api/analytics/income-vs-expense?start_date=2025-01-01&end_date=2025-01-31

Returns:
- Total income
- Total expense
- Net amount
- Savings rate
- Transaction counts
```

### Trend Analysis
```
GET /api/analytics/trend?start_date=2025-01-01&end_date=2025-12-31&group_by=month

group_by options: day, week, month, year
Returns data points for charts
```

### Monthly Comparison
```
GET /api/analytics/monthly-comparison?months=6

Returns last N months with percentage changes
```

### Yearly Report
```
GET /api/analytics/yearly-report?year=2025

Returns:
- Annual totals
- Monthly breakdown
- Top categories
- Net savings
```

### Category Trend
```
GET /api/analytics/category-trend/1?start_date=2025-01-01&end_date=2025-01-31

Returns spending trend for specific category
```

---

## Existing Endpoints

### Auth
```
POST /api/register
POST /api/login
```

### Users
```
GET    /api/users?page=1&page_size=10
POST   /api/users
PUT    /api/users/:id
DELETE /api/users/:id
```

### Banks
```
GET    /api/banks?page=1&page_size=10
POST   /api/banks
DELETE /api/banks/:id
```

### Categories
```
GET    /api/categories
POST   /api/categories (protected)
GET    /api/my-categories (protected)
DELETE /api/categories/:id
```

### Transactions
```
POST /api/transaction (protected)
GET  /api/transaction/initial-data
```

---

## Response Format

### Success
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {...}
}
```

### Error
```json
{
  "success": false,
  "message": "Error description",
  "data": null
}
```

### Paginated Response
```json
{
  "success": true,
  "message": "Data retrieved",
  "data": {
    "data": [...],
    "page": 1,
    "page_size": 10,
    "total_items": 50,
    "total_pages": 5
  }
}
```

---

## Budget Status Values

- `safe`: Under alert threshold (< 80% by default)
- `warning`: At or above alert threshold (>= 80%)
- `exceeded`: Over budget (>= 100%)

---

## Transaction Types

- `1`: Income
- `2`: Expense

---

## Query Parameters

### Pagination
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10, max: 100)
- `search`: Global search
- `sort_by`: Field to sort by
- `sort_dir`: asc or desc (default: desc)

### Budget Filters
- `category_id`: Filter by category
- `period`: monthly or yearly
- `is_active`: true or false

### Analytics
- `start_date`: Start date (required, format: 2025-01-01)
- `end_date`: End date (required, format: 2025-01-31)
- `group_by`: day, week, month, year
- `months`: Number of months (for comparisons)
- `year`: Year (for yearly report)

---

## Common Workflows

### 1. Setup Monthly Budgets
```
1. GET /api/my-categories (get user categories)
2. POST /api/budgets (create budget for each category)
3. GET /api/budgets/status (view all budgets)
```

### 2. Check Financial Health
```
1. GET /api/analytics/dashboard (overview)
2. GET /api/analytics/spending-by-category (details)
3. GET /api/budgets/status (budget compliance)
4. GET /api/budget-alerts (warnings)
```

### 3. Monthly Review
```
1. GET /api/analytics/income-vs-expense (current month)
2. GET /api/analytics/monthly-comparison (vs last month)
3. GET /api/analytics/spending-by-category (breakdown)
4. Adjust budgets based on findings
```

### 4. Year-End Analysis
```
1. GET /api/analytics/yearly-report?year=2025
2. GET /api/analytics/monthly-comparison?months=12
3. Plan budgets for next year
```

---

## Tips

1. Always include date ranges for analytics
2. Use pagination for large datasets
3. Check unread alerts regularly
4. Dashboard endpoint gives comprehensive overview
5. Set realistic alert thresholds (70-90%)
6. Review and adjust budgets monthly
7. Use category trends to identify patterns
8. Compare spending across banks to optimize
