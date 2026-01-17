package services

import (
	"errors"
	"fmt"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"my-api/utils"
	"time"
	"gorm.io/gorm"
)

type BudgetService interface {
	CreateBudget(userID uint, req *dto.CreateBudgetRequest) (*dto.BudgetResponse, error)
	GetBudgetByID(id uint, userID uint) (*dto.BudgetWithSpendingResponse, error)
	GetAllBudgets(userID uint, filter *dto.BudgetFilterRequest) (*dto.PaginationResponse, error)
	UpdateBudget(id uint, userID uint, req *dto.UpdateBudgetRequest) (*dto.BudgetResponse, error)
	DeleteBudget(id uint, userID uint) error
	GetBudgetStatus(userID uint) ([]dto.BudgetWithSpendingResponse, error)
	CheckBudgetAlerts(userID uint) error
	GetUserAlerts(userID uint, unreadOnly bool) ([]dto.BudgetAlertResponse, error)
	GetUserAlertsPaginated(userID uint, filter *dto.AlertFilterRequest) (*dto.PaginationResponse, error)
	MarkAlertAsRead(alertID uint, userID uint) error
	MarkAllAlertsAsRead(userID uint) error
}

type budgetService struct {
	repo repositories.BudgetRepository
}

func NewBudgetService(repo repositories.BudgetRepository) BudgetService {
	return &budgetService{repo: repo}
}

func (s *budgetService) CreateBudget(userID uint, req *dto.CreateBudgetRequest) (*dto.BudgetResponse, error) {
	endDate := s.calculateEndDate(req.StartDate.Time, req.Period)

	// Check for overlapping budgets
	existing, _ := s.repo.FindBudgetByCategory(userID, req.CategoryID, req.StartDate.Time, endDate.Time)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("budget already exists for this category in the specified period")
	}

	alertAt := 80
	if req.AlertAt > 0 {
		alertAt = req.AlertAt
	}

	budget := &models.Budget{
		UserID:      userID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Period:      req.Period,
		StartDate:   req.StartDate,
		EndDate:     endDate,
		IsActive:    true,
		AlertAt:     alertAt,
		Description: req.Description,
	}

	if err := s.repo.Create(budget); err != nil {
		return nil, err
	}

	budget, _ = s.repo.FindByID(budget.ID, userID)
	return s.toBudgetResponse(budget), nil
}

func (s *budgetService) GetBudgetByID(id uint, userID uint) (*dto.BudgetWithSpendingResponse, error) {
	budget, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("budget not found")
		}
		return nil, err
	}

	return s.toBudgetWithSpendingResponse(budget), nil
}

func (s *budgetService) GetAllBudgets(userID uint, filter *dto.BudgetFilterRequest) (*dto.PaginationResponse, error) {
	filter.SetDefaults()

	budgets, total, err := s.repo.FindAll(userID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.BudgetWithSpendingResponse, len(budgets))
	for i, budget := range budgets {
		responses[i] = *s.toBudgetWithSpendingResponse(&budget)
	}

	return dto.NewPaginationResponse(responses, filter.Page, filter.PageSize, total), nil
}

func (s *budgetService) UpdateBudget(id uint, userID uint, req *dto.UpdateBudgetRequest) (*dto.BudgetResponse, error) {
	budget, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("budget not found")
		}
		return nil, err
	}

	if req.Amount > 0 {
		budget.Amount = req.Amount
	}
	if req.AlertAt > 0 {
		budget.AlertAt = req.AlertAt
	}
	if req.Description != "" {
		budget.Description = req.Description
	}
	if req.IsActive != nil {
		budget.IsActive = *req.IsActive
	}

	if err := s.repo.Update(budget); err != nil {
		return nil, err
	}

	return s.toBudgetResponse(budget), nil
}

func (s *budgetService) DeleteBudget(id uint, userID uint) error {
	_, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("budget not found")
		}
		return err
	}

	return s.repo.Delete(id, userID)
}

func (s *budgetService) GetBudgetStatus(userID uint) ([]dto.BudgetWithSpendingResponse, error) {
	budgets, err := s.repo.FindActiveBudgets(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.BudgetWithSpendingResponse, len(budgets))
	for i, budget := range budgets {
		responses[i] = *s.toBudgetWithSpendingResponse(&budget)
	}

	return responses, nil
}

