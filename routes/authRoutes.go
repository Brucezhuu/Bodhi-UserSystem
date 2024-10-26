package routes

import (
	"UserSystem/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/", controllers.Health)

	protected := r.Group("/protected")
	protected.Use(controllers.AuthMiddleware())
	{
		protected.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "This is a protected route"})
		})
	}
}
