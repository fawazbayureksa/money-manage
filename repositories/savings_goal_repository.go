package repositories

import (
	"gorm.io/gorm"
	"my-api/dto"
	"my-api/models"
)

type SavingsGoalRepository interface {
	Create(goal *models.SavingsGoal) error
	FindByID(id uint, userID uint) (*models.SavingsGoal, error)
	FindAll(userID uint, filter *dto.SavingsGoalFilterRequest) ([]models.SavingsGoal, int64, error)
	Update(goal *models.SavingsGoal) error
	Delete(id uint, userID uint) error

	// Contributions
	CreateContribution(contribution *models.SavingsContribution) error
	FindContributionByID(id uint, goalID uint, userID uint) (*models.SavingsContribution, error)
	FindContributions(goalID uint, userID uint, filter *dto.ContributionFilterRequest) ([]models.SavingsContribution, int64, error)
	DeleteContribution(id uint, goalID uint, userID uint) error
}

type savingsGoalRepository struct {
	db *gorm.DB
}

func NewSavingsGoalRepository(db *gorm.DB) SavingsGoalRepository {
	return &savingsGoalRepository{db: db}
}

func (r *savingsGoalRepository) Create(goal *models.SavingsGoal) error {
	return r.db.Create(goal).Error
}

func (r *savingsGoalRepository) FindByID(id uint, userID uint) (*models.SavingsGoal, error) {
	var goal models.SavingsGoal
	err := r.db.Preload("Asset").
		Where("id = ? AND user_id = ?", id, userID).
		First(&goal).Error
	return &goal, err
}

func (r *savingsGoalRepository) FindAll(userID uint, filter *dto.SavingsGoalFilterRequest) ([]models.SavingsGoal, int64, error) {
	var goals []models.SavingsGoal
	var total int64

	query := r.db.Model(&models.SavingsGoal{}).Where("user_id = ?", userID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
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

	err := query.Preload("Asset").Find(&goals).Error
	return goals, total, err
}

func (r *savingsGoalRepository) Update(goal *models.SavingsGoal) error {
	return r.db.Save(goal).Error
}

func (r *savingsGoalRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.SavingsGoal{}).Error
}

func (r *savingsGoalRepository) CreateContribution(contribution *models.SavingsContribution) error {
	return r.db.Create(contribution).Error
}

func (r *savingsGoalRepository) FindContributionByID(id uint, goalID uint, userID uint) (*models.SavingsContribution, error) {
	var contribution models.SavingsContribution
	err := r.db.Where("id = ? AND goal_id = ? AND user_id = ?", id, goalID, userID).
		First(&contribution).Error
	return &contribution, err
}

func (r *savingsGoalRepository) FindContributions(goalID uint, userID uint, filter *dto.ContributionFilterRequest) ([]models.SavingsContribution, int64, error) {
	var contributions []models.SavingsContribution
	var total int64

	query := r.db.Model(&models.SavingsContribution{}).
		Where("goal_id = ? AND user_id = ?", goalID, userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "date"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	query = query.Order(sortBy + " " + filter.SortDir)
	query = query.Offset(filter.GetOffset()).Limit(filter.PageSize)

	err := query.Find(&contributions).Error
	return contributions, total, err
}

func (r *savingsGoalRepository) DeleteContribution(id uint, goalID uint, userID uint) error {
	return r.db.Where("id = ? AND goal_id = ? AND user_id = ?", id, goalID, userID).
		Delete(&models.SavingsContribution{}).Error
}
