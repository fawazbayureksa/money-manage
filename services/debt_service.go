package services

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
)

type DebtService interface {
	// Debt CRUD
	CreateDebt(userID uint, req *dto.CreateDebtRequest) (*dto.DebtResponse, error)
	GetDebtByID(id uint, userID uint) (*dto.DebtResponse, error)
	GetAllDebts(userID uint, filter *dto.DebtFilterRequest) (*dto.PaginationResponse, error)
	UpdateDebt(id uint, userID uint, req *dto.UpdateDebtRequest) (*dto.DebtResponse, error)
	DeleteDebt(id uint, userID uint) error

	// Payments
	RecordPayment(debtID uint, userID uint, req *dto.CreateDebtPaymentRequest) (*dto.DebtPaymentResponse, error)
	GetPayments(debtID uint, userID uint, filter *dto.DebtPaymentFilterRequest) (*dto.PaginationResponse, error)
	DeletePayment(paymentID uint, userID uint) error

	// Analytics & Strategy
	GetDebtSummary(userID uint) (*dto.DebtSummaryResponse, error)
	GetPayoffStrategies(userID uint) ([]dto.PayoffStrategyResponse, error)
	GetPayoffProjection(debtID uint, userID uint, extraPayment int) (*dto.PayoffProjectionResponse, error)

	// Milestones
	GetMilestones(userID uint, unreadOnly bool) ([]dto.DebtMilestoneResponse, error)
	MarkMilestoneAsRead(milestoneID uint, userID uint) error
	MarkAllMilestonesAsRead(userID uint) error
}

type debtService struct {
	repo repositories.DebtRepository
}

func NewDebtService(repo repositories.DebtRepository) DebtService {
	return &debtService{repo: repo}
}

// --- Debt CRUD ---

func (s *debtService) CreateDebt(userID uint, req *dto.CreateDebtRequest) (*dto.DebtResponse, error) {
	dueDay := 1
	if req.DueDay > 0 {
		dueDay = req.DueDay
	}

	debt := &models.Debt{
		UserID:         userID,
		Name:           req.Name,
		DebtType:       req.DebtType,
		OriginalAmount: req.OriginalAmount,
		CurrentBalance: req.CurrentBalance,
		InterestRate:   req.InterestRate,
		MinimumPayment: req.MinimumPayment,
		DueDay:         dueDay,
		StartDate:      req.StartDate,
		LenderName:     req.LenderName,
		Notes:          req.Notes,
		IsActive:       true,
	}

	if err := s.repo.Create(debt); err != nil {
		return nil, err
	}

	return s.toDebtResponse(debt)
}

func (s *debtService) GetDebtByID(id uint, userID uint) (*dto.DebtResponse, error) {
	debt, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	return s.toDebtResponse(debt)
}

func (s *debtService) GetAllDebts(userID uint, filter *dto.DebtFilterRequest) (*dto.PaginationResponse, error) {
	filter.SetDefaults()

	debts, total, err := s.repo.FindAll(userID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DebtResponse, len(debts))
	for i, debt := range debts {
		resp, err := s.toDebtResponse(&debt)
		if err != nil {
			return nil, err
		}
		responses[i] = *resp
	}

	return dto.NewPaginationResponse(responses, filter.Page, filter.PageSize, total), nil
}

func (s *debtService) UpdateDebt(id uint, userID uint, req *dto.UpdateDebtRequest) (*dto.DebtResponse, error) {
	debt, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	if req.Name != "" {
		debt.Name = req.Name
	}
	if req.CurrentBalance != nil {
		debt.CurrentBalance = *req.CurrentBalance
	}
	if req.InterestRate != nil {
		debt.InterestRate = *req.InterestRate
	}
	if req.MinimumPayment != nil {
		debt.MinimumPayment = *req.MinimumPayment
	}
	if req.DueDay != nil {
		debt.DueDay = *req.DueDay
	}
	if req.LenderName != "" {
		debt.LenderName = req.LenderName
	}
	if req.Notes != "" {
		debt.Notes = req.Notes
	}
	if req.IsActive != nil {
		debt.IsActive = *req.IsActive
	}

	if err := s.repo.Update(debt); err != nil {
		return nil, err
	}

	return s.toDebtResponse(debt)
}

