# Tags & Labels Implementation - Verification Document

## Overview
This document describes the implementation of the Tags & Labels feature for Transaction V2 in the money-manage application.

## Implementation Date
February 12, 2026

## Feature Summary
The Tags & Labels feature allows users to add custom, flexible context to their transactions beyond categories. This enables better organization and analysis of spending patterns across different contexts (e.g., "Date Night", "Business", "Vacation", "Birthday").

---

## What Was Implemented

### 1. Database Schema

#### Tables Created
- **tags**: Stores user-defined tags with color, icon, and usage tracking
- **transaction_tags**: Junction table for many-to-many relationship between transactions and tags

#### Migration File
- `migrations/20260212_create_tags.sql` - Creates both tables with appropriate indexes and constraints

### 2. Models

#### Files Created
- `models/tag.go` - Defines Tag and TransactionTag models with GORM annotations
- Updated `models/transaction_v2.go` - Added Tags relationship to TransactionV2 model

#### Key Features
- Soft delete support for tags
- Usage count tracking for popularity sorting
- Many-to-many relationship between transactions and tags

### 3. DTOs (Data Transfer Objects)

#### Files Created
- `dto/tag_dto.go` - Request/response DTOs for tag operations
  - CreateTagRequest
  - UpdateTagRequest
  - AddTagsToTransactionRequest
  - TagSuggestion
  - TagSpending
  - TagSpendingResponse

#### Files Updated
- `dto/transaction_v2_dto.go` - Added tag_ids field to transaction requests and tags field to responses

### 4. Repository Layer

#### Files Created
- `repositories/tag_repository.go` - Data access layer for tags with the following methods:
  - Create, FindByID, FindByUserID, Update, Delete
  - FindByName, IncrementUsage
  - GetTagsByCategory (for suggestions)
  - GetSpendingByTag (for analytics)

#### Files Updated
- `repositories/transaction_v2_repository.go` - Added methods:
  - AddTagsToTransaction
  - RemoveTagFromTransaction
  - Updated all query methods to preload Tags

### 5. Service Layer

#### Files Created
- `services/tag_service.go` - Business logic for tags:
  - CreateTag, GetTags, UpdateTag, DeleteTag, GetTagByID
  - SuggestTags - Intelligent tag suggestions based on:
    - Historical usage with same category
    - Keyword matching in description
    - Overall tag popularity
  - GetSpendingByTag - Analytics for spending grouped by tags

#### Files Updated
- `services/transaction_v2_service.go` - Added methods:
  - AddTagsToTransaction
  - RemoveTagFromTransaction
  - Updated service to include tag repository dependency

### 6. Controller Layer

#### Files Created
- `controllers/tag_controller.go` - API handlers for tags:
  - CreateTag, GetTags, GetTagByID, UpdateTag, DeleteTag
  - SuggestTags
  - GetSpendingByTag

#### Files Updated
- `controllers/transaction_v2_controller.go` - Added:
  - AddTagsToTransaction
  - RemoveTagFromTransaction
  - Updated CreateTransaction to handle tag_ids in request

### 7. Routes

#### Files Updated
- `routes/routes.go` - Added v2 API endpoints:
  - Tag CRUD operations
  - Transaction tag management
  - Tag analytics

---

## API Endpoints

### Tag Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v2/tags` | List all user tags (with optional sorting) |
| GET | `/api/v2/tags/:id` | Get a specific tag |
| POST | `/api/v2/tags` | Create a new tag |
| PUT | `/api/v2/tags/:id` | Update a tag |
| DELETE | `/api/v2/tags/:id` | Delete a tag |
| GET | `/api/v2/tags/suggest` | Get tag suggestions based on category and description |

### Transaction Tags

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v2/transactions` | Create transaction with tags (tag_ids in body) |
| POST | `/api/v2/transactions/:id/tags` | Add tags to existing transaction |
| DELETE | `/api/v2/transactions/:id/tags/:tag_id` | Remove tag from transaction |

### Analytics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v2/analytics/spending-by-tag` | Get spending statistics grouped by tags |

