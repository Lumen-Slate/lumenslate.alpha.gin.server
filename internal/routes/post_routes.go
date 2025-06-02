package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(r *gin.Engine) {
	p := r.Group("/posts")
	{
		p.GET("", controller.GetAllPosts)
		p.GET(":id", controller.GetPost)
		p.POST("", controller.CreatePost)
		p.PUT(":id", controller.UpdatePost)
		p.PATCH(":id", controller.PatchPost)
		p.DELETE(":id", controller.DeletePost)
	}
}