func (s *debtService) DeleteDebt(id uint, userID uint) error {
	_, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("debt not found")
		}
		return err
	}

	return s.repo.Delete(id, userID)
}

// --- Payments ---

func (s *debtService) RecordPayment(debtID uint, userID uint, req *dto.CreateDebtPaymentRequest) (*dto.DebtPaymentResponse, error) {
	debt, err := s.repo.FindByID(debtID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	if req.Amount > debt.CurrentBalance {
		return nil, errors.New("payment amount exceeds current balance")
	}

	payment := &models.DebtPayment{
		DebtID:      debtID,
		UserID:      userID,
		Amount:      req.Amount,
		PaymentDate: req.PaymentDate,
		Notes:       req.Notes,
	}

	if err := s.repo.CreatePayment(payment); err != nil {
		return nil, err
	}

	// Update current balance
	debt.CurrentBalance -= req.Amount
	if debt.CurrentBalance <= 0 {
		debt.CurrentBalance = 0
		debt.IsActive = false
	}

	if err := s.repo.Update(debt); err != nil {
		return nil, err
	}

	// Check for milestones
	s.checkMilestones(debt, userID)

	return &dto.DebtPaymentResponse{
		ID:          payment.ID,
		DebtID:      payment.DebtID,
		Amount:      payment.Amount,
		PaymentDate: payment.PaymentDate,
		Notes:       payment.Notes,
		CreatedAt:   payment.CreatedAt,
		DebtName:    debt.Name,
	}, nil
}

func (s *debtService) GetPayments(debtID uint, userID uint, filter *dto.DebtPaymentFilterRequest) (*dto.PaginationResponse, error) {
	// Verify debt belongs to user
	_, err := s.repo.FindByID(debtID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	filter.SetDefaults()

	payments, total, err := s.repo.FindPaymentsByDebtID(debtID, userID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DebtPaymentResponse, len(payments))
	for i, p := range payments {
		responses[i] = dto.DebtPaymentResponse{
			ID:          p.ID,
			DebtID:      p.DebtID,
			Amount:      p.Amount,
			PaymentDate: p.PaymentDate,
			Notes:       p.Notes,
			CreatedAt:   p.CreatedAt,
		}
	}

	return dto.NewPaginationResponse(responses, filter.Page, filter.PageSize, total), nil
}

func (s *debtService) DeletePayment(paymentID uint, userID uint) error {
	return s.repo.DeletePayment(paymentID, userID)
}

// --- Analytics & Strategy ---

func (s *debtService) GetDebtSummary(userID uint) (*dto.DebtSummaryResponse, error) {
	debts, err := s.repo.FindActiveDebts(userID)
	if err != nil {
		return nil, err
	}

	// Also get all debts for total count
	allFilter := &dto.DebtFilterRequest{}
	allFilter.SetDefaults()
	allFilter.PageSize = 1000
	allDebts, _, err := s.repo.FindAll(userID, allFilter)
	if err != nil {
		return nil, err
	}

	totalPaid, err := s.repo.GetTotalPaidForUser(userID)
	if err != nil {
		return nil, err
	}

	var totalOwed, totalOriginal, totalMinPayment int
	var highestInterest float64
	var totalMonthlyInterest float64

	for _, debt := range debts {
		totalOwed += debt.CurrentBalance
		totalOriginal += debt.OriginalAmount
		totalMinPayment += debt.MinimumPayment
		if debt.InterestRate > highestInterest {
			highestInterest = debt.InterestRate
		}
		// Monthly interest = balance * (annual rate / 12 / 100)
		monthlyInterest := float64(debt.CurrentBalance) * (debt.InterestRate / 12.0 / 100.0)
		totalMonthlyInterest += monthlyInterest
	}

	// Calculate overall progress across all debts (including paid-off ones)
	var allOriginal int
	for _, debt := range allDebts {
		allOriginal += debt.OriginalAmount
	}

	overallProgress := 0.0
	if allOriginal > 0 {
		overallProgress = float64(totalPaid) / float64(allOriginal) * 100
		if overallProgress > 100 {
			overallProgress = 100
		}
	}

	return &dto.DebtSummaryResponse{
		TotalDebts:        len(allDebts),
		ActiveDebts:       len(debts),
		TotalOwed:         totalOwed,
		TotalOriginal:     allOriginal,
		TotalPaid:         totalPaid,
		OverallProgress:   math.Round(overallProgress*100) / 100,
		TotalMinPayment:   totalMinPayment,
		HighestInterest:   highestInterest,
		EstimatedInterest: int(math.Round(totalMonthlyInterest)),
	}, nil
}

func (s *debtService) GetPayoffStrategies(userID uint) ([]dto.PayoffStrategyResponse, error) {
	debts, err := s.repo.FindActiveDebts(userID)
	if err != nil {
		return nil, err
	}

	if len(debts) == 0 {
		return []dto.PayoffStrategyResponse{}, nil
	}

	strategies := make([]dto.PayoffStrategyResponse, 0, 2)

	// Avalanche strategy: highest interest rate first
	avalancheDebts := make([]models.Debt, len(debts))
	copy(avalancheDebts, debts)
	sort.Slice(avalancheDebts, func(i, j int) bool {
		return avalancheDebts[i].InterestRate > avalancheDebts[j].InterestRate
	})

	avalancheResult := s.buildStrategyResponse(avalancheDebts, "avalanche",
		"Pay off debts with the highest interest rate first. Saves the most money on interest over time.")
	strategies = append(strategies, avalancheResult)

	// Snowball strategy: lowest balance first
	snowballDebts := make([]models.Debt, len(debts))
	copy(snowballDebts, debts)
	sort.Slice(snowballDebts, func(i, j int) bool {
		return snowballDebts[i].CurrentBalance < snowballDebts[j].CurrentBalance
	})

	snowballResult := s.buildStrategyResponse(snowballDebts, "snowball",
		"Pay off debts with the smallest balance first. Provides quick wins for motivation.")
	strategies = append(strategies, snowballResult)

	return strategies, nil
}

func (s *debtService) GetPayoffProjection(debtID uint, userID uint, extraPayment int) (*dto.PayoffProjectionResponse, error) {
	debt, err := s.repo.FindByID(debtID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	if debt.CurrentBalance <= 0 {
		return &dto.PayoffProjectionResponse{
			DebtID:              debt.ID,
			DebtName:            debt.Name,
			CurrentBalance:      debt.CurrentBalance,
			InterestRate:        debt.InterestRate,
			MonthlyPayment:      0,
			EstimatedMonths:     0,
			EstimatedPayoffDate: "Already paid off",
			TotalInterest:       0,
			TotalPayment:        0,
			Schedule:            []dto.PayoffProjectionMonth{},
		}, nil
	}

	monthlyPayment := debt.MinimumPayment + extraPayment
	if monthlyPayment <= 0 {
		monthlyPayment = debt.CurrentBalance / 12
		if monthlyPayment <= 0 {
			monthlyPayment = debt.CurrentBalance
		}
	}

	monthlyRate := debt.InterestRate / 12.0 / 100.0
	balance := float64(debt.CurrentBalance)
	totalInterest := 0.0
	totalPayment := 0.0
	months := 0
	maxMonths := 360 // 30 years cap

	schedule := make([]dto.PayoffProjectionMonth, 0)

	for balance > 0 && months < maxMonths {
		months++
		interest := balance * monthlyRate
		totalInterest += interest

		payment := float64(monthlyPayment)
		if payment > balance+interest {
			payment = balance + interest
		}

		principal := payment - interest
		if principal < 0 {
			// Payment doesn't even cover interest
			principal = 0
			balance += interest - payment
		} else {
			balance -= principal
		}
		totalPayment += payment

		if balance < 0.5 {
			balance = 0
		}

		schedule = append(schedule, dto.PayoffProjectionMonth{
			Month:            months,
			Payment:          int(math.Round(payment)),
			Principal:        int(math.Round(principal)),
			Interest:         int(math.Round(interest)),
			RemainingBalance: int(math.Round(balance)),
		})
	}

	payoffDate := time.Now().AddDate(0, months, 0).Format("2006-01-02")

	return &dto.PayoffProjectionResponse{
		DebtID:              debt.ID,
		DebtName:            debt.Name,
		CurrentBalance:      debt.CurrentBalance,
		InterestRate:        debt.InterestRate,
		MonthlyPayment:      monthlyPayment,
		EstimatedMonths:     months,
		EstimatedPayoffDate: payoffDate,
		TotalInterest:       int(math.Round(totalInterest)),
		TotalPayment:        int(math.Round(totalPayment)),
		Schedule:            schedule,
	}, nil
}

// --- Milestones ---

func (s *debtService) GetMilestones(userID uint, unreadOnly bool) ([]dto.DebtMilestoneResponse, error) {
	milestones, err := s.repo.GetUserMilestones(userID, unreadOnly)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DebtMilestoneResponse, len(milestones))
	for i, m := range milestones {
		responses[i] = dto.DebtMilestoneResponse{
			ID:            m.ID,
			DebtID:        m.DebtID,
			MilestoneType: m.MilestoneType,
			TargetValue:   m.TargetValue,
			Message:       m.Message,
			IsRead:        m.IsRead,
			CreatedAt:     m.CreatedAt.Time,
		}
		if m.Debt.ID > 0 {
			responses[i].DebtName = m.Debt.Name
		}
	}

	return responses, nil
}

func (s *debtService) MarkMilestoneAsRead(milestoneID uint, userID uint) error {
	return s.repo.MarkMilestoneAsRead(milestoneID, userID)
}

func (s *debtService) MarkAllMilestonesAsRead(userID uint) error {
	return s.repo.MarkAllMilestonesAsRead(userID)
}

// --- Helper functions ---

func (s *debtService) toDebtResponse(debt *models.Debt) (*dto.DebtResponse, error) {
	totalPaid, err := s.repo.GetTotalPaid(debt.ID)
	if err != nil {
		return nil, err
	}

	percentagePaid := 0.0
	if debt.OriginalAmount > 0 {
		percentagePaid = float64(totalPaid) / float64(debt.OriginalAmount) * 100
		if percentagePaid > 100 {
			percentagePaid = 100
		}
	}

	return &dto.DebtResponse{
		ID:              debt.ID,
		Name:            debt.Name,
		DebtType:        debt.DebtType,
		OriginalAmount:  debt.OriginalAmount,
		CurrentBalance:  debt.CurrentBalance,
		InterestRate:    debt.InterestRate,
		MinimumPayment:  debt.MinimumPayment,
		DueDay:          debt.DueDay,
		StartDate:       debt.StartDate,
		LenderName:      debt.LenderName,
		Notes:           debt.Notes,
		IsActive:        debt.IsActive,
		CreatedAt:       debt.CreatedAt,
		TotalPaid:       totalPaid,
		PercentagePaid:  math.Round(percentagePaid*100) / 100,
		RemainingAmount: debt.CurrentBalance,
	}, nil
}

func (s *debtService) checkMilestones(debt *models.Debt, userID uint) {
	totalPaid, err := s.repo.GetTotalPaid(debt.ID)
	if err != nil {
		return
	}

	percentagePaid := 0.0
	if debt.OriginalAmount > 0 {
		percentagePaid = float64(totalPaid) / float64(debt.OriginalAmount) * 100
	}

	milestones := []int{25, 50, 75, 100}
	for _, target := range milestones {
		if percentagePaid >= float64(target) {
			// Check if milestone already exists
			existing, err := s.repo.FindMilestoneByDebtAndTarget(debt.ID, "percentage_paid", target)
			if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				var message string
				if target == 100 {
					message = fmt.Sprintf("🎉 Congratulations! You've completely paid off '%s'!", debt.Name)
				} else {
					message = fmt.Sprintf("🎯 Great progress! You've paid off %d%% of '%s'!", target, debt.Name)
				}

				milestone := &models.DebtMilestone{
					DebtID:        debt.ID,
					UserID:        userID,
					MilestoneType: "percentage_paid",
					TargetValue:   target,
					Message:       message,
				}
				s.repo.CreateMilestone(milestone)
			} else if existing != nil {
				continue
			}
		}
	}

	// Debt-free milestone
	if debt.CurrentBalance <= 0 {
		existing, err := s.repo.FindMilestoneByDebtAndTarget(debt.ID, "debt_free", 0)
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			milestone := &models.DebtMilestone{
				DebtID:        debt.ID,
				UserID:        userID,
				MilestoneType: "debt_free",
				TargetValue:   0,
				Message:       fmt.Sprintf("🏆 You are now debt-free on '%s'! Total paid: %d", debt.Name, totalPaid),
			}
			s.repo.CreateMilestone(milestone)
		} else if existing != nil {
			return
		}
	}
}

