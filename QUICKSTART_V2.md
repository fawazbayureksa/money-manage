# Quick Start Guide - Transaction V2 API

This guide helps you quickly set up and test the Transaction V2 API with automatic balance management.

---

## Prerequisites

- Go 1.20+ installed
- Database access with migration permissions
- Valid JWT token for API authentication

---

## Step 1: Run Database Migration

```bash
# Navigate to project root
cd /path/to/my-api

# Apply migration (add asset_id to transactions table)
go run cmd/migrate/main.go -action=up
```

**Expected Output:**
```
Starting migration: Add AssetID to Transactions
Column asset_id already exists in transactions table
Migrating existing transactions from bank_id to asset_id...
Migrated 150 transactions to use asset_id
Created 5 default assets for banks
Migration completed successfully!
```

---

## Step 2: Start the API Server

```bash
go run main.go
```

The API will be available at `http://localhost:8080`

---

## Step 3: Test V2 API

### 3.1 Get User's Assets

```bash
curl -X GET http://localhost:8080/api/wallets \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "Assets fetched successfully",
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "name": "Main Checking",
      "type": "bank",
      "balance": 5000.00,
      "currency": "USD",
      "bank_name": "Chase Bank",
      "account_no": "****1234"
    }
  ]
}
```

Note the `asset_id` (e.g., `1`) for the next steps.

---

### 3.2 Create Income Transaction

```bash
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Salary Deposit",
    "category_id": 3,
    "asset_id": 1,
    "amount": 1000,
    "transaction_type": "Income",
    "date": "2025-01-15"
  }'
```

**Success Response (201):**
```json
{
  "success": true,
  "message": "Transaction created successfully",
  "data": {
    "id": 1,
    "description": "Salary Deposit",
    "amount": 1000,
    "transaction_type": 1,
    "asset_id": 1,
    "user_id": 1,
    "category_id": 3
  }
}
```

---

### 3.3 Verify Balance Increased

```bash
curl -X GET http://localhost:8080/api/wallets/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Expected:** Balance should now be `6000.00` (5000 + 1000)

---

### 3.4 Create Expense Transaction

```bash
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Grocery Shopping",
    "category_id": 2,
    "asset_id": 1,
    "amount": 200,
    "transaction_type": "Expense",
    "date": "2025-01-15"
  }'
```

---

### 3.5 Verify Balance Decreased

```bash
curl -X GET http://localhost:8080/api/wallets/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Expected:** Balance should now be `5800.00` (6000 - 200)

---

### 3.6 Test Insufficient Balance

Try to create an expense larger than current balance:

```bash
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Large Expense",
    "category_id": 2,
    "asset_id": 1,
    "amount": 10000,
    "transaction_type": "Expense",
    "date": "2025-01-15"
  }'
```

**Expected Error (400):**
```json
{
  "success": false,
  "message": "Insufficient balance in the selected asset"
}
```

---

### 3.7 Get Transactions with Asset Filter

```bash
curl -X GET "http://localhost:8080/api/v2/transactions?asset_id=1&page=1&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response includes asset information:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "description": "Salary Deposit",
      "amount": 1000,
      "transaction_type": 1,
      "date": "2025-01-15T00:00:00Z",
      "category_name": "Salary",
      "bank_name": "Chase Bank",
      "asset_id": 1,
      "asset_name": "Main Checking",
      "asset_type": "bank",
      "asset_balance": 5800.00,
      "asset_currency": "USD"
    }
  ]
}
```

---

### 3.8 Get Asset Transactions

```bash
curl -X GET "http://localhost:8080/api/v2/assets/1/transactions?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response includes asset summary:**
```json
{
  "success": true,
  "data": {
    "asset_id": 1,
    "asset_name": "Main Checking",
    "asset_type": "bank",
    "current_balance": 5800.00,
    "currency": "USD",
    "total_income": 1000.00,
    "total_expense": 200.00,
    "transactions": [...]
  }
}
```

---

### 3.9 Update Transaction

```bash
curl -X PUT http://localhost:8080/api/v2/transactions/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500
  }'
