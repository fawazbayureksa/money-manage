# Budget Asset Support - Frontend Adjustments Guide

This guide outlines the specific changes needed to add asset support to your existing budget implementation.

---

## Quick Summary of Changes

The backend now supports:
- `asset_id` field in Budget model (nullable)
- Global budgets (all assets) when `asset_id = null`
- Asset-specific budgets when `asset_id` is set
- Spending tracking filtered by asset

---

## 1. API Client Updates

### 1.1 Update Type Definitions

```typescript
// types/budget.ts - UPDATE EXISTING TYPE
export interface Budget {
  id: number;
  user_id: number;
  category_id: number;
  category_name: string;
  asset_id: number | null;        // ADD THIS
  asset_name: string;             // ADD THIS
  amount: number;
  period: 'monthly' | 'yearly';
  start_date: string;
  end_date: string;
  is_active: boolean;
  alert_at: number;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateBudgetRequest {
  category_id: number;
  asset_id?: number | null;       // ADD THIS (optional)
  amount: number;
  period: 'monthly' | 'yearly';
  start_date: string;
  alert_at?: number;
  description?: string;
}

export interface UpdateBudgetRequest {
  amount?: number;
  alert_at?: number;
  description?: string;
  is_active?: boolean;
  asset_id?: number | null;       // ADD THIS (optional)
}
```

### 1.2 Update API Service Methods

```typescript
// services/api.ts - UPDATE EXISTING FUNCTIONS

// EXISTING: Create Budget - ADD asset_id to request body
export const createBudget = async (data: CreateBudgetRequest): Promise<Budget> => {
  const response = await api.post('/budgets', {
    category_id: data.category_id,
    asset_id: data.asset_id,      // ADD THIS LINE
    amount: data.amount,
    period: data.period,
    start_date: data.start_date,
    alert_at: data.alert_at,
    description: data.description,
  });
  return response.data;
};

// EXISTING: Get Budgets - ADD asset_id to query params
export const getBudgets = async (params?: {
  category_id?: number;
  asset_id?: number | null;       // ADD THIS
  period?: string;
  is_active?: boolean;
  status?: string;
  page?: number;
  page_size?: number;
}): Promise<PaginatedResponse<Budget>> => {
  const response = await api.get('/budgets', { params });
  return response.data;
};

// EXISTING: Update Budget - ADD asset_id to request body
export const updateBudget = async (id: number, data: UpdateBudgetRequest): Promise<Budget> => {
  const response = await api.put(`/budgets/${id}`, {
    amount: data.amount,
    alert_at: data.alert_at,
    description: data.description,
    is_active: data.is_active,
    asset_id: data.asset_id,      // ADD THIS LINE
  });
  return response.data;
};
```

---

## 2. Form Component Adjustments

### 2.1 Add Asset Selection to Create/Edit Form

```tsx
// components/budget/BudgetForm.tsx - EXISTING COMPONENT

// ADD to imports
import { useAssets } from '@/hooks/useAssets'; // or your existing asset hook

const BudgetForm = ({ initialData, onSubmit, onCancel }) => {
  // EXISTING STATE
  const [categoryId, setCategoryId] = useState(initialData?.category_id || '');
  const [amount, setAmount] = useState(initialData?.amount || '');
  const [period, setPeriod] = useState(initialData?.period || 'monthly');
  const [startDate, setStartDate] = useState(initialData?.start_date || '');

  // ADD THIS NEW STATE
  const [selectedAssetId, setSelectedAssetId] = useState<number | null>(
    initialData?.asset_id || null
  );

  // ADD THIS HOOK
  const { assets, loading: assetsLoading } = useAssets(); // your existing asset hook

  // UPDATE SUBMIT HANDLER
  const handleSubmit = (e) => {
    e.preventDefault();

    const data = {
      category_id: categoryId,
      asset_id: selectedAssetId,  // ADD THIS LINE
      amount: parseInt(amount),
      period,
      start_date: startDate,
      alert_at: 80,
      description: '',
    };

    onSubmit(data);
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* EXISTING FIELDS */}
      <div>
        <label>Category</label>
        <select value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
          {/* existing options */}
        </select>
      </div>

      {/* ADD THIS NEW SECTION - Asset Selection */}
      <div>
        <label>Asset Scope</label>
        <select
          value={selectedAssetId || ''}
          onChange={(e) => setSelectedAssetId(e.target.value ? parseInt(e.target.value) : null)}
        >
          <option value="">All Assets (Global)</option>
          {!assetsLoading && assets?.map((asset) => (
            <option key={asset.id} value={asset.id}>
              {asset.name}
            </option>
          ))}
        </select>
        <small>Leave empty to track spending across all assets</small>
      </div>

      {/* EXISTING FIELDS - Amount, Period, Date */}
      <div>
        <label>Amount</label>
        <input
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
        />
      </div>

      {/* ... rest of existing form ... */}

      <button type="submit">
        {initialData ? 'Update' : 'Create'} Budget
      </button>
      <button type="button" onClick={onCancel}>
        Cancel
      </button>
    </form>
  );
};
```

