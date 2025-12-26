# 🚀 Money Management API - Clean Architecture

A professional REST API for personal finance management built with Go, featuring clean architecture similar to Laravel's structure.

## 📋 Project Overview

This is a **portfolio-ready** Go project demonstrating:
- Clean Architecture (Repository + Service + Controller patterns)
- RESTful API design
- JWT Authentication
- Pagination & Filtering
- Security best practices
- Production-ready code structure

## 🛠️ Tech Stack

- **Language**: Go 1.24
- **Framework**: Gin (HTTP router)
- **ORM**: GORM
- **Database**: MySQL
- **Auth**: JWT (JSON Web Tokens)
- **Password**: bcrypt hashing

## 🏗️ Architecture

```
┌──────────────┐
│  Controller  │  ← HTTP handling only
└──────┬───────┘
       ↓
┌──────────────┐
│   Service    │  ← Business logic
└──────┬───────┘
       ↓
┌──────────────┐
│  Repository  │  ← Database operations
└──────┬───────┘
       ↓
┌──────────────┐
│   Database   │
└──────────────┘
```

**Similar to Laravel:**
- DTOs = Request/Resource Classes
- Services = Service Layer
- Repositories = Repository Pattern
- Controllers = Thin Controllers

## 📁 Folder Structure

```
my-api/
├── dto/                    # Data Transfer Objects
│   ├── pagination.go
│   ├── user_dto.go
│   └── bank_dto.go
├── repositories/          # Database Operations
│   ├── user_repository.go
│   └── bank_repository.go
├── services/              # Business Logic
│   ├── user_service.go
│   └── bank_service.go
├── controllers/           # HTTP Handlers
│   ├── user_controller.go
│   ├── bank_controller.go
│   ├── auth_controller.go
│   ├── category_controller.go
│   └── transaction_controller.go
├── models/                # Database Models
│   ├── user.go
│   ├── bank.go
│   ├── category.go
│   └── transaction.go
├── middleware/            # HTTP Middleware
│   └── auth_middleware.go
├── routes/                # Route Definitions
│   └── routes.go
├── config/                # Configuration
│   └── database.go
├── utils/                 # Utilities
│   ├── response.go
│   ├── hash.go
│   └── genere_token.go
└── db/migrations/         # Database Migrations
```

## 🚀 Quick Start

### 1. Clone & Install
```bash
git clone <repository-url>
cd my-api
go mod download
```

### 2. Configure Environment
Create `.env` file:
```env
DB_NAME=your_database
DB_USER=your_username
DB_PASS=your_password
DB_HOST=localhost
DB_PORT=3306
```

### 3. Run Migrations
```bash
# Run migrations using your preferred tool
# Or let AutoMigrate handle it on first run
```

### 4. Start Server
```bash
go run main.go
# Server runs on http://localhost:8080
```

## 🔌 API Endpoints

### Authentication
```
POST   /api/register      Register new user
POST   /api/login         Login and get JWT token
```

### Users (with pagination & filtering)
```
GET    /api/users         Get all users
POST   /api/users         Create user
PUT    /api/users/:id     Update user
DELETE /api/users/:id     Delete user
```

### Banks (with pagination & filtering)
```
GET    /api/banks         Get all banks
POST   /api/banks         Create bank
DELETE /api/banks/:id     Delete bank
```

### Categories
```
GET    /api/categories           Get all categories
POST   /api/categories           Create category (protected)
GET    /api/my-categories        Get user's categories (protected)
DELETE /api/categories/:id       Delete category
```

### Transactions
```
POST   /api/transaction          Create transaction (protected)
GET    /api/transaction/initial-data  Get initial data
```

### Budgets (protected)
```
POST   /api/budgets              Create budget
GET    /api/budgets              Get all budgets (paginated, filterable)
GET    /api/budgets/:id          Get budget with spending data
PUT    /api/budgets/:id          Update budget
DELETE /api/budgets/:id          Delete budget
GET    /api/budgets/status       Get active budgets status
GET    /api/budget-alerts        Get budget alerts
PUT    /api/budget-alerts/:id/read  Mark alert as read
```

### Analytics (protected)
```
GET    /api/analytics/dashboard             Complete dashboard summary
GET    /api/analytics/spending-by-category  Category breakdown
GET    /api/analytics/spending-by-bank      Bank usage analysis
GET    /api/analytics/income-vs-expense     Income vs Expense
GET    /api/analytics/trend                 Trend analysis
GET    /api/analytics/monthly-comparison    Month-to-month comparison
GET    /api/analytics/yearly-report         Annual financial report
GET    /api/analytics/category-trend/:id    Category trends
```

