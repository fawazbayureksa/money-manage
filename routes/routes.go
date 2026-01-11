package routes

import (
	"github.com/gin-gonic/gin"
	"my-api/config"
	"my-api/controllers"
	"my-api/middleware"
	"my-api/repositories"
	"my-api/services"
)

func SetupRouter(router *gin.Engine) {
	// Initialize repositories
	userRepo := repositories.NewUserRepository(config.DB)
	bankRepo := repositories.NewBankRepository(config.DB)
	budgetRepo := repositories.NewBudgetRepository(config.DB)
	analyticsRepo := repositories.NewAnalyticsRepository(config.DB)
	transactionRepo := repositories.NewTransactionRepository(config.DB)

	// Initialize services
	userService := services.NewUserService(userRepo)
	bankService := services.NewBankService(bankRepo)
	budgetService := services.NewBudgetService(budgetRepo)
	analyticsService := services.NewAnalyticsService(analyticsRepo, budgetRepo)
	transactionService := services.NewTransactionService(transactionRepo)

	// Initialize controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)
	bankController := controllers.NewBankController(bankService)
	budgetController := controllers.NewBudgetController(budgetService)
	analyticsController := controllers.NewAnalyticsController(analyticsService)
	transactionController := controllers.NewTransactionController(transactionService, budgetService)

	api := router.Group("/api")
	{
		// Auth routes
		api.POST("/register", authController.Register)
		api.POST("/login", authController.Login)

		// User routes
		api.GET("/users", userController.GetUsers)
		api.POST("/users", userController.CreateUser)
		api.PUT("/users/:id", userController.UpdateUser)
		api.DELETE("/users/:id", userController.DeleteUser)

		// Bank routes
		api.GET("/banks", bankController.GetBanks)
		api.POST("/banks", bankController.CreateBank)
		api.DELETE("/banks/:id", bankController.DeleteBank)

		// Category routes
		api.GET("/categories", controllers.GetCategories)
		api.GET("/transaction/initial-data", controllers.GetInitialData)
		api.DELETE("/categories/:id", controllers.DeleteCategory)
	}

	// Protected routes
	authorized := router.Group("/api")
	authorized.Use(middleware.AuthMiddleware())
	{
		// Auth routes (protected)
		authorized.POST("/logout", authController.Logout)

		// Transaction routes
		authorized.GET("/transactions", transactionController.GetTransactions)
		authorized.GET("/transactions/:id", transactionController.GetTransactionByID)
		authorized.POST("/transaction", transactionController.CreateTransaction)
		authorized.DELETE("/transactions/:id", transactionController.DeleteTransaction)
		
		// Category routes
		authorized.GET("/my-categories", controllers.GetCategoriesByUser)
		authorized.POST("/categories", controllers.CreateCategory)

		// Budget routes
		authorized.POST("/budgets", budgetController.CreateBudget)
		authorized.GET("/budgets", budgetController.GetBudgets)
		authorized.GET("/budgets/status", budgetController.GetBudgetStatus)
		authorized.GET("/budgets/:id", budgetController.GetBudget)
		authorized.PUT("/budgets/:id", budgetController.UpdateBudget)
		authorized.DELETE("/budgets/:id", budgetController.DeleteBudget)
		authorized.GET("/budget-alerts", budgetController.GetAlerts)
		authorized.PUT("/budget-alerts/:id/read", budgetController.MarkAlertAsRead)
		authorized.PUT("/budget-alerts/read-all", budgetController.MarkAllAlertsAsRead)

		// Analytics routes
		authorized.GET("/analytics/dashboard", analyticsController.GetDashboardSummary)
		authorized.GET("/analytics/spending-by-category", analyticsController.GetSpendingByCategory)
		authorized.GET("/analytics/spending-by-bank", analyticsController.GetSpendingByBank)
		authorized.GET("/analytics/income-vs-expense", analyticsController.GetIncomeVsExpense)
		authorized.GET("/analytics/trend", analyticsController.GetTrendAnalysis)
		authorized.GET("/analytics/monthly-comparison", analyticsController.GetMonthlyComparison)
		authorized.GET("/analytics/yearly-report", analyticsController.GetYearlyReport)
		authorized.GET("/analytics/category-trend/:category_id", analyticsController.GetCategoryTrend)
	}
}
