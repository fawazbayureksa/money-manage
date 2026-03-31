package repositories

import (
	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
	"time"
)

type DebtRepository interface {
	// Debt CRUD
	Create(debt *models.Debt) error
	FindByID(id uint, userID uint) (*models.Debt, error)
	FindAll(userID uint, filter *dto.DebtFilterRequest) ([]models.Debt, int64, error)
	Update(debt *models.Debt) error
	Delete(id uint, userID uint) error
	FindActiveDebts(userID uint) ([]models.Debt, error)

	// Debt Payments
	CreatePayment(payment *models.DebtPayment) error
	FindPaymentsByDebtID(debtID uint, userID uint, filter *dto.DebtPaymentFilterRequest) ([]models.DebtPayment, int64, error)
	GetTotalPaid(debtID uint) (int, error)
	GetTotalPaidForUser(userID uint) (int, error)
	DeletePayment(paymentID uint, userID uint) error

	// Milestones
	CreateMilestone(milestone *models.DebtMilestone) error
	FindMilestoneByDebtAndTarget(debtID uint, milestoneType string, targetValue int) (*models.DebtMilestone, error)
	GetUserMilestones(userID uint, unreadOnly bool) ([]models.DebtMilestone, error)
	MarkMilestoneAsRead(milestoneID uint, userID uint) error
	MarkAllMilestonesAsRead(userID uint) error
}

type debtRepository struct {
	db *gorm.DB
}

func NewDebtRepository(db *gorm.DB) DebtRepository {
	return &debtRepository{db: db}
}

// --- Debt CRUD ---

func (r *debtRepository) Create(debt *models.Debt) error {
	return r.db.Create(debt).Error
}

func (r *debtRepository) FindByID(id uint, userID uint) (*models.Debt, error) {
	var debt models.Debt
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		First(&debt).Error
	return &debt, err
}

func (r *debtRepository) FindAll(userID uint, filter *dto.DebtFilterRequest) ([]models.Debt, int64, error) {
	var debts []models.Debt
	var total int64

	query := r.db.Model(&models.Debt{}).Where("user_id = ?", userID)

	if filter.DebtType != "" {
		query = query.Where("debt_type = ?", filter.DebtType)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		query = query.Where("name LIKE ? OR lender_name LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	query = query.Order(sortBy + " " + filter.SortDir)
	query = query.Offset(filter.GetOffset()).Limit(filter.PageSize)

	err := query.Find(&debts).Error
	return debts, total, err
}

func (r *debtRepository) Update(debt *models.Debt) error {
	return r.db.Save(debt).Error
}

func (r *debtRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Debt{}).Error
}

func (r *debtRepository) FindActiveDebts(userID uint) ([]models.Debt, error) {
	var debts []models.Debt
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("interest_rate DESC").
		Find(&debts).Error
	return debts, err
}

// --- Debt Payments ---

func (r *debtRepository) CreatePayment(payment *models.DebtPayment) error {
	return r.db.Create(payment).Error
}

func (r *debtRepository) FindPaymentsByDebtID(debtID uint, userID uint, filter *dto.DebtPaymentFilterRequest) ([]models.DebtPayment, int64, error) {
	var payments []models.DebtPayment
	var total int64

	query := r.db.Model(&models.DebtPayment{}).Where("debt_id = ? AND user_id = ?", debtID, userID)

	if filter.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", filter.StartDate)
		if err == nil {
			query = query.Where("payment_date >= ?", startDate)
		}
	}
	if filter.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", filter.EndDate)
		if err == nil {
			query = query.Where("payment_date <= ?", endDate)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "payment_date"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	query = query.Order(sortBy + " " + filter.SortDir)
	query = query.Offset(filter.GetOffset()).Limit(filter.PageSize)

	err := query.Find(&payments).Error
	return payments, total, err
}

func (r *debtRepository) GetTotalPaid(debtID uint) (int, error) {
	var total int64
	err := r.db.Model(&models.DebtPayment{}).
		Where("debt_id = ?", debtID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return int(total), err
}

func (r *debtRepository) GetTotalPaidForUser(userID uint) (int, error) {
	var total int64
	err := r.db.Model(&models.DebtPayment{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return int(total), err
}

func (r *debtRepository) DeletePayment(paymentID uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", paymentID, userID).Delete(&models.DebtPayment{}).Error
}

// --- Milestones ---

func (r *debtRepository) CreateMilestone(milestone *models.DebtMilestone) error {
	return r.db.Create(milestone).Error
}

func (r *debtRepository) FindMilestoneByDebtAndTarget(debtID uint, milestoneType string, targetValue int) (*models.DebtMilestone, error) {
	var milestone models.DebtMilestone
	err := r.db.Where("debt_id = ? AND milestone_type = ? AND target_value = ?",
		debtID, milestoneType, targetValue).First(&milestone).Error
	return &milestone, err
}

func (r *debtRepository) GetUserMilestones(userID uint, unreadOnly bool) ([]models.DebtMilestone, error) {
	var milestones []models.DebtMilestone
	query := r.db.Preload("Debt").Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	err := query.Order("created_at DESC").Find(&milestones).Error
	return milestones, err
}

func (r *debtRepository) MarkMilestoneAsRead(milestoneID uint, userID uint) error {
	return r.db.Model(&models.DebtMilestone{}).
		Where("id = ? AND user_id = ?", milestoneID, userID).
		Update("is_read", true).Error
}

func (r *debtRepository) MarkAllMilestonesAsRead(userID uint) error {
	return r.db.Model(&models.DebtMilestone{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
