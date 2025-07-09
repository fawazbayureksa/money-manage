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

func CreateCategory(c *gin.Context) {
    var categories models.Category
    if err := c.ShouldBindJSON(&categories); err != nil {
         utils.JSONError(c, http.StatusBadRequest, "Failed to Create Category")
        return
    }
    config.DB.Create(&categories)
    utils.JSONSuccess(c, "Category Create successfully", categories)
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