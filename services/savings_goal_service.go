package services

import (
	"errors"
	"math"
	"time"

	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"my-api/utils"
)

type SavingsGoalService interface {
	CreateGoal(userID uint, req *dto.CreateSavingsGoalRequest) (*dto.SavingsGoalResponse, error)
	GetGoalByID(id uint, userID uint) (*dto.SavingsGoalResponse, error)
	GetAllGoals(userID uint, filter *dto.SavingsGoalFilterRequest) (*dto.PaginationResponse, error)
	UpdateGoal(id uint, userID uint, req *dto.UpdateSavingsGoalRequest) (*dto.SavingsGoalResponse, error)
	DeleteGoal(id uint, userID uint) error

	AddContribution(goalID uint, userID uint, req *dto.AddContributionRequest) (*dto.SavingsContributionResponse, error)
	GetContributions(goalID uint, userID uint, filter *dto.ContributionFilterRequest) (*dto.PaginationResponse, error)
	DeleteContribution(id uint, goalID uint, userID uint) error
}

type savingsGoalService struct {
	repo repositories.SavingsGoalRepository
}

func NewSavingsGoalService(repo repositories.SavingsGoalRepository) SavingsGoalService {
	return &savingsGoalService{repo: repo}
}

func (s *savingsGoalService) CreateGoal(userID uint, req *dto.CreateSavingsGoalRequest) (*dto.SavingsGoalResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "IDR"
	}

	goal := &models.SavingsGoal{
		UserID:       userID,
		AssetID:      req.AssetID,
		Name:         req.Name,
		Description:  req.Description,
		TargetAmount: req.TargetAmount,
		Currency:     currency,
		Deadline:     req.Deadline,
		Status:       "active",
		Icon:         req.Icon,
		Color:        req.Color,
	}

	if err := s.repo.Create(goal); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(goal.ID, userID)
	if err != nil {
		return nil, err
	}
	return s.toGoalResponse(created), nil
}

func (s *savingsGoalService) GetGoalByID(id uint, userID uint) (*dto.SavingsGoalResponse, error) {
	goal, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("savings goal not found")
		}
		return nil, err
	}
	return s.toGoalResponse(goal), nil
}

func (s *savingsGoalService) GetAllGoals(userID uint, filter *dto.SavingsGoalFilterRequest) (*dto.PaginationResponse, error) {
	filter.SetDefaults()

	goals, total, err := s.repo.FindAll(userID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.SavingsGoalResponse, len(goals))
	for i, goal := range goals {
		responses[i] = *s.toGoalResponse(&goal)
	}

	return dto.NewPaginationResponse(responses, filter.Page, filter.PageSize, total), nil
}

func (s *savingsGoalService) UpdateGoal(id uint, userID uint, req *dto.UpdateSavingsGoalRequest) (*dto.SavingsGoalResponse, error) {
	goal, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("savings goal not found")
		}
		return nil, err
	}

	if req.Name != "" {
		goal.Name = req.Name
	}
	if req.Description != "" {
		goal.Description = req.Description
	}
	if req.TargetAmount > 0 {
		goal.TargetAmount = req.TargetAmount
	}
	if req.Currency != "" {
		goal.Currency = req.Currency
	}
	if req.Deadline != nil {
		goal.Deadline = *req.Deadline
	}
	if req.AssetID != nil {
		goal.AssetID = req.AssetID
	}
	if req.Icon != "" {
		goal.Icon = req.Icon
	}
	if req.Color != "" {
		goal.Color = req.Color
	}
	if req.Status != "" {
		goal.Status = req.Status
	}

	if err := s.repo.Update(goal); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(goal.ID, userID)
	if err != nil {
		return nil, err
	}
	return s.toGoalResponse(updated), nil
}

func (s *savingsGoalService) DeleteGoal(id uint, userID uint) error {
	_, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("savings goal not found")
		}
		return err
	}
	return s.repo.Delete(id, userID)
}

