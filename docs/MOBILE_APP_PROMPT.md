# React Native Money Management App - AI Implementation Guide

## Project Overview

Build a React Native mobile application for money management that connects to an existing Go backend API. The app is user-focused (not admin) and includes authentication, category management, transaction tracking, budget management, and budget alerts.

**Backend API Base URL**: `http://your-server:8080` (replace with actual URL)

**Tech Stack**:
- React Native (Expo or React Native CLI)
- React Navigation for routing
- Axios for API calls
- AsyncStorage for token persistence
- Context API or Redux for state management

for ui please use 
npm install react-native-paper

From v5 there is a need to install react-native-safe-area-context for handling safe area.

npm install react-native-safe-area-context

Additionaly for iOS platform there is a requirement to link the native parts of the library:

npx pod-install

Specifically MaterialDesignIcons icon pack needs to be included in the project, because some components use those internally (e.g. AppBar.BackAction on Android).

npm install @react-native-vector-icons/material-design-icons

---

## SECTION 1: Project Setup & Authentication Structure

### Objective
Set up the React Native project with basic structure, navigation, and authentication screens (Login & Register).

### Requirements

1. **Project Initialization**
   - Create new React Native project
   - Install dependencies:
     - `@react-navigation/native`
     - `@react-navigation/stack`
     - `@react-navigation/bottom-tabs`
     - `axios`
     - `@react-native-async-storage/async-storage`
     - `react-native-vector-icons`

2. **Folder Structure**
   ```
   src/
   ├── api/
   │   └── client.js          # Axios instance with interceptors
   ├── screens/
   │   ├── Auth/
   │   │   ├── LoginScreen.js
   │   │   └── RegisterScreen.js
   ├── context/
   │   └── AuthContext.js     # Authentication state management
   ├── navigation/
   │   ├── AuthNavigator.js
   │   └── AppNavigator.js
   └── utils/
       └── storage.js         # AsyncStorage helpers
   ```

3. **API Client Setup** (`src/api/client.js`)
   ```javascript
   import axios from 'axios';
   import AsyncStorage from '@react-native-async-storage/async-storage';

   const API_BASE_URL = 'http://your-server:8080';

   const apiClient = axios.create({
     baseURL: API_BASE_URL,
     timeout: 10000,
     headers: {
       'Content-Type': 'application/json',
     },
   });

   // Request interceptor to add token
   apiClient.interceptors.request.use(
     async (config) => {
       const token = await AsyncStorage.getItem('auth_token');
       if (token) {
         config.headers.Authorization = `Bearer ${token}`;
       }
       return config;
     },
     (error) => Promise.reject(error)
   );

   // Response interceptor for error handling
   apiClient.interceptors.response.use(
     (response) => response,
     async (error) => {
       if (error.response?.status === 401) {
         await AsyncStorage.removeItem('auth_token');
         // Navigate to login
       }
       return Promise.reject(error);
     }
   );

   export default apiClient;
   ```

4. **Authentication Context** (`src/context/AuthContext.js`)
   - Store user state and token
   - Provide login, register, logout functions
   - Auto-check token on app start

5. **Login Screen** (`src/screens/Auth/LoginScreen.js`)
   - Email input field
   - Password input field (secure text entry)
   - Login button
   - Link to Register screen
   - Show loading state during API call
   - Handle errors with user-friendly messages

6. **Register Screen** (`src/screens/Auth/RegisterScreen.js`)
   - Username input field
   - Email input field
   - Password input field (minimum 6 characters)
   - Register button
   - Link to Login screen
   - Show loading state
   - Handle validation errors

### API Endpoints

**Register**
```
POST /register
Body: {
  "username": "string",
  "email": "string",
  "password": "string"
}
Response: {
  "status": true,
  "message": "User registered successfully",
  "data": {
    "token": "jwt-token-here"
  }
}
```

**Login**
```
POST /login
Body: {
  "email": "string",
  "password": "string"
}
Response: {
  "status": true,
  "message": "Login successful",
  "data": {
    "token": "jwt-token-here"
  }
}
```

### Deliverables
- Working authentication flow (Login/Register)
- Token storage in AsyncStorage
- Navigation between Auth and Main app
- Error handling with alerts/toasts

---

## SECTION 2: Category Management

### Objective
Create screens for viewing, adding, and managing transaction categories.

### Requirements

1. **New Files**
   ```
   src/
   ├── screens/
   │   └── Category/
   │       ├── CategoryListScreen.js
   │       └── AddCategoryScreen.js
   └── api/
       └── categoryService.js
   ```

