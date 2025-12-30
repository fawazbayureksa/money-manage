package repositories

import (
	"my-api/dto"
	"my-api/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll(filter *dto.UserFilterRequest) ([]models.User, int64, error)
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(filter *dto.UserFilterRequest) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}
	if filter.IsAdmin != nil {
		query = query.Where("is_admin = ?", *filter.IsAdmin)
	}
	if filter.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "id"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	query = query.Order(sortBy + " " + filter.SortDir)

	// Apply pagination
	query = query.Offset(filter.GetOffset()).Limit(filter.PageSize)

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}
