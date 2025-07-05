package controllers

import (
    "my-api/config"
    "my-api/models"
    "github.com/gin-gonic/gin"
    "net/http"
     "my-api/utils"
)

func GetBank(c *gin.Context) {
    var banks []models.Bank
    config.DB.Find(&banks)
	utils.JSONSuccess(c, "Banks successfully get data", banks)
}

func CreateBank(c *gin.Context) {
    var bank models.Bank
    if err := c.ShouldBindJSON(&bank); err != nil {
         utils.JSONError(c, http.StatusBadRequest, "Failed to Create Bank")
        return
    }
    config.DB.Create(&bank)
    utils.JSONSuccess(c, "Bank Create successfully", bank)
}

func DeleteBank(c *gin.Context) {
    var bank models.Bank
    if err := config.DB.Where("id = ?", c.Param("id")).First(&bank).Error; err != nil {
        utils.JSONError(c, http.StatusBadRequest, "Bank not found")
        return
    }
    config.DB.Delete(&bank)
    utils.JSONSuccess(c, "Bank successfully deleted", bank)
}