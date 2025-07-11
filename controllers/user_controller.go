package controllers

import (
    "my-api/config"
    "my-api/models"
    "github.com/gin-gonic/gin"
    "net/http"
    "my-api/utils" 
)

func GetUsers(c *gin.Context) {
    var users []models.User
    config.DB.Find(&users)
    // c.JSON(http.StatusOK, users)
      utils.JSONSuccess(c, "User Get successfully", users)
}

func CreateUser(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
         utils.JSONError(c, http.StatusBadRequest, "Failed to fetch users")
        return
    }
    config.DB.Create(&user)
    utils.JSONSuccess(c, "User Create successfully", user)
}

func UpdateUser(c *gin.Context) {
    // Get the user ID from the URL
    id := c.Param("id")

    // Find the existing user
    var user models.User
    if err := config.DB.First(&user, id).Error; err != nil {
        utils.JSONError(c, http.StatusNotFound, "User not found")
        return
    }

    // Bind the JSON input to a new user struct
    var updatedUser models.User
    if err := c.ShouldBindJSON(&updatedUser); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Invalid input data")
        return
    }

    // Update the user fields
    user.Name = updatedUser.Name
    user.Email = updatedUser.Email
    user.Address = updatedUser.Address


    hashedPassword, err := utils.HashPassword(user.Password)
   
    if err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to hash password")
        return
    }

    user.Password = hashedPassword

    // Save the updated user to the database
    if err := config.DB.Save(&user).Error; err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to update user")
        return
    }

    // Return the updated user
    utils.JSONSuccess(c, "User updated successfully", user)
}

func DeleteUser(c *gin.Context) {
    // Get the user ID from the URL
    id := c.Param("id")

    // Find the user to delete
    var user models.User
    if err := config.DB.First(&user, id).Error; err != nil {
          utils.JSONError(c, http.StatusNotFound, "User not found")
        return
    }

    // Delete the user from the database
    if err := config.DB.Delete(&user).Error; err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to delete user")
        return
    }

    // Return a success message
     utils.JSONSuccess(c, "User Deleted successfully", nil)
}