### 2.2 If Using a Reusable AssetPicker Component

```tsx
// components/common/AssetPicker.tsx - CREATE OR USE EXISTING
interface AssetPickerProps {
  assets: Asset[];
  selectedId: number | null;
  onChange: (id: number | null) => void;
  includeAll?: boolean;
  label?: string;
}

export const AssetPicker = ({
  assets,
  selectedId,
  onChange,
  includeAll = true,
  label = 'Asset'
}: AssetPickerProps) => {
  return (
    <div>
      <label>{label}</label>
      <select
        value={selectedId || ''}
        onChange={(e) => onChange(e.target.value ? parseInt(e.target.value) : null)}
      >
        {includeAll && (
          <option value="">All Assets (Global)</option>
        )}
        {assets.map((asset) => (
          <option key={asset.id} value={asset.id}>
            {asset.name} ({formatCurrency(asset.balance)})
          </option>
        ))}
      </select>
    </div>
  );
};

// Then in BudgetForm:
import { AssetPicker } from '@/components/common/AssetPicker';

<AssetPicker
  assets={assets}
  selectedId={selectedAssetId}
  onChange={setSelectedAssetId}
  includeAll={true}
  label="Asset Scope"
/>
```

---

## 3. Budget List Component Adjustments

### 3.1 Add Asset Filter to Budget List

```tsx
// components/budget/BudgetList.tsx - EXISTING COMPONENT

const BudgetList = () => {
  // EXISTING STATE
  const [budgets, setBudgets] = useState([]);
  const [loading, setLoading] = useState(false);

  // ADD THIS NEW STATE
  const [filterAssetId, setFilterAssetId] = useState<number | null>(null);

  // ADD THIS HOOK
  const { assets } = useAssets();

  // UPDATE FETCH FUNCTION
  const fetchBudgets = async () => {
    setLoading(true);
    try {
      // ADD asset_id to params
      const params: any = {};
      if (filterAssetId !== null) {
        params.asset_id = filterAssetId;
      }

      const data = await getBudgets(params);
      setBudgets(data.data);
    } catch (error) {
      console.error('Failed to fetch budgets', error);
    } finally {
      setLoading(false);
    }
  };

  // FETCH WHEN FILTER CHANGES
  useEffect(() => {
    fetchBudgets();
  }, [filterAssetId]);

  return (
    <div>
      {/* ADD THIS FILTER SECTION */}
      <div className="budget-filters">
        <label>Filter by Asset:</label>
        <select
          value={filterAssetId || ''}
          onChange={(e) => setFilterAssetId(e.target.value ? parseInt(e.target.value) : null)}
        >
          <option value="">All Assets</option>
          {assets?.map((asset) => (
            <option key={asset.id} value={asset.id}>
              {asset.name}
            </option>
          ))}
        </select>
      </div>

      {/* EXISTING BUDGET CARDS */}
      {loading ? (
        <div>Loading...</div>
      ) : (
        <div className="budget-grid">
          {budgets.map((budget) => (
            <BudgetCard key={budget.id} budget={budget} />
          ))}
        </div>
      )}
    </div>
  );
};
```

### 3.2 Update BudgetCard to Show Asset

```tsx
// components/budget/BudgetCard.tsx - EXISTING COMPONENT

const BudgetCard = ({ budget }) => {
  return (
    <div className="budget-card">
      {/* EXISTING: Category Name */}
      <h3>{budget.category_name}</h3>

      {/* ADD THIS: Show Asset if exists */}
      {budget.asset_id && (
        <div className="budget-asset">
          <span>📊 {budget.asset_name}</span>
        </div>
      )}

      {/* ADD THIS: Show if global budget */}
      {!budget.asset_id && (
        <div className="budget-asset-global">
          <span>🌐 All Assets</span>
        </div>
      )}

      {/* EXISTING: Progress Bar, Amount, etc. */}
      <div className="budget-progress">
        <ProgressBar value={budget.percentage_used} />
      </div>

      <div className="budget-amount">
        {formatCurrency(budget.spent_amount)} / {formatCurrency(budget.amount)}
      </div>

      {/* EXISTING: Actions */}
      <button onClick={() => onEdit(budget)}>Edit</button>
      <button onClick={() => onDelete(budget.id)}>Delete</button>
    </div>
  );
};
```

