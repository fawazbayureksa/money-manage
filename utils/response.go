package utils

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

func JSONSuccess(c *gin.Context, message string, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

func JSONError(c *gin.Context, status int, message string) {
    c.JSON(status, Response{
        Success: false,
        Message: message,
        Data:    nil,
    })
}
