package services

import (
	"errors"
	"fmt"
	"sort"
	"time"
	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"my-api/utils"
)

// ─── Interface ────────────────────────────────────────────────────────────────

type DebtService interface {
	CreateDebt(userID uint, req *dto.CreateDebtRequest) (*dto.DebtResponse, error)
	GetDebt(id, userID uint, includePayments, includeMilestones bool) (*dto.DebtDetailResponse, error)
	GetAllDebts(userID uint, status string) (*dto.DebtsListResponse, error)
	UpdateDebt(id, userID uint, req *dto.UpdateDebtRequest) (*dto.DebtResponse, error)
	DeleteDebt(id, userID uint) error
	RecordPayment(debtID, userID uint, req *dto.RecordDebtPaymentRequest) (*dto.RecordDebtPaymentResponse, error)
	UpdateBalance(debtID, userID uint, newBalance int) error
	GetPayoffStrategies(userID uint, extraPayment int) (*dto.StrategiesResponse, error)
	GetPayoffTimeline(debtID, userID uint, strategy string, extraPayment int) (*dto.TimelineResponse, error)
	ProcessMonthlyInterest() error
	SendPaymentReminders() error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type debtService struct {
	repo              repositories.DebtRepository
	transactionV2Repo repositories.TransactionV2Repository
}

func NewDebtService(repo repositories.DebtRepository, txV2Repo repositories.TransactionV2Repository) DebtService {
	return &debtService{
		repo:              repo,
		transactionV2Repo: txV2Repo,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *debtService) CreateDebt(userID uint, req *dto.CreateDebtRequest) (*dto.DebtResponse, error) {
	interestType := "fixed"
	if req.InterestType != "" {
		interestType = req.InterestType
	}

	debt := &models.Debt{
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		DebtType:       req.DebtType,
		CreditorName:   req.CreditorName,
		OriginalAmount: req.OriginalAmount,
		CurrentBalance: req.CurrentBalance,
		InterestRate:   req.InterestRate,
		InterestType:   interestType,
		MinimumPayment: req.MinimumPayment,
		PaymentDueDay:  req.PaymentDueDay,
		StartDate:      req.StartDate,
		PaymentAssetID: req.PaymentAssetID,
		Notes:          req.Notes,
		Status:         "active",
	}

	if err := s.repo.Create(debt); err != nil {
		return nil, err
	}

	return s.toDebtResponse(debt, true), nil
}

// ─── Get Single ───────────────────────────────────────────────────────────────

func (s *debtService) GetDebt(id, userID uint, includePayments, includeMilestones bool) (*dto.DebtDetailResponse, error) {
	debt, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	resp := &dto.DebtDetailResponse{
		DebtResponse: *s.toDebtResponse(debt, true),
	}

	if includePayments {
		payments, _ := s.repo.FindPayments(debt.ID)
		resp.Payments = toPaymentHistoryItems(payments)
	}

	if includeMilestones {
		resp.Milestones = toMilestoneItems(debt.Milestones)
	}

	return resp, nil
}

// ─── Get All ──────────────────────────────────────────────────────────────────

func (s *debtService) GetAllDebts(userID uint, status string) (*dto.DebtsListResponse, error) {
	debts, err := s.repo.FindAllByUser(userID, status)
	if err != nil {
		return nil, err
	}

	summary := buildSummary(debts)
	responses := make([]dto.DebtResponse, len(debts))
	for i, d := range debts {
		responses[i] = *s.toDebtResponse(&d, false)
	}

	return &dto.DebtsListResponse{
		Summary: summary,
		Debts:   responses,
	}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *debtService) UpdateDebt(id, userID uint, req *dto.UpdateDebtRequest) (*dto.DebtResponse, error) {
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
	if req.Description != "" {
		debt.Description = req.Description
	}
	if req.CreditorName != "" {
		debt.CreditorName = req.CreditorName
	}
	if req.InterestRate != nil {
		debt.InterestRate = *req.InterestRate
	}
	if req.InterestType != "" {
		debt.InterestType = req.InterestType
	}
	if req.MinimumPayment != nil {
		debt.MinimumPayment = *req.MinimumPayment
	}
	if req.PaymentDueDay != nil {
		debt.PaymentDueDay = *req.PaymentDueDay
	}
	if req.PaymentAssetID != nil {
		debt.PaymentAssetID = req.PaymentAssetID
	}
	if req.IncludeInNetWorth != nil {
		debt.IncludeInNetWorth = *req.IncludeInNetWorth
	}
	if req.AutoTrackInterest != nil {
		debt.AutoTrackInterest = *req.AutoTrackInterest
	}
	if req.Status != "" {
		debt.Status = req.Status
	}
	if req.Notes != "" {
		debt.Notes = req.Notes
	}
	if req.ExpectedPayoffDate != nil {
		debt.ExpectedPayoffDate = req.ExpectedPayoffDate
	}

	if err := s.repo.Update(debt); err != nil {
		return nil, err
	}

	return s.toDebtResponse(debt, true), nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *debtService) DeleteDebt(id, userID uint) error {
	_, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("debt not found")
		}
		return err
	}
	return s.repo.Delete(id, userID)
}

