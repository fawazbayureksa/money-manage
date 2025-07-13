package controllers

import (
    "my-api/config"
    "my-api/models"
    "github.com/gin-gonic/gin"
     "my-api/utils"
)

func GetInitialData(c *gin.Context) {
	var banks []models.Bank
    var categories []models.Category
    var users []models.User

    config.DB.Find(&banks)
    config.DB.Find(&categories)
    config.DB.Find(&users)

    data := gin.H{
        "banks": banks,
        "categories": categories,
        "users": users,
    }

    utils.JSONSuccess(c, "Initial data successfully fetched", data)
}