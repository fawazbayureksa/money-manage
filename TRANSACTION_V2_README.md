# Transaction V2 API Implementation

## Summary
Implemented Transaction V2 API with automatic asset balance management while maintaining backward compatibility with existing V1 API.

## What's New

### V2 API Features
- ✅ **Asset-based transactions** - Link transactions to assets instead of banks
- ✅ **Automatic balance sync** - Asset balance updates automatically on create/update/delete
- ✅ **Balance validation** - Prevents transactions exceeding available balance
- ✅ **Asset filtering** - Filter transactions by specific assets
- ✅ **Asset transaction history** - Get all transactions for a specific asset
- ✅ **Balance rollback** - Automatic balance reversal when deleting transactions
- ✅ **Concurrent-safe** - Row locking prevents race conditions

### API Versioning
- **V1 API** (Legacy): `/api/transactions` - No balance management
- **V2 API** (New): `/api/v2/transactions` - Full balance management

## Files Created

### Models
- `models/transaction_v2.go` - Transaction model with asset support
- `models/migration_add_asset_id.go` - Migration script for adding asset_id

### DTOs
- `dto/transaction_v2_dto.go` - Request/response DTOs for V2 API

### Repositories
- `repositories/transaction_v2_repository.go` - V2 repository with balance management

### Services
- `services/transaction_v2_service.go` - V2 service layer

### Controllers
- `controllers/transaction_v2_controller.go` - V2 API endpoints

### Migration
- `cmd/migrate/main.go` - Migration runner

### Documentation
- `docs/frontend_integration_guide.md` - Complete frontend integration guide

## Running the Migration

### Step 1: Apply Database Migration

```bash
# Run migration to add asset_id to transactions table
go run cmd/migrate/main.go -action=up
```

### Step 2: Verify Migration

Check that `asset_id` column was added:
```sql
DESCRIBE transactions;
```

### Rollback Migration (if needed)

```bash
go run cmd/migrate/main.go -action=down
```

## API Endpoints

### V2 Transaction Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v2/transactions` | Get paginated transactions (with filters) |
| GET | `/api/v2/transactions/:id` | Get transaction by ID |
| POST | `/api/v2/transactions` | Create new transaction |
| PUT | `/api/v2/transactions/:id` | Update transaction |
| DELETE | `/api/v2/transactions/:id` | Delete transaction |
| GET | `/api/v2/assets/:id/transactions` | Get transactions for specific asset |

### Request/Response Examples

#### Create Transaction (V2)
```bash
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Salary Deposit",
    "category_id": 3,
    "asset_id": 1,
    "amount": 3000,
    "transaction_type": "Income",
    "date": "2025-01-15"
  }'
```

#### Get Transactions with Asset Filter
```bash
curl -X GET "http://localhost:8080/api/v2/transactions?asset_id=1&page=1&limit=20" \
  -H "Authorization: Bearer <token>"
```

#### Get Asset Transactions
```bash
curl -X GET "http://localhost:8080/api/v2/assets/1/transactions?page=1&limit=50" \
  -H "Authorization: Bearer <token>"
```

## Response Format

### Transaction Response (V2)
```json
{
  "success": true,
  "message": "Transactions fetched successfully",
  "data": [
    {
      "id": 1,
      "description": "Salary Deposit",
      "amount": 3000,
      "transaction_type": 1,
      "date": "2025-01-15T09:00:00Z",
      "category_name": "Salary",
      "bank_name": "Chase Bank",
      "asset_id": 1,
      "asset_name": "Main Checking",
      "asset_type": "bank",
      "asset_balance": 5000.00,
      "asset_currency": "USD"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 150,
    "total_pages": 8
  }
}
```

## Error Handling

### Common Errors

| Error | Status | Description |
|-------|--------|-------------|
| `Insufficient balance` | 400 | Expense amount exceeds asset balance |
| `Asset not found` | 404 | Asset doesn't exist |
| `Asset does not belong to you` | 403 | User doesn't own the asset |
| `User not authenticated` | 401 | Invalid or missing JWT token |

## Testing

### Manual Testing

```bash
# 1. Get current asset balance
curl -X GET http://localhost:8080/api/wallets/1 -H "Authorization: Bearer <token>"

# 2. Create income transaction
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"description":"Test","category_id":1,"asset_id":1,"amount":100,"transaction_type":"Income","date":"2025-01-15"}'

# 3. Verify balance increased
curl -X GET http://localhost:8080/api/wallets/1 -H "Authorization: Bearer <token>"

# 4. Create expense transaction
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"description":"Test","category_id":1,"asset_id":1,"amount":50,"transaction_type":"Expense","date":"2025-01-15"}'

# 5. Verify balance decreased
curl -X GET http://localhost:8080/api/wallets/1 -H "Authorization: Bearer <token>"

# 6. Delete transaction
curl -X DELETE http://localhost:8080/api/v2/transactions/<id> -H "Authorization: Bearer <token>"

# 7. Verify balance rolled back
curl -X GET http://localhost:8080/api/wallets/1 -H "Authorization: Bearer <token>"
```

