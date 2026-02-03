# Frontend Migration Guide: Spending by Bank → Spending by Asset

## 📋 Overview

This guide provides a step-by-step plan to migrate from the legacy "Spending by Bank" feature to the new "Spending by Asset/Wallet" feature in your frontend application.

---

## 🎯 Migration Goals

| Goal | Benefit |
|------|---------|
| Replace bank-based analytics with asset-based | Align with wallet system |
| Improve data accuracy | Show actual wallet usage |
| Better user experience | More relevant financial insights |
| Support multi-currency | Assets have currency info |

---

## 📊 API Comparison

### Old Endpoint (Deprecated)

```
GET /api/analytics/spending-by-bank
```

**Query Parameters:**
- `start_date` (required): YYYY-MM-DD
- `end_date` (required): YYYY-MM-DD
- `asset_id` (optional): Filter by specific asset

**Response Structure:**
```json
{
  "success": true,
  "message": "Spending by bank retrieved successfully",
  "data": [
    {
      "bank_id": 1,
      "bank_name": "Chase Bank",
      "total_amount": 3200000,
      "percentage": 45.5,
      "count": 150
    }
  ]
}
```

**Limitations:**
- ❌ Only shows expense total (no income)
- ❌ No net amount calculation
- ❌ No currency information
- ❌ Bank is now legacy (replaced by assets)

### New Endpoint (Recommended)

```
GET /api/analytics/spending-by-asset
```

**Query Parameters:**
- `start_date` (required): YYYY-MM-DD
- `end_date` (required): YYYY-MM-DD

**Response Structure:**
```json
{
  "success": true,
  "message": "Spending by asset retrieved successfully",
  "data": [
    {
      "asset_id": 1,
      "asset_name": "Main Wallet",
      "asset_type": "cash",
      "asset_currency": "IDR",
      "total_income": 5000000,
      "total_expense": 3200000,
      "net_amount": 1800000,
      "percentage": 45.5,
      "transaction_count": 150
    },
    {
      "asset_id": 2,
      "asset_name": "Savings Account",
      "asset_type": "bank",
      "asset_currency": "USD",
      "total_income": 2000,
      "total_expense": 500,
      "net_amount": 1500,
      "percentage": 25.2,
      "transaction_count": 75
    }
  ]
}
```

**Advantages:**
- ✅ Shows both income AND expense
- ✅ Calculates net amount automatically
- ✅ Includes currency for proper formatting
- ✅ Shows asset type (cash, bank, e-wallet, etc.)
- ✅ More transaction details

---

## 🔄 Migration Strategy

### Option 1: Immediate Replacement (Recommended)

**Timeline:** 1-2 days

Replace all bank-based analytics with asset-based in one update.

**Pros:**
- Clean codebase
- No confusion
- Immediate benefits

**Cons:**
- Requires testing all analytics pages

### Option 2: Gradual Migration

**Timeline:** 1 week

Keep both options temporarily, gradually phase out bank analytics.

**Pros:**
- Lower risk
- Can A/B test

**Cons:**
- Temporary code duplication
- More maintenance

---

## 📝 Implementation Checklist

### Phase 1: Type Definitions & API Setup

- [ ] **1.1** Update TypeScript types
- [ ] **1.2** Create new API service methods
- [ ] **1.3** Add currency formatting utilities

### Phase 2: UI Components

- [ ] **2.1** Update spending chart component
- [ ] **2.2** Update spending list/table component
- [ ] **2.3** Add currency badge component
- [ ] **2.4** Update color scheme (wallet types vs banks)

### Phase 3: Replace API Calls

- [ ] **3.1** Analytics Dashboard
- [ ] **3.2** Reports Page
- [ ] **3.3** Transaction Statistics
- [ ] **3.4** Mobile App (if applicable)

### Phase 4: Testing & Cleanup

