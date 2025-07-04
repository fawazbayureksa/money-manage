package controllers

import (
    "my-api/config"
    "my-api/models"
    "github.com/gin-gonic/gin"
    "net/http"
    "my-api/utils" 
)

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
    var user models.User

    if err := c.ShouldBindJSON(&user); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
        return
    }
    
    // Check if the user already exists
    
    if err := config.DB.Where("email =?", user.Email).First(&user).Error; err == nil {
        utils.JSONError(c, http.StatusConflict, "User already exists")
        return
    }
    
    // Hash the user's password
    hashedPassword, err := utils.HashPassword(user.Password)
   
    if err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to hash password")
        return
    }

    user.Password = hashedPassword

    // Save the user to the database
    config.DB.Create(&user)
    utils.JSONSuccess(c, "User registered successfully", user)
}

func Login(c *gin.Context) {
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

	// ✅ Compare input.Password with user.Password (hashed)
    if err := utils.CheckPasswordHash(input.Password,user.Password); !err {
        utils.JSONError(c, http.StatusUnauthorized, "Incorrect password")
        return
    }

    // Generate a JWT token for the user
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to generate token")
        return
	}

	utils.JSONSuccess(c, "Login successful", gin.H{"user": user, "token": token})
}