2. **Category Service** (`src/api/categoryService.js`)
   ```javascript
   import apiClient from './client';

   export const getCategories = async () => {
     const response = await apiClient.get('/categories');
     return response.data;
   };

   export const createCategory = async (categoryData) => {
     const response = await apiClient.post('/categories', categoryData);
     return response.data;
   };

   export const deleteCategory = async (categoryId) => {
     const response = await apiClient.delete(`/categories/${categoryId}`);
     return response.data;
   };
   ```

3. **Category List Screen**
   - Display all user categories in a list
   - Show category name
   - Add floating action button (FAB) to create new category
   - Swipe-to-delete functionality (optional)
   - Empty state message when no categories
   - Pull-to-refresh

4. **Add Category Screen**
   - Category name input field
   - Submit button
   - Navigate back to list on success
   - Show success message

### API Endpoints

**Get Categories**
```
GET /categories
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Success",
  "data": [
    {
      "id": 1,
      "name": "Food",
      "user_id": 1,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

**Create Category**
```
POST /categories
Headers: Authorization: Bearer {token}
Body: {
  "name": "string"
}
Response: {
  "status": true,
  "message": "Category created successfully",
  "data": {
    "id": 1,
    "name": "Food"
  }
}
```

**Delete Category**
```
DELETE /categories/:id
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Category deleted successfully"
}
```

### Deliverables
- Category list with user's categories
- Add new category functionality
- Delete category option
- Proper error handling

---

## SECTION 3: Transaction Management

### Objective
Create screens to add transactions and view transaction history.

### Requirements

1. **New Files**
   ```
   src/
   ├── screens/
   │   └── Transaction/
   │       ├── AddTransactionScreen.js
   │       └── TransactionListScreen.js
   └── api/
       └── transactionService.js
   ```

2. **Transaction Service** (`src/api/transactionService.js`)
   ```javascript
   import apiClient from './client';

   export const getTransactions = async (params) => {
     const response = await apiClient.get('/transactions', { params });
     return response.data;
   };

   export const createTransaction = async (transactionData) => {
     const response = await apiClient.post('/transactions', transactionData);
     return response.data;
   };
   ```

3. **Add Transaction Screen**
   - Transaction type selector (Income/Expense) - toggle or radio buttons
   - Amount input (numeric keyboard)
   - Category picker (dropdown/modal from user's categories)
   - Bank picker (dropdown/modal from user's banks)
   - Description/note input (optional, multiline)
   - Date picker (default to today)
   - Submit button
   - Clear validation errors

4. **Transaction List Screen**
   - Display transactions sorted by date (newest first)
   - Show for each transaction:
     - Type indicator (Income: green, Expense: red)
     - Amount with currency
     - Category name
     - Date
     - Bank name (optional)
   - Group by date (optional enhancement)
   - Filter by date range (optional)
   - FAB to add new transaction

### API Endpoints

**Get Transactions**
```
GET /transactions?page=1&page_size=20
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Success",
  "data": [
    {
      "id": 1,
      "amount": 50000,
      "description": "Lunch",
      "transaction_type": 2,  // 1=income, 2=expense
      "date": "2025-01-15",
      "user_id": 1,
      "category_id": 1,
      "bank_id": 1,
      "Category": {
        "id": 1,
        "name": "Food"
      },
      "Bank": {
        "id": 1,
        "bank_name": "BCA"
      }
    }
  ],
  "pagination": {
    "current_page": 1,
    "page_size": 20,
    "total_records": 50,
    "total_pages": 3
  }
}
```

**Create Transaction**
```
POST /transactions
Headers: Authorization: Bearer {token}
Body: {
  "amount": 50000,
  "description": "Lunch at restaurant",
  "transaction_type": 2,  // 1 for income, 2 for expense
  "date": "2025-01-15",
  "category_id": 1,
  "bank_id": 1  // optional
}
Response: {
  "status": true,
  "message": "Transaction created successfully",
  "data": {
    "id": 1,
    "amount": 50000,
    "transaction_type": 2
  }
}
```

### Deliverables
- Add transaction form with all fields
- Transaction list with filtering
- Type-based color coding
- Date formatting
- Validation for required fields

---

## SECTION 4: Budget Management

### Objective
Create screens to set budgets for categories and view budget status.

### Requirements

1. **New Files**
   ```
   src/
   ├── screens/
   │   └── Budget/
   │       ├── BudgetListScreen.js
   │       ├── AddBudgetScreen.js
   │       └── BudgetDetailScreen.js
   └── api/
       └── budgetService.js
   ```

2. **Budget Service** (`src/api/budgetService.js`)
   ```javascript
   import apiClient from './client';

## SECTION 4: Budget Management

### Objective
Create screens to set budgets for categories and view budget status.

### Requirements

1. **New Files**
   ```
   src/
   ├── screens/
   │   └── Budget/
   │       ├── BudgetListScreen.js
   │       ├── AddBudgetScreen.js
   │       └── BudgetDetailScreen.js
   └── api/
       └── budgetService.js
   ```