## Frontend Integration

See `docs/frontend_integration_guide.md` for:
- Complete API documentation
- React/Next.js integration examples
- React Native integration examples
- Error handling patterns
- Migration guide from V1 to V2

## Migration from V1

### Steps for Frontend Teams

1. **Update API endpoints**
   ```typescript
   // Before (V1)
   const response = await fetch('/api/transactions');
   
   // After (V2)
   const response = await fetch('/api/v2/transactions');
   ```

2. **Update request payload**
   ```typescript
   // Before (V1)
   { BankID: 1, ... }
   
   // After (V2)
   { AssetID: 1, ... }
   ```

3. **Handle new response fields**
   ```typescript
   // V2 includes asset information
   {
     asset_id: 1,
     asset_name: "Main Checking",
     asset_balance: 5000.00,
     asset_currency: "USD"
   }
   ```

4. **Add balance validation**
   ```typescript
   try {
     await createTransaction(data);
   } catch (error) {
     if (error.message.includes('Insufficient balance')) {
       alert('Not enough funds in selected asset');
     }
   }
   ```

## Database Schema Changes

### Transactions Table
Added `asset_id` column:
```sql
ALTER TABLE transactions 
ADD COLUMN asset_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
ADD INDEX idx_transactions_asset_id (asset_id),
ADD CONSTRAINT fk_transactions_asset 
  FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE RESTRICT;
```

### Data Migration
Existing transactions are migrated automatically:
- Bank-to-asset mapping created
- `asset_id` populated based on `bank_id`
- Default assets created for banks without matching assets

## Technical Details

### Balance Management Flow

**Create Transaction:**
1. Lock asset row (SELECT FOR UPDATE)
2. Validate asset ownership
3. Check sufficient balance (for expenses)
4. Update asset balance (income: +, expense: -)
5. Create transaction record
6. Commit transaction

**Update Transaction:**
1. Lock asset row
2. Revert old transaction effect from balance
3. Validate sufficient balance for new amount
4. Apply new transaction effect
5. Update transaction record
6. Commit transaction

**Delete Transaction:**
1. Lock asset row
2. Get transaction details
3. Revert transaction effect (income: -, expense: +)
4. Delete transaction record
5. Commit transaction

### Concurrency Safety
- Row-level locking prevents race conditions
- Database transactions ensure atomicity
- All balance updates are transactional

## Backward Compatibility

### V1 API Still Works
- Existing V1 endpoints remain unchanged
- V1 transactions don't affect asset balances
- Frontend teams can migrate gradually

### Migration Strategy
1. Deploy migration script
2. Deploy V2 API code
3. Frontend teams migrate to V2 at their pace
4. Monitor and test V2 endpoints
5. Deprecate V1 after migration complete

## Benefits

### For Users
- Real-time balance tracking
- Prevents overdrafts
- Clear transaction-to-asset relationship
- Better financial visibility

### For Developers
- Automatic balance management
- Less manual balance calculations
- Cleaner data model
- Enhanced API capabilities

## Support

### Documentation
- API Guide: `docs/frontend_integration_guide.md`
- Integration Plan: `docs/asset_transaction_enhancement_plan.md`

### Troubleshooting
- See `docs/frontend_integration_guide.md` for common issues
- Check migration logs for data migration issues
- Verify database constraints after migration

## Future Enhancements

See `docs/asset_transaction_enhancement_plan.md` for planned features:
- Asset transfers between wallets
- Recurring transactions
- Per-asset budgets
- Transaction splitting
- Transaction tags
- Balance alerts
- Balance snapshots

## Checklist

### Pre-deployment
- [ ] Run migration on staging database
- [ ] Verify data migration accuracy
- [ ] Test V2 endpoints thoroughly
- [ ] Verify V1 endpoints still work
- [ ] Update API documentation
- [ ] Prepare rollback plan

### Post-deployment
- [ ] Monitor API performance
- [ ] Check balance calculation accuracy
- [ ] Track error rates
- [ ] Collect user feedback
- [ ] Plan frontend migration timeline

---

**Implementation Date:** January 2025  
**Version:** 2.0.0