- [ ] **4.1** Test with different date ranges
- [ ] **4.2** Test with multiple currencies
- [ ] **4.3** Remove old bank analytics code
- [ ] **4.4** Update documentation

---

## 💻 Code Implementation

### Step 1: Update Type Definitions

```typescript
// types/analytics.ts

// ❌ OLD - Remove this
export interface SpendingByBankResponse {
  bank_id: number;
  bank_name: string;
  total_amount: number;
  percentage: number;
  count: number;
}

// ✅ NEW - Add this
export interface SpendingByAssetResponse {
  asset_id: number;
  asset_name: string;
  asset_type: string;
  asset_currency: string;
  total_income: number;
  total_expense: number;
  net_amount: number;
  percentage: number;
  transaction_count: number;
}

// Asset type for better type safety
export type AssetType = 'cash' | 'bank' | 'e-wallet' | 'investment' | 'credit-card';

// Helper type for color mapping
export const ASSET_TYPE_COLORS: Record<AssetType, string> = {
  'cash': '#4CAF50',
  'bank': '#2196F3',
  'e-wallet': '#FF9800',
  'investment': '#9C27B0',
  'credit-card': '#F44336'
};

export const ASSET_TYPE_ICONS: Record<AssetType, string> = {
  'cash': '💵',
  'bank': '🏦',
  'e-wallet': '📱',
  'investment': '📈',
  'credit-card': '💳'
};
```

### Step 2: Create API Service

```typescript
// services/analyticsApi.ts

import { api } from './api';
import { SpendingByAssetResponse } from '../types/analytics';

// ❌ OLD - Can be removed after migration
export const getSpendingByBank = async (
  startDate: string,
  endDate: string
): Promise<SpendingByBankResponse[]> => {
  const response = await api.get('/analytics/spending-by-bank', {
    params: { start_date: startDate, end_date: endDate }
  });
  return response.data.data;
};

// ✅ NEW - Add this
export const getSpendingByAsset = async (
  startDate: string,
  endDate: string
): Promise<SpendingByAssetResponse[]> => {
  const response = await api.get('/analytics/spending-by-asset', {
    params: { 
      start_date: startDate, 
      end_date: endDate 
    }
  });
  return response.data.data;
};
```

### Step 3: Update React Component

#### Option A: React with Hooks

```tsx
// components/Analytics/SpendingChart.tsx

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { getSpendingByAsset } from '../../services/analyticsApi';
import { ASSET_TYPE_COLORS, ASSET_TYPE_ICONS } from '../../types/analytics';

interface SpendingChartProps {
  startDate: string;
  endDate: string;
}

export const SpendingChart: React.FC<SpendingChartProps> = ({ 
  startDate, 
  endDate 
}) => {
  // ❌ OLD - Replace this
  // const { data, isLoading } = useQuery({
  //   queryKey: ['spendingByBank', startDate, endDate],
  //   queryFn: () => getSpendingByBank(startDate, endDate)
  // });

  // ✅ NEW - Use this
  const { data: assets, isLoading, error } = useQuery({
    queryKey: ['spendingByAsset', startDate, endDate],
    queryFn: () => getSpendingByAsset(startDate, endDate)
  });

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;
  if (!assets || assets.length === 0) {
    return <EmptyState message="No transactions found for this period" />;
  }

  return (
    <div className="spending-chart">
      <h2>💰 Spending by Wallet</h2>
      
      {/* Chart visualization */}
      <div className="chart-container">
        <PieChart data={assets.map(asset => ({
          name: asset.asset_name,
          value: asset.total_expense,
          color: ASSET_TYPE_COLORS[asset.asset_type as AssetType]
        }))} />
      </div>

      {/* List view */}
      <div className="asset-list">
        {assets.map(asset => (
          <div key={asset.asset_id} className="asset-card">
            <div className="asset-header">
              <span className="asset-icon">
                {ASSET_TYPE_ICONS[asset.asset_type as AssetType]}
              </span>
              <h3>{asset.asset_name}</h3>
              <span className="asset-type-badge">{asset.asset_type}</span>
            </div>
            
            <div className="asset-stats">
              <div className="stat income">
                <label>Income</label>
                <span className="amount positive">
                  +{formatCurrency(asset.total_income, asset.asset_currency)}
                </span>
              </div>
              
              <div className="stat expense">
                <label>Expense</label>
                <span className="amount negative">
                  -{formatCurrency(asset.total_expense, asset.asset_currency)}
                </span>
              </div>
              
              <div className="stat net">
                <label>Net</label>
                <span className={`amount ${asset.net_amount >= 0 ? 'positive' : 'negative'}`}>
                  {formatCurrency(asset.net_amount, asset.asset_currency)}
                </span>
              </div>
            </div>

            <div className="asset-footer">
              <span className="percentage">{asset.percentage.toFixed(1)}% of total</span>
              <span className="transaction-count">
                {asset.transaction_count} transactions
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Helper function for currency formatting
const formatCurrency = (amount: number, currency: string): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount);
};
```

