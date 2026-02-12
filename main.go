package main

import (
	"log"
	"time"

	"my-api/config"
	// "my-api/models"
	"my-api/routes"
	"my-api/utils"

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

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		utils.LogInfof("%s | %3d | %13v | %15s | %s",
			method,
			statusCode,
			latency,
			clientIP,
			path,
		)
	}
}

func main() {
	// Initialize logger
	utils.InitLogger()

	utils.LogInfo("Starting Money Manage API...")

	config.ConnectDatabase()
	// models.AutoMigrate()

	r := gin.Default()

	r.Use(CORSMiddleware())
	r.Use(LoggerMiddleware())

	routes.SetupRouter(r)
	utils.LogInfo("Routes configured successfully")

	utils.LogInfo("Server starting on port 8080...")
	if err := r.Run(":8080"); err != nil {
		utils.LogErrorf("Failed to start server: %v", err)
		log.Fatal(err)
	}
}