// ─── Record Payment ───────────────────────────────────────────────────────────

func (s *debtService) RecordPayment(debtID, userID uint, req *dto.RecordDebtPaymentRequest) (*dto.RecordDebtPaymentResponse, error) {
	debt, err := s.repo.FindByID(debtID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	if debt.Status != "active" {
		return nil, errors.New("cannot record payment on inactive debt")
	}

	balanceBefore := debt.CurrentBalance
	balanceAfter := balanceBefore - req.PrincipalAmount
	if balanceAfter < 0 {
		balanceAfter = 0
	}

	payment := &models.DebtPayment{
		DebtID:          debtID,
		UserID:          userID,
		Amount:          req.Amount,
		PaymentType:     req.PaymentType,
		PrincipalAmount: req.PrincipalAmount,
		InterestAmount:  req.InterestAmount,
		FeesAmount:      req.FeesAmount,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		PaymentDate:     req.PaymentDate,
		Notes:           req.Notes,
	}

	// Create a linked transaction if source_asset_id is provided
	if req.SourceAssetID != nil {
		tx := &models.TransactionV2{
			UserID:          userID,
			Description:     fmt.Sprintf("Debt payment: %s", debt.Name),
			Amount:          req.Amount,
			TransactionType: 2, // Expense
			AssetID:         uint64(*req.SourceAssetID),
			Date:            req.PaymentDate,
		}
		if err := s.transactionV2Repo.CreateWithBalanceUpdate(tx); err == nil {
			txID := uint(tx.ID)
			payment.TransactionID = &txID
		}
	}

	if err := s.repo.CreatePayment(payment); err != nil {
		return nil, err
	}

	// Update debt
	debt.CurrentBalance = balanceAfter
	debt.CurrentMonthPaid = true

	if balanceAfter == 0 {
		debt.Status = "paid_off"
		now := utils.CustomTime{Time: time.Now()}
		debt.ActualPayoffDate = &now
	}

	s.repo.Update(debt)

	// Check milestones asynchronously (best effort)
	go s.checkMilestones(debt)

	return &dto.RecordDebtPaymentResponse{
		PaymentID:     payment.ID,
		Amount:        payment.Amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		IsPaidOff:     balanceAfter == 0,
	}, nil
}

// ─── Update Balance ───────────────────────────────────────────────────────────

func (s *debtService) UpdateBalance(debtID, userID uint, newBalance int) error {
	debt, err := s.repo.FindByID(debtID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("debt not found")
		}
		return err
	}

	debt.CurrentBalance = newBalance
	if newBalance == 0 {
		debt.Status = "paid_off"
		now := utils.CustomTime{Time: time.Now()}
		debt.ActualPayoffDate = &now
	}

	return s.repo.Update(debt)
}

// ─── Payoff Strategies ────────────────────────────────────────────────────────

