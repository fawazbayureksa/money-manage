# Clean Architecture Implementation

This project now follows Clean Architecture principles similar to Laravel's structure.

## 📁 Folder Structure

```
my-api/
├── dto/                    # Data Transfer Objects (like Laravel's Requests/Resources)
│   ├── pagination.go      # Pagination helpers
│   ├── user_dto.go        # User DTOs
│   └── bank_dto.go        # Bank DTOs
├── repositories/          # Database Operations (like Laravel's Repositories)
│   ├── user_repository.go
│   └── bank_repository.go
├── services/              # Business Logic (like Laravel's Services)
│   ├── user_service.go
│   └── bank_service.go
├── controllers/           # HTTP Handlers (Thin controllers)
│   ├── user_controller.go
│   └── bank_controller.go
├── models/                # Database Models
├── middleware/            # HTTP Middleware
├── routes/                # Route definitions with DI
└── config/                # Configuration
```

## 🏗️ Architecture Layers

### 1. **DTO Layer** (Data Transfer Objects)
- Request validation structures
- Response structures
- Separates API contracts from database models
- Never exposes sensitive data (like passwords)

### 2. **Repository Layer**
- **Responsibility**: Database operations only
- Direct interaction with GORM
- Returns models
- Example:
  ```go
  type UserRepository interface {
      FindAll(filter *dto.UserFilterRequest) ([]models.User, int64, error)
      FindByID(id uint) (*models.User, error)
      Create(user *models.User) error
  }
  ```

### 3. **Service Layer**
- **Responsibility**: Business logic
- Uses repositories to access data
- Validates business rules
- Converts models to DTOs
- Example:
  ```go
  type UserService interface {
      GetAllUsers(filter *dto.UserFilterRequest) (*dto.PaginationResponse, error)
      CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error)
  }
  ```

### 4. **Controller Layer**
- **Responsibility**: HTTP handling only
- Thin controllers
- Parse request parameters
- Call services
- Return HTTP responses
- Example:
  ```go
  func (ctrl *UserController) GetUsers(c *gin.Context) {
      var filter dto.UserFilterRequest
      c.ShouldBindQuery(&filter)
      
      result, err := ctrl.service.GetAllUsers(&filter)
      utils.JSONSuccess(c, "Users retrieved", result)
  }
  ```

## 🔍 API Usage Examples

### **Get Users with Pagination**

```bash
# Basic pagination
GET /api/users?page=1&page_size=10

# With search
GET /api/users?search=john&page=1&page_size=10

# With filtering
GET /api/users?name=john&is_admin=true&page=1

# With sorting
GET /api/users?sort_by=name&sort_dir=asc&page=1
```

**Response:**
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": {
    "data": [
      {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com",
        "address": "123 Main St",
        "is_verified": false,
        "is_admin": false
      }
    ],
    "page": 1,
    "page_size": 10,
    "total_items": 25,
    "total_pages": 3
  }
}
```

### **Get Banks with Filtering**

```bash
# Basic pagination
GET /api/banks?page=1&page_size=10

# Search by bank name
GET /api/banks?search=mandiri&page=1

# Filter by specific bank name
GET /api/banks?bank_name=BCA&page=1

# Filter by color
GET /api/banks?color=blue&page=1

# With sorting
GET /api/banks?sort_by=bank_name&sort_dir=asc
```

### **Query Parameters**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `page` | int | Page number | 1 |
| `page_size` | int | Items per page (max 100) | 10 |
| `search` | string | Global search | - |
| `sort_by` | string | Field to sort by | id |
| `sort_dir` | string | Sort direction (asc/desc) | desc |

#### User-specific filters:
- `name`: Filter by user name
- `email`: Filter by email
- `is_admin`: Filter by admin status (true/false)

#### Bank-specific filters:
- `bank_name`: Filter by bank name
- `color`: Filter by color

## 🚀 Benefits of This Architecture

### ✅ **Separation of Concerns**
- Each layer has a single responsibility
- Easy to understand and maintain
- Similar to Laravel's architecture

### ✅ **Testability**
- Easy to mock repositories in service tests
- Easy to mock services in controller tests
- Each layer can be tested independently

### ✅ **Reusability**
- Services can be reused across different controllers
- Repositories can be reused across different services

### ✅ **Security**
- DTOs prevent password exposure
- Input validation at DTO level
- Clear separation of internal/external data

### ✅ **Flexibility**
- Easy to swap database implementations
- Easy to add new business rules
- Easy to change API responses without affecting database

## 🔄 Adding New Features

### Step 1: Create DTOs
```go
// dto/new_entity_dto.go
type NewEntityResponse struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

type CreateNewEntityRequest struct {
    Name string `json:"name" binding:"required"`
}
```

### Step 2: Create Repository
```go
// repositories/new_entity_repository.go
type NewEntityRepository interface {
    FindAll() ([]models.NewEntity, error)
    Create(entity *models.NewEntity) error
}
```

### Step 3: Create Service
```go
// services/new_entity_service.go
type NewEntityService interface {
    GetAll() ([]dto.NewEntityResponse, error)
    Create(req *dto.CreateNewEntityRequest) (*dto.NewEntityResponse, error)
}
```

### Step 4: Create Controller
```go
// controllers/new_entity_controller.go
type NewEntityController struct {
    service services.NewEntityService
}

func (ctrl *NewEntityController) GetAll(c *gin.Context) {
    // Handle HTTP request
}
```

### Step 5: Wire it up in Routes
```go
// routes/routes.go
repo := repositories.NewNewEntityRepository(config.DB)
service := services.NewNewEntityService(repo)
controller := controllers.NewNewEntityController(service)

api.GET("/entities", controller.GetAll)
```

## 📝 Code Quality Improvements

### ✅ Implemented
- [x] Repository pattern for database operations
- [x] Service layer for business logic
- [x] DTOs for request/response
- [x] Dependency injection in routes
- [x] Pagination support
- [x] Filtering support
- [x] Sorting support
- [x] Password never exposed in responses
- [x] Fixed unexported field bug (isVerified → IsVerified)
- [x] Proper error handling
- [x] Input validation

### 🎯 Next Steps
- [ ] Add unit tests for services
- [ ] Add integration tests for controllers
- [ ] Add JWT secret to environment variables
- [ ] Add rate limiting middleware
- [ ] Add request logging
- [ ] Add database transactions for complex operations
- [ ] Add soft deletes
- [ ] Add database indexes
- [ ] Add API documentation (Swagger)

## 🎓 Comparison with Laravel

| Laravel | Go (This Project) |
|---------|-------------------|
| Request Classes | DTOs |
| Resource Classes | DTOs |
| Repositories | Repositories |
| Services | Services |
| Controllers | Controllers |
| Routes with DI | Routes with DI |
| Eloquent ORM | GORM |
| Validation Rules | Binding Tags |

This architecture makes Go development feel familiar to Laravel developers while maintaining Go's performance and type safety! 🚀
