package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterThreadRoutes(r *gin.RouterGroup) {
	p := r.Group("/threads")
	{
		p.GET("", controller.GetAllThreads)
		p.GET(":id", controller.GetThread)
		p.POST("", controller.CreateThread)
		p.PUT(":id", controller.UpdateThread)
		p.PATCH(":id", controller.PatchThread)
		p.DELETE(":id", controller.DeleteThread)
	}
}