#### Option B: Vue 3 Composition API

```vue
<!-- components/Analytics/SpendingChart.vue -->

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuery } from '@tanstack/vue-query';
import { getSpendingByAsset } from '@/services/analyticsApi';
import { ASSET_TYPE_COLORS, ASSET_TYPE_ICONS } from '@/types/analytics';

interface Props {
  startDate: string;
  endDate: string;
}

const props = defineProps<Props>();

// ❌ OLD - Remove this
// const { data: banks, isLoading } = useQuery({
//   queryKey: ['spendingByBank', props.startDate, props.endDate],
//   queryFn: () => getSpendingByBank(props.startDate, props.endDate)
// });

// ✅ NEW - Use this
const { data: assets, isLoading, error } = useQuery({
  queryKey: computed(() => ['spendingByAsset', props.startDate, props.endDate]),
  queryFn: () => getSpendingByAsset(props.startDate, props.endDate)
});

const chartData = computed(() => {
  if (!assets.value) return [];
  return assets.value.map(asset => ({
    name: asset.asset_name,
    value: asset.total_expense,
    color: ASSET_TYPE_COLORS[asset.asset_type]
  }));
});

const formatCurrency = (amount: number, currency: string) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 0
  }).format(amount);
};
</script>

<template>
  <div class="spending-chart">
    <h2>💰 Spending by Wallet</h2>
    
    <div v-if="isLoading">Loading...</div>
    <div v-else-if="error">Error loading data</div>
    <div v-else-if="!assets || assets.length === 0">
      No transactions found
    </div>
    
    <div v-else>
      <!-- Chart -->
      <PieChart :data="chartData" />
      
      <!-- Asset cards -->
      <div class="asset-list">
        <div 
          v-for="asset in assets" 
          :key="asset.asset_id" 
          class="asset-card"
        >
          <div class="asset-header">
            <span class="asset-icon">{{ ASSET_TYPE_ICONS[asset.asset_type] }}</span>
            <h3>{{ asset.asset_name }}</h3>
            <span class="asset-type-badge">{{ asset.asset_type }}</span>
          </div>
          
          <div class="asset-stats">
            <div class="stat income">
              <label>Income</label>
              <span class="amount positive">
                +{{ formatCurrency(asset.total_income, asset.asset_currency) }}
              </span>
            </div>
            
            <div class="stat expense">
              <label>Expense</label>
              <span class="amount negative">
                -{{ formatCurrency(asset.total_expense, asset.asset_currency) }}
              </span>
            </div>
            
            <div class="stat net">
              <label>Net</label>
              <span 
                :class="['amount', asset.net_amount >= 0 ? 'positive' : 'negative']"
              >
                {{ formatCurrency(asset.net_amount, asset.asset_currency) }}
              </span>
            </div>
          </div>

          <div class="asset-footer">
            <span class="percentage">{{ asset.percentage.toFixed(1) }}% of total</span>
            <span class="transaction-count">
              {{ asset.transaction_count }} transactions
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
```

