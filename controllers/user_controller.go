package controllers

import (
	"my-api/dto"
	"my-api/services"
	"my-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service services.UserService
}

func NewUserController(service services.UserService) *UserController {
	return &UserController{service: service}
}

// GetUsers godoc
// @Summary Get all users with pagination and filtering
// @Description Get users list with optional filters and pagination
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search by name or email"
// @Param name query string false "Filter by name"
// @Param email query string false "Filter by email"
// @Param is_admin query bool false "Filter by admin status"
// @Param sort_by query string false "Sort by field" default(id)
// @Param sort_dir query string false "Sort direction (asc/desc)" default(desc)
// @Success 200 {object} dto.PaginationResponse
// @Router /api/users [get]
func (ctrl *UserController) GetUsers(c *gin.Context) {
	var filter dto.UserFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	result, err := ctrl.service.GetAllUsers(&filter)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Users retrieved successfully", result)
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.UserResponse
// @Router /api/users [post]
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	user, err := ctrl.service.CreateUser(&req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "User created successfully", user)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update user details by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User data"
// @Success 200 {object} dto.UserResponse
// @Router /api/users/{id} [put]
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	user, err := ctrl.service.UpdateUser(uint(id), &req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "User updated successfully", user)
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200
// @Router /api/users/{id} [delete]
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := ctrl.service.DeleteUser(uint(id)); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, "User deleted successfully", nil)
}