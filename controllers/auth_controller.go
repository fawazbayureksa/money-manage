package controllers

import (
    "my-api/config"
    "my-api/dto"
    "my-api/models"
    "my-api/services"
    "github.com/gin-gonic/gin"
    "net/http"
    "my-api/utils" 
)

type AuthController struct {
    userService services.UserService
}

func NewAuthController(userService services.UserService) *AuthController {
    return &AuthController{userService: userService}
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func (ctrl *AuthController) Register(c *gin.Context) {
    var req dto.CreateUserRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
        return
    }

    // Use the user service to create the user (handles password hashing)
    user, err := ctrl.userService.CreateUser(&req)
    if err != nil {
        if err.Error() == "user with this email already exists" {
            utils.JSONError(c, http.StatusConflict, err.Error())
            return
        }
        utils.JSONError(c, http.StatusInternalServerError, err.Error())
        return
    }

    utils.JSONSuccess(c, "User registered successfully", user)
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
        return
    }

	// Find the user by email
	var user models.User
    if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
        utils.JSONError(c, http.StatusUnauthorized, "Invalid email")
        return
    }

	// Compare input.Password with user.Password (hashed)
    if !utils.CheckPasswordHash(input.Password, user.Password) {
        utils.JSONError(c, http.StatusUnauthorized, "Incorrect password")
        return
    }

    // Generate a JWT token for the user
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to generate token")
        return
	}

	// Create response without password
	userResponse := dto.UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Address:    user.Address,
		IsVerified: user.IsVerified,
		IsAdmin:    user.IsAdmin,
	}

	utils.JSONSuccess(c, "Login successful", gin.H{"user": userResponse, "token": token})
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	// In JWT-based auth, logout is typically handled client-side by removing the token
	// This endpoint confirms the logout action
	userID, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Optionally, you can log the logout action or invalidate refresh tokens here
	// For now, we just return a success response
	utils.JSONSuccess(c, "Logout successful", gin.H{"user_id": userID})
}