#### Option C: React Native (Mobile)

```tsx
// screens/Analytics/SpendingByAssetScreen.tsx

import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  ActivityIndicator
} from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { getSpendingByAsset } from '../../services/analyticsApi';
import { ASSET_TYPE_COLORS, ASSET_TYPE_ICONS } from '../../types/analytics';

interface Props {
  startDate: string;
  endDate: string;
}

export const SpendingByAssetScreen: React.FC<Props> = ({ startDate, endDate }) => {
  const { data: assets, isLoading } = useQuery({
    queryKey: ['spendingByAsset', startDate, endDate],
    queryFn: () => getSpendingByAsset(startDate, endDate)
  });

  if (isLoading) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" color="#2196F3" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>💰 Spending by Wallet</Text>
      
      <FlatList
        data={assets}
        keyExtractor={(item) => item.asset_id.toString()}
        renderItem={({ item }) => (
          <View style={styles.assetCard}>
            <View style={styles.header}>
              <Text style={styles.icon}>
                {ASSET_TYPE_ICONS[item.asset_type]}
              </Text>
              <View style={styles.headerText}>
                <Text style={styles.assetName}>{item.asset_name}</Text>
                <Text style={styles.assetType}>{item.asset_type}</Text>
              </View>
            </View>

            <View style={styles.statsContainer}>
              <View style={styles.stat}>
                <Text style={styles.statLabel}>Income</Text>
                <Text style={[styles.statValue, styles.positive]}>
                  +{formatCurrency(item.total_income, item.asset_currency)}
                </Text>
              </View>

              <View style={styles.stat}>
                <Text style={styles.statLabel}>Expense</Text>
                <Text style={[styles.statValue, styles.negative]}>
                  -{formatCurrency(item.total_expense, item.asset_currency)}
                </Text>
              </View>

              <View style={styles.stat}>
                <Text style={styles.statLabel}>Net</Text>
                <Text 
                  style={[
                    styles.statValue, 
                    item.net_amount >= 0 ? styles.positive : styles.negative
                  ]}
                >
                  {formatCurrency(item.net_amount, item.asset_currency)}
                </Text>
              </View>
            </View>

            <View style={styles.footer}>
              <Text style={styles.percentage}>
                {item.percentage.toFixed(1)}% of total
              </Text>
              <Text style={styles.count}>
                {item.transaction_count} transactions
              </Text>
            </View>
          </View>
        )}
        contentContainerStyle={styles.listContent}
      />
    </View>
  );
};

const formatCurrency = (amount: number, currency: string): string => {
  // Mobile-friendly formatting
  if (amount >= 1000000) {
    return `${(amount / 1000000).toFixed(1)}M ${currency}`;
  } else if (amount >= 1000) {
    return `${(amount / 1000).toFixed(1)}K ${currency}`;
  }
  return `${amount} ${currency}`;
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
    padding: 16
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center'
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 16,
    color: '#333'
  },
  listContent: {
    paddingBottom: 20
  },
  assetCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16
  },
  icon: {
    fontSize: 32,
    marginRight: 12
  },
  headerText: {
    flex: 1
  },
  assetName: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4
  },
  assetType: {
    fontSize: 14,
    color: '#666',
    textTransform: 'capitalize'
  },
  statsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 12
  },
  stat: {
    flex: 1,
    alignItems: 'center'
  },
  statLabel: {
    fontSize: 12,
    color: '#999',
    marginBottom: 4
  },
  statValue: {
    fontSize: 16,
    fontWeight: '600'
  },
  positive: {
    color: '#4CAF50'
  },
  negative: {
    color: '#F44336'
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: '#E0E0E0'
  },
  percentage: {
    fontSize: 14,
    color: '#666'
  },
  count: {
    fontSize: 14,
    color: '#666'
  }
});
```

