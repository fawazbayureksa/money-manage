# API Testing Examples

## Test the Clean Architecture Implementation

### 1. Get Users with Pagination (Default)
```bash
curl -X GET "http://localhost:8080/api/users?page=1&page_size=5"
```

### 2. Get Users with Search
```bash
curl -X GET "http://localhost:8080/api/users?search=john&page=1&page_size=10"
```

### 3. Get Users with Filtering
```bash
# Filter by admin status
curl -X GET "http://localhost:8080/api/users?is_admin=true&page=1"

# Filter by name
curl -X GET "http://localhost:8080/api/users?name=john&page=1"
```

### 4. Get Users with Sorting
```bash
# Sort by name ascending
curl -X GET "http://localhost:8080/api/users?sort_by=name&sort_dir=asc&page=1"

# Sort by email descending
curl -X GET "http://localhost:8080/api/users?sort_by=email&sort_dir=desc&page=1"
```

### 5. Create User (with validation)
```bash
curl -X POST "http://localhost:8080/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "address": "123 Main St",
    "password": "password123"
  }'
```

### 6. Get Banks with Pagination
```bash
curl -X GET "http://localhost:8080/api/banks?page=1&page_size=10"
```

### 7. Get Banks with Search
```bash
curl -X GET "http://localhost:8080/api/banks?search=mandiri&page=1"
```

### 8. Get Banks with Filtering
```bash
# Filter by bank name
curl -X GET "http://localhost:8080/api/banks?bank_name=BCA&page=1"

# Filter by color
curl -X GET "http://localhost:8080/api/banks?color=blue&page=1"
```

### 9. Create Bank
```bash
curl -X POST "http://localhost:8080/api/banks" \
  -H "Content-Type: application/json" \
  -d '{
    "bank_name": "Bank Mandiri",
    "color": "#003399",
    "image": "mandiri.png"
  }'
```

### 10. Advanced: Multiple Filters + Pagination + Sorting
```bash
# Get users with name containing "john", who are admins, sorted by email
curl -X GET "http://localhost:8080/api/users?name=john&is_admin=true&sort_by=email&sort_dir=asc&page=1&page_size=20"
```

## Expected Response Format

All endpoints now return a consistent format:

### Success Response
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": {
    "data": [...],
    "page": 1,
    "page_size": 10,
    "total_items": 25,
    "total_pages": 3
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Invalid input data",
  "data": null
}
```

## Key Improvements

✅ **Clean Architecture**: Repository → Service → Controller
✅ **Pagination**: All list endpoints support pagination
✅ **Filtering**: Multiple filter options per entity
✅ **Sorting**: Sort by any field, ascending or descending
✅ **Search**: Global search across multiple fields
✅ **Security**: Passwords never exposed in responses
✅ **Validation**: Request validation at DTO level
✅ **Type Safety**: Proper error handling and type conversion

## Using Postman/Thunder Client

Import these as environment variables:
- `base_url`: http://localhost:8080
- `api_prefix`: /api

Then test endpoints like:
- `{{base_url}}{{api_prefix}}/users?page=1&page_size=10`
- `{{base_url}}{{api_prefix}}/banks?search=mandiri`