---

## 4. React Native Adjustments

### 4.1 Update BudgetFormScreen

```tsx
// screens/BudgetFormScreen.tsx - EXISTING SCREEN

const BudgetFormScreen = ({ route, navigation }) => {
  // EXISTING STATE
  const [categoryId, setCategoryId] = useState('');
  const [amount, setAmount] = useState('');
  const [period, setPeriod] = useState('monthly');

  // ADD THIS NEW STATE
  const [selectedAssetId, setSelectedAssetId] = useState<number | null>(null);

  // ADD THIS HOOK
  const { assets } = useAssets(); // your existing asset hook

  // UPDATE SUBMIT HANDLER
  const handleSubmit = async () => {
    try {
      const data = {
        category_id: categoryId,
        asset_id: selectedAssetId,  // ADD THIS LINE
        amount: parseInt(amount),
        period,
        start_date: startDate,
      };

      await createBudget(data);
      navigation.goBack();
    } catch (error) {
      Alert.alert('Error', 'Failed to create budget');
    }
  };

  return (
    <ScrollView style={styles.container}>
      {/* EXISTING FIELDS */}
      <Text>Category</Text>
      <CategoryPicker
        selected={categoryId}
        onSelect={setCategoryId}
      />

      {/* ADD THIS NEW SECTION */}
      <Text>Asset Scope</Text>
      <Picker
        selectedValue={selectedAssetId || ''}
        onValueChange={(value) => setSelectedAssetId(value ? parseInt(value) : null)}
      >
        <Picker.Item label="All Assets (Global)" value="" />
        {assets?.map((asset) => (
          <Picker.Item
            key={asset.id}
            label={asset.name}
            value={asset.id}
          />
        ))}
      </Picker>

      {/* EXISTING FIELDS */}
      <Text>Amount</Text>
      <TextInput
        value={amount}
        onChangeText={setAmount}
        keyboardType="numeric"
        style={styles.input}
      />

      <TouchableOpacity style={styles.button} onPress={handleSubmit}>
        <Text style={styles.buttonText}>
          {route.params?.budgetId ? 'Update' : 'Create'} Budget
        </Text>
      </TouchableOpacity>
    </ScrollView>
  );
};
```

### 4.2 Update BudgetCard Component

```tsx
// components/BudgetCard.tsx - EXISTING COMPONENT

const BudgetCard = ({ budget, onEdit, onDelete }) => {
  return (
    <View style={styles.card}>
      {/* EXISTING: Category */}
      <Text style={styles.category}>{budget.category_name}</Text>

      {/* ADD THIS: Show Asset */}
      {budget.asset_id && (
        <View style={styles.assetBadge}>
          <Text style={styles.assetText}>{budget.asset_name}</Text>
        </View>
      )}

      {/* ADD THIS: Show Global */}
      {!budget.asset_id && (
        <View style={styles.globalBadge}>
          <Text style={styles.globalText}>🌐 All Assets</Text>
        </View>
      )}

      {/* EXISTING: Progress, Amount */}
      <ProgressBar progress={budget.percentage_used} />

      <Text style={styles.amount}>
        {formatCurrency(budget.spent_amount)} / {formatCurrency(budget.amount)}
      </Text>

      {/* EXISTING: Actions */}
      <View style={styles.actions}>
        <Button title="Edit" onPress={() => onEdit(budget)} />
        <Button title="Delete" onPress={() => onDelete(budget.id)} />
      </View>
    </View>
  );
};

// ADD THESE STYLES
const styles = StyleSheet.create({
  // EXISTING STYLES...

  assetBadge: {
    backgroundColor: '#E3F2FD',
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 12,
    alignSelf: 'flex-start',
    marginTop: 4,
  },
  assetText: {
    color: '#1976D2',
    fontSize: 12,
    fontWeight: '500',
  },
  globalBadge: {
    backgroundColor: '#F5F5F5',
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 12,
    alignSelf: 'flex-start',
    marginTop: 4,
  },
  globalText: {
    color: '#666',
    fontSize: 12,
  },
});
```

---

## 5. iOS (Swift/SwiftUI) Adjustments

### 5.1 Update Budget Model