### Step 4: CSS Styling

```css
/* styles/analytics.css */

/* Asset Card Styles */
.spending-chart {
  padding: 24px;
  background: #f5f5f5;
  border-radius: 8px;
}

.asset-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  margin-top: 24px;
}

.asset-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s, box-shadow 0.2s;
}

.asset-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.asset-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.asset-icon {
  font-size: 32px;
}

.asset-header h3 {
  flex: 1;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.asset-type-badge {
  padding: 4px 12px;
  background: #e3f2fd;
  color: #1976d2;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: capitalize;
}

.asset-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.stat {
  text-align: center;
}

.stat label {
  display: block;
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat .amount {
  font-size: 20px;
  font-weight: 600;
}

.amount.positive {
  color: #4caf50;
}

.amount.negative {
  color: #f44336;
}

.asset-footer {
  display: flex;
  justify-content: space-between;
  padding-top: 16px;
  border-top: 1px solid #e0e0e0;
  font-size: 14px;
  color: #666;
}

.percentage {
  font-weight: 500;
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .asset-list {
    grid-template-columns: 1fr;
  }

  .asset-stats {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .stat {
    display: flex;
    justify-content: space-between;
    text-align: left;
  }

  .stat label {
    margin-bottom: 0;
  }
}

/* Loading & Empty States */
.loading-spinner {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-state-icon {
  font-size: 64px;
  margin-bottom: 16px;
}
```

---

## 🎨 UI/UX Improvements

### Before (Bank-based)
```
┌─────────────────────────────────────┐
│ Spending by Bank                    │
├─────────────────────────────────────┤
│ Chase Bank                          │
│ Total: $3,200 (45.5%)              │
│ 150 transactions                    │
└─────────────────────────────────────┘
```

### After (Asset-based)
```
┌─────────────────────────────────────┐
│ 💰 Spending by Wallet               │
├─────────────────────────────────────┤
│ 💵 Main Wallet          [cash]      │
│ ─────────────────────────────────   │
│ Income:   +5,000,000 IDR           │
│ Expense:  -3,200,000 IDR           │
│ Net:      +1,800,000 IDR           │
│ ─────────────────────────────────   │
│ 45.5% of total • 150 transactions   │
└─────────────────────────────────────┘
```

### Key Improvements
1. **More Information** - Shows income, expense, and net
2. **Visual Clarity** - Icons for asset types
3. **Currency Support** - Proper formatting per currency
4. **Better Context** - Asset type badges
5. **Net Amount** - Quick overview of money flow

---

## 🧪 Testing Plan

### Manual Testing Checklist

- [ ] **Date Range Testing**
  - [ ] Last 7 days
  - [ ] Last 30 days
  - [ ] Last 3 months
  - [ ] Custom date range

- [ ] **Multi-Currency Testing**
  - [ ] Assets with IDR
  - [ ] Assets with USD
  - [ ] Mixed currency display

- [ ] **Edge Cases**
  - [ ] No transactions in period
  - [ ] Single asset
  - [ ] Many assets (10+)
  - [ ] Zero balance assets

- [ ] **Responsive Design**
  - [ ] Desktop (1920x1080)
  - [ ] Tablet (768x1024)
  - [ ] Mobile (375x667)

### Automated Testing (Jest Example)

