# Advanced Features Documentation

## Budget Management

Complete budget tracking system with alerts and spending monitoring.

### Features

1. Create monthly/yearly budgets per category
2. Track spending against budgets in real-time
3. Automatic alerts when reaching thresholds
4. Budget status monitoring
5. Multiple budgets per user

### API Endpoints

#### Create Budget
```
POST /api/budgets
Authorization: Bearer {token}

Body:
{
  "category_id": 1,
  "amount": 5000000,
  "period": "monthly",
  "start_date": "2025-01-01T00:00:00Z",
  "alert_at": 80,
  "description": "Food budget for January"
}

Response:
{
  "success": true,
  "message": "Budget created successfully",
  "data": {
    "id": 1,
    "category_id": 1,
    "category_name": "Food",
    "amount": 5000000,
    "period": "monthly",
    "start_date": "2025-01-01T00:00:00Z",
    "end_date": "2025-01-31T23:59:59Z",
    "is_active": true,
    "alert_at": 80,
    "description": "Food budget for January"
  }
}
```

#### Get All Budgets with Pagination
```
GET /api/budgets?page=1&page_size=10
GET /api/budgets?category_id=1
GET /api/budgets?period=monthly
GET /api/budgets?is_active=true
Authorization: Bearer {token}

Response:
{
  "success": true,
  "message": "Budgets retrieved successfully",
  "data": {
    "data": [...],
    "page": 1,
    "page_size": 10,
    "total_items": 25,
    "total_pages": 3
  }
}
```

#### Get Budget with Spending
```
GET /api/budgets/:id
Authorization: Bearer {token}

Response:
{
  "success": true,
  "message": "Budget retrieved successfully",
  "data": {
    "id": 1,
    "category_id": 1,
    "category_name": "Food",
    "amount": 5000000,
    "spent_amount": 3500000,
    "remaining_amount": 1500000,
    "percentage_used": 70.0,
    "status": "safe",
    "days_remaining": 10,
    "period": "monthly",
    "is_active": true
  }
}
```

Status values:
- `safe`: Under alert threshold
- `warning`: Reached alert threshold
- `exceeded`: Over budget

#### Update Budget
```
PUT /api/budgets/:id
Authorization: Bearer {token}

Body:
{
  "amount": 6000000,
  "alert_at": 85,
  "description": "Updated food budget"
}
```

#### Delete Budget
```
DELETE /api/budgets/:id
Authorization: Bearer {token}
```

#### Get Budget Status
```
GET /api/budgets/status
Authorization: Bearer {token}

Response: All active budgets with spending data
```

#### Get Budget Alerts
```
GET /api/budget-alerts
GET /api/budget-alerts?unread_only=true
Authorization: Bearer {token}

Response:
{
  "success": true,
  "message": "Alerts retrieved successfully",
  "data": [
    {
      "id": 1,
      "budget_id": 1,
      "percentage": 85,
      "spent_amount": 4250000,
      "message": "You have spent 85% of your monthly budget for Food",
      "is_read": false,
      "created_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

#### Mark Alert as Read
```
PUT /api/budget-alerts/:id/read
Authorization: Bearer {token}
```

---

## Analytics & Reports

Comprehensive financial analytics and reporting features.

### API Endpoints

#### Dashboard Summary
```
GET /api/analytics/dashboard
Authorization: Bearer {token}

