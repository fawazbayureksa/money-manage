# Tags & Labels API Quick Reference

## Base URL
All endpoints are prefixed with `/api/v2/` and require authentication token.

---

## Tag Management Endpoints

### 1. Create Tag
```http
POST /api/v2/tags
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Date Night",
  "color": "#EC4899",
  "icon": "❤️"
}
```

### 2. List All Tags
```http
GET /api/v2/tags?sort=usage
Authorization: Bearer {token}
```
- `sort` parameter: `usage` (by popularity) or `name` (alphabetical, default)

### 3. Get Tag by ID
```http
GET /api/v2/tags/:id
Authorization: Bearer {token}
```

### 4. Update Tag
```http
PUT /api/v2/tags/:id
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Beach Vacation 2026",
  "color": "#06B6D4",
  "icon": "🏖️"
}
```

### 5. Delete Tag
```http
DELETE /api/v2/tags/:id
Authorization: Bearer {token}
```

---

## Transaction with Tags

### 1. Create Transaction with Tags
```http
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

### 2. Add Tags to Existing Transaction
```http
POST /api/v2/transactions/:id/tags
Authorization: Bearer {token}
Content-Type: application/json

{
  "tag_ids": [1, 3]
}
```

### 3. Remove Tag from Transaction
```http
DELETE /api/v2/transactions/:id/tags/:tag_id
Authorization: Bearer {token}
```

---

## Tag Suggestions

### Get Tag Suggestions
```http
GET /api/v2/tags/suggest?category_id=3&description=dinner
Authorization: Bearer {token}
```

**Parameters:**
- `category_id` (required): Category ID for the transaction
- `description` (optional): Transaction description for keyword matching

---

## Analytics

### Get Spending by Tag
```http
GET /api/v2/analytics/spending-by-tag?start_date=2026-01-01&end_date=2026-02-28
Authorization: Bearer {token}
```

**Parameters:**
- `start_date` (required): Start date in YYYY-MM-DD format
- `end_date` (required): End date in YYYY-MM-DD format

---

## Testing Sequence

### 1. Basic Tag Operations
```bash
# 1. Create a tag
curl -X POST http://localhost:8080/api/v2/tags \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Date Night","color":"#EC4899","icon":"❤️"}'

# 2. Get all tags
curl -X GET http://localhost:8080/api/v2/tags \
  -H "Authorization: Bearer YOUR_TOKEN"

# 3. Get tags sorted by usage
curl -X GET "http://localhost:8080/api/v2/tags?sort=usage" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 2. Transaction with Tags
```bash
# 1. Create transaction with tags
curl -X POST http://localhost:8080/api/v2/transactions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Dinner at Restaurant",
    "amount": 500000,
    "transaction_type": "Expense",
    "category_id": 3,
    "asset_id": 1,
    "date": "2026-02-12",
    "tag_ids": [1]
  }'

# 2. Get transaction to verify tags
curl -X GET http://localhost:8080/api/v2/transactions/TRANSACTION_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# 3. Add more tags to transaction
curl -X POST http://localhost:8080/api/v2/transactions/TRANSACTION_ID/tags \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tag_ids": [2, 3]}'
```

### 3. Tag Suggestions
```bash
# Get tag suggestions for a category
curl -X GET "http://localhost:8080/api/v2/tags/suggest?category_id=3&description=dinner" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. Analytics
```bash
# Get spending by tag
curl -X GET "http://localhost:8080/api/v2/analytics/spending-by-tag?start_date=2026-01-01&end_date=2026-02-28" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Response Formats

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error description",
  "data": null
}
```

---

## Tag Object Structure
```json
{
  "id": 1,
  "user_id": 1,
  "name": "Date Night",
  "color": "#EC4899",
  "icon": "❤️",
  "usage_count": 12,
  "created_at": "2026-02-12T10:30:00Z",
  "updated_at": "2026-02-12T10:30:00Z"
}
```

---

## Transaction Response with Tags
```json
{
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
    }
  ]
}
```

---

## Common HTTP Status Codes
- `200 OK`: Successful GET, PUT, DELETE
- `201 Created`: Successful POST (resource created)
- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: User doesn't have permission
- `404 Not Found`: Resource not found
