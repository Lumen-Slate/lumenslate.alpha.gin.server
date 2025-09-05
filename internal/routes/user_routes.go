package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.RouterGroup) {
	user := router.Group("/users")
	userController := &controller.UserController{}
	{
		user.POST("/", userController.CreateUser)
		user.GET("/:id", userController.GetUser)
		user.PUT("/:id", userController.UpdateUser)
		user.PATCH("/:id", userController.PatchUser)
		user.DELETE("/:id", userController.DeleteUser)
		user.GET("/", userController.ListUsers)
	}
}
