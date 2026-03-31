package main

import (
	"fmt"
	"log"
	"time"

	"my-api/config"
	// "my-api/models"
	"my-api/repositories"
	"my-api/routes"
	"my-api/services"
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

	// Start background scheduler for daily/weekly budget summaries
	go runBudgetScheduler()

	utils.LogInfo("Server starting on port 8080...")
	if err := r.Run(":8080"); err != nil {
		utils.LogErrorf("Failed to start server: %v", err)
		log.Fatal(err)
	}
}

// runBudgetScheduler runs daily and weekly budget summary generation in the background.
// Daily summaries fire once per day; weekly reports fire every Sunday.
func runBudgetScheduler() {
	budgetRepo := repositories.NewBudgetRepository(config.DB)
	budgetService := services.NewBudgetService(budgetRepo)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	var lastDailySummaryDate string
	var lastWeeklyReportWeek string

	for t := range ticker.C {
		now := t

		// Daily summary — run once a day after 20:00 (8 PM) server time
		dateKey := now.Format("2006-01-02")
		if now.Hour() >= 20 && dateKey != lastDailySummaryDate {
			if err := budgetService.GenerateDailySummaries(); err != nil {
				utils.LogErrorf("Daily summary generation failed: %v", err)
			} else {
				utils.LogInfof("Daily budget summaries generated for %s", dateKey)
				lastDailySummaryDate = dateKey
			}
		}

		// Weekly report — run once a week on Sunday after 20:00
		weekKey := now.Format("2006-W") + fmt.Sprintf("%02d", int(now.Weekday()))
		if now.Weekday() == time.Sunday && now.Hour() >= 20 && weekKey != lastWeeklyReportWeek {
			if err := budgetService.GenerateWeeklyReports(); err != nil {
				utils.LogErrorf("Weekly report generation failed: %v", err)
			} else {
				utils.LogInfof("Weekly budget reports generated for week of %s", dateKey)
				lastWeeklyReportWeek = weekKey
			}
		}
	}
}
