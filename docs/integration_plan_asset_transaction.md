# Implementation Plan: Integrating Asset (Wallet) and Transaction Models

## Overview
The current Transaction model references a BankID, but to enable proper wallet/balance management, we need to link transactions directly to assets (wallets). This will allow automatic balance updates when transactions are created.

## Current Models Analysis
- **Asset**: Represents user wallets/accounts with balance tracking
- **Transaction**: Records financial activities but currently lacks direct asset linkage

## Proposed Changes

### 1. Modify Transaction Model (`models/transaction.go`)
- Add `AssetID` field as foreign key to Asset
- Add Asset relation for GORM preloading
- Remove or deprecate BankID if Asset fully replaces bank reference

```go
type Transaction struct {
    // ... existing fields ...
    AssetID uint64 `gorm:"not null;index" json:"asset_id"`
    
    // Relations
    // ... existing relations ...
    Asset Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}
```

### 2. Database Migration
Create migration file to:
- Add `asset_id` column to transactions table
- Add foreign key constraint to assets table
- Ensure data integrity

### 3. Balance Update Logic
Implement automatic balance updates in transaction handlers:
- **Income transactions** (TransactionType=1): Add amount to asset balance
- **Expense transactions** (TransactionType=2): Subtract amount from asset balance
- Add validation to prevent negative balances for expenses

### 4. API Updates
- Update transaction creation/update endpoints to accept `asset_id`
- Add validation that asset belongs to the authenticated user
- Return asset balance in transaction responses

### 5. Business Logic Changes
- Update transaction creation service to:
  1. Validate asset ownership
  2. Check sufficient balance for expenses
  3. Update asset balance atomically with transaction creation
- Consider using database transactions for consistency

### 6. Testing
- Add unit tests for balance updates
- Add integration tests for transaction-asset interactions
- Test edge cases (insufficient funds, concurrent transactions)

### 7. Migration Strategy
- For existing transactions, assign appropriate assets based on BankID
- Run data migration to populate asset_id for historical transactions
- Update existing API consumers

## Benefits
- Automatic balance tracking
- Better financial data integrity
- Simplified wallet management
- Enhanced reporting capabilities

## Risks
- Data migration complexity for existing transactions
- Potential race conditions in concurrent balance updates
- Need for atomic operations to maintain consistency