func (s *budgetService) CheckBudgetAlerts(userID uint) error {
	budgets, err := s.repo.FindActiveBudgets(userID)
	if err != nil {
		return err
	}

	for _, budget := range budgets {
		spent, _ := s.repo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
		percentage := float64(spent) / float64(budget.Amount) * 100

		if percentage >= float64(budget.AlertAt) {
			// Check if alert already exists for this budget at this percentage level
			existingAlerts, _ := s.repo.GetUserAlerts(userID, true)
			alertExists := false
			for _, existing := range existingAlerts {
				if existing.BudgetID == budget.ID && existing.Percentage >= int(percentage)-5 {
					alertExists = true
					break
				}
			}

			// Only create alert if it doesn't exist
			if !alertExists {
				statusMsg := "reached"
				if percentage >= 100 {
					statusMsg = "exceeded"
				}
				alert := &models.BudgetAlert{
					BudgetID:    budget.ID,
					UserID:      userID,
					Percentage:  int(percentage),
					SpentAmount: spent,
					Message:     fmt.Sprintf("You have %s %.0f%% of your %s budget", statusMsg, percentage, budget.Category.CategoryName),
				}
				s.repo.CreateAlert(alert)
			}
		}
	}

	return nil
}

func (s *budgetService) GetUserAlerts(userID uint, unreadOnly bool) ([]dto.BudgetAlertResponse, error) {
	alerts, err := s.repo.GetUserAlerts(userID, unreadOnly)
	if err != nil {
		return nil, err
	}

	return s.toAlertResponses(alerts), nil
}

func (s *budgetService) GetUserAlertsPaginated(userID uint, filter *dto.AlertFilterRequest) (*dto.PaginationResponse, error) {
	filter.SetDefaults()

	alerts, total, err := s.repo.GetUserAlertsPaginated(userID, filter)
	if err != nil {
		return nil, err
	}

	responses := s.toAlertResponses(alerts)
	return dto.NewPaginationResponse(responses, filter.Page, filter.PageSize, total), nil
}

func (s *budgetService) toAlertResponses(alerts []models.BudgetAlert) []dto.BudgetAlertResponse {
	responses := make([]dto.BudgetAlertResponse, len(alerts))
	for i, alert := range alerts {
		response := dto.BudgetAlertResponse{
			ID:          alert.ID,
			BudgetID:    alert.BudgetID,
			Percentage:  alert.Percentage,
			SpentAmount: alert.SpentAmount,
			Message:     alert.Message,
			IsRead:      alert.IsRead,
			CreatedAt:   alert.CreatedAt.Time,
		}
		
		// Include budget and category information if available
		if alert.Budget.ID > 0 {
			response.CategoryID = alert.Budget.CategoryID
			response.BudgetAmount = alert.Budget.Amount
			if alert.Budget.Category.ID > 0 {
				response.CategoryName = alert.Budget.Category.CategoryName
			}
		}
		
		responses[i] = response
	}
	return responses
}

func (s *budgetService) MarkAlertAsRead(alertID uint, userID uint) error {
	return s.repo.MarkAlertAsRead(alertID, userID)
}

func (s *budgetService) MarkAllAlertsAsRead(userID uint) error {
	return s.repo.MarkAllAlertsAsRead(userID)
}

// Helper functions
func (s *budgetService) calculateEndDate(startDate time.Time, period string) utils.CustomTime {
	var endDate time.Time
	switch period {
	case "monthly":
		endDate = startDate.AddDate(0, 1, -1)
	case "yearly":
		endDate = startDate.AddDate(1, 0, -1)
	default:
		endDate = startDate.AddDate(0, 1, -1)
	}
	return utils.CustomTime{Time: endDate}
}

func (s *budgetService) toBudgetResponse(budget *models.Budget) *dto.BudgetResponse {
	categoryName := ""
	if budget.Category.ID > 0 {
		categoryName = budget.Category.CategoryName
	}

	return &dto.BudgetResponse{
		ID:           budget.ID,
		CategoryID:   budget.CategoryID,
		CategoryName: categoryName,
		Amount:       budget.Amount,
		Period:       budget.Period,
		StartDate:    budget.StartDate,
		EndDate:      budget.EndDate,
		IsActive:     budget.IsActive,
		AlertAt:      budget.AlertAt,
		Description:  budget.Description,
		CreatedAt:    budget.CreatedAt,
	}
}

func (s *budgetService) toBudgetWithSpendingResponse(budget *models.Budget) *dto.BudgetWithSpendingResponse {
	baseResponse := s.toBudgetResponse(budget)
	spent, _ := s.repo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
	
	remaining := budget.Amount - spent
	percentageUsed := float64(spent) / float64(budget.Amount) * 100
	
	status := "safe"
	if percentageUsed >= 100 {
		status = "exceeded"
	} else if percentageUsed >= float64(budget.AlertAt) {
		status = "warning"
	}

	daysRemaining := int(budget.EndDate.Sub(time.Now()).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	return &dto.BudgetWithSpendingResponse{
		BudgetResponse:  *baseResponse,
		SpentAmount:     spent,
		RemainingAmount: remaining,
		PercentageUsed:  percentageUsed,
		Status:          status,
		DaysRemaining:   daysRemaining,
	}
}
