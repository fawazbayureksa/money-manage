# Budget Alerts API - Frontend Implementation Guide

## Overview

The Budget Alerts API provides paginated access to user budget alerts with filtering capabilities. This document covers how to integrate the API into your frontend application.

---

## API Endpoint

### GET `/api/budgets/alerts`

Retrieves paginated budget alerts for the authenticated user.

---

## Request Parameters (Query String)

| Parameter     | Type    | Required | Default | Description                                    |
|--------------|---------|----------|---------|------------------------------------------------|
| `page`       | integer | No       | 1       | Page number (starts from 1)                    |
| `page_size`  | integer | No       | 10      | Number of items per page (max: 100)            |
| `unread_only`| boolean | No       | false   | Filter to show only unread alerts              |
| `budget_id`  | integer | No       | -       | Filter alerts by specific budget ID            |
| `sort_by`    | string  | No       | created_at | Field to sort by                            |
| `sort_dir`   | string  | No       | desc    | Sort direction: `asc` or `desc`               |

---

## Response Structure

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Alerts retrieved successfully",
  "data": {
    "data": [
      {
        "id": 1,
        "budget_id": 5,
        "percentage": 85,
        "spent_amount": 850000,
        "message": "You have reached 85% of your Food budget",
        "is_read": false,
        "created_at": "2026-01-15T10:30:00Z",
        "category_id": 3,
        "category_name": "Food",
        "budget_amount": 1000000
      },
      {
        "id": 2,
        "budget_id": 6,
        "percentage": 100,
        "spent_amount": 500000,
        "message": "You have exceeded 100% of your Transportation budget",
        "is_read": true,
        "created_at": "2026-01-14T08:15:00Z",
        "category_id": 4,
        "category_name": "Transportation",
        "budget_amount": 500000
      }
    ],
    "page": 1,
    "page_size": 10,
    "total_items": 25,
    "total_pages": 3
  }
}
```

### Error Response (401 Unauthorized)

```json
{
  "success": false,
  "message": "User not authenticated"
}
```

---

## Frontend Implementation Examples

### 1. React/Next.js with TypeScript

#### Types Definition

```typescript
// types/budget.ts

