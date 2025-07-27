package routes

import (
    "github.com/gin-gonic/gin"
    "my-api/controllers"
    "my-api/middleware"
    // ... other imports
)

func SetupRouter(router *gin.Engine) {

    api := router.Group("/api")
    {
        api.POST("/register", controllers.Register)
        api.POST("/login", controllers.Login)

        api.GET("/users", controllers.GetUsers)
        api.POST("/users", controllers.CreateUser)
        api.PUT("/users/:id", controllers.UpdateUser)
        api.DELETE("/users/:id", controllers.DeleteUser)
        
        api.GET("/banks", controllers.GetBank)
        api.POST("/banks", controllers.CreateBank)
        api.DELETE("/banks/:id", controllers.DeleteBank)

        api.GET("/categories", controllers.GetCategories)
        api.GET("/transaction/initial-data", controllers.GetInitialData)
    }
    
    
    authorized := router.Group("/api")
    authorized.Use(middleware.AuthMiddleware())
    {
        authorized.POST("/transaction", controllers.CreateTransaction)
        authorized.GET("/my-categories", controllers.GetCategoriesByUser)
        authorized.POST("/categories", controllers.CreateCategory)
    }
    api.DELETE("/categories/:id", controllers.DeleteCategory)

    return 
}
