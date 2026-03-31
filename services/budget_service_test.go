package services

import (
	"testing"
	"time"

	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
	"my-api/utils"
)

// --- Mock BudgetRepository ---

type mockBudgetRepository struct {
	budgets     map[uint]*models.Budget
	alerts      map[uint]*models.BudgetAlert
	nextBudgetID uint
	nextAlertID  uint
	createError  error
	updateError  error
	deleteError  error
	spentAmounts map[uint]int // budgetID -> spent amount
}

func newMockBudgetRepo() *mockBudgetRepository {
	return &mockBudgetRepository{
		budgets:      make(map[uint]*models.Budget),
		alerts:       make(map[uint]*models.BudgetAlert),
		nextBudgetID: 1,
		nextAlertID:  1,
		spentAmounts: make(map[uint]int),
	}
}

func (m *mockBudgetRepository) Create(budget *models.Budget) error {
	if m.createError != nil {
		return m.createError
	}
	budget.ID = m.nextBudgetID
	m.nextBudgetID++
	copy := *budget
	m.budgets[copy.ID] = &copy
	return nil
}

func (m *mockBudgetRepository) FindByID(id uint, userID uint) (*models.Budget, error) {
	if b, ok := m.budgets[id]; ok && b.UserID == userID {
		return b, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockBudgetRepository) FindAll(userID uint, filter *dto.BudgetFilterRequest) ([]models.Budget, int64, error) {
	var result []models.Budget
	for _, b := range m.budgets {
		if b.UserID == userID {
			result = append(result, *b)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockBudgetRepository) Update(budget *models.Budget) error {
	if m.updateError != nil {
		return m.updateError
	}
	copy := *budget
	m.budgets[copy.ID] = &copy
	return nil
}

func (m *mockBudgetRepository) Delete(id uint, userID uint) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.budgets, id)
	return nil
}

func (m *mockBudgetRepository) FindActiveBudgets(userID uint) ([]models.Budget, error) {
	var result []models.Budget
	for _, b := range m.budgets {
		if b.UserID == userID && b.IsActive {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (m *mockBudgetRepository) GetSpentAmount(budgetID uint, startDate, endDate time.Time) (int, error) {
	if v, ok := m.spentAmounts[budgetID]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *mockBudgetRepository) FindBudgetByCategory(userID, categoryID uint, startDate, endDate time.Time, assetID *uint64) (*models.Budget, error) {
	// Return nil to indicate no overlapping budget by default.
	return nil, nil
}

func (m *mockBudgetRepository) CreateAlert(alert *models.BudgetAlert) error {
	alert.ID = m.nextAlertID
	m.nextAlertID++
	copy := *alert
	m.alerts[copy.ID] = &copy
	return nil
}

func (m *mockBudgetRepository) GetUserAlerts(userID uint, unreadOnly bool) ([]models.BudgetAlert, error) {
	var result []models.BudgetAlert
	for _, a := range m.alerts {
		if a.UserID == userID {
			if unreadOnly && a.IsRead {
				continue
			}
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockBudgetRepository) GetUserAlertsPaginated(userID uint, filter *dto.AlertFilterRequest) ([]models.BudgetAlert, int64, error) {
	all, _ := m.GetUserAlerts(userID, filter.UnreadOnly)
	return all, int64(len(all)), nil
}

func (m *mockBudgetRepository) MarkAlertAsRead(alertID uint, userID uint) error {
	if a, ok := m.alerts[alertID]; ok && a.UserID == userID {
		a.IsRead = true
	}
	return nil
}

func (m *mockBudgetRepository) MarkAllAlertsAsRead(userID uint) error {
	for _, a := range m.alerts {
		if a.UserID == userID {
			a.IsRead = true
		}
	}
	return nil
}

// Helper to build a CreateBudgetRequest with a start date.
func newBudgetRequest(categoryID uint, amount int, period string) *dto.CreateBudgetRequest {
	return &dto.CreateBudgetRequest{
		CategoryID: categoryID,
		Amount:     amount,
		Period:     period,
		StartDate:  utils.CustomTime{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
}

// --- Tests ---

func TestBudgetService_CreateBudget(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	resp, err := svc.CreateBudget(1, newBudgetRequest(10, 5000, "monthly"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected budget response, got nil")
	}
	if resp.Amount != 5000 {
		t.Errorf("Expected amount 5000, got %d", resp.Amount)
	}
	if resp.Period != "monthly" {
		t.Errorf("Expected period 'monthly', got '%s'", resp.Period)
	}
}

func TestBudgetService_CreateBudget_DefaultAlertAt(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	resp, err := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.AlertAt != 80 {
		t.Errorf("Expected default AlertAt 80, got %d", resp.AlertAt)
	}
}

func TestBudgetService_CreateBudget_CustomAlertAt(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	req := newBudgetRequest(10, 1000, "monthly")
	req.AlertAt = 90

	resp, err := svc.CreateBudget(1, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.AlertAt != 90 {
		t.Errorf("Expected AlertAt 90, got %d", resp.AlertAt)
	}
}

func TestBudgetService_CreateBudget_MonthlyEndDate(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	resp, err := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// StartDate is 2024-01-01; monthly period → EndDate should be 2024-01-31.
	if resp.EndDate.Month() != 1 || resp.EndDate.Day() != 31 {
		t.Errorf("Expected EndDate 2024-01-31, got %v", resp.EndDate)
	}
}

func TestBudgetService_CreateBudget_YearlyEndDate(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	resp, err := svc.CreateBudget(1, newBudgetRequest(10, 1000, "yearly"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// StartDate is 2024-01-01; yearly period → EndDate should be 2024-12-31.
	if resp.EndDate.Year() != 2024 || resp.EndDate.Month() != 12 || resp.EndDate.Day() != 31 {
		t.Errorf("Expected EndDate 2024-12-31, got %v", resp.EndDate)
	}
}

func TestBudgetService_CreateBudget_RepoError(t *testing.T) {
	repo := newMockBudgetRepo()
	repo.createError = gorm.ErrInvalidData
	svc := NewBudgetService(repo)

	_, err := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	if err == nil {
		t.Error("Expected error from repository")
	}
}

func TestBudgetService_GetBudgetByID(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	created, _ := svc.CreateBudget(1, newBudgetRequest(10, 2000, "monthly"))

	resp, err := svc.GetBudgetByID(created.ID, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Amount != 2000 {
		t.Errorf("Expected amount 2000, got %d", resp.Amount)
	}
}

func TestBudgetService_GetBudgetByID_NotFound(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	_, err := svc.GetBudgetByID(999, 1)
	if err == nil {
		t.Error("Expected error for non-existent budget")
	}
	if err.Error() != "budget not found" {
		t.Errorf("Expected 'budget not found', got '%v'", err)
	}
}

func TestBudgetService_GetBudgetByID_SpendingStatus(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	created, _ := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	repo.spentAmounts[created.ID] = 500 // 50% spent

	resp, err := svc.GetBudgetByID(created.ID, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.SpentAmount != 500 {
		t.Errorf("Expected SpentAmount 500, got %d", resp.SpentAmount)
	}
	if resp.RemainingAmount != 500 {
		t.Errorf("Expected RemainingAmount 500, got %d", resp.RemainingAmount)
	}
	if resp.Status != "safe" {
		t.Errorf("Expected status 'safe', got '%s'", resp.Status)
	}
}

func TestBudgetService_GetBudgetByID_StatusWarning(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	created, _ := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	repo.spentAmounts[created.ID] = 850 // 85% — over default AlertAt of 80

	resp, _ := svc.GetBudgetByID(created.ID, 1)
	if resp.Status != "warning" {
		t.Errorf("Expected status 'warning', got '%s'", resp.Status)
	}
}

func TestBudgetService_GetBudgetByID_StatusExceeded(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	created, _ := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	repo.spentAmounts[created.ID] = 1100 // over budget

	resp, _ := svc.GetBudgetByID(created.ID, 1)
	if resp.Status != "exceeded" {
		t.Errorf("Expected status 'exceeded', got '%s'", resp.Status)
	}
}

func TestBudgetService_GetAllBudgets(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	svc.CreateBudget(1, newBudgetRequest(11, 2000, "yearly"))
	svc.CreateBudget(2, newBudgetRequest(10, 500, "monthly")) // different user

	result, err := svc.GetAllBudgets(1, &dto.BudgetFilterRequest{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalItems != 2 {
		t.Errorf("Expected 2 budgets for user 1, got %d", result.TotalItems)
	}
}

func TestBudgetService_UpdateBudget_Amount(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	created, _ := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))

	updated, err := svc.UpdateBudget(created.ID, 1, &dto.UpdateBudgetRequest{Amount: 2000})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if updated.Amount != 2000 {
		t.Errorf("Expected amount 2000, got %d", updated.Amount)
	}
}

func TestBudgetService_UpdateBudget_IsActive(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	created, _ := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	inactive := false

	updated, err := svc.UpdateBudget(created.ID, 1, &dto.UpdateBudgetRequest{IsActive: &inactive})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if updated.IsActive {
		t.Error("Expected IsActive to be false after update")
	}
}

func TestBudgetService_UpdateBudget_NotFound(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	_, err := svc.UpdateBudget(999, 1, &dto.UpdateBudgetRequest{Amount: 500})
	if err == nil {
		t.Error("Expected error for missing budget")
	}
	if err.Error() != "budget not found" {
		t.Errorf("Expected 'budget not found', got '%v'", err)
	}
}

func TestBudgetService_DeleteBudget(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	created, _ := svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	if err := svc.DeleteBudget(created.ID, 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err := svc.GetBudgetByID(created.ID, 1)
	if err == nil {
		t.Error("Expected error for deleted budget")
	}
}

func TestBudgetService_DeleteBudget_NotFound(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	err := svc.DeleteBudget(999, 1)
	if err == nil {
		t.Error("Expected error for non-existent budget")
	}
	if err.Error() != "budget not found" {
		t.Errorf("Expected 'budget not found', got '%v'", err)
	}
}

func TestBudgetService_GetBudgetStatus(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	svc.CreateBudget(1, newBudgetRequest(10, 1000, "monthly"))
	svc.CreateBudget(1, newBudgetRequest(11, 2000, "monthly"))

	statuses, err := svc.GetBudgetStatus(1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(statuses) != 2 {
		t.Errorf("Expected 2 budget statuses, got %d", len(statuses))
	}
}

func TestBudgetService_GetUserAlerts(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	// Create an alert directly in the mock.
	repo.alerts[1] = &models.BudgetAlert{
		ID:       1,
		UserID:   1,
		BudgetID: 1,
		Message:  "Test alert",
		IsRead:   false,
		CreatedAt: utils.CustomTime{Time: time.Now()},
	}
	repo.nextAlertID = 2

	alerts, err := svc.GetUserAlerts(1, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
}

func TestBudgetService_GetUserAlerts_UnreadOnly(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	repo.alerts[1] = &models.BudgetAlert{ID: 1, UserID: 1, IsRead: false, CreatedAt: utils.CustomTime{Time: time.Now()}}
	repo.alerts[2] = &models.BudgetAlert{ID: 2, UserID: 1, IsRead: true, CreatedAt: utils.CustomTime{Time: time.Now()}}
	repo.nextAlertID = 3

	unread, err := svc.GetUserAlerts(1, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(unread) != 1 {
		t.Errorf("Expected 1 unread alert, got %d", len(unread))
	}
}

func TestBudgetService_MarkAlertAsRead(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	repo.alerts[1] = &models.BudgetAlert{ID: 1, UserID: 1, IsRead: false, CreatedAt: utils.CustomTime{Time: time.Now()}}
	repo.nextAlertID = 2

	if err := svc.MarkAlertAsRead(1, 1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !repo.alerts[1].IsRead {
		t.Error("Expected alert to be marked as read")
	}
}

func TestBudgetService_MarkAllAlertsAsRead(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	for i := uint(1); i <= 3; i++ {
		repo.alerts[i] = &models.BudgetAlert{ID: i, UserID: 1, IsRead: false, CreatedAt: utils.CustomTime{Time: time.Now()}}
	}
	repo.nextAlertID = 4

	if err := svc.MarkAllAlertsAsRead(1); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	for i := uint(1); i <= 3; i++ {
		if !repo.alerts[i].IsRead {
			t.Errorf("Alert %d should be marked as read", i)
		}
	}
}

func TestBudgetService_GetUserAlertsPaginated(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewBudgetService(repo)

	for i := uint(1); i <= 5; i++ {
		repo.alerts[i] = &models.BudgetAlert{ID: i, UserID: 1, IsRead: false, CreatedAt: utils.CustomTime{Time: time.Now()}}
	}
	repo.nextAlertID = 6

	result, err := svc.GetUserAlertsPaginated(1, &dto.AlertFilterRequest{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalItems != 5 {
		t.Errorf("Expected 5 total alerts, got %d", result.TotalItems)
	}
}