## 📊 API Features

### Pagination
All list endpoints support pagination:
```bash
GET /api/users?page=1&page_size=10
```

**Parameters:**
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 10, max: 100)

### Filtering
Filter by specific fields:
```bash
# Users
GET /api/users?name=john&is_admin=true

# Banks
GET /api/banks?bank_name=mandiri&color=blue
```

### Search
Global search across multiple fields:
```bash
GET /api/users?search=john
GET /api/banks?search=mandiri
```

### Sorting
Sort by any field:
```bash
GET /api/users?sort_by=name&sort_dir=asc
GET /api/banks?sort_by=bank_name&sort_dir=desc
```

### Combined Example
```bash
GET /api/users?search=john&is_admin=true&sort_by=email&sort_dir=asc&page=2&page_size=20
```

## 📝 Response Format

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

## 🔒 Security Features

- ✅ Password hashing with bcrypt
- ✅ JWT authentication for protected routes
- ✅ Passwords never exposed in API responses
- ✅ Input validation at DTO level
- ✅ CORS support

## 🎯 Key Features

### Clean Architecture
- **Repository Pattern**: Database operations isolated
- **Service Layer**: Business logic separated
- **Thin Controllers**: HTTP handling only
- **DTOs**: Request/response validation

### Code Quality
- ✅ Dependency Injection
- ✅ Interface-based design
- ✅ Single Responsibility Principle
- ✅ Easy to test and mock
- ✅ Reusable components

### API Features
- ✅ Pagination support
- ✅ Dynamic filtering
- ✅ Flexible sorting
- ✅ Global search
- ✅ Consistent error handling
- ✅ Proper HTTP status codes

## 🧪 Testing

```bash
# Example: Testing with curl
curl -X GET "http://localhost:8080/api/users?page=1&page_size=5"

curl -X POST "http://localhost:8080/api/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com","password":"secret123"}'
```

See [API_TESTING.md](API_TESTING.md) for more examples.

## 📚 Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Complete architecture guide
- **[API_TESTING.md](API_TESTING.md)** - API testing examples
- **[BEFORE_AFTER_COMPARISON.md](BEFORE_AFTER_COMPARISON.md)** - Code comparison
- **[ARCHITECTURE_DIAGRAM.md](ARCHITECTURE_DIAGRAM.md)** - Visual diagrams
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** - Implementation details

## 🎓 What Makes This Portfolio-Ready?

### Professional Architecture ✅
- Clean separation of concerns
- Similar to industry-standard Laravel structure
- Easy to understand and maintain

### Best Practices ✅
- Repository pattern for data access
- Service layer for business logic
- DTOs for input/output validation
- Dependency injection

### Production Features ✅
- Pagination for large datasets
- Filtering and search functionality
- Proper error handling
- Security considerations

### Code Quality ✅
- Consistent code structure
- Self-documenting code
- Easy to test
- Scalable design

## 🚧 Future Enhancements

- [ ] Unit tests (services)
- [ ] Integration tests (controllers)
- [ ] Swagger documentation
- [ ] Rate limiting
- [ ] Request logging
- [ ] Database transactions
- [ ] Soft deletes
- [ ] Database indexes
- [ ] Caching layer (Redis)
- [ ] Docker support

## 🤝 Contributing

This is a portfolio project. Feel free to fork and use it as a reference for your own projects!

## 📄 License

MIT License

## 👤 Author

**Fawwaz Bayureksa**
- GitHub: [@fawazbayureksa](https://github.com/fawazbayureksa)

---

## 📖 Learning Resources

This project demonstrates:
- Go web development with Gin
- Clean Architecture in Go
- RESTful API design
- JWT authentication
- GORM for database operations
- Pagination & filtering implementation
- Repository and Service patterns


## Advanced Features Implemented

### Budget Management
- Create and manage monthly/yearly budgets per category
- Real-time spending tracking against budgets
- Automatic alerts when reaching thresholds
- Budget status monitoring (safe/warning/exceeded)
- Multiple budgets per user with pagination

### Analytics & Reports
- Comprehensive dashboard with key financial metrics
- Spending analysis by category and bank
- Income vs Expense tracking with savings rate
- Trend analysis (daily/weekly/monthly/yearly)
- Month-to-month comparison with percentage changes
- Yearly financial reports
- Category-specific spending trends

See ADVANCED_FEATURES.md for complete documentation.

## Next Steps:
- Fix security issues (JWT secret in env, rate limiting)
- Write unit and integration tests
- Add Swagger/OpenAPI documentation
- Implement soft deletes
- Add database indexes for performance
- Docker containerization
- CI/CD pipeline setup