# API Documentation - Money Management System

## Base URL
```
http://localhost:8080/api
```

## Table of Contents
1. [Authentication](#authentication)
2. [User Management](#user-management)
3. [Bank Management](#bank-management)
4. [Category Management](#category-management)
5. [Transaction Management](#transaction-management)
6. [Budget Management](#budget-management)
7. [Budget Alerts](#budget-alerts)
8. [Analytics](#analytics)

---

## Authentication

### Register
Create a new user account.

**Endpoint:** `POST /api/register`

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "is_verified": true,
    "is_admin": false,
    "created_at": "2025-12-27T10:00:00Z",
    "updated_at": "2025-12-27T10:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid input data
- `409 Conflict` - User already exists

---

### Login
Authenticate user and receive JWT token.

**Endpoint:** `POST /api/login`

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "is_verified": true,
      "is_admin": false
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid input data
- `401 Unauthorized` - Invalid email or incorrect password

---

### Logout
Logout user (client should delete token).

**Endpoint:** `POST /api/logout`

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Logout successful",
  "data": {
    "user_id": 1
  }
}
```

**Error Responses:**
- `401 Unauthorized` - User not authenticated

---

## User Management

### Get All Users
Retrieve list of all users (no auth required).

**Endpoint:** `GET /api/users`

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com"
    }
  ]
}
```

---

### Create User
Create a new user (no auth required).

**Endpoint:** `POST /api/users`

**Request Body:**
```json
{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "password": "password123"
}
```

---

### Update User
Update user details (no auth required).

**Endpoint:** `PUT /api/users/:id`

**Request Body:**
```json
{
  "name": "Jane Updated",
  "email": "jane.updated@example.com"
}
```

---

### Delete User
Delete a user (no auth required).

**Endpoint:** `DELETE /api/users/:id`

---

## Bank Management

### Get All Banks
Retrieve list of all banks (no auth required).

**Endpoint:** `GET /api/banks`

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Banks retrieved successfully",
  "data": [
    {
      "id": 1,
      "bank_name": "Bank ABC",
      "account_number": "1234567890",
      "account_holder": "John Doe",
      "balance": 1000000
    }
  ]
}
```

---

### Create Bank
Add a new bank account (no auth required).

**Endpoint:** `POST /api/banks`

**Request Body:**
```json
{
  "bank_name": "Bank XYZ",
  "account_number": "0987654321",
  "account_holder": "John Doe",
  "balance": 5000000
}
```

---

### Delete Bank
Delete a bank account (no auth required).

**Endpoint:** `DELETE /api/banks/:id`

---

## Category Management

### Get All Categories
Retrieve all categories (no auth required).

**Endpoint:** `GET /api/categories`

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Categories retrieved successfully",
  "data": [
    {
      "id": 1,
      "category_name": "Food",
      "description": "Food and dining expenses",
      "created_at": "2025-12-27T10:00:00Z"
    }
  ]
}
```

---

### Get User's Categories
Retrieve categories for authenticated user.

**Endpoint:** `GET /api/my-categories`

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Categories retrieved successfully",
  "data": [
    {
      "id": 1,
      "category_name": "Food",
      "user_id": 1
    }
  ]
}
```

---

### Create Category
Create a new category.

**Endpoint:** `POST /api/categories`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "category_name": "Transportation",
  "description": "Vehicle and transport expenses"
}
```

---

### Delete Category
Delete a category (no auth required).

**Endpoint:** `DELETE /api/categories/:id`

---

### Get Initial Data
Get all banks, categories, and users for transaction creation.

**Endpoint:** `GET /api/transaction/initial-data`

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Initial data successfully fetched",
  "data": {
    "banks": [...],
    "categories": [...],
    "users": [...]
  }
}
```

---

## Transaction Management

### Get Transactions
Retrieve paginated list of transactions with optional filters.

**Endpoint:** `GET /api/transactions`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `page` (optional, default: 1) - Page number
- `limit` (optional, default: 10) - Items per page
- `start_date` (optional) - Filter by start date (YYYY-MM-DD)
- `end_date` (optional) - Filter by end date (YYYY-MM-DD)
- `transaction_type` (optional) - Filter by type (1=Income, 2=Expense)
- `category_id` (optional) - Filter by category
- `bank_id` (optional) - Filter by bank

**Example:**
```
GET /api/transactions?page=1&limit=10&transaction_type=2&start_date=2025-01-01
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Transactions fetched successfully",
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "description": "Grocery shopping",
      "amount": 150000,
      "transaction_type": 2,
      "date": "2025-12-26T00:00:00Z",
      "category_id": 1,
      "bank_id": 1,
      "Category": {
        "id": 1,
        "category_name": "Food"
      },
      "Bank": {
        "id": 1,
        "bank_name": "Bank ABC"
      }
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items": 50,
    "items_per_page": 10
  }
}
```

---

### Get Transaction by ID
Retrieve a specific transaction.

**Endpoint:** `GET /api/transactions/:id`

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Transaction fetched successfully",
  "data": {
    "id": 1,
    "description": "Grocery shopping",
    "amount": 150000,
    "transaction_type": 2,
    "date": "2025-12-26T00:00:00Z"
  }
}
```

---

### Create Transaction
Create a new transaction.

**Endpoint:** `POST /api/transaction`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "Description": "Salary payment",
  "Amount": "5000000",
  "TransactionType": "Income",
  "Date": "2025-12-27T10:00:00Z",
  "CategoryID": 1,
  "BankID": 1
}
```

**Alternative format:**
```json
{
  "Description": "Coffee shop",
  "Amount": 50000,
  "TransactionType": 2,
  "Date": "2025-12-27",
  "CategoryID": 1,
  "BankID": 1
}
```

**Notes:**
- `TransactionType` can be "Income"/"Expense" or 1/2
- `Amount` can be string or number
- `Date` supports ISO 8601 or YYYY-MM-DD format
- **Creating an expense transaction automatically triggers budget alert checking**

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Transaction created successfully",
  "data": {
    "id": 1,
    "description": "Coffee shop",
    "amount": 50000,
    "transaction_type": 2
  }
}
```

---

## Budget Management

### Create Budget
Create a new budget for a category.

**Endpoint:** `POST /api/budgets`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 500000,
  "period": "monthly",
  "start_date": "2025-01-01",
  "alert_at": 80,
  "description": "Monthly food budget"
}
```

**Fields:**
- `category_id` (required) - Category to budget for
- `amount` (required) - Budget amount
- `period` (required) - "monthly" or "yearly"
- `start_date` (required) - Budget start date (YYYY-MM-DD)
- `alert_at` (optional, default: 80) - Alert threshold percentage (1-100)
- `description` (optional) - Budget description

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Budget created successfully",
  "data": {
    "id": 1,
    "category_id": 1,
    "category_name": "Food",
    "amount": 500000,
    "period": "monthly",
    "start_date": "2025-01-01T00:00:00Z",
    "end_date": "2025-01-31T00:00:00Z",
    "is_active": true,
    "alert_at": 80,
    "description": "Monthly food budget"
  }
}
```