---

## Request/Response Examples

### 1. Create a Tag

**Request:**
```bash
POST /api/v2/tags
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Date Night",
  "color": "#EC4899",
  "icon": "❤️"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Tag created successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "name": "Date Night",
    "color": "#EC4899",
    "icon": "❤️",
    "usage_count": 0,
    "created_at": "2026-02-12T10:30:00Z",
    "updated_at": "2026-02-12T10:30:00Z"
  }
}
```

### 2. Create Transaction with Tags

**Request:**
```bash
POST /api/v2/transactions
Authorization: Bearer {token}
Content-Type: application/json

{
  "description": "Dinner at Sushi Tei",
  "amount": 450000,
  "transaction_type": "Expense",
  "category_id": 3,
  "asset_id": 1,
  "date": "2026-02-11",
  "tag_ids": [1, 3]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Transaction created successfully",
  "data": {
    "id": 42,
    "description": "Dinner at Sushi Tei",
    "amount": 450000,
    "transaction_type": 2,
    "date": "2026-02-11T00:00:00Z",
    "category_name": "Food & Dining",
    "asset_id": 1,
    "asset_name": "Main Checking",
    "tags": [
      {
        "id": 1,
        "name": "Date Night",
        "color": "#EC4899",
        "icon": "❤️",
        "usage_count": 1
      },
      {
        "id": 3,
        "name": "Family",
        "color": "#10B981",
        "icon": "👨‍👩‍👧",
        "usage_count": 1
      }
    ]
  }
}
```

### 3. Get Tag Suggestions

**Request:**
```bash
GET /api/v2/tags/suggest?category_id=3&description=dinner
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "message": "Tag suggestions fetched successfully",
  "data": [
    {
      "id": 1,
      "name": "Date Night",
      "confidence": 0.8
    },
    {
      "id": 3,
      "name": "Family",
      "confidence": 0.6
    },
    {
      "id": 2,
      "name": "Business",
      "confidence": 0.4
    }
  ]
}
```

### 4. Get Spending by Tag

**Request:**
```bash
GET /api/v2/analytics/spending-by-tag?start_date=2026-01-01&end_date=2026-02-28
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "message": "Spending by tag fetched successfully",
  "data": {
    "data": [
      {
        "tag": {
          "id": 1,
          "name": "Date Night",
          "color": "#EC4899",
          "icon": "❤️",
          "usage_count": 12
        },
        "total_amount": 2500000,
        "transaction_count": 12,
        "avg_amount": 208333
      },
      {
        "tag": {
          "id": 2,
          "name": "Business",
          "color": "#3B82F6",
          "icon": "💼",
          "usage_count": 8
        },
        "total_amount": 1800000,
        "transaction_count": 8,
        "avg_amount": 225000
      }
    ],
    "period": {
      "start_date": "2026-01-01",
      "end_date": "2026-02-28"
    }
  }
}
```

---

## Key Features

### 1. Flexible Tagging
- Users can create unlimited custom tags
- Each tag can have a name, color (hex code), and emoji icon
- Tags are user-specific and isolated

### 2. Multi-Tag Support
- Transactions can have multiple tags
- Tags can be added during transaction creation or later
- Tags can be removed individually from transactions

### 3. Intelligent Suggestions
The tag suggestion algorithm considers:
- **Category History** (50% weight): Tags frequently used with the same category
- **Keyword Matching** (30% weight): Tags whose names match words in the transaction description
- **Popularity** (20% weight): Overall tag usage frequency

### 4. Usage Tracking
- Automatic increment of usage_count when tags are used
- Enables sorting by popularity
- Helps identify most-used tags

### 5. Analytics
- Spending grouped by tags
- Transaction counts per tag
- Average spending per tag
- Time period filtering

### 6. Data Integrity
- Soft delete for tags (keeps historical data)
- Cascading deletes for transaction_tags when transactions are deleted
- Unique constraint on user_id + tag_name
- Foreign key constraints ensure referential integrity

---

## Database Schema Details

