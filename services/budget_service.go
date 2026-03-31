package services

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"my-api/utils"
	"time"
)

type BudgetService interface {
	CreateBudget(userID uint, req *dto.CreateBudgetRequest) (*dto.BudgetResponse, error)
	GetBudgetByID(id uint, userID uint) (*dto.BudgetWithSpendingResponse, error)
	GetAllBudgets(userID uint, filter *dto.BudgetFilterRequest) (*dto.PaginationResponse, error)
	UpdateBudget(id uint, userID uint, req *dto.UpdateBudgetRequest) (*dto.BudgetResponse, error)
	DeleteBudget(id uint, userID uint) error
	GetBudgetStatus(userID uint) ([]dto.BudgetWithSpendingResponse, error)
	CheckBudgetAlerts(userID uint) error
	CheckVelocityAlerts(userID uint) error
	CheckAnomalyAlert(userID uint, transaction *models.TransactionV2) error
	GenerateDailySummaries() error
	GenerateWeeklyReports() error
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
	existing, _ := s.repo.FindBudgetByCategory(userID, req.CategoryID, req.StartDate.Time, endDate.Time, nil)
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

// standardThresholds lists the fixed percentage levels to alert at.
var standardThresholds = []int{50, 75, 90, 100}

func (s *budgetService) CheckBudgetAlerts(userID uint) error {
	budgets, err := s.repo.FindActiveBudgets(userID)
	if err != nil {
		return err
	}

	// Load all existing threshold alerts for this user once (avoids N+1 queries)
	existingAlerts, _ := s.repo.GetUserAlerts(userID, false)

	for _, budget := range budgets {
		spent, _ := s.repo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
		if spent == 0 {
			continue
		}
		percentage := float64(spent) / float64(budget.Amount) * 100

		// Build deduplicated set of thresholds to evaluate (standard + user's custom AlertAt)
		thresholdsToCheck := make([]int, len(standardThresholds))
		copy(thresholdsToCheck, standardThresholds)
		isStandard := false
		for _, t := range standardThresholds {
			if t == budget.AlertAt {
				isStandard = true
				break
			}
		}
		if !isStandard {
			thresholdsToCheck = append(thresholdsToCheck, budget.AlertAt)
		}

		for _, threshold := range thresholdsToCheck {
			if int(percentage) < threshold {
				continue
			}

			// Skip if we already have a threshold alert for this budget at or above this level
			alreadyAlerted := false
			for _, existing := range existingAlerts {
				if existing.BudgetID == budget.ID &&
					existing.AlertType == models.AlertTypeThreshold &&
					existing.Percentage >= threshold {
					alreadyAlerted = true
					break
				}
			}
			if alreadyAlerted {
				continue
			}

			// Build message with top contributing transactions
			statusMsg := "reached"
			if percentage >= 100 {
				statusMsg = "exceeded"
			}
			message := fmt.Sprintf("You have %s %d%% of your %s budget", statusMsg, threshold, budget.Category.CategoryName)

			topTxns, _ := s.repo.GetTopTransactionsForBudget(&budget, 3)
			if len(topTxns) > 0 {
				message += ". Top transactions: "
				for i, t := range topTxns {
					if i > 0 {
						message += ", "
					}
					message += fmt.Sprintf("%s (%d)", t.Description, t.Amount)
				}
			}

			alert := &models.BudgetAlert{
				BudgetID:    budget.ID,
				UserID:      userID,
				AlertType:   models.AlertTypeThreshold,
				Percentage:  int(percentage),
				SpentAmount: spent,
				Message:     message,
			}
			if err := s.repo.CreateAlert(alert); err == nil {
				// Keep the in-memory slice consistent so later thresholds for the same budget see this alert
				existingAlerts = append(existingAlerts, *alert)
			}
		}
	}

	return nil
}

func (s *budgetService) CheckVelocityAlerts(userID uint) error {
	budgets, err := s.repo.FindActiveBudgets(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, budget := range budgets {
		daysInPeriod := budget.EndDate.Sub(budget.StartDate.Time).Hours() / 24
		daysElapsed := now.Sub(budget.StartDate.Time).Hours() / 24

		if daysElapsed < 1 {
			continue // Not enough data to calculate a meaningful rate (also handles budgets not yet started)
		}

		spent, _ := s.repo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
		if spent == 0 {
			continue
		}

		dailyRate := float64(spent) / daysElapsed
		projectedTotal := dailyRate * daysInPeriod

		if projectedTotal <= float64(budget.Amount) {
			continue // On track
		}

		// Calculate how many days until budget is exhausted
		remaining := float64(budget.Amount - spent)
		if remaining <= 0 {
			continue // Already exceeded — threshold alert handles this
		}
		daysUntilExceeded := remaining / dailyRate
		expectedExceedDate := now.AddDate(0, 0, int(daysUntilExceeded))

		// Skip if the budget period itself ends before the projected exceed date
		if expectedExceedDate.After(budget.EndDate.Time) {
			continue
		}

		// Deduplicate: only send one velocity alert per budget every 3 days
		since := now.AddDate(0, 0, -3)
		recentAlert, _ := s.repo.GetRecentAlertByBudgetAndType(budget.ID, models.AlertTypeVelocity, since)
		if recentAlert != nil {
			continue
		}

		percentage := int(float64(spent) / float64(budget.Amount) * 100)
		message := fmt.Sprintf(
			"Velocity alert for %s: at %.0f/day you'll exceed your budget around %s (in ~%d days)",
			budget.Category.CategoryName, dailyRate, expectedExceedDate.Format("Jan 2"), int(daysUntilExceeded),
		)

		alert := &models.BudgetAlert{
			BudgetID:    budget.ID,
			UserID:      userID,
			AlertType:   models.AlertTypeVelocity,
			Percentage:  percentage,
			SpentAmount: spent,
			Message:     message,
		}
		s.repo.CreateAlert(alert)
	}

	return nil
}

func (s *budgetService) CheckAnomalyAlert(userID uint, transaction *models.TransactionV2) error {
	// Only analyze expense transactions that have a category
	if transaction.TransactionType != 2 || transaction.CategoryID == nil {
		return nil
	}

	// Need a budget for this category to associate the alert with
	now := time.Now()
	budget, err := s.repo.FindBudgetByCategory(userID, *transaction.CategoryID, now, now, nil)
	if err != nil || budget == nil || budget.ID == 0 {
		return nil
	}

	// Get 30-day average transaction amount for the category
	since := now.AddDate(0, 0, -30)
	avg, err := s.repo.GetAverageCategoryTransactionAmount(userID, *transaction.CategoryID, since)
	if err != nil || avg == 0 {
		return nil
	}

	// Alert only when the transaction is 3× above average
	if float64(transaction.Amount) <= avg*3 {
		return nil
	}

	// Deduplicate: one anomaly alert per budget per 24 hours
	since24h := now.Add(-24 * time.Hour)
	recentAlert, _ := s.repo.GetRecentAlertByBudgetAndType(budget.ID, models.AlertTypeAnomaly, since24h)
	if recentAlert != nil {
		return nil
	}

	multiplier := float64(transaction.Amount) / avg
	categoryName := budget.Category.CategoryName
	message := fmt.Sprintf(
		"Unusual spending in %s: %s (%d) is %.1fx your typical amount (30-day avg: %.0f)",
		categoryName, transaction.Description, transaction.Amount, multiplier, avg,
	)

	spent, _ := s.repo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
	alert := &models.BudgetAlert{
		BudgetID:    budget.ID,
		UserID:      userID,
		AlertType:   models.AlertTypeAnomaly,
		Percentage:  int(float64(spent) / float64(budget.Amount) * 100),
		SpentAmount: transaction.Amount,
		Message:     message,
	}
	s.repo.CreateAlert(alert)

	return nil
}

func (s *budgetService) GenerateDailySummaries() error {
	userIDs, err := s.repo.GetUsersWithActiveBudgets()
	if err != nil {
		return err
	}

	today := time.Now().Truncate(24 * time.Hour)
	todayEnd := today.Add(24*time.Hour - time.Second)

	for _, userID := range userIDs {
		// Skip if a daily summary was already generated for this user today
		recentAlert, _ := s.repo.GetRecentAlertByUserAndType(userID, models.AlertTypeDailySummary, today)
		if recentAlert != nil {
			continue
		}

		budgets, _ := s.repo.FindActiveBudgets(userID)
		for _, budget := range budgets {
			todaySpent, _ := s.repo.GetSpentAmount(budget.ID, today, todayEnd)
			if todaySpent == 0 {
				continue
			}

			daysInPeriod := int(budget.EndDate.Sub(budget.StartDate.Time).Hours()/24) + 1
			if daysInPeriod < 1 {
				daysInPeriod = 1
			}
			dailyTarget := budget.Amount / daysInPeriod

			status := "on track"
			if dailyTarget > 0 {
				if todaySpent > dailyTarget*2 {
					status = "over daily target"
				} else if todaySpent > dailyTarget {
					status = "slightly over daily target"
				}
			}

			totalSpent, _ := s.repo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
			percentUsed := 0
			if budget.Amount > 0 {
				percentUsed = int(float64(totalSpent) / float64(budget.Amount) * 100)
			}

			message := fmt.Sprintf(
				"Daily summary for %s: spent %d today (daily target: %d) — %s. Budget used: %d%%",
				budget.Category.CategoryName, todaySpent, dailyTarget, status, percentUsed,
			)

			alert := &models.BudgetAlert{
				BudgetID:    budget.ID,
				UserID:      userID,
				AlertType:   models.AlertTypeDailySummary,
				Percentage:  percentUsed,
				SpentAmount: todaySpent,
				Message:     message,
			}
			s.repo.CreateAlert(alert)
		}
	}

	return nil
}

func (s *budgetService) GenerateWeeklyReports() error {
	userIDs, err := s.repo.GetUsersWithActiveBudgets()
	if err != nil {
		return err
	}

	now := time.Now()
	// Week boundaries: Monday 00:00 → Sunday 23:59:59
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7).Add(-time.Second)
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	prevWeekEnd := weekStart.Add(-time.Second)

	for _, userID := range userIDs {
		// Skip if already reported this week
		recentAlert, _ := s.repo.GetRecentAlertByUserAndType(userID, models.AlertTypeWeeklyReport, weekStart)
		if recentAlert != nil {
			continue
		}

		budgets, _ := s.repo.FindActiveBudgets(userID)
		for _, budget := range budgets {
			thisWeekSpent, _ := s.repo.GetSpentAmount(budget.ID, weekStart, weekEnd)
			prevWeekSpent, _ := s.repo.GetSpentAmount(budget.ID, prevWeekStart, prevWeekEnd)

			if thisWeekSpent == 0 && prevWeekSpent == 0 {
				continue
			}

			// Approximate weekly target (budget.Amount is for the full period)
			daysInPeriod := int(budget.EndDate.Sub(budget.StartDate.Time).Hours()/24) + 1
			if daysInPeriod < 1 {
				daysInPeriod = 1
			}
			weeklyTarget := int(float64(budget.Amount) * 7 / float64(daysInPeriod))

			changeMsg := "no change"
			if prevWeekSpent > 0 {
				changePct := (float64(thisWeekSpent-prevWeekSpent) / float64(prevWeekSpent)) * 100
				if changePct > 0 {
					changeMsg = fmt.Sprintf("up %.0f%% vs last week", changePct)
				} else if changePct < 0 {
					changeMsg = fmt.Sprintf("down %.0f%% vs last week", -changePct)
				}
			}

			totalSpent, _ := s.repo.GetSpentAmount(budget.ID, budget.StartDate.Time, budget.EndDate.Time)
			percentUsed := 0
			if budget.Amount > 0 {
				percentUsed = int(float64(totalSpent) / float64(budget.Amount) * 100)
			}

			message := fmt.Sprintf(
				"Weekly report for %s: spent %d this week (target: %d/week), %s. Overall budget used: %d%%",
				budget.Category.CategoryName, thisWeekSpent, weeklyTarget, changeMsg, percentUsed,
			)

			alert := &models.BudgetAlert{
				BudgetID:    budget.ID,
				UserID:      userID,
				AlertType:   models.AlertTypeWeeklyReport,
				Percentage:  percentUsed,
				SpentAmount: thisWeekSpent,
				Message:     message,
			}
			s.repo.CreateAlert(alert)
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
			AlertType:   alert.AlertType,
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
