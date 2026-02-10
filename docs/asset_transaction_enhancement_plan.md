# Asset-Transaction Integration Enhancement Plan

## Overview
This document outlines recommended enhancements to the asset-transaction integration to transform it from a basic transaction-asset link into a comprehensive wallet management system.

---

## Priority Levels

### 🔴 Essential - Core Functionality
Features that provide critical wallet management capabilities.

### 🟡 Valuable - Enhanced Experience
Features that significantly improve user experience and data utility.

### 🟢 Nice-to-have - Future Enhancements
Advanced features that add premium functionality.

---

## Essential Features

### 1. Asset Transfers
**Description**: Move money between assets/wallets, creating linked debit and credit transactions.

**Use Cases**:
- Transfer from savings to checking account
- Pay credit card from bank account
- Move money between investment accounts

**Implementation**:
- Create `Transfer` model with source and target assets
- Auto-generate two transactions: debit from source, credit to target
- Track transfer status (pending/completed/failed)
- Support scheduled transfers (future date)

**API Endpoints**:
```
POST   /api/transfers
GET    /api/transfers
GET    /api/transfers/:id
PUT    /api/transfers/:id/cancel
```

**Database Schema**:
```go
type Transfer struct {
    ID              uint64          `gorm:"primaryKey"`
    UserID          uint64          `gorm:"not null;index"`
    SourceAssetID   uint64          `gorm:"not null;index"`
    TargetAssetID   uint64          `gorm:"not null;index"`
    Amount          float64         `gorm:"not null"`
    Currency        string          `gorm:"not null"`
    ExchangeRate    float64         `gorm:"default:1.0"`
    Status          string          `gorm:"not null;default:pending"` // pending, completed, failed
    ScheduledDate   *time.Time      `json:"scheduled_date"`
    CompletedDate   *time.Time      `json:"completed_date"`
    Description     string          `gorm:"size:255"`
    SourceTransactionID uint64      `gorm:"index"`
    TargetTransactionID uint64      `gorm:"index"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
    
    SourceAsset     Asset           `gorm:"foreignKey:SourceAssetID"`
    TargetAsset     Asset           `gorm:"foreignKey:TargetAssetID"`
}
```

---

### 2. Asset Transaction History with Balance Timeline
**Description**: API to retrieve all transactions for a specific asset with running balance after each transaction.

**Use Cases**:
- See transaction history with balance progression
- Debug balance discrepancies
- Generate account statements

**API Endpoint**:
```
GET /api/assets/:id/transactions?include_balance=true
```

**Response Example**:
```json
{
  "asset_id": 1,
  "asset_name": "Main Checking",
  "current_balance": 5000.00,
  "transactions": [
    {
      "id": 123,
      "description": "Salary Deposit",
      "amount": 3000.00,
      "transaction_type": 1,
      "date": "2025-01-15T09:00:00Z",
      "balance_after": 6000.00
    },
    {
      "id": 122,
      "description": "Grocery Shopping",
      "amount": -150.00,
      "transaction_type": 2,
      "date": "2025-01-14T14:30:00Z",
      "balance_after": 3000.00
    }
  ]
}
```

---

### 3. Transaction Rollback on Delete
**Description**: When a transaction is deleted, automatically revert the asset balance.

**Implementation**:
- Add soft delete flag to transactions
- Create `RollbackTransaction` service method
- Inverse operation based on transaction type
- Log rollback action in audit trail
- Prevent deletion of old transactions (configurable limit)

**Logic**:
```go
func (s *transactionService) DeleteTransaction(id, userID uint) error {
    tx, err := s.repo.GetByID(id, userID)
    if err != nil {
        return err
    }
    
    // Get asset
    asset, err := s.assetRepo.GetByID(tx.AssetID)
    if err != nil {
        return err
    }
    
    // Revert balance
    if tx.TransactionType == 1 { // Income
        asset.Balance -= tx.Amount
    } else { // Expense
        asset.Balance += tx.Amount
    }
    
    // Use database transaction
    return s.db.Transaction(func(txDB *gorm.DB) error {
        if err := s.assetRepo.Update(asset); err != nil {
            return err
        }
        return s.repo.SoftDelete(id)
    })
}
```

---

### 4. Automatic Balance Synchronization
**Description**: Update asset balances automatically when transactions are created, updated, or deleted.

**Implementation**:
- Add `AssetID` field to Transaction model
- Update transaction creation service to sync balance
- Use database transactions for atomicity
- Add concurrent access protection with row locking

**Service Method**:
```go
func (s *transactionService) CreateTransaction(tx *models.Transaction) error {
    return s.db.Transaction(func(db *gorm.DB) error {
        // Lock asset row for update
        asset, err := s.assetRepo.GetByIDForUpdate(tx.AssetID)
        if err != nil {
            return err
        }
        
        // Validate sufficient balance for expense
        if tx.TransactionType == 2 && asset.Balance < tx.Amount {
            return errors.New("insufficient balance")
        }
        
        // Update balance
        if tx.TransactionType == 1 {
            asset.Balance += tx.Amount
        } else {
            asset.Balance -= tx.Amount
        }
        
        // Create transaction and update asset atomically
        if err := s.assetRepo.Update(asset); err != nil {
            return err
        }
        return s.repo.Create(tx)
    })
}
```

---

### 5. Real-time Balance Endpoints
**Description**: Quick endpoints to get current balances without loading full transaction history.

**API Endpoints**:
```
GET    /api/assets/:id/balance
GET    /api/assets/summary
```

**Response**:
```json
{
  "asset_id": 1,
  "balance": 5000.00,
  "currency": "USD",
  "last_updated": "2025-01-23T10:30:00Z",
  "pending_transactions": 250.00
}
```

---

## Valuable Features

### 6. Recurring Transactions
**Description**: Automatically create future transactions on a schedule (daily, weekly, monthly, yearly).

**Use Cases**:
- Salary deposits
- Rent/mortgage payments
- Subscription bills
- Investment contributions

**API Endpoints**:
```
POST   /api/recurring-transactions
GET    /api/recurring-transactions
GET    /api/recurring-transactions/:id
PUT    /api/recurring-transactions/:id
DELETE /api/recurring-transactions/:id
POST   /api/recurring-transactions/:id/skip
POST   /api/recurring-transactions/:id/pause
POST   /api/recurring-transactions/:id/resume
```

**Database Schema**:
```go
type RecurringTransaction struct {
    ID              uint64      `gorm:"primaryKey"`
    UserID          uint64      `gorm:"not null;index"`
    AssetID         uint64      `gorm:"not null;index"`
    CategoryID      uint        `gorm:"not null;index"`
    Description     string      `gorm:"size:255;not null"`
    Amount          float64     `gorm:"not null"`
    TransactionType int         `gorm:"not null"` // 1=income, 2=expense
    Frequency       string      `gorm:"not null"` // daily, weekly, biweekly, monthly, quarterly, yearly
    NextDueDate     time.Time   `gorm:"not null;index"`
    LastProcessed   *time.Time
    EndDate         *time.Time  `json:"end_date"`
    TotalOccurences *int        `json:"total_occurrences"`
    CurrentCount    int         `gorm:"default:0"`
    Status          string      `gorm:"not null;default:active"` // active, paused, completed, cancelled
    CreatedAt       time.Time
    UpdatedAt       time.Time
    
    Asset           Asset       `gorm:"foreignKey:AssetID"`
    Category        Category    `gorm:"foreignKey:CategoryID"`
}
```

---

### 7. Per-Asset Budgets
**Description**: Set spending limits for individual assets/wallets.

**Use Cases**:
- Limit spending on credit card
- Set weekly cash withdrawal limit
- Track savings goals per account

**API Endpoints**:
```
POST   /api/asset-budgets
GET    /api/asset-budgets
GET    /api/asset-budgets/:id
PUT    /api/asset-budgets/:id
DELETE /api/asset-budgets/:id
GET    /api/asset-budgets/:id/progress
```

**Database Schema**:
```go
type AssetBudget struct {
    ID              uint64      `gorm:"primaryKey"`
    UserID          uint64      `gorm:"not null;index"`
    AssetID         uint64      `gorm:"not null;index:unique_asset_period"`
    CategoryID      *uint       `gorm:"index:unique_asset_period"` // null for overall budget
    Amount          float64     `gorm:"not null"`
    Period          string      `gorm:"not null;default:monthly"` // weekly, monthly, yearly
    StartDate       time.Time   `gorm:"not null"`
    EndDate         *time.Time
    AlertThreshold  float64     `gorm:"default:80"` // percentage
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

---

### 8. Transaction Notes & Memos
**Description**: Add detailed notes to transactions for context and documentation.

**Use Cases**:
- Add receipt numbers
- Note payment references
- Document business expenses

**Implementation**:
- Add `Notes` field to Transaction model
- Support rich text (markdown optional)
- Full-text search on notes

**API Update**:
```json
{
  "description": "Grocery Shopping",
  "notes": "Weekly groceries at Walmart. Receipt #1234567",
  "amount": 150.00
}
```

---

### 9. Multi-Currency Transfers with Conversion
**Description**: Support transfers between assets with different currencies using exchange rates.

**Implementation**:
- Store exchange rate for each transfer
- Fetch real-time rates from external API (optional)
- Track historical exchange rates for accuracy

**API Endpoint**:
```
POST /api/transfers
```

**Request**:
```json
{
  "source_asset_id": 1,
  "target_asset_id": 2,
  "amount": 1000.00,
  "source_currency": "USD",
  "target_currency": "EUR",
  "exchange_rate": 0.85,
  "description": "Transfer to European account"
}
```

---

## Nice-to-have Features

### 10. Transaction Splitting
**Description**: Split a single transaction across multiple categories or budget items.

**Use Cases**:
- Split shopping bill into categories (groceries, household, clothing)
- Divide business expenses by project
- Allocate spending to different budgets

**Database Schema**:
```go
type TransactionSplit struct {
    ID              uint64    `gorm:"primaryKey"`
    TransactionID   uint64    `gorm:"not null;index"`
    CategoryID      uint      `gorm:"not null;index"`
    Amount          float64   `gorm:"not null"`
    Description     string    `gorm:"size:255"`
    
    Transaction     Transaction `gorm:"foreignKey:TransactionID"`
    Category        Category    `gorm:"foreignKey:CategoryID"`
}
```

---

### 11. Transaction Tags
**Description**: Add tags to transactions for flexible organization and filtering.

**Use Cases**:
- Tag transactions by purpose (#vacation, #work, #home)
- Filter by custom groups
- Generate reports by tag

**Database Schema**:
```go
type Tag struct {
    ID          uint64    `gorm:"primaryKey"`
    UserID      uint64    `gorm:"not null;index"`
    Name        string    `gorm:"size:50;not null"`
    Color       string    `gorm:"size:7"` // hex color code
    CreatedAt   time.Time
}

type TransactionTag struct {
    TransactionID   uint64 `gorm:"not null;primaryKey;autoIncrement:false"`
    TagID          uint64 `gorm:"not null;primaryKey;autoIncrement:false"`
    
    Transaction    Transaction `gorm:"foreignKey:TransactionID"`
    Tag            Tag         `gorm:"foreignKey:TagID"`
}
```

---

### 12. Receipt Attachments
**Description**: Store URLs or file references to receipts/invoices for transactions.

**Implementation**:
- Add `ReceiptURL` field to Transaction model
- Support multiple receipts (separate table)
- Integration with cloud storage (S3, etc.)

**Database Schema**:
```go
type Receipt struct {
    ID              uint64      `gorm:"primaryKey"`
    TransactionID   uint64      `gorm:"not null;index"`
    FileName        string      `gorm:"size:255"`
    FileURL         string      `gorm:"size:500;not null"`
    FileType        string      `gorm:"size:50"`
    FileSize        int64
    UploadedAt      time.Time
    
    Transaction     Transaction `gorm:"foreignKey:TransactionID"`
}
```

---

### 13. Balance Alerts
**Description**: Notify users when asset balances drop below or exceed thresholds.

**Use Cases**:
- Low balance warning
- Overdraft prevention
- Savings goal notifications

**API Endpoints**:
```
POST   /api/balance-alerts
GET    /api/balance-alerts
PUT    /api/balance-alerts/:id
DELETE /api/balance-alerts/:id
```

**Database Schema**:
```go
type BalanceAlert struct {
    ID              uint64      `gorm:"primaryKey"`
    UserID          uint64      `gorm:"not null;index"`
    AssetID         uint64      `gorm:"not null;index"`
    AlertType       string      `gorm:"not null"` // below, above, equals
    Threshold       float64     `gorm:"not null"`
    IsActive        bool        `gorm:"default:true"`
    LastTriggered   *time.Time
    NotificationMethod string    `gorm:"not null;default:email"` // email, push, sms
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

---

### 14. Balance Snapshots
**Description**: Track daily/weekly asset balance snapshots for trend analysis.

**Use Cases**:
- Historical balance charts
- Net worth tracking
- Growth rate calculations

**Database Schema**:
```go
type BalanceSnapshot struct {
    ID              uint64      `gorm:"primaryKey"`
    AssetID         uint64      `gorm:"not null;index"`
    UserID          uint64      `gorm:"not null;index"`
    Balance         float64     `gorm:"not null"`
    Currency        string      `gorm:"not null"`
    SnapshotDate    time.Time   `gorm:"not null;index"`
    CreatedAt       time.Time
    
    Asset           Asset       `gorm:"foreignKey:AssetID"`
    
    // Unique constraint to prevent duplicate snapshots
    // UNIQUE(asset_id, snapshot_date)
}
```

**API Endpoints**:
```
GET /api/assets/:id/snapshots?period=month
GET /api/users/:id/net-worth?period=year
```

---

### 15. Transaction Search
**Description**: Advanced search across transactions with full-text and filters.

**API Endpoint**:
```
GET /api/transactions/search
```

**Query Parameters**:
```
?q=search+term
&asset_id=1
&category_id=2
&min_amount=100
&max_amount=500
&start_date=2025-01-01
&end_date=2025-01-31
&tags=vacation,travel
&notes=refund
```

---

## Implementation Phases

### Phase 1: Core Integration (2-3 weeks)
1. Add `AssetID` to Transaction model
2. Implement automatic balance synchronization
3. Add transaction rollback on delete
4. Create asset transaction history API
5. Add real-time balance endpoints

### Phase 2: Essential Features (2-3 weeks)
1. Implement asset transfers
2. Add database migrations
3. Create transfer service and API
4. Add transfer status tracking

### Phase 3: Enhanced Experience (3-4 weeks)
1. Recurring transactions system
2. Per-asset budgets
3. Transaction notes
4. Multi-currency transfer support

### Phase 4: Advanced Features (4-6 weeks)
1. Transaction splitting
2. Transaction tags
3. Receipt attachments
4. Balance alerts
5. Balance snapshots
6. Advanced search

---

## Technical Considerations

### Database Transactions
- Use database transactions for all balance updates
- Implement row locking with `SELECT ... FOR UPDATE`
- Ensure atomicity across asset and transaction operations

### Concurrency
- Handle race conditions in balance updates
- Consider using optimistic locking
- Implement retry logic for failed operations

### Performance
- Add indexes on frequently queried fields
- Consider materialized views for analytics
- Implement caching for balance snapshots

### Security
- Validate asset ownership before operations
- Add audit logging for balance changes
- Implement rate limiting for transfers

---

## API Examples

### Create Transaction with Asset
```http
POST /api/transactions
Content-Type: application/json
Authorization: Bearer <token>

{
  "description": "Salary Deposit",
  "amount": 3000.00,
  "transaction_type": "Income",
  "date": "2025-01-15T09:00:00Z",
  "asset_id": 1,
  "category_id": 3
}
```

### Create Transfer
```http
POST /api/transfers
Content-Type: application/json
Authorization: Bearer <token>

{
  "source_asset_id": 2,
  "target_asset_id": 1,
  "amount": 500.00,
  "description": "Transfer from savings to checking",
  "scheduled_date": "2025-01-25"
}
```

### Get Asset with Balance Timeline
```http
GET /api/assets/1/transactions?include_balance=true&limit=50
Authorization: Bearer <token>
```

---

## Migration Strategy

### 1. Data Migration for Existing Transactions
- Map existing `BankID` transactions to corresponding assets
- Create default assets for banks without matching assets
- Run migration script in batches for large datasets

### 2. API Versioning
- Maintain `/api/v1/transactions` with `BankID` for backward compatibility
- Introduce `/api/v2/transactions` with `AssetID`
- Provide deprecation timeline for v1

### 3. Client Updates
- Update frontend to select asset instead of bank
- Add transfer management UI
- Implement balance display components

---

## Testing Strategy

### Unit Tests
- Balance calculation logic
- Transfer amount conversion
- Recurring transaction scheduling

### Integration Tests
- Asset-transaction synchronization
- Concurrent transaction handling
- Database transaction rollback

### Edge Cases
- Insufficient balance scenarios
- Transfer between same asset
- Delete historical transactions
- Multiple recurring instances on same day

---

## Benefits

### User Experience
- Real-time balance tracking
- Seamless asset management
- Better financial visibility

### Business Value
- Reduced manual reconciliation
- Improved data accuracy
- Enhanced reporting capabilities

### Scalability
- Foundation for advanced features
- Support for multi-currency
- Extensible architecture
