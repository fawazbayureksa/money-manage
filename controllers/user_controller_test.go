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

// --- Mock UserService ---

type mockUserService struct {
	users         map[uint]*dto.UserResponse
	nextID        uint
	createError   error
	updateError   error
	deleteError   error
	getAllResponse *dto.PaginationResponse
}

func newMockUserService() *mockUserService {
	return &mockUserService{
		users:  make(map[uint]*dto.UserResponse),
		nextID: 1,
	}
}

func (m *mockUserService) GetAllUsers(filter *dto.UserFilterRequest) (*dto.PaginationResponse, error) {
	if m.getAllResponse != nil {
		return m.getAllResponse, nil
	}
	var data []dto.UserResponse
	for _, u := range m.users {
		data = append(data, *u)
	}
	return &dto.PaginationResponse{
		Data:       data,
		Page:       1,
		PageSize:   10,
		TotalItems: int64(len(data)),
		TotalPages: 1,
	}, nil
}

func (m *mockUserService) GetUserByID(id uint) (*dto.UserResponse, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserService) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	if m.createError != nil {
		return nil, m.createError
	}
	user := &dto.UserResponse{
		ID:    m.nextID,
		Name:  req.Name,
		Email: req.Email,
	}
	m.nextID++
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserService) UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	if m.updateError != nil {
		return nil, m.updateError
	}
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	return u, nil
}

func (m *mockUserService) DeleteUser(id uint) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, ok := m.users[id]; !ok {
		return errors.New("user not found")
	}
	delete(m.users, id)
	return nil
}

// --- Helpers ---

func init() {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()
}

func setupUserRouter(svc *mockUserService) *gin.Engine {
	r := gin.New()
	ctrl := NewUserController(svc)
	r.GET("/users", ctrl.GetUsers)
	r.POST("/users", ctrl.CreateUser)
	r.PUT("/users/:id", ctrl.UpdateUser)
	r.DELETE("/users/:id", ctrl.DeleteUser)
	return r
}

func toJSON(v interface{}) *bytes.Buffer {
	data, _ := json.Marshal(v)
	return bytes.NewBuffer(data)
}

// --- Tests ---

func TestUserController_GetUsers(t *testing.T) {
	svc := newMockUserService()
	svc.users[1] = &dto.UserResponse{ID: 1, Name: "Alice", Email: "alice@example.com"}
	svc.users[2] = &dto.UserResponse{ID: 2, Name: "Bob", Email: "bob@example.com"}

	r := setupUserRouter(svc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp utils.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Error("Expected success response")
	}
}

func TestUserController_CreateUser_Success(t *testing.T) {
	svc := newMockUserService()
	r := setupUserRouter(svc)

	body := toJSON(map[string]string{
		"name":     "Charlie",
		"email":    "charlie@example.com",
		"password": "secret123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp utils.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Errorf("Expected success, body: %s", w.Body.String())
	}
}

func TestUserController_CreateUser_InvalidBody(t *testing.T) {
	svc := newMockUserService()
	r := setupUserRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestUserController_CreateUser_MissingRequired(t *testing.T) {
	svc := newMockUserService()
	r := setupUserRouter(svc)

	// Missing 'name' and 'password' which are required.
	body := toJSON(map[string]string{"email": "only@example.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing fields, got %d", w.Code)
	}
}

func TestUserController_CreateUser_ServiceError(t *testing.T) {
	svc := newMockUserService()
	svc.createError = errors.New("user with this email already exists")
	r := setupUserRouter(svc)

	body := toJSON(map[string]string{
		"name":     "Dave",
		"email":    "dave@example.com",
		"password": "secret123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestUserController_UpdateUser_Success(t *testing.T) {
	svc := newMockUserService()
	svc.users[1] = &dto.UserResponse{ID: 1, Name: "OldName", Email: "old@example.com"}
	r := setupUserRouter(svc)

	body := toJSON(map[string]string{"name": "NewName"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/1", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestUserController_UpdateUser_InvalidID(t *testing.T) {
	svc := newMockUserService()
	r := setupUserRouter(svc)

	body := toJSON(map[string]string{"name": "X"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/notanid", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid ID, got %d", w.Code)
	}
}

func TestUserController_UpdateUser_InvalidBody(t *testing.T) {
	svc := newMockUserService()
	r := setupUserRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid body, got %d", w.Code)
	}
}

func TestUserController_UpdateUser_ServiceError(t *testing.T) {
	svc := newMockUserService()
	svc.users[1] = &dto.UserResponse{ID: 1, Name: "User", Email: "user@example.com"}
	svc.updateError = errors.New("email already in use")
	r := setupUserRouter(svc)

	body := toJSON(map[string]string{"email": "taken@example.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/1", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestUserController_DeleteUser_Success(t *testing.T) {
	svc := newMockUserService()
	svc.users[5] = &dto.UserResponse{ID: 5, Name: "Eve", Email: "eve@example.com"}
	r := setupUserRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/5", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestUserController_DeleteUser_InvalidID(t *testing.T) {
	svc := newMockUserService()
	r := setupUserRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid ID, got %d", w.Code)
	}
}

func TestUserController_DeleteUser_NotFound(t *testing.T) {
	svc := newMockUserService()
	r := setupUserRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/99", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for not-found user, got %d", w.Code)
	}
}
