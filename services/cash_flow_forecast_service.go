package services

import (
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"time"
)

// CashFlowForecastService projects future balances and supports what-if scenarios.
type CashFlowForecastService interface {
	GetForecast(userID uint, req *dto.CashFlowForecastRequest) (*dto.CashFlowForecastResponse, error)
	GetScenario(userID uint, req *dto.CashFlowScenarioRequest) (*dto.CashFlowForecastResponse, error)
}

type cashFlowForecastService struct {
	recurringRepo   repositories.RecurringTransactionRepository
	transactionRepo repositories.TransactionV2Repository
	assetRepo       *repositories.AssetRepository
}

func NewCashFlowForecastService(
	recurringRepo repositories.RecurringTransactionRepository,
	transactionRepo repositories.TransactionV2Repository,
	assetRepo *repositories.AssetRepository,
) CashFlowForecastService {
	return &cashFlowForecastService{
		recurringRepo:   recurringRepo,
		transactionRepo: transactionRepo,
		assetRepo:       assetRepo,
	}
}

// GetForecast returns a day-by-day cash flow projection.
func (s *cashFlowForecastService) GetForecast(userID uint, req *dto.CashFlowForecastRequest) (*dto.CashFlowForecastResponse, error) {
	days := req.Days
	if days == 0 {
		days = 30
	}

	currentBalance, err := s.getCurrentBalance(userID, req.AssetID)
	if err != nil {
		return nil, err
	}

	avgDailyExpense, err := s.getAvgDailyExpense(userID, req.AssetID)
	if err != nil {
		return nil, err
	}

	recurring, err := s.recurringRepo.FindActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.buildForecast(currentBalance, avgDailyExpense, days, recurring, nil, req.AssetID), nil
}

// GetScenario runs a forecast with additional what-if transactions injected.
func (s *cashFlowForecastService) GetScenario(userID uint, req *dto.CashFlowScenarioRequest) (*dto.CashFlowForecastResponse, error) {
	days := req.Days
	if days == 0 {
		days = 30
	}

	currentBalance, err := s.getCurrentBalance(userID, req.AssetID)
	if err != nil {
		return nil, err
	}

	avgDailyExpense, err := s.getAvgDailyExpense(userID, req.AssetID)
	if err != nil {
		return nil, err
	}

	recurring, err := s.recurringRepo.FindActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.buildForecast(currentBalance, avgDailyExpense, days, recurring, req.Transactions, req.AssetID), nil
}

// ---- Core Forecast Builder ----