func (s *debtService) GetPayoffStrategies(userID uint, extraPayment int) (*dto.StrategiesResponse, error) {
	debts, err := s.repo.FindActiveDebts(userID)
	if err != nil {
		return nil, err
	}
	if len(debts) == 0 {
		return nil, errors.New("no active debts found")
	}

	totalMinimum := 0
	for _, d := range debts {
		totalMinimum += d.MinimumPayment
	}

	strategies := make([]dto.PayoffStrategyItem, 0, 3)

	// 1. Minimum payments only
	minStrategy := s.calculatePayoff(debts, 0, "minimum")
	strategies = append(strategies, minStrategy)

	var recommendation dto.StrategyRecommendation

	if extraPayment > 0 {
		// 2. Avalanche (highest interest first)
		avalanche := s.calculatePayoff(debts, extraPayment, "avalanche")
		avalanche.InterestSaved = minStrategy.TotalInterest - avalanche.TotalInterest
		avalanche.MonthsSaved = monthsBetweenDates(avalanche.PayoffDate, minStrategy.PayoffDate)
		strategies = append(strategies, avalanche)

		// 3. Snowball (smallest balance first)
		snowball := s.calculatePayoff(debts, extraPayment, "snowball")
		snowball.InterestSaved = minStrategy.TotalInterest - snowball.TotalInterest
		snowball.MonthsSaved = monthsBetweenDates(snowball.PayoffDate, minStrategy.PayoffDate)
		strategies = append(strategies, snowball)

		recommendation = dto.StrategyRecommendation{
			Strategy: "Avalanche",
			Reason: fmt.Sprintf(
				"Saves the most money (Rp %s in interest) while paying off debts %d months faster",
				formatRupiah(avalanche.InterestSaved), avalanche.MonthsSaved,
			),
		}
	} else {
		recommendation = dto.StrategyRecommendation{
			Strategy: "Add Extra Payments",
			Reason:   "Consider adding extra payments to accelerate debt payoff and save on interest",
		}
	}

	return &dto.StrategiesResponse{
		CurrentMonthlyPayment: totalMinimum,
		ExtraAvailable:        extraPayment,
		Strategies:            strategies,
		Recommendation:        recommendation,
	}, nil
}

// ─── Payoff Timeline ──────────────────────────────────────────────────────────

func (s *debtService) GetPayoffTimeline(debtID, userID uint, strategy string, extraPayment int) (*dto.TimelineResponse, error) {
	debt, err := s.repo.FindByID(debtID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}

	timeline := make([]dto.TimelineMonth, 0)
	balance := debt.CurrentBalance
	totalInterest := 0
	currentDate := time.Now()

	for balance > 0 {
		monthlyInterest := int(float64(balance) * (debt.InterestRate / 100 / 12))
		balance += monthlyInterest
		totalInterest += monthlyInterest

		payment := debt.MinimumPayment + extraPayment
		if balance < payment {
			payment = balance
		}
		principalPaid := payment - monthlyInterest
		if principalPaid < 0 {
			principalPaid = 0
		}
		balance -= payment
		if balance < 0 {
			balance = 0
		}

		timeline = append(timeline, dto.TimelineMonth{
			Month:            currentDate.Format("Jan 2006"),
			Payment:          payment,
			PrincipalPaid:    principalPaid,
			InterestPaid:     monthlyInterest,
			RemainingBalance: balance,
		})

		currentDate = currentDate.AddDate(0, 1, 0)

		// Safety limit
		if currentDate.Year() > time.Now().Year()+30 {
			break
		}
	}

	return &dto.TimelineResponse{
		DebtID:        debtID,
		Strategy:      strategy,
		PayoffDate:    currentDate.Format("Jan 2006"),
		TotalInterest: totalInterest,
		Timeline:      timeline,
	}, nil
}

// ─── Background Jobs ──────────────────────────────────────────────────────────

func (s *debtService) ProcessMonthlyInterest() error {
	debts, err := s.repo.FindAllActiveWithAutoTrack()
	if err != nil {
		return err
	}

	for _, debt := range debts {
		interest := debt.MonthlyInterest()
		if interest > 0 {
			if err := s.repo.AddInterest(debt.ID, interest); err != nil {
				continue
			}
		}
		// Reset current_month_paid flag for the new month
		s.repo.ResetMonthlyPaidFlag(debt.ID)
	}

	return nil
}

func (s *debtService) SendPaymentReminders() error {
	debts, err := s.repo.FindDebtsWithDueDateIn(3)
	if err != nil {
		return err
	}

	for _, debt := range debts {
		// In a real scenario you'd send a push/email notification here.
		// For now we just log a placeholder.
		_ = fmt.Sprintf(
			"[Reminder] User %d — %s payment of Rp %s due in %d days",
			debt.UserID, debt.Name, formatRupiah(debt.MinimumPayment), debt.DaysUntilDue(),
		)
	}

	return nil
}

// ─── Internal Helpers ─────────────────────────────────────────────────────────

type simulatedDebt struct {
	ID             uint
	Name           string
	Balance        int
	InterestRate   float64
	MinimumPayment int
}