func (s *debtService) buildStrategyResponse(debts []models.Debt, strategy, description string) dto.PayoffStrategyResponse {
	strategyDebts := make([]dto.PayoffStrategyDebt, len(debts))
	for i, d := range debts {
		strategyDebts[i] = dto.PayoffStrategyDebt{
			ID:             d.ID,
			Name:           d.Name,
			CurrentBalance: d.CurrentBalance,
			InterestRate:   d.InterestRate,
			MinimumPayment: d.MinimumPayment,
			PayoffOrder:    i + 1,
		}
	}

	// Estimate total months and interest with this strategy
	totalMonths, totalInterest := s.simulatePayoff(debts)

	return dto.PayoffStrategyResponse{
		Strategy:          strategy,
		Description:       description,
		Debts:             strategyDebts,
		EstimatedMonths:   totalMonths,
		EstimatedInterest: totalInterest,
	}
}

func (s *debtService) simulatePayoff(debts []models.Debt) (int, int) {
	if len(debts) == 0 {
		return 0, 0
	}

	// Create working copies
	type debtState struct {
		balance    float64
		rate       float64
		minPayment float64
	}

	states := make([]debtState, len(debts))
	totalMinPayment := 0.0
	for i, d := range debts {
		states[i] = debtState{
			balance:    float64(d.CurrentBalance),
			rate:       d.InterestRate / 12.0 / 100.0,
			minPayment: float64(d.MinimumPayment),
		}
		totalMinPayment += float64(d.MinimumPayment)
	}

	if totalMinPayment <= 0 {
		// If no minimum payments set, assume paying off in 24 months
		for i := range states {
			states[i].minPayment = states[i].balance / 24.0
		}
	}

	months := 0
	totalInterest := 0.0
	maxMonths := 360

	for months < maxMonths {
		allPaid := true
		for _, st := range states {
			if st.balance > 0 {
				allPaid = false
				break
			}
		}
		if allPaid {
			break
		}

		months++
		extraBudget := 0.0

		for i := range states {
			if states[i].balance <= 0 {
				extraBudget += states[i].minPayment
				continue
			}

			interest := states[i].balance * states[i].rate
			totalInterest += interest
			states[i].balance += interest

			payment := states[i].minPayment
			if i == 0 {
				// First debt in priority gets extra budget from paid-off debts
				payment += extraBudget
				extraBudget = 0
			}

			if payment > states[i].balance {
				payment = states[i].balance
			}
			states[i].balance -= payment

			if states[i].balance < 0.5 {
				states[i].balance = 0
			}
		}
	}

	return months, int(math.Round(totalInterest))
}
