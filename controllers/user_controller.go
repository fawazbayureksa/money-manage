package controllers

import (
    "my-api/config"
    "my-api/models"
    "github.com/gin-gonic/gin"
    "net/http"
)

func GetUsers(c *gin.Context) {
    var users []models.User
    config.DB.Find(&users)
    c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    config.DB.Create(&user)
    c.JSON(http.StatusCreated, user)
}

func UpdateUser(c *gin.Context) {
    // Get the user ID from the URL
    id := c.Param("id")

    // Find the existing user
    var user models.User
    if err := config.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Bind the JSON input to a new user struct
    var updatedUser models.User
    if err := c.ShouldBindJSON(&updatedUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update the user fields
    user.Name = updatedUser.Name
    user.Email = updatedUser.Email
    user.Address = updatedUser.Address

    // Save the updated user to the database
    if err := config.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    // Return the updated user
    c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
    // Get the user ID from the URL
    id := c.Param("id")

    // Find the user to delete
    var user models.User
    if err := config.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Delete the user from the database
    if err := config.DB.Delete(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    // Return a success message
    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