2. **Budget Service** (`src/api/budgetService.js`)
   ```javascript
   import apiClient from './client';

   export const getBudgets = async () => {
     const response = await apiClient.get('/budgets');
     return response.data;
   };

   export const getBudgetStatus = async () => {
     const response = await apiClient.get('/budgets/status');
     return response.data;
   };

   export const createBudget = async (budgetData) => {
     const response = await apiClient.post('/budgets', budgetData);
     return response.data;
   };
   ```

3. **Budget List Screen**
   - Display all active budgets
   - Show for each budget:
     - Category name
     - Budget amount
     - Spent amount
     - Remaining amount
     - Progress bar (visual representation)
     - Percentage used
     - Status indicator (Safe/Warning/Exceeded)
   - Color coding:
     - Green: < 70% used
     - Yellow: 70-100% used
     - Red: > 100% used
   - FAB to add new budget
   - Pull-to-refresh

4. **Add Budget Screen**
   - Category picker (only categories without active budget)
   - Budget amount input (numeric)
   - Period selector (Monthly/Yearly)
   - Start date picker
   - Alert threshold slider (50-100%, default 80%)
   - Submit button

5. **Budget Detail Screen** (optional enhancement)
   - Show budget details
   - List transactions in this category
   - Visual chart of spending over time

### API Endpoints

**Get Budget Status**
```
GET /budgets/status
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Success",
  "data": [
    {
      "id": 1,
      "category_id": 1,
      "category_name": "Food",
      "amount": 500000,
      "period": "monthly",
      "start_date": "2025-01-01",
      "end_date": "2025-01-31",
      "alert_at": 80,
      "spent_amount": 350000,
      "remaining_amount": 150000,
      "percentage_used": 70,
      "status": "warning"  // safe, warning, exceeded
    }
  ]
}
```

**Create Budget**
```
POST /budgets
Headers: Authorization: Bearer {token}
Body: {
  "category_id": 1,
  "amount": 500000,
  "period": "monthly",  // monthly or yearly
  "start_date": "2025-01-01",
  "alert_at": 80  // percentage (50-100)
}
Response: {
  "status": true,
  "message": "Budget created successfully",
  "data": {
    "id": 1,
    "category_id": 1,
    "amount": 500000
  }
}
```

**Get All Budgets**
```
GET /budgets
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Success",
  "data": [
    {
      "id": 1,
      "category_id": 1,
      "amount": 500000,
      "period": "monthly",
      "start_date": "2025-01-01",
      "alert_at": 80
    }
  ]
}
```

### Deliverables
- Budget list with visual progress indicators
- Add budget form
- Real-time budget status
- Color-coded warnings
- Progress bars for each budget

---

## SECTION 5: Budget Alerts & Notifications

### Objective
Display budget alerts and notify users when they exceed thresholds.

### Requirements

1. **New Files**
   ```
   src/
   ├── screens/
   │   └── Alerts/
   │       └── AlertListScreen.js
   └── components/
       └── AlertBadge.js
   ```

2. **Alert List Screen**
   - Display all budget alerts
   - Show for each alert:
     - Category name
     - Alert message (e.g., "You've reached 80% of your Food budget")
     - Timestamp
     - Current spending vs budget
   - Mark as read functionality
   - Filter: All/Unread
   - Badge on navigation tab showing unread count
   - Empty state when no alerts

3. **Alert Badge Component**
   - Red badge with count
   - Show on Alerts tab icon
   - Update in real-time

4. **Home Screen Enhancement** (optional)
   - Show recent alerts at the top
   - Quick summary of budget status

### API Endpoints

**Get Budget Alerts**
```
GET /budgets/alerts?is_read=false
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Success",
  "data": [
    {
      "id": 1,
      "budget_id": 1,
      "user_id": 1,
      "message": "You've reached 80% of your Food budget for January 2025",
      "alert_type": "threshold",  // threshold or exceeded
      "is_read": false,
      "created_at": "2025-01-20T10:30:00Z",
      "Budget": {
        "id": 1,
        "category_id": 1,
        "amount": 500000,
        "Category": {
          "id": 1,
          "name": "Food"
        }
      }
    }
  ]
}
```

**Mark Alert as Read**
```
PUT /budgets/alerts/:id/read
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Alert marked as read"
}
```

### Deliverables
- Alert list screen
- Unread alert count badge
- Mark as read functionality
- Alert filtering
- Real-time alert updates

---

## SECTION 6: Navigation & Polish

### Objective
Set up bottom tab navigation and add finishing touches.

### Requirements

1. **Bottom Tab Navigation**
   - Tab 1: Home/Dashboard
   - Tab 2: Transactions
   - Tab 3: Add Transaction (center, highlighted)
   - Tab 4: Budgets
   - Tab 5: Alerts (with badge)

