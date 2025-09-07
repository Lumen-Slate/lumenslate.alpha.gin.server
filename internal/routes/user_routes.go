package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.RouterGroup) {
	user := router.Group("/users")
	{
		user.POST("", controller.CreateUser)
		user.GET(":id", controller.GetUser)
		user.PUT(":id", controller.UpdateUser)
		user.PATCH(":id", controller.PatchUser)
		user.DELETE(":id", controller.DeleteUser)
		user.GET("", controller.ListUsers)
	}
}
