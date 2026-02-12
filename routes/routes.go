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
	assetRepo := repositories.NewAssetRepository(config.DB)
	transactionV2Repo := repositories.NewTransactionV2Repository(config.DB)
	tagRepo := repositories.NewTagRepository(config.DB)

	// Initialize services
	userService := services.NewUserService(userRepo)
	bankService := services.NewBankService(bankRepo)
	budgetService := services.NewBudgetService(budgetRepo)
	analyticsService := services.NewAnalyticsService(analyticsRepo, budgetRepo)
	transactionService := services.NewTransactionService(transactionRepo)
	assetService := services.NewAssetService(assetRepo)
	transactionV2Service := services.NewTransactionV2Service(transactionV2Repo, assetRepo, tagRepo)
	tagService := services.NewTagService(tagRepo)

	// Initialize controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)
	bankController := controllers.NewBankController(bankService)
	budgetController := controllers.NewBudgetController(budgetService)
	analyticsController := controllers.NewAnalyticsController(analyticsService)
	transactionController := controllers.NewTransactionController(transactionService, budgetService)
	transactionV2Controller := controllers.NewTransactionV2Controller(transactionV2Service, budgetService)
	assetController := controllers.NewAssetController(assetService)
	tagController := controllers.NewTagController(tagService)

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

		// Transaction routes (v1 - Legacy, uses BankID)
		authorized.GET("/transactions", transactionController.GetTransactions)
		authorized.GET("/transactions/:id", transactionController.GetTransactionByID)
		authorized.POST("/transaction", transactionController.CreateTransaction)
		authorized.DELETE("/transactions/:id", transactionController.DeleteTransaction)

		// Transaction routes (v2 - New, uses AssetID with balance sync)
		v2 := authorized.Group("/v2")
		{
			// Transaction endpoints
			v2.GET("/transactions", transactionV2Controller.GetTransactions)
			v2.GET("/transactions/:id", transactionV2Controller.GetTransactionByID)
			v2.POST("/transactions", transactionV2Controller.CreateTransaction)
			v2.PUT("/transactions/:id", transactionV2Controller.UpdateTransaction)
			v2.DELETE("/transactions/:id", transactionV2Controller.DeleteTransaction)
			v2.GET("/assets/:id/transactions", transactionV2Controller.GetAssetTransactions)

			// Transaction tag endpoints
			v2.POST("/transactions/:id/tags", transactionV2Controller.AddTagsToTransaction)
			v2.DELETE("/transactions/:id/tags/:tag_id", transactionV2Controller.RemoveTagFromTransaction)

			// Tag management endpoints
			v2.GET("/tags", tagController.GetTags)
			v2.GET("/tags/:id", tagController.GetTagByID)
			v2.POST("/tags", tagController.CreateTag)
			v2.PUT("/tags/:id", tagController.UpdateTag)
			v2.DELETE("/tags/:id", tagController.DeleteTag)
			v2.GET("/tags/suggest", tagController.SuggestTags)

			// Analytics endpoints
			v2.GET("/analytics/spending-by-tag", tagController.GetSpendingByTag)
		}

		// Category routes
		authorized.GET("/my-categories", controllers.GetCategoriesByUser)
		authorized.POST("/categories", controllers.CreateCategory)

		// Wallet routes (protected)
		authorized.GET("/wallets", assetController.ListAssets)
		authorized.GET("/wallets/:id", assetController.GetAsset)
		authorized.POST("/wallets", assetController.CreateAsset)
		authorized.PUT("/wallets/:id", assetController.UpdateAsset)
		authorized.DELETE("/wallets/:id", assetController.DeleteAsset)
		authorized.GET("/wallets/summary", assetController.Summary)

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
		authorized.GET("/analytics/spending-by-asset", analyticsController.GetSpendingByAsset)
		authorized.GET("/analytics/income-vs-expense", analyticsController.GetIncomeVsExpense)
		authorized.GET("/analytics/trend", analyticsController.GetTrendAnalysis)
		authorized.GET("/analytics/monthly-comparison", analyticsController.GetMonthlyComparison)
		authorized.GET("/analytics/yearly-report", analyticsController.GetYearlyReport)
		authorized.GET("/analytics/category-trend/:category_id", analyticsController.GetCategoryTrend)
	}
}
