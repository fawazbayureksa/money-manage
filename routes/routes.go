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

	// Initialize services
	userService := services.NewUserService(userRepo)
	bankService := services.NewBankService(bankRepo)

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	bankController := controllers.NewBankController(bankService)

	api := router.Group("/api")
	{
		// Auth routes (no changes needed for now)
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)

		// User routes with new controller structure
		api.GET("/users", userController.GetUsers)
		api.POST("/users", userController.CreateUser)
		api.PUT("/users/:id", userController.UpdateUser)
		api.DELETE("/users/:id", userController.DeleteUser)

		// Bank routes with new controller structure
		api.GET("/banks", bankController.GetBanks)
		api.POST("/banks", bankController.CreateBank)
		api.DELETE("/banks/:id", bankController.DeleteBank)

		// Category routes (no changes for now)
		api.GET("/categories", controllers.GetCategories)
		api.GET("/transaction/initial-data", controllers.GetInitialData)
		api.DELETE("/categories/:id", controllers.DeleteCategory)
	}

	// Protected routes
	authorized := router.Group("/api")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("/transaction", controllers.CreateTransaction)
		authorized.GET("/my-categories", controllers.GetCategoriesByUser)
		authorized.POST("/categories", controllers.CreateCategory)
	}
}