```typescript
// __tests__/SpendingChart.test.tsx

import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SpendingChart } from '../components/Analytics/SpendingChart';
import { getSpendingByAsset } from '../services/analyticsApi';

jest.mock('../services/analyticsApi');

const mockData = [
  {
    asset_id: 1,
    asset_name: 'Main Wallet',
    asset_type: 'cash',
    asset_currency: 'IDR',
    total_income: 5000000,
    total_expense: 3200000,
    net_amount: 1800000,
    percentage: 45.5,
    transaction_count: 150
  }
];

describe('SpendingChart', () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } }
  });

  beforeEach(() => {
    (getSpendingByAsset as jest.Mock).mockResolvedValue(mockData);
  });

  it('should display asset spending data', async () => {
    render(
      <QueryClientProvider client={queryClient}>
        <SpendingChart startDate="2026-01-01" endDate="2026-02-28" />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Main Wallet')).toBeInTheDocument();
      expect(screen.getByText('cash')).toBeInTheDocument();
      expect(screen.getByText(/5,000,000/)).toBeInTheDocument();
    });
  });

  it('should format currency correctly', async () => {
    render(
      <QueryClientProvider client={queryClient}>
        <SpendingChart startDate="2026-01-01" endDate="2026-02-28" />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByText(/IDR/)).toBeInTheDocument();
    });
  });
});
```

---

## 🚀 Deployment Steps

### Step 1: Prepare
```bash
# 1. Create feature branch
git checkout -b feature/spending-by-asset-migration

# 2. Update dependencies (if needed)
npm install

# 3. Update types
# Create/update types/analytics.ts
```

### Step 2: Implement
```bash
# 1. Add new API service
# 2. Create/update components
# 3. Add styling
# 4. Write tests
```

### Step 3: Test
```bash
# Run tests
npm test

# Run linting
npm run lint

# Test in development
npm run dev
```

### Step 4: Deploy
```bash
# 1. Commit changes
git add .
git commit -m "feat: migrate from spending-by-bank to spending-by-asset"

# 2. Push to repository
git push origin feature/spending-by-asset-migration

# 3. Create pull request
# 4. Code review
# 5. Merge to main
# 6. Deploy to production
```

---

## 📊 Success Metrics

Track these metrics after deployment:

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Load Time | < 2s | Browser DevTools |
| Error Rate | < 1% | Error monitoring (Sentry) |
| User Engagement | +20% | Analytics (time on page) |
| API Response Time | < 500ms | Backend monitoring |

---

## 🆘 Troubleshooting

### Issue 1: 400 Bad Request Error

**Symptom:** API returns 400 error

**Cause:** Missing required date parameters

**Solution:**
```typescript
// ❌ Wrong
getSpendingByAsset();

// ✅ Correct
getSpendingByAsset('2026-01-01', '2026-02-28');
```

### Issue 2: Currency Not Formatting

**Symptom:** Shows "NaN" or raw numbers

**Cause:** Invalid currency code

**Solution:**
```typescript
const formatCurrency = (amount: number, currency: string) => {
  try {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD', // Fallback
      minimumFractionDigits: 0
    }).format(amount);
  } catch (error) {
    // Fallback for invalid currency
    return `${amount} ${currency}`;
  }
};
```

### Issue 3: No Data Displayed

**Symptom:** Empty list despite having transactions

**Cause:** Transactions might have NULL asset_id

**Solution:** Check backend - ensure all transactions have asset_id populated

---

## 📚 Additional Resources

- [Backend Analytics API Documentation](./ANALYTICS_ASSET_MIGRATION.md)
- [Asset/Wallet API Guide](./frontend-wallet-implementation.md)
- [Transaction V2 Integration](./frontend_integration_guide.md)
- [MDN Intl.NumberFormat](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/NumberFormat)

---

## ✅ Final Checklist

Before marking this migration complete:

- [ ] All API calls updated to use `/analytics/spending-by-asset`
- [ ] Type definitions updated
- [ ] Components display all new data fields (income, expense, net)
- [ ] Currency formatting works for all currencies
- [ ] Asset type icons/badges display correctly
- [ ] Mobile responsive design verified
- [ ] Tests written and passing
- [ ] Old bank analytics code removed
- [ ] Documentation updated
- [ ] Code reviewed and merged
- [ ] Deployed to production
- [ ] User feedback collected

---

**Estimated Timeline:** 2-3 days for full migration

**Effort:** Medium

**Risk Level:** Low (backward compatible API, no data loss)
