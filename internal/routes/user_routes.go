package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine, uc *controller.UserController) {
	user := r.Group("/api/v1/users")
	{
		user.POST("/", uc.CreateUser)
		user.GET("/:id", uc.GetUser)
		user.PUT("/:id", uc.UpdateUser)
		user.DELETE("/:id", uc.DeleteUser)
		user.GET("/", uc.ListUsers)
	}
}
