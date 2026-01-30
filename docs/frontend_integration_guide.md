# Transaction V2 API - Frontend Integration Guide

## Overview
This guide provides comprehensive documentation for integrating the Transaction V2 API with automatic asset balance management into web and mobile applications.

## Key Features
- ✅ Automatic balance synchronization when transactions are created, updated, or deleted
- ✅ Transaction filtering by asset
- ✅ Asset-specific transaction history
- ✅ Backward compatible with existing V1 API
- ✅ Insufficient balance validation

---

## API Versioning

### V1 API (Legacy)
Uses `BankID` - No automatic balance management

```
GET    /api/transactions
GET    /api/transactions/:id
POST   /api/transaction
DELETE /api/transactions/:id
```

### V2 API (New)
Uses `AssetID` - Automatic balance management

```
GET    /api/v2/transactions
GET    /api/v2/transactions/:id
POST   /api/v2/transactions
PUT    /api/v2/transactions/:id
DELETE /api/v2/transactions/:id
GET    /api/v2/assets/:id/transactions
```

---

## Authentication
All V2 endpoints require authentication via JWT token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

---

## API Endpoints

### 1. Get Transactions (Filtered)
Retrieves paginated transactions with optional filters.

**Endpoint:** `GET /api/v2/transactions`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| page | integer | No | Page number (default: 1) |
| limit | integer | No | Items per page (default: 10) |
| start_date | string | No | Filter transactions from date (YYYY-MM-DD) |
| end_date | string | No | Filter transactions until date (YYYY-MM-DD) |
| transaction_type | string | No | Filter by type: "Income" or "Expense" |
| category_id | integer | No | Filter by category ID |
| asset_id | integer | No | Filter by asset ID |

**Example Request:**
```bash
curl -X GET "https://api.example.com/api/v2/transactions?page=1&limit=20&asset_id=1&start_date=2025-01-01" \
  -H "Authorization: Bearer <token>"
```

**Response:**
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

**Transaction Types:**
- `1` = Income
- `2` = Expense

---

### 2. Get Transaction by ID
Retrieves a specific transaction with asset details.

**Endpoint:** `GET /api/v2/transactions/:id`

**Example Request:**
```bash
curl -X GET "https://api.example.com/api/v2/transactions/123" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "success": true,
  "message": "Transaction fetched successfully",
  "data": {
    "id": 123,
    "description": "Grocery Shopping",
    "amount": 150,
    "transaction_type": 2,
    "date": "2025-01-14T14:30:00Z",
    "category_name": "Groceries",
    "bank_name": "Chase Bank",
    "asset_id": 1,
    "asset_name": "Main Checking",
    "asset_type": "bank",
    "asset_balance": 4850.00,
    "asset_currency": "USD"
  }
}
```

---

### 3. Create Transaction
Creates a new transaction and automatically updates asset balance.

**Endpoint:** `POST /api/v2/transactions`

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| description | string | Yes | Transaction description |
| category_id | integer | Yes | Category ID |
| asset_id | integer | Yes | Asset ID where transaction occurs |
| amount | integer | Yes | Transaction amount (positive integer) |
| transaction_type | string | Yes | "Income" or "Expense" |
| date | string | Yes | Transaction date (YYYY-MM-DD or ISO 8601) |

**Example Request:**
```bash
curl -X POST "https://api.example.com/api/v2/transactions" \
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

**Success Response (201 Created):**
```json
{
  "success": true,
  "message": "Transaction created successfully",
  "data": {
    "id": 124,
    "description": "Salary Deposit",
    "amount": 3000,
    "transaction_type": 1,
    "date": "2025-01-15T00:00:00Z",
    "asset_id": 1,
    "user_id": 1,
    "category_id": 3
  }
}
```

**Error Response (Insufficient Balance):**
```json
{
  "success": false,
  "message": "Insufficient balance in the selected asset"
}
```

**Error Response (Asset Not Found):**
```json
{
  "success": false,
  "message": "Asset does not belong to you"
}
```

---

### 4. Update Transaction
Updates a transaction and recalculates asset balance.

**Endpoint:** `PUT /api/v2/transactions/:id`

**Request Body:** (All fields optional)
```json
{
  "description": "Updated Description",
  "category_id": 4,
  "asset_id": 2,
  "amount": 200,
  "transaction_type": "Expense",
  "date": "2025-01-16"
}
```

**Example Request:**
```bash
curl -X PUT "https://api.example.com/api/v2/transactions/124" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 3500
  }'