func (s *debtService) calculatePayoff(debts []models.Debt, extraPayment int, strategy string) dto.PayoffStrategyItem {
	// Deep copy for simulation
	sim := make([]simulatedDebt, len(debts))
	for i, d := range debts {
		sim[i] = simulatedDebt{
			ID:             d.ID,
			Name:           d.Name,
			Balance:        d.CurrentBalance,
			InterestRate:   d.InterestRate,
			MinimumPayment: d.MinimumPayment,
		}
	}

	// Sort based on strategy
	switch strategy {
	case "avalanche":
		sort.Slice(sim, func(i, j int) bool {
			return sim[i].InterestRate > sim[j].InterestRate
		})
	case "snowball":
		sort.Slice(sim, func(i, j int) bool {
			return sim[i].Balance < sim[j].Balance
		})
	}

	currentDate := time.Now()
	totalInterest := 0
	milestones := make([]dto.PayoffMilestoneItem, 0)

	for {
		allPaidOff := true
		availableExtra := extraPayment

		for i := range sim {
			if sim[i].Balance <= 0 {
				continue
			}
			allPaidOff = false

			// Monthly interest
			monthlyInterest := int(float64(sim[i].Balance) * (sim[i].InterestRate / 100 / 12))
			totalInterest += monthlyInterest
			sim[i].Balance += monthlyInterest

			// Minimum payment
			payment := sim[i].MinimumPayment
			if sim[i].Balance < payment {
				payment = sim[i].Balance
			}
			sim[i].Balance -= payment

			// Extra payment to the target debt (first one with remaining balance)
			if availableExtra > 0 && sim[i].Balance > 0 {
				extra := availableExtra
				if sim[i].Balance < extra {
					extra = sim[i].Balance
				}
				sim[i].Balance -= extra
				availableExtra -= extra
			}

			// Record payoff milestone
			if sim[i].Balance <= 0 && strategy != "minimum" {
				milestones = append(milestones, dto.PayoffMilestoneItem{
					Date:  currentDate.Format("Jan 2006"),
					Event: fmt.Sprintf("%s paid off!", sim[i].Name),
				})
			}
		}

		if allPaidOff {
			break
		}

		currentDate = currentDate.AddDate(0, 1, 0)
		if currentDate.Year() > time.Now().Year()+30 {
			break
		}
	}

	// Build order list (original balances)
	order := make([]dto.DebtOrderItem, len(sim))
	originalBalances := make(map[uint]int, len(debts))
	for _, d := range debts {
		originalBalances[d.ID] = d.CurrentBalance
	}
	for i, d := range sim {
		order[i] = dto.DebtOrderItem{
			DebtID:       d.ID,
			Name:         d.Name,
			Balance:      originalBalances[d.ID],
			InterestRate: d.InterestRate,
		}
	}

	var name, description string
	switch strategy {
	case "avalanche":
		name = "Avalanche (Highest Interest First)"
		description = "Pay minimum on all debts, put extra toward the highest interest debt"
	case "snowball":
		name = "Snowball (Smallest Balance First)"
		description = "Pay minimum on all debts, put extra toward the smallest balance"
	default:
		name = "Minimum Payments Only"
		description = "Pay only minimum payments on all debts"
	}

	return dto.PayoffStrategyItem{
		Name:          name,
		Description:   description,
		PayoffDate:    currentDate.Format("Jan 2006"),
		TotalInterest: totalInterest,
		Order:         order,
		Milestones:    milestones,
	}
}

func (s *debtService) checkMilestones(debt *models.Debt) {
	percentPaid := debt.PaidOffPercentage()

	thresholds := []struct {
		Type    string
		Limit   float64
		Message string
	}{
		{"25_percent", 25.0, "You've paid off 25%% of %s! Keep going! 💪"},
		{"50_percent", 50.0, "Halfway there! 50%% of %s is paid off! 🎉"},
		{"75_percent", 75.0, "Amazing! 75%% of %s is gone! Almost there! 🚀"},
		{"paid_off", 100.0, "Congratulations! %s is completely paid off! 🏆"},
	}

	for _, t := range thresholds {
		if percentPaid >= t.Limit {
			existing, _ := s.repo.FindMilestone(debt.ID, t.Type)
			if existing != nil {
				continue // already recorded
			}

			now := time.Now()
			milestone := &models.DebtMilestone{
				DebtID:        debt.ID,
				UserID:        debt.UserID,
				MilestoneType: t.Type,
				Description:   fmt.Sprintf(t.Message, debt.Name),
				ReachedAt:     &now,
			}
			s.repo.CreateMilestone(milestone)
		}
	}
}