```

This will:
- Revert old amount from balance (-1000)
- Apply new amount to balance (+1500)
- New balance: 6300.00

---

### 3.10 Delete Transaction (Balance Rollback)

```bash
curl -X DELETE http://localhost:8080/api/v2/transactions/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

This will:
- If it was an income: Subtract amount from balance
- If it was an expense: Add amount back to balance

---

## Compare V1 vs V2

### V1 API (Legacy)
```bash
# Uses BankID, no balance management
curl -X POST http://localhost:8080/api/transaction \
  -H "Authorization: Bearer TOKEN" \
  -d '{"BankID": 1, ...}'
```

### V2 API (New)
```bash
# Uses AssetID, automatic balance management
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer TOKEN" \
  -d '{"AssetID": 1, ...}'
```

---

## Rollback Migration

If you need to rollback the migration:

```bash
go run cmd/migrate/main.go -action=down
```

**Warning:** This will:
- Remove `asset_id` column from transactions table
- Remove foreign key constraint
- V2 API will stop working

---

## Common Issues & Solutions

### Issue: Migration fails with "Column already exists"

**Solution:** The migration has already been run. Check if V2 endpoints work.

---

### Issue: "Insufficient balance" error

**Solution:** Check the current asset balance:
```bash
curl -X GET http://localhost:8080/api/wallets/1 \
  -H "Authorization: Bearer TOKEN"
```

---

### Issue: "Asset does not belong to you"

**Solution:** Ensure you're using your own asset_id. Get your assets:
```bash
curl -X GET http://localhost:8080/api/wallets \
  -H "Authorization: Bearer TOKEN"
```

---

### Issue: Balance not updating

**Solution:** Ensure you're using V2 endpoints (`/api/v2/transactions`), not V1 (`/api/transactions`)

---

## Testing Checklist

- [ ] Migration runs successfully
- [ ] V2 API returns transactions
- [ ] Income transaction increases balance
- [ ] Expense transaction decreases balance
- [ ] Insufficient balance error triggers
- [ ] Asset filter works
- [ ] Asset transactions endpoint works
- [ ] Transaction update recalculates balance
- [ ] Transaction delete rolls back balance
- [ ] V1 API still works (backward compatibility)

---

## Next Steps

1. **Frontend Integration**
   - Read `docs/frontend_integration_guide.md`
   - Update your frontend to use V2 endpoints
   - Replace BankID with AssetID
   - Add balance validation

2. **Testing**
   - Test with concurrent transactions
   - Test with large datasets
   - Monitor API performance

3. **Deployment**
   - Deploy to staging first
   - Run full test suite
   - Monitor for issues
   - Gradual frontend migration

---

## Support

- **API Documentation:** `docs/frontend_integration_guide.md`
- **Implementation Guide:** `TRANSACTION_V2_README.md`
- **Enhancement Plan:** `docs/asset_transaction_enhancement_plan.md`

---

## Quick Reference

### V2 Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v2/transactions` | GET | Get transactions (filtered) |
| `/api/v2/transactions/:id` | GET | Get transaction by ID |
| `/api/v2/transactions` | POST | Create transaction |
| `/api/v2/transactions/:id` | PUT | Update transaction |
| `/api/v2/transactions/:id` | DELETE | Delete transaction |
| `/api/v2/assets/:id/transactions` | GET | Get asset transactions |

### Transaction Types

| Value | Type |
|-------|------|
| 1 | Income |
| 2 | Expense |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `asset_id` | number | Asset ID |
| `asset_name` | string | Asset name |
| `asset_balance` | number | Current asset balance |
| `asset_currency` | string | Asset currency (e.g., "USD") |
| `asset_type` | string | Asset type (e.g., "bank") |

---

**Ready to integrate?** Check out the complete frontend guide at `docs/frontend_integration_guide.md`