func (s *savingsGoalService) AddContribution(goalID uint, userID uint, req *dto.AddContributionRequest) (*dto.SavingsContributionResponse, error) {
	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("savings goal not found")
		}
		return nil, err
	}

	if goal.Status == "cancelled" {
		return nil, errors.New("cannot add contribution to a cancelled goal")
	}

	date := req.Date
	if date.IsZero() {
		date = utils.CustomTime{Time: time.Now()}
	}

	contribution := &models.SavingsContribution{
		GoalID: goalID,
		UserID: userID,
		Amount: req.Amount,
		Note:   req.Note,
		Date:   date,
	}

	if err := s.repo.CreateContribution(contribution); err != nil {
		return nil, err
	}

	// Update goal's current amount
	goal.CurrentAmount += req.Amount
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.Status = "completed"
	}
	if err := s.repo.Update(goal); err != nil {
		return nil, err
	}

	return s.toContributionResponse(contribution), nil
}

func (s *savingsGoalService) GetContributions(goalID uint, userID uint, filter *dto.ContributionFilterRequest) (*dto.PaginationResponse, error) {
	// Verify goal belongs to user
	_, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("savings goal not found")
		}
		return nil, err
	}

	filter.SetDefaults()

	contributions, total, err := s.repo.FindContributions(goalID, userID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.SavingsContributionResponse, len(contributions))
	for i, c := range contributions {
		responses[i] = *s.toContributionResponse(&c)
	}

	return dto.NewPaginationResponse(responses, filter.Page, filter.PageSize, total), nil
}

func (s *savingsGoalService) DeleteContribution(id uint, goalID uint, userID uint) error {
	contribution, err := s.repo.FindContributionByID(id, goalID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contribution not found")
		}
		return err
	}

	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteContribution(id, goalID, userID); err != nil {
		return err
	}

	// Rollback goal's current amount
	goal.CurrentAmount -= contribution.Amount
	if goal.CurrentAmount < 0 {
		goal.CurrentAmount = 0
	}
	// Revert status if it was completed
	if goal.Status == "completed" && goal.CurrentAmount < goal.TargetAmount {
		goal.Status = "active"
	}
	if err := s.repo.Update(goal); err != nil {
		return err
	}

	return nil
}

// toGoalResponse converts a SavingsGoal model to a SavingsGoalResponse DTO
func (s *savingsGoalService) toGoalResponse(goal *models.SavingsGoal) *dto.SavingsGoalResponse {
	remaining := goal.TargetAmount - goal.CurrentAmount
	if remaining < 0 {
		remaining = 0
	}

	var progress float64
	if goal.TargetAmount > 0 {
		progress = math.Min(float64(goal.CurrentAmount)/float64(goal.TargetAmount)*100, 100)
	}

	now := time.Now()
	daysRemaining := 0
	monthsRemaining := 0
	monthlyTarget := 0

	if !goal.Deadline.IsZero() && goal.Deadline.After(now) {
		daysRemaining = int(goal.Deadline.Sub(now).Hours() / 24)
		// Calculate whole months remaining using year/month components
		years := goal.Deadline.Year() - now.Year()
		months := int(goal.Deadline.Month()) - int(now.Month())
		monthsRemaining = years*12 + months
		if goal.Deadline.Day() > now.Day() {
			monthsRemaining++
		}
		if monthsRemaining < 1 {
			monthsRemaining = 1
		}
		if remaining > 0 {
			monthlyTarget = int(math.Ceil(float64(remaining) / float64(monthsRemaining)))
		}
	}

	assetName := ""
	if goal.Asset != nil {
		assetName = goal.Asset.Name
	}

	return &dto.SavingsGoalResponse{
		ID:                  goal.ID,
		Name:                goal.Name,
		Description:         goal.Description,
		TargetAmount:        goal.TargetAmount,
		CurrentAmount:       goal.CurrentAmount,
		RemainingAmount:     remaining,
		Currency:            goal.Currency,
		Deadline:            goal.Deadline,
		Status:              goal.Status,
		Icon:                goal.Icon,
		Color:               goal.Color,
		ProgressPercentage:  progress,
		MonthlyTargetAmount: monthlyTarget,
		MonthsRemaining:     monthsRemaining,
		DaysRemaining:       daysRemaining,
		AssetID:             goal.AssetID,
		AssetName:           assetName,
		CreatedAt:           goal.CreatedAt,
	}
}

func (s *savingsGoalService) toContributionResponse(c *models.SavingsContribution) *dto.SavingsContributionResponse {
	return &dto.SavingsContributionResponse{
		ID:        c.ID,
		GoalID:    c.GoalID,
		Amount:    c.Amount,
		Note:      c.Note,
		Date:      c.Date,
		CreatedAt: c.CreatedAt,
	}
}
