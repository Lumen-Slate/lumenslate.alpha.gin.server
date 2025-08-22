package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterVariableRoutes(r *gin.RouterGroup) {
	v := r.Group("/variables")
	{
		v.GET("", controller.GetAllVariables)
		v.GET(":id", controller.GetVariable)
		v.POST("", controller.CreateVariable)
		v.POST("/bulk", controller.CreateBulkVariables)
		v.PUT(":id", controller.UpdateVariable)
		v.PATCH(":id", controller.PatchVariable)
		v.DELETE(":id", controller.DeleteVariable)
	}
}