```swift
// Models/Budget.swift - EXISTING MODEL

struct Budget: Identifiable, Codable {
    let id: Int
    let categoryId: Int
    let categoryName: String
    let assetId: Int?              // ADD THIS
    let assetName: String?         // ADD THIS
    let amount: Int
    let period: String
    let startDate: String
    let endDate: String
    let isActive: Bool
    let alertAt: Int
    let description: String?
    let createdAt: String
    let updatedAt: String

    // ADD THIS HELPER
    var isGlobal: Bool {
        return assetId == nil || assetId == 0
    }

    // ADD TO CODING KEYS
    enum CodingKeys: String, CodingKey {
        case id, amount, period, description, isActive, createdAt, updatedAt
        case categoryId = "category_id"
        case categoryName = "category_name"
        case assetId = "asset_id"      // ADD
        case assetName = "asset_name"  // ADD
        case startDate = "start_date"
        case endDate = "end_date"
        case alertAt = "alert_at"
    }
}
```

### 5.2 Update CreateBudgetRequest

```swift
// Models/CreateBudgetRequest.swift - EXISTING MODEL

struct CreateBudgetRequest: Encodable {
    let categoryId: Int
    let assetId: Int?                // ADD THIS
    let amount: Int
    let period: String
    let startDate: String
    let alertAt: Int?
    let description: String?

    // ADD TO CODING KEYS
    enum CodingKeys: String, CodingKey {
        case amount, period, description
        case categoryId = "category_id"
        case assetId = "asset_id"      // ADD
        case startDate = "start_date"
        case alertAt = "alert_at"
    }
}
```

### 5.3 Update BudgetFormView

```swift
// Views/BudgetFormView.swift - EXISTING VIEW

struct BudgetFormView: View {
    @StateObject private var viewModel: BudgetFormViewModel

    // ADD ASSETS PROPERTY
    let assets: [Asset]

    var body: some View {
        Form {
            // EXISTING: Category Picker
            Section {
                Picker("Category", selection: $viewModel.categoryId) {
                    ForEach(viewModel.categories) { category in
                        Text(category.name).tag(category.id)
                    }
                }
            }

            // ADD THIS NEW SECTION
            Section {
                Picker("Asset Scope", selection: $viewModel.assetId) {
                    Text("All Assets (Global)").tag(nil as Int?)
                    ForEach(assets) { asset in
                        Text(asset.name).tag(asset.id as Int?)
                    }
                }
            } header: {
                Text("Asset")
            } footer: {
                Text("Select a specific asset or leave empty for a global budget")
            }

            // EXISTING: Amount, Period, Date, etc.
            Section {
                HStack {
                    Text("Amount")
                    Spacer()
                    TextField("0", value: $viewModel.amount, format: .number)
                        .keyboardType(.decimalPad)
                }
            }

            // ... rest of existing form ...

            Section {
                Button(action: {
                    viewModel.submit()
                    // dismiss or navigation
                }) {
                    Text("Create Budget")
                        .frame(maxWidth: .infinity)
                }
                .disabled(!viewModel.isValid)
            }
        }
        .navigationTitle("Create Budget")
    }
}
```

### 5.4 Update BudgetFormViewModel

```swift
// ViewModels/BudgetFormViewModel.swift - EXISTING VIEW MODEL

class BudgetFormViewModel: ObservableObject {
    @Published var categoryId: Int?
    @Published var assetId: Int?          // ADD THIS
    @Published var amount: Double = 0
    @Published var period: String = "monthly"
    // ... other properties ...

    var isValid: Bool {
        categoryId != nil && amount > 0
    }

    func submit() {
        guard let categoryId = categoryId, amount > 0 else { return }

        let request = CreateBudgetRequest(
            categoryId: categoryId,
            assetId: assetId,          // ADD THIS
            amount: Int(amount),
            period: period,
            startDate: ISO8601DateFormatter().string(from: startDate),
            alertAt: Int(alertAt),
            description: description.isEmpty ? nil : description
        )

        // ... existing submit logic ...
    }
}
```

### 5.5 Update BudgetCardView