```

**Balance Behavior:**
- Old transaction effect is reverted from asset balance
- New transaction effect is applied to asset balance
- Validates sufficient balance for new expense

---

### 5. Delete Transaction
Deletes a transaction and rolls back asset balance.

**Endpoint:** `DELETE /api/v2/transactions/:id`

**Example Request:**
```bash
curl -X DELETE "https://api.example.com/api/v2/transactions/124" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "success": true,
  "message": "Transaction deleted successfully"
}
```

**Balance Behavior:**
- Income: Subtracts amount from asset balance
- Expense: Adds amount back to asset balance

---

### 6. Get Asset Transactions
Retrieves all transactions for a specific asset with summary.

**Endpoint:** `GET /api/v2/assets/:id/transactions`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| page | integer | No | Page number (default: 1) |
| limit | integer | No | Items per page (default: 50) |

**Example Request:**
```bash
curl -X GET "https://api.example.com/api/v2/assets/1/transactions?page=1&limit=20" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "success": true,
  "message": "Asset transactions fetched successfully",
  "data": {
    "asset_id": 1,
    "asset_name": "Main Checking",
    "asset_type": "bank",
    "current_balance": 5000.00,
    "currency": "USD",
    "total_income": 8000.00,
    "total_expense": 3000.00,
    "transactions": [
      {
        "id": 123,
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
    ]
  }
}
```

---

## Web Integration Guide

### React/Next.js Example

#### 1. API Client Setup

```typescript
// lib/api/v2/transactions.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'https://api.example.com/api/v2';

interface Transaction {
  id: number;
  description: string;
  amount: number;
  transaction_type: number;
  date: string;
  category_name: string;
  bank_name?: string;
  asset_id: number;
  asset_name?: string;
  asset_type?: string;
  asset_balance?: number;
  asset_currency?: string;
}

interface CreateTransactionRequest {
  description: string;
  category_id: number;
  asset_id: number;
  amount: number;
  transaction_type: 'Income' | 'Expense';
  date: string;
}

interface UpdateTransactionRequest {
  description?: string;
  category_id?: number;
  asset_id?: number;
  amount?: number;
  transaction_type?: 'Income' | 'Expense';
  date?: string;
}

interface TransactionsResponse {
  success: boolean;
  message: string;
  data: Transaction[];
  pagination: {
    page: number;
    page_size: number;
    total_items: number;
    total_pages: number;
  };
}

class TransactionV2API {
  private getAuthHeaders() {
    const token = localStorage.getItem('token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    };
  }

  async getTransactions(params?: {
    page?: number;
    limit?: number;
    start_date?: string;
    end_date?: string;
    transaction_type?: string;
    category_id?: number;
    asset_id?: number;
  }): Promise<TransactionsResponse> {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, String(value));
        }
      });
    }

    const response = await fetch(`${API_BASE}/transactions?${queryParams}`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch transactions');
    }

    return response.json();
  }

  async getTransactionById(id: number): Promise<{ success: boolean; data: Transaction }> {
    const response = await fetch(`${API_BASE}/transactions/${id}`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch transaction');
    }

    return response.json();
  }

  async createTransaction(data: CreateTransactionRequest): Promise<{ success: boolean; data: Transaction }> {
    const response = await fetch(`${API_BASE}/transactions`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to create transaction');
    }

    return response.json();
  }

  async updateTransaction(id: number, data: UpdateTransactionRequest): Promise<{ success: boolean; data: Transaction }> {
    const response = await fetch(`${API_BASE}/transactions/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to update transaction');
    }

    return response.json();
  }

  async deleteTransaction(id: number): Promise<{ success: boolean }> {
    const response = await fetch(`${API_BASE}/transactions/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to delete transaction');
    }

    return response.json();
  }

  async getAssetTransactions(assetId: number, page = 1, limit = 20) {
    const response = await fetch(`${API_BASE}/assets/${assetId}/transactions?page=${page}&limit=${limit}`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch asset transactions');
    }

    return response.json();
  }
}

export const transactionV2API = new TransactionV2API();
export type { Transaction, CreateTransactionRequest, UpdateTransactionRequest };
```

#### 2. React Component Example

```typescript
// components/TransactionForm.tsx
import { useState } from 'react';
import { transactionV2API, CreateTransactionRequest } from '@/lib/api/v2/transactions';