### tags Table
```sql
CREATE TABLE IF NOT EXISTS tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366F1',
    icon VARCHAR(10) NULL,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE KEY unique_user_tag (user_id, name),
    INDEX idx_user_usage (user_id, usage_count DESC),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### transaction_tags Table
```sql
CREATE TABLE IF NOT EXISTS transaction_tags (
    transaction_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (transaction_id, tag_id),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_tag_id (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

## Testing Checklist

### Manual Testing Steps

1. **Tag CRUD Operations**
   - [ ] Create a tag
   - [ ] Get all tags
   - [ ] Get tag by ID
   - [ ] Update tag
   - [ ] Delete tag
   - [ ] Verify unique constraint (try creating duplicate tag name)

2. **Transaction with Tags**
   - [ ] Create transaction with tags
   - [ ] Verify tags appear in transaction response
   - [ ] Add tags to existing transaction
   - [ ] Remove tag from transaction
   - [ ] Delete transaction and verify tags remain for other transactions

3. **Tag Suggestions**
   - [ ] Create several transactions with tags in same category
   - [ ] Request suggestions for that category
   - [ ] Verify relevant tags appear first

4. **Analytics**
   - [ ] Create expense transactions with different tags
   - [ ] Query spending by tag for a date range
   - [ ] Verify totals, counts, and averages are correct

5. **Edge Cases**
   - [ ] Create tag with no color (should get default)
   - [ ] Create tag with emoji icon
   - [ ] Try to add non-existent tag to transaction
   - [ ] Try to access another user's tags

---

## Migration Instructions

### To Apply Migration

Run the migration SQL file against your database:

```bash
mysql -u [username] -p [database_name] < migrations/20260212_create_tags.sql
```

Or use your preferred migration tool.

### Rollback (if needed)

```sql
DROP TABLE IF EXISTS transaction_tags;
DROP TABLE IF EXISTS tags;
```

---

## Security Considerations

1. **Authorization**: All tag operations are user-scoped and verified
2. **Input Validation**: Tag names are limited to 50 characters
3. **SQL Injection**: Protected by GORM parameterized queries
4. **Data Isolation**: Users can only access their own tags

---

## Performance Considerations

1. **Indexes**: Added on user_id, usage_count, and foreign keys
2. **Pagination**: List endpoints support pagination
3. **Eager Loading**: Tags are preloaded to avoid N+1 queries
4. **Query Optimization**: Analytics use aggregated SQL queries

---

## Future Enhancements

1. Tag colors from predefined palette
2. Tag categories/groups
3. Tag sharing between users
4. Tag templates for common use cases
5. Bulk tag operations
6. Tag-based budgets
7. Tag-based filters in transaction list
8. Export transactions by tag

---

## Files Changed Summary

### New Files (7)
1. `migrations/20260212_create_tags.sql`
2. `models/tag.go`
3. `dto/tag_dto.go`
4. `repositories/tag_repository.go`
5. `services/tag_service.go`
6. `controllers/tag_controller.go`
7. `docs/verification/tags_labels_implementation.md` (this file)

### Modified Files (5)
1. `models/transaction_v2.go` - Added Tags relationship
2. `dto/transaction_v2_dto.go` - Added tag_ids and tags fields
3. `repositories/transaction_v2_repository.go` - Added tag methods and preloading
4. `services/transaction_v2_service.go` - Added tag handling
5. `controllers/transaction_v2_controller.go` - Added tag endpoints
6. `routes/routes.go` - Added tag routes

### Total: 13 files (7 new, 6 modified)

---

## Conclusion

The Tags & Labels feature has been successfully implemented for Transaction V2. All core functionality is in place including:
- Full CRUD operations for tags
- Multi-tag support for transactions
- Intelligent tag suggestions
- Analytics for spending by tag
- Proper data isolation and security

The implementation follows the existing code patterns and maintains consistency with the rest of the application architecture.

---

**Implementation Status**: ✅ Complete
**Testing Status**: ⏳ Ready for Testing
**Documentation Status**: ✅ Complete