```swift
// Views/BudgetCardView.swift - EXISTING VIEW

struct BudgetCardView: View {
    let budget: Budget
    let onEdit: () -> Void
    let onDelete: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(budget.categoryName)
                        .font(.headline)

                    // ADD THIS: Show Asset
                    if let assetName = budget.assetName {
                        Text(assetName)
                            .font(.caption)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color.blue.opacity(0.1))
                            .foregroundColor(.blue)
                            .cornerRadius(8)
                    }

                    // ADD THIS: Show Global
                    if budget.isGlobal {
                        Text("🌐 All Assets")
                            .font(.caption)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color.gray.opacity(0.1))
                            .foregroundColor(.gray)
                            .cornerRadius(8)
                    }
                }

                Spacer()

                Text("\(Int(budget.percentageUsed))%")
                    .font(.headline)
                    .foregroundColor(statusColor)
            }

            // EXISTING: Progress Bar, Amounts, Actions
            ProgressView(value: budget.percentageUsed / 100)
                .tint(statusColor)

            HStack {
                Text("Spent: \(formatCurrency(budget.spentAmount))")
                Spacer()
                Text("Budget: \(formatCurrency(budget.amount))")
            }

            HStack {
                Button("Edit") { onEdit() }
                Button("Delete", role: .destructive) { onDelete() }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .cornerRadius(12)
        .shadow(radius: 2)
    }
}
```

---

## 6. Testing Adjustments

### 6.1 Update Existing Tests

```typescript
// __tests__/budget/BudgetForm.test.tsx - EXISTING TEST

describe('BudgetForm', () => {
  it('should submit budget with asset_id', async () => {
    const onSubmit = jest.fn();

    render(<BudgetForm assets={mockAssets} onSubmit={onSubmit} />);

    // EXISTING TESTS...

    // ADD THIS TEST
    await userEvent.selectOptions(
      screen.getByLabelText('Asset Scope'),
      'Checking Account'
    );

    await userEvent.click(screen.getByText('Create Budget'));

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        asset_id: 1,
        category_id: expect.any(Number),
        amount: expect.any(Number),
      })
    );
  });

  it('should submit budget without asset_id (global)', async () => {
    const onSubmit = jest.fn();

    render(<BudgetForm assets={mockAssets} onSubmit={onSubmit} />);

    // Don't select any asset (leave as "All Assets")

    await userEvent.click(screen.getByText('Create Budget'));

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        asset_id: null,
        category_id: expect.any(Number),
        amount: expect.any(Number),
      })
    );
  });
});
```

---

## 7. Migration Strategy

### 7.1 No Breaking Changes Required

The backend is **backward compatible**:
- `asset_id` is optional (nullable)
- Existing budgets have `asset_id = null` (global budgets)
- Old API calls without `asset_id` still work

### 7.2 Rollout Plan

1. **Deploy Backend First** ✅ (Already done)
2. **Update API Types** - No breaking changes, just new fields
3. **Add Asset Picker to Forms** - Optional field, can be rolled out gradually
4. **Add Asset Filter to Lists** - Optional enhancement
5. **Show Asset in Budget Cards** - Display enhancement

### 7.3 Gradual Feature Rollout

```typescript
// Feature flag for gradual rollout
const SHOW_ASSET_SELECTOR = process.env.REACT_APP_SHOW_ASSET_SELECTOR === 'true';

{SHOW_ASSET_SELECTOR && (
  <AssetSelector assets={assets} selected={selectedAssetId} onChange={setSelectedAssetId} />
)}
```

---

## Summary Checklist

### Web (React/Next.js)
- [ ] Update `Budget` type to include `asset_id` and `asset_name`
- [ ] Update `CreateBudgetRequest` to include `asset_id`
- [ ] Add asset selection to `BudgetForm` component
- [ ] Add asset filter to `BudgetList` component
- [ ] Show asset name or "All Assets" in `BudgetCard`
- [ ] Update API service to send `asset_id` in requests

### Mobile (React Native)
- [ ] Update `Budget` interface
- [ ] Add asset picker to `BudgetFormScreen`
- [ ] Show asset in `BudgetCard` component
- [ ] Add asset filter to budget list

### iOS (Swift/SwiftUI)
- [ ] Add `assetId` and `assetName` to `Budget` struct
- [ ] Update `CreateBudgetRequest` with `assetId`
- [ ] Add asset picker to `BudgetFormView`
- [ ] Show asset in `BudgetCardView`
- [ ] Update `BudgetFormViewModel`

---

## Quick Reference: What to Change

| File | Change |
|------|--------|
| Types/Interfaces | Add `asset_id?: number \| null` and `asset_name: string` |
| BudgetForm | Add `<AssetSelector />` component |
| BudgetCard | Display `asset_name` or "All Assets" |
| API Service | Include `asset_id` in request body/params |
| Tests | Test with and without `asset_id` |

That's it! Your existing budget implementation just needs these adjustments to support asset filtering.