export interface BudgetAlert {
  id: number;
  budget_id: number;
  percentage: number;
  spent_amount: number;
  message: string;
  is_read: boolean;
  created_at: string;
  category_id?: number;
  category_name?: string;
  budget_amount?: number;
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

export interface AlertsResponse {
  success: boolean;
  message: string;
  data: {
    data: BudgetAlert[];
    page: number;
    page_size: number;
    total_items: number;
    total_pages: number;
  };
}

export interface AlertFilterParams {
  page?: number;
  page_size?: number;
  unread_only?: boolean;
  budget_id?: number;
  sort_by?: 'created_at' | 'percentage';
  sort_dir?: 'asc' | 'desc';
}
```

#### API Service

```typescript
// services/alertService.ts

import { AlertsResponse, AlertFilterParams } from '../types/budget';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const alertService = {
  async getAlerts(params: AlertFilterParams = {}, token: string): Promise<AlertsResponse> {
    const queryParams = new URLSearchParams();
    
    if (params.page) queryParams.append('page', params.page.toString());
    if (params.page_size) queryParams.append('page_size', params.page_size.toString());
    if (params.unread_only) queryParams.append('unread_only', 'true');
    if (params.budget_id) queryParams.append('budget_id', params.budget_id.toString());
    if (params.sort_by) queryParams.append('sort_by', params.sort_by);
    if (params.sort_dir) queryParams.append('sort_dir', params.sort_dir);

    const response = await fetch(
      `${API_BASE_URL}/api/budgets/alerts?${queryParams.toString()}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error('Failed to fetch alerts');
    }

    return response.json();
  },

  async markAsRead(alertId: number, token: string): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/api/budgets/alerts/${alertId}/read`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error('Failed to mark alert as read');
    }
  },

  async markAllAsRead(token: string): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/api/budgets/alerts/read-all`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error('Failed to mark all alerts as read');
    }
  },
};
```

#### React Hook

```typescript
// hooks/useAlerts.ts

import { useState, useEffect, useCallback } from 'react';
import { alertService } from '../services/alertService';
import { BudgetAlert, AlertFilterParams, PaginationMeta } from '../types/budget';

interface UseAlertsReturn {
  alerts: BudgetAlert[];
  pagination: PaginationMeta | null;
  loading: boolean;
  error: string | null;
  fetchAlerts: (params?: AlertFilterParams) => Promise<void>;
  markAsRead: (alertId: number) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  setPage: (page: number) => void;
  filters: AlertFilterParams;
  setFilters: (filters: AlertFilterParams) => void;
}

export function useAlerts(token: string): UseAlertsReturn {
  const [alerts, setAlerts] = useState<BudgetAlert[]>([]);
  const [pagination, setPagination] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<AlertFilterParams>({
    page: 1,
    page_size: 10,
    unread_only: false,
    sort_dir: 'desc',
  });

  const fetchAlerts = useCallback(async (params?: AlertFilterParams) => {
    setLoading(true);
    setError(null);

    try {
      const finalParams = { ...filters, ...params };
      const response = await alertService.getAlerts(finalParams, token);
      
      setAlerts(response.data.data);
      setPagination({
        page: response.data.page,
        page_size: response.data.page_size,
        total_items: response.data.total_items,
        total_pages: response.data.total_pages,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch alerts');
    } finally {
      setLoading(false);
    }
  }, [filters, token]);

  const markAsRead = async (alertId: number) => {
    try {
      await alertService.markAsRead(alertId, token);
      setAlerts(prev =>
        prev.map(alert =>
          alert.id === alertId ? { ...alert, is_read: true } : alert
        )
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark as read');
    }
  };

  const markAllAsRead = async () => {
    try {
      await alertService.markAllAsRead(token);
      setAlerts(prev => prev.map(alert => ({ ...alert, is_read: true })));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark all as read');
    }
  };

  const setPage = (page: number) => {
    setFilters(prev => ({ ...prev, page }));
  };

  useEffect(() => {
    fetchAlerts();
  }, [fetchAlerts]);

  return {
    alerts,
    pagination,
    loading,
    error,
    fetchAlerts,
    markAsRead,
    markAllAsRead,
    setPage,
    filters,
    setFilters,
  };
}
```

#### React Component Example

```tsx
// components/AlertsList.tsx

import React from 'react';
import { useAlerts } from '../hooks/useAlerts';
import { useAuth } from '../hooks/useAuth'; // Assume you have auth hook

export function AlertsList() {
  const { token } = useAuth();
  const {
    alerts,
    pagination,
    loading,
    error,
    markAsRead,
    markAllAsRead,
    setPage,
    filters,
    setFilters,
  } = useAlerts(token);

  if (loading) {
    return <div className="loading">Loading alerts...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="alerts-container">
      {/* Filters */}
      <div className="alerts-filters">
        <label>
          <input
            type="checkbox"
            checked={filters.unread_only || false}
            onChange={(e) =>
              setFilters({ ...filters, unread_only: e.target.checked, page: 1 })
            }
          />
          Show unread only
        </label>
        
        <button onClick={markAllAsRead} disabled={alerts.every(a => a.is_read)}>
          Mark all as read
        </button>
      </div>

      {/* Alerts List */}
      <div className="alerts-list">
        {alerts.length === 0 ? (
          <p>No alerts found</p>
        ) : (
          alerts.map((alert) => (
            <div
              key={alert.id}
              className={`alert-item ${alert.is_read ? 'read' : 'unread'}`}
              onClick={() => !alert.is_read && markAsRead(alert.id)}
            >
              <div className="alert-header">
                <span className={`percentage ${alert.percentage >= 100 ? 'exceeded' : 'warning'}`}>
                  {alert.percentage}%
                </span>
                <span className="category">{alert.category_name}</span>
                {!alert.is_read && <span className="unread-badge">New</span>}
              </div>
              <p className="alert-message">{alert.message}</p>
              <div className="alert-footer">
                <span className="spent">
                  Spent: {alert.spent_amount.toLocaleString()} / {alert.budget_amount?.toLocaleString()}
                </span>
                <span className="date">
                  {new Date(alert.created_at).toLocaleDateString()}
                </span>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Pagination */}
      {pagination && pagination.total_pages > 1 && (
        <div className="pagination">
          <button
            disabled={pagination.page === 1}
            onClick={() => setPage(pagination.page - 1)}
          >
            Previous
          </button>
          
          <span>
            Page {pagination.page} of {pagination.total_pages}
            ({pagination.total_items} total)
          </span>
          
          <button
            disabled={pagination.page === pagination.total_pages}
            onClick={() => setPage(pagination.page + 1)}
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}
```

---

### 2. Flutter/Dart Implementation

#### Model Class

```dart
// models/budget_alert.dart

class BudgetAlert {
  final int id;
  final int budgetId;
  final int percentage;
  final int spentAmount;
  final String message;
  final bool isRead;
  final DateTime createdAt;
  final int? categoryId;
  final String? categoryName;
  final int? budgetAmount;

  BudgetAlert({
    required this.id,
    required this.budgetId,
    required this.percentage,
    required this.spentAmount,
    required this.message,
    required this.isRead,
    required this.createdAt,
    this.categoryId,
    this.categoryName,
    this.budgetAmount,
  });

  factory BudgetAlert.fromJson(Map<String, dynamic> json) {
    return BudgetAlert(
      id: json['id'],
      budgetId: json['budget_id'],
      percentage: json['percentage'],
      spentAmount: json['spent_amount'],
      message: json['message'],
      isRead: json['is_read'],
      createdAt: DateTime.parse(json['created_at']),
      categoryId: json['category_id'],
      categoryName: json['category_name'],
      budgetAmount: json['budget_amount'],
    );
  }
}

class PaginatedAlerts {
  final List<BudgetAlert> data;
  final int page;
  final int pageSize;
  final int totalItems;
  final int totalPages;

  PaginatedAlerts({
    required this.data,
    required this.page,
    required this.pageSize,
    required this.totalItems,
    required this.totalPages,
  });

  factory PaginatedAlerts.fromJson(Map<String, dynamic> json) {
    return PaginatedAlerts(
      data: (json['data'] as List)
          .map((item) => BudgetAlert.fromJson(item))
          .toList(),
      page: json['page'],
      pageSize: json['page_size'],
      totalItems: json['total_items'],
      totalPages: json['total_pages'],
    );
  }
}
```

#### API Service

```dart
// services/alert_service.dart

import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/budget_alert.dart';

class AlertService {
  final String baseUrl;
  final String token;

  AlertService({required this.baseUrl, required this.token});

  Future<PaginatedAlerts> getAlerts({
    int page = 1,
    int pageSize = 10,
    bool unreadOnly = false,
    int? budgetId,
    String sortBy = 'created_at',
    String sortDir = 'desc',
  }) async {
    final queryParams = {
      'page': page.toString(),
      'page_size': pageSize.toString(),
      'sort_by': sortBy,
      'sort_dir': sortDir,
    };

    if (unreadOnly) queryParams['unread_only'] = 'true';
    if (budgetId != null) queryParams['budget_id'] = budgetId.toString();

    final uri = Uri.parse('$baseUrl/api/budgets/alerts')
        .replace(queryParameters: queryParams);

    final response = await http.get(
      uri,
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body);
      return PaginatedAlerts.fromJson(json['data']);
    } else {
      throw Exception('Failed to fetch alerts');
    }
  }

  Future<void> markAsRead(int alertId) async {
    final response = await http.put(
      Uri.parse('$baseUrl/api/budgets/alerts/$alertId/read'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to mark alert as read');
    }
  }

  Future<void> markAllAsRead() async {
    final response = await http.put(
      Uri.parse('$baseUrl/api/budgets/alerts/read-all'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to mark all alerts as read');
    }
  }
}
```

---

### 3. Vanilla JavaScript / Fetch API

```javascript
// alertsApi.js

const API_BASE_URL = 'http://localhost:8080';

export const alertsApi = {
  async getAlerts(params = {}, token) {
    const queryParams = new URLSearchParams();
    
    if (params.page) queryParams.append('page', params.page);
    if (params.pageSize) queryParams.append('page_size', params.pageSize);
    if (params.unreadOnly) queryParams.append('unread_only', 'true');
    if (params.budgetId) queryParams.append('budget_id', params.budgetId);
    if (params.sortBy) queryParams.append('sort_by', params.sortBy);
    if (params.sortDir) queryParams.append('sort_dir', params.sortDir);

    const response = await fetch(
      `${API_BASE_URL}/api/budgets/alerts?${queryParams.toString()}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  },

