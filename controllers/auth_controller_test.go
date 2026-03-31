package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"my-api/dto"
	"my-api/utils"
)

// mockAuthUserService wraps mockUserService and adds only the Register path.
// The Login endpoint uses config.DB directly so it cannot be unit-tested here.
type mockAuthUserService struct {
	createErr error
	created   *dto.UserResponse
}

func (m *mockAuthUserService) GetAllUsers(f *dto.UserFilterRequest) (*dto.PaginationResponse, error) {
	return nil, nil
}
func (m *mockAuthUserService) GetUserByID(id uint) (*dto.UserResponse, error) { return nil, nil }
func (m *mockAuthUserService) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.created != nil {
		return m.created, nil
	}
	return &dto.UserResponse{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
	}, nil
}
func (m *mockAuthUserService) UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	return nil, nil
}
func (m *mockAuthUserService) DeleteUser(id uint) error { return nil }

func setupAuthRouter(svc *mockAuthUserService) *gin.Engine {
	r := gin.New()
	ctrl := NewAuthController(svc)
	r.POST("/register", ctrl.Register)
	// Logout requires user_id in context (set by auth middleware in production).
	r.POST("/logout", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		ctrl.Logout(c)
	})
	return r
}

// --- Tests ---

func TestAuthController_Register_Success(t *testing.T) {
	svc := &mockAuthUserService{}
	r := setupAuthRouter(svc)

	body := toJSON(map[string]string{
		"name":     "Alice",
		"email":    "alice@test.com",
		"password": "secret123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp utils.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Error("Expected success response")
	}
}

func TestAuthController_Register_InvalidBody(t *testing.T) {
	svc := &mockAuthUserService{}
	r := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestAuthController_Register_MissingFields(t *testing.T) {
	svc := &mockAuthUserService{}
	r := setupAuthRouter(svc)

	// Missing name and password.
	body := toJSON(map[string]string{"email": "x@example.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing fields, got %d", w.Code)
	}
}

func TestAuthController_Register_EmailConflict(t *testing.T) {
	svc := &mockAuthUserService{
		createErr: errors.New("user with this email already exists"),
	}
	r := setupAuthRouter(svc)

	body := toJSON(map[string]string{
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "secret123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected 409, got %d", w.Code)
	}
}

func TestAuthController_Register_InternalError(t *testing.T) {
	svc := &mockAuthUserService{
		createErr: errors.New("database failure"),
	}
	r := setupAuthRouter(svc)

	body := toJSON(map[string]string{
		"name":     "Carol",
		"email":    "carol@example.com",
		"password": "secret123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestAuthController_Logout_Success(t *testing.T) {
	svc := &mockAuthUserService{}
	r := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/logout", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp utils.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Error("Expected success response from logout")
	}
}