Response:
{
  "success": true,
  "message": "Dashboard summary retrieved successfully",
  "data": {
    "current_month": {
      "total_income": 10000000,
      "total_expense": 6500000,
      "net_amount": 3500000,
      "income_count": 5,
      "expense_count": 45,
      "savings_rate": 35.0
    },
    "last_month": {
      "total_income": 9500000,
      "total_expense": 6000000,
      "net_amount": 3500000,
      "savings_rate": 36.84
    },
    "top_categories": [
      {
        "category_id": 1,
        "category_name": "Food",
        "total_amount": 2500000,
        "percentage": 38.46,
        "count": 20
      }
    ],
    "recent_transactions": [...],
    "budget_summary": {
      "total_budgets": 5,
      "active_budgets": 5,
      "exceeded_budgets": 1,
      "warning_budgets": 2,
      "total_budgeted": 15000000,
      "total_spent": 12000000,
      "average_utilization": 80.0
    }
  }
}
```

#### Spending by Category
```
GET /api/analytics/spending-by-category?start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": [
    {
      "category_id": 1,
      "category_name": "Food",
      "total_amount": 2500000,
      "percentage": 38.46,
      "count": 20
    },
    {
      "category_id": 2,
      "category_name": "Transportation",
      "total_amount": 1500000,
      "percentage": 23.08,
      "count": 15
    }
  ]
}
```

#### Spending by Bank
```
GET /api/analytics/spending-by-bank?start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": [
    {
      "bank_id": 1,
      "bank_name": "BCA",
      "total_amount": 4000000,
      "percentage": 61.54,
      "count": 30
    }
  ]
}
```

#### Income vs Expense
```
GET /api/analytics/income-vs-expense?start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": {
    "total_income": 10000000,
    "total_expense": 6500000,
    "net_amount": 3500000,
    "income_count": 5,
    "expense_count": 45,
    "savings_rate": 35.0
  }
}
```

#### Trend Analysis
```
GET /api/analytics/trend?start_date=2025-01-01&end_date=2025-12-31&group_by=month
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": {
    "period": "month",
    "data_points": [
      {
        "date": "2025-01",
        "income": 10000000,
        "expense": 6500000,
        "net": 3500000
      },
      {
        "date": "2025-02",
        "income": 10500000,
        "expense": 7000000,
        "net": 3500000
      }
    ],
    "summary": {...}
  }
}
```

group_by options: `day`, `week`, `month`, `year`

#### Monthly Comparison
```
GET /api/analytics/monthly-comparison?months=6
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": [
    {
      "month": "2024-08",
      "income": 9000000,
      "expense": 5500000,
      "net": 3500000,
      "income_change": 0,
      "expense_change": 0
    },
    {
      "month": "2024-09",
      "income": 9500000,
      "expense": 6000000,
      "net": 3500000,
      "income_change": 5.56,
      "expense_change": 9.09
    }
  ]
}
```

#### Yearly Report
```
GET /api/analytics/yearly-report?year=2025
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": {
    "year": 2025,
    "total_income": 120000000,
    "total_expense": 78000000,
    "net_savings": 42000000,
    "monthly_breakdown": [...],
    "top_expense_categories": [...],
    "top_income_categories": [...]
  }
}
```

#### Category Trend
```
GET /api/analytics/category-trend/:category_id?start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": {
    "category_id": 1,
    "category_name": "Food",
    "data_points": [
      {
        "date": "2025-01-01",
        "expense": 150000
      }
    ],
    "total_amount": 2500000,
    "average_amount": 83333.33
  }
}
```

---

## Use Cases

### Budget Management Use Cases

1. Monthly Budget Planning
```
Create budgets for each spending category
Set alert thresholds (e.g., 80%)
Monitor spending throughout the month
```

2. Budget Alerts
```
Receive alerts when reaching 80% of budget
Get notified when exceeding budget
Track multiple budgets simultaneously
```

3. Budget Analysis
```
Compare budgeted vs actual spending
Identify categories needing adjustment
Track budget compliance over time
```

### Analytics Use Cases

1. Financial Health Check
```
View dashboard summary
Check income vs expense ratio
Monitor savings rate
Review top spending categories
```

2. Spending Pattern Analysis
```
Analyze spending by category
Identify spending trends
Compare month-to-month changes
Find areas to reduce expenses
```

3. Long-term Planning
```
Review yearly reports
Analyze multi-month trends
Set realistic budgets based on history
Track financial goals progress
```

4. Bank Account Management
```
See which bank accounts are used most
Balance spending across accounts
Identify preferred payment methods
```

---

## Data Models

### Budget
```
id: uint
user_id: uint
category_id: uint
amount: int
period: string (monthly/yearly)
start_date: datetime
end_date: datetime
is_active: bool
alert_at: int (percentage)
description: string
created_at: datetime
updated_at: datetime
```

### BudgetAlert
```
id: uint
budget_id: uint
user_id: uint
percentage: int
spent_amount: int
message: string
is_read: bool
created_at: datetime
```

---

## Implementation Notes

1. Budget calculations are real-time based on transactions
2. Alerts are triggered when creating/updating transactions
3. All analytics queries are optimized with indexes
4. Date ranges are inclusive
5. All amounts are in integer (cents/smallest currency unit)
6. Pagination available for all list endpoints
7. All endpoints require authentication

---

## Frontend Integration Tips

1. Dashboard: Call `/api/analytics/dashboard` on login
2. Budget Widget: Use `/api/budgets/status` for quick overview
3. Charts: Use trend/monthly-comparison endpoints
4. Notifications: Poll `/api/budget-alerts?unread_only=true`
5. Reports: Generate PDFs from yearly-report endpoint

---

## Performance Considerations

1. Analytics queries use database indexes on:
   - user_id
   - category_id
   - bank_id
   - date
   - transaction_type

2. Dashboard summary caches recent transactions
3. Budget calculations cache spending totals
4. Use date ranges wisely to avoid slow queries
5. Pagination recommended for large datasets