  async markAsRead(alertId, token) {
    const response = await fetch(
      `${API_BASE_URL}/api/budgets/alerts/${alertId}/read`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  },

  async markAllAsRead(token) {
    const response = await fetch(
      `${API_BASE_URL}/api/budgets/alerts/read-all`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  },
};
```

---

## Common Use Cases

### 1. Initial Load with Default Pagination

```typescript
// Load first page with 10 items
const response = await alertService.getAlerts({}, token);
```

### 2. Load Only Unread Alerts

```typescript
const response = await alertService.getAlerts({
  unread_only: true,
  page: 1,
  page_size: 20,
}, token);
```

### 3. Filter Alerts by Specific Budget

```typescript
const response = await alertService.getAlerts({
  budget_id: 5,
  page: 1,
}, token);
```

### 4. Load Alerts Sorted by Percentage (Highest First)

```typescript
const response = await alertService.getAlerts({
  sort_by: 'percentage',
  sort_dir: 'desc',
}, token);
```

### 5. Infinite Scroll / Load More

```typescript
const loadMore = async () => {
  if (pagination.page < pagination.total_pages) {
    const response = await alertService.getAlerts({
      page: pagination.page + 1,
    }, token);
    
    setAlerts(prev => [...prev, ...response.data.data]);
    setPagination({
      page: response.data.page,
      page_size: response.data.page_size,
      total_items: response.data.total_items,
      total_pages: response.data.total_pages,
    });
  }
};
```

---

## Related Endpoints

| Method | Endpoint                        | Description                      |
|--------|--------------------------------|----------------------------------|
| GET    | `/api/budgets/alerts`          | Get paginated alerts             |
| PUT    | `/api/budgets/alerts/:id/read` | Mark specific alert as read      |
| PUT    | `/api/budgets/alerts/read-all` | Mark all alerts as read          |

---

## Error Handling

| Status Code | Description              | Action                          |
|-------------|--------------------------|----------------------------------|
| 200         | Success                  | Process the data                 |
| 400         | Invalid query parameters | Check parameter types/values     |
| 401         | Unauthorized             | Redirect to login                |
| 500         | Server error             | Show error message, retry later  |

---

## Best Practices

1. **Debounce Filter Changes**: When implementing search/filter, debounce API calls to avoid excessive requests.

2. **Cache Results**: Consider caching alert results and invalidating when marking as read.

3. **Optimistic Updates**: When marking alerts as read, update UI immediately before API confirmation.

4. **Handle Loading States**: Show skeleton loaders or spinners during API calls.

5. **Error Boundaries**: Implement proper error handling to show user-friendly messages.

6. **Polling for New Alerts**: Consider implementing polling or WebSocket for real-time alert updates.

```typescript
// Example polling implementation
useEffect(() => {
  const interval = setInterval(() => {
    fetchAlerts({ unread_only: true });
  }, 30000); // Poll every 30 seconds

  return () => clearInterval(interval);
}, []);
```
