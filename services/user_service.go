package services

import (
	"errors"
	"my-api/dto"
	"my-api/models"
	"my-api/repositories"
	"my-api/utils"
	"gorm.io/gorm"
)

type UserService interface {
	GetAllUsers(filter *dto.UserFilterRequest) (*dto.PaginationResponse, error)
	GetUserByID(id uint) (*dto.UserResponse, error)
	CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id uint) error
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAllUsers(filter *dto.UserFilterRequest) (*dto.PaginationResponse, error) {
	filter.SetDefaults()

	users, total, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs (excluding passwords)
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(&user)
	}

	return dto.NewPaginationResponse(userResponses, filter.Page, filter.PageSize, total), nil
}

func (s *userService) GetUserByID(id uint) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Check if user already exists
	existingUser, _ := s.repo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Address:  req.Address,
		Password: hashedPassword,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if email is being changed and if it's already taken
	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := s.repo.FindByEmail(req.Email)
		if existingUser != nil {
			return nil, errors.New("email already in use")
		}
		user.Email = req.Email
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Address != "" {
		user.Address = req.Address
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

// Helper function to convert model to response DTO
func (s *userService) toUserResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Address:    user.Address,
		IsVerified: user.IsVerified,
		IsAdmin:    user.IsAdmin,
	}
}