---

### Get All Budgets
Retrieve paginated list of budgets with filters.

**Endpoint:** `GET /api/budgets`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `page` (optional, default: 1)
- `page_size` (optional, default: 10)
- `category_id` (optional) - Filter by category
- `period` (optional) - Filter by period (monthly/yearly)
- `is_active` (optional) - Filter by active status (true/false)
- `search` (optional) - Search in description
- `sort_by` (optional, default: created_at)
- `sort_dir` (optional, default: desc)

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Budgets retrieved successfully",
  "data": {
    "items": [
      {
        "id": 1,
        "category_name": "Food",
        "amount": 500000,
        "spent_amount": 300000,
        "remaining_amount": 200000,
        "percentage_used": 60,
        "status": "safe",
        "days_remaining": 15
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total_items": 5,
      "total_pages": 1
    }
  }
}
```

**Status values:**
- `safe` - Under alert threshold
- `warning` - At or above alert threshold but under 100%
- `exceeded` - Over 100% of budget

---

### Get Budget by ID
Retrieve a specific budget with spending details.

**Endpoint:** `GET /api/budgets/:id`

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Budget retrieved successfully",
  "data": {
    "id": 1,
    "category_id": 1,
    "category_name": "Food",
    "amount": 500000,
    "period": "monthly",
    "start_date": "2025-01-01T00:00:00Z",
    "end_date": "2025-01-31T00:00:00Z",
    "is_active": true,
    "alert_at": 80,
    "description": "Monthly food budget",
    "spent_amount": 300000,
    "remaining_amount": 200000,
    "percentage_used": 60,
    "status": "safe",
    "days_remaining": 15
  }
}
```

---

### Update Budget
Update an existing budget.

**Endpoint:** `PUT /api/budgets/:id`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "amount": 600000,
  "alert_at": 75,
  "description": "Updated monthly food budget",
  "is_active": true
}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Budget updated successfully",
  "data": {
    "id": 1,
    "amount": 600000,
    "alert_at": 75
  }
}
```

---

### Delete Budget
Delete a budget.

**Endpoint:** `DELETE /api/budgets/:id`

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Budget deleted successfully",
  "data": null
}
```

---

### Get Budget Status
Get status of all active budgets.

