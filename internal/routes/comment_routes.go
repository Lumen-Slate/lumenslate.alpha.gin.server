package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterCommentRoutes(r *gin.Engine) {
	comment := r.Group("/comments")
	{
		comment.GET("", controller.GetAllComments)
		comment.GET(":id", controller.GetComment)
		comment.POST("", controller.CreateComment)
		comment.PUT(":id", controller.UpdateComment)
		comment.PATCH(":id", controller.PatchComment)
		comment.DELETE(":id", controller.DeleteComment)
	}
}