interface TransactionFormProps {
  onSuccess: () => void;
}

export default function TransactionForm({ onSuccess }: TransactionFormProps) {
  const [formData, setFormData] = useState<CreateTransactionRequest>({
    description: '',
    category_id: 0,
    asset_id: 0,
    amount: 0,
    transaction_type: 'Income',
    date: new Date().toISOString().split('T')[0],
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await transactionV2API.createTransaction(formData);
      onSuccess();
      // Reset form
      setFormData({
        ...formData,
        description: '',
        amount: 0,
      });
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div>
        <label className="block text-sm font-medium mb-1">Description</label>
        <input
          type="text"
          value={formData.description}
          onChange={(e) => setFormData({ ...formData, description: e.target.value })}
          className="w-full px-3 py-2 border rounded"
          required
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Asset</label>
        <select
          value={formData.asset_id}
          onChange={(e) => setFormData({ ...formData, asset_id: parseInt(e.target.value) })}
          className="w-full px-3 py-2 border rounded"
          required
        >
          <option value="">Select Asset</option>
          <option value="1">Main Checking ($5,000.00)</option>
          <option value="2">Savings ($10,000.00)</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Amount</label>
        <input
          type="number"
          value={formData.amount}
          onChange={(e) => setFormData({ ...formData, amount: parseInt(e.target.value) })}
          className="w-full px-3 py-2 border rounded"
          min="1"
          required
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Type</label>
        <select
          value={formData.transaction_type}
          onChange={(e) => setFormData({ ...formData, transaction_type: e.target.value as 'Income' | 'Expense' })}
          className="w-full px-3 py-2 border rounded"
          required
        >
          <option value="Income">Income</option>
          <option value="Expense">Expense</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Date</label>
        <input
          type="date"
          value={formData.date}
          onChange={(e) => setFormData({ ...formData, date: e.target.value })}
          className="w-full px-3 py-2 border rounded"
          required
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 disabled:opacity-50"
      >
        {loading ? 'Creating...' : 'Create Transaction'}
      </button>
    </form>
  );
}
```

#### 3. Transaction List Component

```typescript
// components/TransactionList.tsx
import { useEffect, useState } from 'react';
import { transactionV2API, Transaction } from '@/lib/api/v2/transactions';

export default function TransactionList({ assetId }: { assetId?: number }) {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);

  useEffect(() => {
    loadTransactions();
  }, [page, assetId]);

  const loadTransactions = async () => {
    setLoading(true);
    try {
      const response = await transactionV2API.getTransactions({
        page,
        limit: 20,
        asset_id: assetId,
      });
      setTransactions(response.data);
    } catch (err) {
      console.error('Failed to load transactions:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this transaction?')) return;

    try {
      await transactionV2API.deleteTransaction(id);
      loadTransactions();
    } catch (err) {
      console.error('Failed to delete transaction:', err);
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-2">
      {transactions.map((tx) => (
        <div key={tx.id} className="border p-4 rounded flex justify-between items-center">
          <div>
            <div className="font-semibold">{tx.description}</div>
            <div className="text-sm text-gray-500">
              {tx.category_name} • {tx.asset_name}
            </div>
            <div className="text-sm">
              Balance: ${tx.asset_balance?.toFixed(2)}
            </div>
          </div>
          <div className="text-right">
            <div className={`font-bold ${tx.transaction_type === 1 ? 'text-green-600' : 'text-red-600'}`}>
              {tx.transaction_type === 1 ? '+' : '-'}${tx.amount}
            </div>
            <div className="text-sm text-gray-500">{new Date(tx.date).toLocaleDateString()}</div>
            <button
              onClick={() => handleDelete(tx.id)}
              className="text-red-500 hover:text-red-700 text-sm"
            >
              Delete
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
```

---

## Mobile Integration Guide

### React Native Example

#### 1. API Client Setup

```typescript
// api/TransactionV2API.ts
import AsyncStorage from '@react-native-async-storage/async-storage';

const API_BASE = 'https://api.example.com/api/v2';

interface Transaction {
  id: number;
  description: string;
  amount: number;
  transaction_type: number;
  date: string;
  category_name: string;
  bank_name?: string;
  asset_id: number;
  asset_name?: string;
  asset_balance?: number;
}

class TransactionV2API {
  private async getAuthHeaders() {
    const token = await AsyncStorage.getItem('token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    };
  }

  async getTransactions(params?: {
    page?: number;
    limit?: number;
    asset_id?: number;
  }) {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, String(value));
        }
      });
    }

    const response = await fetch(`${API_BASE}/transactions?${queryParams}`, {
      headers: await this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch transactions');
    }

    return response.json();
  }

  async createTransaction(data: {
    description: string;
    category_id: number;
    asset_id: number;
    amount: number;
    transaction_type: 'Income' | 'Expense';
    date: string;
  }) {
    const response = await fetch(`${API_BASE}/transactions`, {
      method: 'POST',
      headers: await this.getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to create transaction');
    }

    return response.json();
  }

  async deleteTransaction(id: number) {
    const response = await fetch(`${API_BASE}/transactions/${id}`, {
      method: 'DELETE',
      headers: await this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to delete transaction');
    }

    return response.json();
  }
}

export default new TransactionV2API();
```

#### 2. Transaction List Screen

```typescript
// screens/TransactionListScreen.tsx
import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
} from 'react-native';
import TransactionV2API from '../api/TransactionV2API';

export default function TransactionListScreen({ route }: any) {
  const { assetId } = route.params || {};
  const [transactions, setTransactions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadTransactions();
  }, [assetId]);

  const loadTransactions = async () => {
    try {
      const response = await TransactionV2API.getTransactions({
        page: 1,
        limit: 50,
        asset_id: assetId,
      });
      setTransactions(response.data);
    } catch (error: any) {
      Alert.alert('Error', error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (id: number) => {
    Alert.alert(
      'Delete Transaction',
      'Are you sure you want to delete this transaction?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              await TransactionV2API.deleteTransaction(id);
              loadTransactions();
            } catch (error: any) {
              Alert.alert('Error', error.message);
            }
          },
        },
      ]
    );
  };

  const renderTransaction = ({ item }: any) => (
    <View className="p-4 border-b border-gray-200 flex-row justify-between items-center">
      <View className="flex-1">
        <Text className="font-semibold text-base">{item.description}</Text>
        <Text className="text-gray-500 text-sm">{item.category_name}</Text>
        <Text className="text-gray-500 text-sm">{item.asset_name}</Text>
        <Text className="text-gray-400 text-xs">
          Balance: ${item.asset_balance?.toFixed(2)}
        </Text>
      </View>
      <View className="items-end">
        <Text
          className={`font-bold text-lg ${
            item.transaction_type === 1 ? 'text-green-600' : 'text-red-600'
          }`}
        >
          {item.transaction_type === 1 ? '+' : '-'}${item.amount}
        </Text>
        <Text className="text-gray-500 text-sm">
          {new Date(item.date).toLocaleDateString()}
        </Text>
        <TouchableOpacity onPress={() => handleDelete(item.id)}>
          <Text className="text-red-500 mt-1">Delete</Text>
        </TouchableOpacity>
      </View>
    </View>
  );

  if (loading) {
    return (
      <View className="flex-1 justify-center items-center">
        <ActivityIndicator size="large" color="#3B82F6" />
      </View>
    );
  }

  return (
    <View className="flex-1 bg-white">
      <FlatList
        data={transactions}
        renderItem={renderTransaction}
        keyExtractor={(item) => item.id.toString()}
      />
    </View>
  );
}
```

---

## Error Handling

### Common Error Codes

| Status Code | Message | Solution |
|-------------|---------|----------|
| 400 | Invalid request payload | Check request body format |
| 400 | Insufficient balance in the selected asset | User has insufficient funds for expense |
| 401 | User not authenticated | User needs to log in again |
| 403 | Asset does not belong to you | User doesn't have permission for this asset |
| 404 | Transaction not found | Transaction ID doesn't exist |
| 500 | Failed to create transaction | Server error, try again later |

### Error Response Format

```json
{
  "success": false,
  "message": "Insufficient balance in the selected asset"
}
```

---

## Migration from V1 to V2

### Frontend Migration Steps

1. **Update API endpoint URLs**
   - Change `/api/transactions` to `/api/v2/transactions`
   - Change `BankID` to `AssetID` in requests

2. **Update form fields**
   - Replace bank selector with asset selector
   - Fetch assets from `/api/wallets`

3. **Handle new responses**
   - Parse `asset_id`, `asset_name`, `asset_balance` from responses
   - Display current asset balance in UI

4. **Update error handling**
   - Handle "Insufficient balance" error explicitly
   - Show user-friendly messages for asset-related errors

### V1 to V2 Request Mapping

| V1 Field | V2 Field | Notes |
|----------|----------|-------|
| BankID | AssetID | Use asset_id instead |
| Amount | Amount | Same |
| TransactionType | TransactionType | Same values (1=Income, 2=Expense) |
| Description | Description | Same |
| CategoryID | CategoryID | Same |
| Date | Date | Same format |

---

## Testing Checklist

### Manual Testing

- [ ] Create income transaction - verify balance increases
- [ ] Create expense transaction - verify balance decreases
- [ ] Create expense exceeding balance - verify error
- [ ] Delete transaction - verify balance rolls back
- [ ] Update transaction - verify balance recalculates correctly
- [ ] Filter transactions by asset
- [ ] Get asset-specific transactions
- [ ] Verify concurrent transactions don't cause balance issues

### Automated Testing

```typescript
// Example test suite
describe('Transaction V2 API', () => {
  it('should create income transaction and update balance', async () => {
    const initialBalance = await getAssetBalance(1);
    const amount = 100;
    
    await transactionV2API.createTransaction({
      description: 'Test Income',
      category_id: 1,
      asset_id: 1,
      amount,
      transaction_type: 'Income',
      date: '2025-01-15',
    });
    
    const newBalance = await getAssetBalance(1);
    expect(newBalance).toBe(initialBalance + amount);
  });

  it('should prevent expense exceeding balance', async () => {
    await expect(
      transactionV2API.createTransaction({
        description: 'Test Expense',
        category_id: 1,
        asset_id: 1,
        amount: 999999,
        transaction_type: 'Expense',
        date: '2025-01-15',
      })
    ).rejects.toThrow('Insufficient balance');
  });

  it('should rollback balance on delete', async () => {
    const tx = await createExpenseTransaction(1, 50);
    const balanceAfterCreate = await getAssetBalance(1);
    
    await transactionV2API.deleteTransaction(tx.id);
    const balanceAfterDelete = await getAssetBalance(1);
    
    expect(balanceAfterDelete).toBe(balanceAfterCreate + 50);
  });
});
```

---

## Best Practices

### For Web Applications

1. **Refresh balances after transactions**
   - Reload asset list or use optimistic UI updates
   - Show current balance on transaction forms

2. **Validate before API calls**
   - Check if amount is positive
   - Validate date format
   - Ensure asset is selected

3. **Handle loading states**
   - Show loading indicators during API calls
   - Disable submit buttons to prevent duplicate submissions

4. **Display helpful errors**
   - Translate API errors to user-friendly messages
   - Show balance when insufficient funds error occurs

### For Mobile Applications

1. **Offline support**
   - Cache transactions locally
   - Sync when connection is restored

2. **Pull-to-refresh**
   - Allow users to refresh transaction lists
   - Refresh balances after pull-to-refresh

3. **Background sync**
   - Sync transactions in the background
   - Use push notifications for balance updates

4. **Optimistic updates**
   - Update UI immediately
   - Rollback on API error

---

## Support & Troubleshooting

### Common Issues

**Q: Why is the balance not updating?**
A: Ensure you're using V2 endpoints (`/api/v2/transactions`). V1 endpoints do not update balances automatically.

**Q: How do I handle concurrent transactions?**
A: The API uses row locking to prevent race conditions. Multiple users can create transactions simultaneously without data corruption.

**Q: Can I create transactions without an asset?**
A: No, V2 API requires an `asset_id`. Use the V1 API if you need BankID-based transactions.

**Q: What happens if I delete an asset with transactions?**
A: The foreign key constraint prevents deleting an asset that has transactions. Delete or reassign transactions first.

---

## Changelog

### Version 2.0.0
- ✅ Added AssetID support
- ✅ Automatic balance synchronization
- ✅ Asset-specific transaction filtering
- ✅ Insufficient balance validation
- ✅ Transaction rollback on delete

---

## Contact & Feedback

For issues, questions, or feature requests:
- GitHub Issues: [repository-url]
- Email: support@example.com
- Documentation: https://docs.example.com
