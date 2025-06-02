package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterVariableRoutes(r *gin.Engine) {
	v := r.Group("/variables")
	{
		v.GET("", controller.GetAllVariables)
		v.GET(":id", controller.GetVariable)
		v.POST("", controller.CreateVariable)
		v.POST("/bulk", controller.CreateBulkVariables) // Bulk creation route
		v.PUT(":id", controller.UpdateVariable)
		v.PATCH(":id", controller.PatchVariable)
		v.DELETE(":id", controller.DeleteVariable)
	}
}