func (s *cashFlowForecastService) buildForecast(
	currentBalance float64,
	avgDailyExpense float64,
	days int,
	recurring []models.RecurringTransaction,
	scenarioTxns []dto.ScenarioTransaction,
	assetID *uint64,
) *dto.CashFlowForecastResponse {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDay := today.AddDate(0, 0, days)

	// Build a map of recurring events per day within [today, endDay]
	recurringEvents := s.expandRecurringEvents(recurring, today, endDay)

	// Build a map of scenario events per day
	scenarioEvents := make(map[string][]dto.ForecastEvent)
	for _, st := range scenarioTxns {
		dayKey := st.Date.Format("2006-01-02")
		scenarioEvents[dayKey] = append(scenarioEvents[dayKey], dto.ForecastEvent{
			Description:     st.Description,
			Amount:          st.Amount,
			TransactionType: st.TransactionType,
			Source:          "scenario",
		})
	}

	projections := make([]dto.ForecastDay, 0, days)
	runningBalance := currentBalance

	minBalance := currentBalance
	minBalanceDate := today.Format("2006-01-02")
	maxBalance := currentBalance
	maxBalanceDate := today.Format("2006-01-02")

	var negativeAlerts []dto.NegativeBalanceAlert

	for i := 0; i < days; i++ {
		day := today.AddDate(0, 0, i)
		dayKey := day.Format("2006-01-02")

		var events []dto.ForecastEvent
		dayIncome := 0
		dayExpense := 0

		// Add recurring events
		if re, ok := recurringEvents[dayKey]; ok {
			events = append(events, re...)
		}

		// Add scenario events
		if se, ok := scenarioEvents[dayKey]; ok {
			events = append(events, se...)
		}

		// Calculate recurring/scenario net
		for _, e := range events {
			if e.TransactionType == 1 {
				dayIncome += e.Amount
			} else {
				dayExpense += e.Amount
			}
		}

		// Add estimated variable spending (represents routine daily expenses such as
		// food and entertainment, separate from the known recurring transactions above).
		estimatedExpense := avgDailyExpense
		if estimatedExpense > 0 {
			dayExpense += int(estimatedExpense)
			events = append(events, dto.ForecastEvent{
				Description:     "Estimated daily spending",
				Amount:          int(estimatedExpense),
				TransactionType: 2,
				Source:          "estimated",
			})
		}

		runningBalance += float64(dayIncome) - float64(dayExpense)

		isNegative := runningBalance < 0

		if runningBalance < minBalance {
			minBalance = runningBalance
			minBalanceDate = dayKey
		}
		if runningBalance > maxBalance {
			maxBalance = runningBalance
			maxBalanceDate = dayKey
		}

		if isNegative {
			triggerEvent := "Balance goes negative"
			if len(events) > 0 {
				triggerEvent = events[len(events)-1].Description
			}
			negativeAlerts = append(negativeAlerts, dto.NegativeBalanceAlert{
				Date:             dayKey,
				ProjectedBalance: runningBalance,
				TriggerEvent:     triggerEvent,
			})
		}

		projections = append(projections, dto.ForecastDay{
			Date:             dayKey,
			Events:           events,
			DayIncome:        dayIncome,
			DayExpense:       dayExpense,
			ProjectedBalance: runningBalance,
			IsNegative:       isNegative,
		})
	}

	return &dto.CashFlowForecastResponse{
		AssetID:          assetID,
		CurrentBalance:   currentBalance,
		ForecastDays:     days,
		StartDate:        today.Format("2006-01-02"),
		EndDate:          endDay.Format("2006-01-02"),
		MinBalance:       minBalance,
		MinBalanceDate:   minBalanceDate,
		MaxBalance:       maxBalance,
		MaxBalanceDate:   maxBalanceDate,
		NegativeAlerts:   negativeAlerts,
		DailyProjections: projections,
		AvgDailyExpense:  avgDailyExpense,
	}
}

// expandRecurringEvents generates all occurrences of each active recurring transaction
// that fall within [from, to] and groups them by date key (YYYY-MM-DD).
func (s *cashFlowForecastService) expandRecurringEvents(
	recurring []models.RecurringTransaction,
	from, to time.Time,
) map[string][]dto.ForecastEvent {
	events := make(map[string][]dto.ForecastEvent)

	for _, r := range recurring {
		// Start from the stored next_occurrence
		cur := time.Date(r.NextOccurrence.Year(), r.NextOccurrence.Month(), r.NextOccurrence.Day(), 0, 0, 0, 0, from.Location())

		for !cur.After(to) {
			if !cur.Before(from) {
				// Check end_date
				if r.EndDate != nil && cur.After(r.EndDate.Time) {
					break
				}
				id := r.ID
				dayKey := cur.Format("2006-01-02")
				events[dayKey] = append(events[dayKey], dto.ForecastEvent{
					Description:     r.Description,
					Amount:          r.Amount,
					TransactionType: r.TransactionType,
					Source:          "recurring",
					RecurringID:     &id,
				})
			}
			cur = advanceOccurrence(r.Frequency, r.DayOfMonth, r.DayOfWeek, cur)
		}
	}

	return events
}

// getCurrentBalance sums the balance of all (or one) asset(s) for the user.
func (s *cashFlowForecastService) getCurrentBalance(userID uint, assetID *uint64) (float64, error) {
	if assetID != nil {
		asset, err := s.assetRepo.GetAssetByID(*assetID)
		if err != nil {
			return 0, err
		}
		return asset.Balance, nil
	}

	assets, err := s.assetRepo.GetAssetsByUser(uint64(userID))
	if err != nil {
		return 0, err
	}
	var total float64
	for _, a := range assets {
		total += a.Balance
	}
	return total, nil
}

// getAvgDailyExpense computes average daily expense from the last 90 days of transactions.
func (s *cashFlowForecastService) getAvgDailyExpense(userID uint, assetID *uint64) (float64, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -90)

	txns, err := s.transactionRepo.FindByDateRange(userID, from, now, assetID)
	if err != nil {
		return 0, nil // non-fatal; default to 0
	}

	var totalExpense int64
	for _, t := range txns {
		if t.TransactionType == 2 { // expense
			totalExpense += int64(t.Amount)
		}
	}

	if len(txns) == 0 {
		return 0, nil
	}

	return float64(totalExpense) / 90.0, nil
}