2. **Home/Dashboard Screen** (simple version)
   - Welcome message with username
   - Quick stats:
     - Total income (this month)
     - Total expenses (this month)
     - Active budgets count
     - Recent transactions (last 5)
   - Quick action buttons:
     - Add Transaction
     - Add Budget
   - Pull-to-refresh

3. **Polish & UX Improvements**
   - Loading states (spinners/skeletons)
   - Error messages (user-friendly)
   - Success toasts/alerts
   - Form validation feedback
   - Empty states with helpful messages
   - Logout button (in settings/profile)
   - Currency formatting (consistent)
   - Date formatting (user-friendly)

4. **Optional Enhancements**
   - Dark mode support
   - Profile screen with user info
   - Edit/delete transactions
   - Edit/delete budgets
   - Bank management (add/list banks)
   - Charts/graphs for spending
   - Search transactions
   - Export data

### Dashboard API Endpoint

**Get Dashboard Summary**
```
GET /analytics/dashboard?start_date=2025-01-01&end_date=2025-01-31
Headers: Authorization: Bearer {token}
Response: {
  "status": true,
  "message": "Success",
  "data": {
    "total_income": 5000000,
    "total_expense": 3500000,
    "net_savings": 1500000,
    "savings_rate": 30,
    "active_budgets_count": 5,
    "exceeded_budgets_count": 1,
    "top_spending_categories": [
      {
        "category_name": "Food",
        "total_amount": 1200000,
        "percentage": 34.3
      }
    ]
  }
}
```

### Deliverables
- Complete bottom tab navigation
- Dashboard with quick stats
- Logout functionality
- Consistent styling across screens
- Loading and error states
- User-friendly messages

---

## SECTION 7: Testing & Deployment Checklist

### Final Testing

1. **Authentication Flow**
   - [ ] Register new user
   - [ ] Login with credentials
   - [ ] Token persists after app restart
   - [ ] Logout clears token

2. **Categories**
   - [ ] View all categories
   - [ ] Create new category
   - [ ] Delete category (if no transactions)

3. **Transactions**
   - [ ] Add income transaction
   - [ ] Add expense transaction
   - [ ] View transaction list
   - [ ] Transactions display correctly

4. **Budgets**
   - [ ] Create budget for category
   - [ ] View budget status with progress
   - [ ] Budget calculations are correct
   - [ ] Cannot create duplicate budget for same category/period

5. **Alerts**
   - [ ] Alerts appear when threshold reached
   - [ ] Unread count badge shows correctly
   - [ ] Mark alert as read works
   - [ ] Filter alerts by read status

6. **General**
   - [ ] All screens handle loading states
   - [ ] Error messages are user-friendly
   - [ ] Navigation flows work smoothly
   - [ ] App doesn't crash on network errors
   - [ ] Pull-to-refresh works where implemented

### Build & Deployment

1. **Android Build**
   - Update app name and icon
   - Configure signing keys
   - Build APK/AAB for testing
   - Test on real device

2. **iOS Build** (if applicable)
   - Update app name and icon
   - Configure provisioning profiles
   - Build IPA for testing
   - Test on real device

3. **Environment Configuration**
   - Store API base URL in environment config
   - Don't hardcode sensitive data

---

## Additional Notes

### Error Handling Pattern
```javascript
try {
  setLoading(true);
  const response = await someApiCall();
  // Handle success
  setData(response.data);
} catch (error) {
  // Handle error
  const errorMessage = error.response?.data?.message || 'Something went wrong';
  Alert.alert('Error', errorMessage);
} finally {
  setLoading(false);
}
```

### Date Formatting Helper
```javascript
export const formatDate = (dateString) => {
  const date = new Date(dateString);
  return date.toLocaleDateString('id-ID', { 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric' 
  });
};
```

### Currency Formatting Helper
```javascript
export const formatCurrency = (amount) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(amount);
};
```

### Implementation Order Recommendation
1. Section 1 (Auth) - Foundation
2. Section 2 (Categories) - Simple CRUD
3. Section 3 (Transactions) - Core feature
4. Section 4 (Budgets) - Complex feature
5. Section 5 (Alerts) - Notifications
6. Section 6 (Navigation & Polish) - Final touches
7. Section 7 (Testing) - Quality assurance

---

## Success Criteria

Your mobile app is complete when:
- Users can register and login
- Users can create and manage categories
- Users can add income/expense transactions
- Users can set budgets for categories
- Users receive alerts when budgets reach thresholds
- All screens have proper loading and error states
- Navigation is intuitive and smooth
- The app works offline with cached token
- The app reconnects to API when network returns

Good luck building! Implement section by section, test each section before moving to the next.
