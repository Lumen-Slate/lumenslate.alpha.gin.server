package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterSubmissionRoutes(r *gin.RouterGroup) {
	sub := r.Group("/submissions")
	{
		sub.GET("", controller.GetAllSubmissions)
		sub.GET(":id", controller.GetSubmission)
		sub.POST("", controller.CreateSubmission)
		sub.PUT(":id", controller.UpdateSubmission)
		sub.PATCH(":id", controller.PatchSubmission)
		sub.DELETE(":id", controller.DeleteSubmission)
	}
}
