package repositories

import (
	"my-api/models"

	"gorm.io/gorm"
)

type DebtRepository interface {
	// Debt CRUD
	Create(debt *models.Debt) error
	FindByID(id, userID uint) (*models.Debt, error)
	FindAllByUser(userID uint, status string) ([]models.Debt, error)
	FindActiveDebts(userID uint) ([]models.Debt, error)
	FindAllActiveWithAutoTrack() ([]models.Debt, error)
	Update(debt *models.Debt) error
	Delete(id, userID uint) error

	// Payments
	CreatePayment(payment *models.DebtPayment) error
	FindPayments(debtID uint) ([]models.DebtPayment, error)

	// Milestones
	CreateMilestone(milestone *models.DebtMilestone) error
	FindMilestone(debtID uint, milestoneType string) (*models.DebtMilestone, error)

	// Background job helpers
	FindDebtsWithDueDateIn(days int) ([]models.Debt, error)
	ResetMonthlyPaidFlag(debtID uint) error
	AddInterest(debtID uint, amount int) error
}

type debtRepository struct {
	db *gorm.DB
}

func NewDebtRepository(db *gorm.DB) DebtRepository {
	return &debtRepository{db: db}
}

// ─── Debt CRUD ───────────────────────────────────────────────────────────────

func (r *debtRepository) Create(debt *models.Debt) error {
	return r.db.Create(debt).Error
}

func (r *debtRepository) FindByID(id, userID uint) (*models.Debt, error) {
	var debt models.Debt
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&debt).Error
	return &debt, err
}

func (r *debtRepository) FindAllByUser(userID uint, status string) ([]models.Debt, error) {
	var debts []models.Debt
	query := r.db.Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&debts).Error
	return debts, err
}

func (r *debtRepository) FindActiveDebts(userID uint) ([]models.Debt, error) {
	var debts []models.Debt
	err := r.db.Where("user_id = ? AND status = ?", userID, "active").
		Order("interest_rate DESC").
		Find(&debts).Error
	return debts, err
}

func (r *debtRepository) FindAllActiveWithAutoTrack() ([]models.Debt, error) {
	var debts []models.Debt
	err := r.db.Where("status = ? AND auto_track_interest = ?", "active", true).
		Find(&debts).Error
	return debts, err
}

func (r *debtRepository) Update(debt *models.Debt) error {
	return r.db.Save(debt).Error
}

func (r *debtRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Debt{}).Error
}

// ─── Payments ────────────────────────────────────────────────────────────────

func (r *debtRepository) CreatePayment(payment *models.DebtPayment) error {
	return r.db.Create(payment).Error
}

func (r *debtRepository) FindPayments(debtID uint) ([]models.DebtPayment, error) {
	var payments []models.DebtPayment
	err := r.db.Where("debt_id = ?", debtID).
		Order("payment_date DESC").
		Find(&payments).Error
	return payments, err
}

// ─── Milestones ──────────────────────────────────────────────────────────────

func (r *debtRepository) CreateMilestone(milestone *models.DebtMilestone) error {
	return r.db.Create(milestone).Error
}

func (r *debtRepository) FindMilestone(debtID uint, milestoneType string) (*models.DebtMilestone, error) {
	var milestone models.DebtMilestone
	err := r.db.Where("debt_id = ? AND milestone_type = ?", debtID, milestoneType).First(&milestone).Error
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// ─── Background Job Helpers ───────────────────────────────────────────────────

// FindDebtsWithDueDateIn returns active debts whose payment_due_day falls within
// the next `days` days from today and have not been paid this month yet.
func (r *debtRepository) FindDebtsWithDueDateIn(days int) ([]models.Debt, error) {
	var debts []models.Debt

	// We check whether payment_due_day - DAY(NOW()) is between 0 and days
	err := r.db.Where(
		"status = ? AND current_month_paid = ? AND (payment_due_day - DAY(NOW())) BETWEEN 0 AND ?",
		"active", false, days,
	).Find(&debts).Error
	return debts, err
}

func (r *debtRepository) ResetMonthlyPaidFlag(debtID uint) error {
	return r.db.Model(&models.Debt{}).
		Where("id = ?", debtID).
		Update("current_month_paid", false).Error
}

func (r *debtRepository) AddInterest(debtID uint, amount int) error {
	return r.db.Model(&models.Debt{}).
		Where("id = ?", debtID).
		UpdateColumn("current_balance", gorm.Expr("current_balance + ?", amount)).Error
}
