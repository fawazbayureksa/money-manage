package controllers

import (
    "my-api/config"
    "my-api/models"
    "github.com/gin-gonic/gin"
    "net/http"
     "my-api/utils"
)

func GetCategories(c *gin.Context) {
    var categories []models.Category
    config.DB.Find(&categories)
	utils.JSONSuccess(c, "Categories successfully get data", categories)
}
func GetCategoriesByUser(c *gin.Context) {
    // Get the user ID from the JWT token
    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }

    var categories []models.Category
    if err := config.DB.Where("user_id = ?", userID).Find(&categories).Error; err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to fetch categories")
        return
    }

    utils.JSONSuccess(c, "Categories successfully retrieved", categories)
}

func CreateCategory(c *gin.Context) {
    var category models.Category
    if err := c.ShouldBindJSON(&category); err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Failed to parse category data")
        return
    }

    userID, exists := c.Get("user_id")
    if !exists {
        utils.JSONError(c, http.StatusUnauthorized, "User not authenticated")
        return
    }

    // Convert userID to uint
    userIDUint, ok := userID.(uint)
    if !ok {
        utils.JSONError(c, http.StatusInternalServerError, "Invalid user ID")
        return
    }

    category.UserID = userIDUint

    if err := config.DB.Create(&category).Error; err != nil {
        utils.JSONError(c, http.StatusInternalServerError, "Failed to create category")
        return
    }

    utils.JSONSuccess(c, "Category created successfully", category)
}

func DeleteCategory(c *gin.Context) {
    var categories models.Category
    if err := config.DB.Where("id = ?", c.Param("id")).First(&categories).Error; err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Categories not found")
        return
    }
    config.DB.Delete(&categories)
    utils.JSONSuccess(c, "Category successfully deleted", categories)
}