**Endpoint:** `GET /api/budgets/status`

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Budget status retrieved successfully",
  "data": [
    {
      "id": 1,
      "category_name": "Food",
      "amount": 500000,
      "spent_amount": 420000,
      "remaining_amount": 80000,
      "percentage_used": 84,
      "status": "warning",
      "days_remaining": 10
    }
  ]
}
```

---

## Budget Alerts

### How Budget Alerts Work

Alerts are **automatically created** when:
1. User creates an expense transaction (TransactionType = 2)
2. System checks all active budgets for the user
3. If spending reaches or exceeds alert threshold (default 80%), an alert is created
4. Duplicate alerts are prevented (one alert per ~5% range)

**Alert Flow:**
```
Expense Transaction → Save to DB → CheckBudgetAlerts()
→ Calculate spending % for active budgets
→ If % ≥ alert_at: Create alert (if not duplicate)
→ Alert available via GET /budget-alerts
```

---

### Get Budget Alerts
Retrieve budget alerts for authenticated user.

**Endpoint:** `GET /api/budget-alerts`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `unread_only` (optional) - Set to "true" to get only unread alerts

**Example:**
```
GET /api/budget-alerts?unread_only=true
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Alerts retrieved successfully",
  "data": [
    {
      "id": 1,
      "budget_id": 1,
      "percentage": 85,
      "spent_amount": 425000,
      "message": "You have reached 85% of your Food budget",
      "is_read": false,
      "created_at": "2025-12-27T10:30:00Z",
      "category_id": 1,
      "category_name": "Food",
      "budget_amount": 500000
    }
  ]
}
```

---

### Mark Alert as Read
Mark a budget alert as read.

**Endpoint:** `PUT /api/budget-alerts/:id/read`

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "status": true,
  "message": "Alert marked as read",
  "data": null
}
```

---

## Analytics

All analytics endpoints require authentication.

### Get Dashboard Summary
Get overview statistics for the dashboard.

**Endpoint:** `GET /api/analytics/dashboard`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `start_date` (optional) - Start date (YYYY-MM-DD)
- `end_date` (optional) - End date (YYYY-MM-DD)

---

### Get Spending by Category
Get spending breakdown by category.

**Endpoint:** `GET /api/analytics/spending-by-category`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `start_date` (optional)
- `end_date` (optional)

---

### Get Spending by Bank
Get spending breakdown by bank account.

**Endpoint:** `GET /api/analytics/spending-by-bank`

**Headers:**
```
Authorization: Bearer {token}
```

---

### Get Income vs Expense
Get income vs expense comparison.

**Endpoint:** `GET /api/analytics/income-vs-expense`

**Headers:**
```
Authorization: Bearer {token}
```

---

### Get Trend Analysis
Get spending trend analysis.

**Endpoint:** `GET /api/analytics/trend`

**Headers:**
```
Authorization: Bearer {token}
```

---

### Get Monthly Comparison
Compare current month with previous month.

**Endpoint:** `GET /api/analytics/monthly-comparison`

**Headers:**
```
Authorization: Bearer {token}
```

---

### Get Yearly Report
Get yearly financial report.

**Endpoint:** `GET /api/analytics/yearly-report`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `year` (optional) - Year to get report for

---

### Get Category Trend
Get spending trend for a specific category.

**Endpoint:** `GET /api/analytics/category-trend/:category_id`

**Headers:**
```
Authorization: Bearer {token}
```

---

## Error Responses

All endpoints may return these error responses:

### 400 Bad Request
```json
{
  "status": false,
  "message": "Invalid input data",
  "data": null
}
```

### 401 Unauthorized
```json
{
  "status": false,
  "message": "User not authenticated",
  "data": null
}
```

### 404 Not Found
```json
{
  "status": false,
  "message": "Resource not found",
  "data": null
}
```

### 409 Conflict
```json
{
  "status": false,
  "message": "Resource already exists",
  "data": null
}
```

### 500 Internal Server Error
```json
{
  "status": false,
  "message": "Internal server error",
  "data": null
}
```

---

## Authentication Flow

1. **Register** - Create account with email/password
2. **Login** - Receive JWT token
3. **Use Token** - Include in Authorization header: `Bearer {token}`
4. **Logout** - Client deletes token

**Token Format:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Important Notes

### Transaction Types
- `1` or `"Income"` - Income transaction
- `2` or `"Expense"` - Expense transaction

### Date Formats
Supported date formats:
- ISO 8601: `2025-12-27T10:00:00Z`
- Simple date: `2025-12-27`

### Budget Periods
- `monthly` - Monthly budget
- `yearly` - Yearly budget

### Budget Alert Thresholds
- Default: 80% of budget amount
- Configurable per budget (1-100)
- Alerts created automatically on expense transactions

### Pagination
Default pagination:
- Page: 1
- Page size: 10

---

## Testing Examples

### Using cURL

**Register:**
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'
```

**Create Transaction (with token):**
```bash
curl -X POST http://localhost:8080/api/transaction \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "Description":"Lunch",
    "Amount":"50000",
    "TransactionType":"Expense",
    "Date":"2025-12-27",
    "CategoryID":1,
    "BankID":1
  }'
```

**Get Budgets:**
```bash
curl -X GET "http://localhost:8080/api/budgets?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Get Unread Alerts:**
```bash
curl -X GET "http://localhost:8080/api/budget-alerts?unread_only=true" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## Postman Collection

Import these examples into Postman for easier testing:

1. Create environment with variables:
   - `base_url`: `http://localhost:8080/api`
   - `token`: (will be set after login)

2. Add tests to Login request to save token:
```javascript
if (pm.response.code === 200) {
    var jsonData = pm.response.json();
    pm.environment.set("token", jsonData.data.token);
}
```

3. Use `{{token}}` in Authorization headers for protected endpoints

---

## Support

For issues or questions:
- Repository: fawazbayureksa/money-manage
- Branch: enhacment

---

**Last Updated:** December 27, 2025
