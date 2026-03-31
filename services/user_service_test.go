package services

import (
	"errors"
	"testing"

	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
)

// --- Mock UserRepository ---

type mockUserRepository struct {
	users       map[uint]*models.User
	nextID      uint
	createError error
	updateError error
	deleteError error
}

func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{
		users:  make(map[uint]*models.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) FindAll(filter *dto.UserFilterRequest) ([]models.User, int64, error) {
	var result []models.User
	for _, u := range m.users {
		result = append(result, *u)
	}
	return result, int64(len(result)), nil
}

func (m *mockUserRepository) FindByID(id uint) (*models.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepository) FindByEmail(email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepository) Create(user *models.User) error {
	if m.createError != nil {
		return m.createError
	}
	user.ID = m.nextID
	m.nextID++
	copy := *user
	m.users[copy.ID] = &copy
	return nil
}

func (m *mockUserRepository) Update(user *models.User) error {
	if m.updateError != nil {
		return m.updateError
	}
	copy := *user
	m.users[copy.ID] = &copy
	return nil
}

func (m *mockUserRepository) Delete(id uint) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.users, id)
	return nil
}

// --- Tests ---

func TestUserService_CreateUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	req := &dto.CreateUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Address:  "123 Main St",
		Password: "password123",
	}

	user, err := svc.CreateUser(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("Expected user response, got nil")
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", user.Email)
	}
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	req := &dto.CreateUserRequest{
		Name:     "User One",
		Email:    "dup@example.com",
		Password: "pass123",
	}
	if _, err := svc.CreateUser(req); err != nil {
		t.Fatalf("First creation failed: %v", err)
	}

	_, err := svc.CreateUser(req)
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
	if err.Error() != "user with this email already exists" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestUserService_CreateUser_RepoError(t *testing.T) {
	repo := newMockUserRepo()
	repo.createError = errors.New("db error")
	svc := NewUserService(repo)

	_, err := svc.CreateUser(&dto.CreateUserRequest{
		Name:     "Test",
		Email:    "test@example.com",
		Password: "pass123",
	})
	if err == nil {
		t.Error("Expected error from repository")
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, _ := svc.CreateUser(&dto.CreateUserRequest{
		Name:     "Jane",
		Email:    "jane@example.com",
		Password: "pass123",
	})

	user, err := svc.GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.Email != "jane@example.com" {
		t.Errorf("Expected email 'jane@example.com', got '%s'", user.Email)
	}
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.GetUserByID(999)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
	if err.Error() != "user not found" {
		t.Errorf("Expected 'user not found', got '%v'", err)
	}
}

func TestUserService_GetAllUsers(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	svc.CreateUser(&dto.CreateUserRequest{Name: "U1", Email: "u1@example.com", Password: "pass"})
	svc.CreateUser(&dto.CreateUserRequest{Name: "U2", Email: "u2@example.com", Password: "pass"})

	result, err := svc.GetAllUsers(&dto.UserFilterRequest{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalItems != 2 {
		t.Errorf("Expected 2 total items, got %d", result.TotalItems)
	}
}

func TestUserService_GetAllUsers_SetsDefaults(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	filter := &dto.UserFilterRequest{}
	result, err := svc.GetAllUsers(filter)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Pagination defaults should be applied.
	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}
	if result.PageSize != 10 {
		t.Errorf("Expected page_size 10, got %d", result.PageSize)
	}
}

func TestUserService_UpdateUser_Name(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, _ := svc.CreateUser(&dto.CreateUserRequest{
		Name:     "OldName",
		Email:    "old@example.com",
		Password: "pass123",
	})

	updated, err := svc.UpdateUser(created.ID, &dto.UpdateUserRequest{Name: "NewName"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if updated.Name != "NewName" {
		t.Errorf("Expected name 'NewName', got '%s'", updated.Name)
	}
}

func TestUserService_UpdateUser_Address(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, _ := svc.CreateUser(&dto.CreateUserRequest{
		Name:     "User",
		Email:    "user@example.com",
		Password: "pass",
	})

	updated, err := svc.UpdateUser(created.ID, &dto.UpdateUserRequest{Address: "New Address"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if updated.Address != "New Address" {
		t.Errorf("Expected address 'New Address', got '%s'", updated.Address)
	}
}

func TestUserService_UpdateUser_Email(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, _ := svc.CreateUser(&dto.CreateUserRequest{
		Name:     "User",
		Email:    "before@example.com",
		Password: "pass",
	})

	updated, err := svc.UpdateUser(created.ID, &dto.UpdateUserRequest{Email: "after@example.com"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if updated.Email != "after@example.com" {
		t.Errorf("Expected email 'after@example.com', got '%s'", updated.Email)
	}
}

func TestUserService_UpdateUser_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	u1, _ := svc.CreateUser(&dto.CreateUserRequest{Name: "U1", Email: "u1@example.com", Password: "pass"})
	svc.CreateUser(&dto.CreateUserRequest{Name: "U2", Email: "u2@example.com", Password: "pass"})

	_, err := svc.UpdateUser(u1.ID, &dto.UpdateUserRequest{Email: "u2@example.com"})
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
	if err.Error() != "email already in use" {
		t.Errorf("Expected 'email already in use', got '%v'", err)
	}
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.UpdateUser(999, &dto.UpdateUserRequest{Name: "X"})
	if err == nil {
		t.Error("Expected error for missing user")
	}
	if err.Error() != "user not found" {
		t.Errorf("Expected 'user not found', got '%v'", err)
	}
}

func TestUserService_UpdateUser_SameEmail(t *testing.T) {
	// Updating with the same email should succeed without conflict.
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, _ := svc.CreateUser(&dto.CreateUserRequest{
		Name:     "Same",
		Email:    "same@example.com",
		Password: "pass",
	})

	_, err := svc.UpdateUser(created.ID, &dto.UpdateUserRequest{Email: "same@example.com"})
	if err != nil {
		t.Errorf("Updating with same email should not fail: %v", err)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, _ := svc.CreateUser(&dto.CreateUserRequest{
		Name:     "ToDelete",
		Email:    "delete@example.com",
		Password: "pass123",
	})

	if err := svc.DeleteUser(created.ID); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// User should no longer be found.
	_, err := svc.GetUserByID(created.ID)
	if err == nil {
		t.Error("Expected error for deleted user")
	}
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	err := svc.DeleteUser(999)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
	if err.Error() != "user not found" {
		t.Errorf("Expected 'user not found', got '%v'", err)
	}
}