func (s *debtService) toDebtResponse(debt *models.Debt, includeProjections bool) *dto.DebtResponse {
	resp := &dto.DebtResponse{
		ID:                debt.ID,
		Name:              debt.Name,
		DebtType:          debt.DebtType,
		CreditorName:      debt.CreditorName,
		OriginalAmount:    debt.OriginalAmount,
		CurrentBalance:    debt.CurrentBalance,
		PaidOffAmount:     debt.PaidOffAmount(),
		PaidOffPercentage: debt.PaidOffPercentage(),
		InterestRate:      debt.InterestRate,
		InterestType:      debt.InterestType,
		MinimumPayment:    debt.MinimumPayment,
		PaymentDueDay:     debt.PaymentDueDay,
		DaysUntilDue:      debt.DaysUntilDue(),
		CurrentMonthPaid:  debt.CurrentMonthPaid,
		Status:            debt.Status,
		StartDate:         debt.StartDate,
		Notes:             debt.Notes,
		CreatedAt:         debt.CreatedAt,
	}

	if includeProjections && debt.Status == "active" {
		proj := s.buildProjections(debt)
		resp.Projections = proj
	}

	return resp
}

func (s *debtService) buildProjections(debt *models.Debt) *dto.DebtProjections {
	balance := debt.CurrentBalance
	months := 0
	totalInterest := 0

	for balance > 0 {
		interest := int(float64(balance) * (debt.InterestRate / 100 / 12))
		balance += interest
		totalInterest += interest

		payment := debt.MinimumPayment
		if balance < payment {
			payment = balance
		}
		balance -= payment
		months++

		if months > 360 { // 30 year safety limit
			break
		}
	}

	payoffDate := time.Now().AddDate(0, months, 0)

	return &dto.DebtProjections{
		PayoffDateMinimum:    payoffDate.Format("2006-01-02"),
		TotalInterestMinimum: totalInterest,
		MonthsRemainingMin:   months,
	}
}

// ─── Package-level Helpers ────────────────────────────────────────────────────

func buildSummary(debts []models.Debt) dto.DebtSummary {
	var totalDebt, totalMin, totalPaidThisMonth, paidCount, unpaidCount int
	var totalRate float64

	for _, d := range debts {
		totalDebt += d.CurrentBalance
		totalMin += d.MinimumPayment
		totalRate += d.InterestRate

		if d.CurrentMonthPaid {
			paidCount++
			totalPaidThisMonth += d.MinimumPayment
		} else {
			unpaidCount++
		}
	}

	avgRate := 0.0
	if len(debts) > 0 {
		avgRate = totalRate / float64(len(debts))
	}

	return dto.DebtSummary{
		TotalDebt:            totalDebt,
		TotalMinimumPayment:  totalMin,
		TotalPaidThisMonth:   totalPaidThisMonth,
		DebtsPaidThisMonth:   paidCount,
		DebtsUnpaidThisMonth: unpaidCount,
		AvgInterestRate:      avgRate,
	}
}

func toPaymentHistoryItems(payments []models.DebtPayment) []dto.DebtPaymentHistoryItem {
	items := make([]dto.DebtPaymentHistoryItem, len(payments))
	for i, p := range payments {
		items[i] = dto.DebtPaymentHistoryItem{
			ID:              p.ID,
			Amount:          p.Amount,
			PaymentType:     p.PaymentType,
			PrincipalAmount: p.PrincipalAmount,
			InterestAmount:  p.InterestAmount,
			FeesAmount:      p.FeesAmount,
			BalanceBefore:   p.BalanceBefore,
			BalanceAfter:    p.BalanceAfter,
			PaymentDate:     p.PaymentDate,
			Notes:           p.Notes,
			CreatedAt:       p.CreatedAt,
		}
	}
	return items
}

func toMilestoneItems(milestones []models.DebtMilestone) []dto.DebtMilestoneItem {
	items := make([]dto.DebtMilestoneItem, len(milestones))
	for i, m := range milestones {
		items[i] = dto.DebtMilestoneItem{
			ID:            m.ID,
			MilestoneType: m.MilestoneType,
			Description:   m.Description,
			ReachedAt:     m.ReachedAt,
			IsCelebrated:  m.IsCelebrated,
		}
	}
	return items
}

func monthsBetweenDates(earlier, later string) int {
	e, err1 := time.Parse("Jan 2006", earlier)
	l, err2 := time.Parse("Jan 2006", later)
	if err1 != nil || err2 != nil {
		return 0
	}
	months := (l.Year()-e.Year())*12 + int(l.Month()-e.Month())
	if months < 0 {
		return 0
	}
	return months
}

func formatRupiah(amount int) string {
	return fmt.Sprintf("%d", amount)
}
