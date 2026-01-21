package main

import (
    "my-api/config"
    "my-api/models"
    "my-api/routes"
    "github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }   
}   


func main() {
    config.ConnectDatabase()
    models.AutoMigrate()

    r := gin.Default()
   
    r.Use(CORSMiddleware())

    routes.SetupRouter(r)
    r.Run(":8081")
}
