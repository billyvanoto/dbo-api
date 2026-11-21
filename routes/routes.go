package routes

import (
	"dbo-api/controllers"
	"dbo-api/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		// Public Auth Routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register) // Insert Login Data
			auth.POST("/login", controllers.Login)       // Get JWT Token
		}

		// Protected Routes
		protected := api.Group("/")
		protected.Use(middlewares.JWTAuth())
		{
			protected.GET("/auth/me", controllers.GetLoginData) // Get Login Data

			// Customer Management
			protected.GET("/customers", controllers.GetCustomers)
			protected.GET("/customers/options", controllers.GetCustomerOptions)
			protected.GET("/customers/:id", controllers.GetCustomerDetail)
			protected.POST("/customers", controllers.CreateCustomer)
			protected.PUT("/customers/:id", controllers.UpdateCustomer)
			protected.DELETE("/customers/:id", controllers.DeleteCustomer)

			// Order Management
			protected.GET("/orders", controllers.GetOrders)
			protected.GET("/orders/:id", controllers.GetOrderDetail)
			protected.POST("/orders", controllers.CreateOrder)
			protected.PUT("/orders/:id", controllers.UpdateOrder)
			protected.DELETE("/orders/:id", controllers.DeleteOrder)
		}
	}
	return r
}
