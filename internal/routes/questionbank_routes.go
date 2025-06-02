package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterQuestionBankRoutes(r *gin.Engine) {
	q := r.Group("/question-banks")
	{
		q.GET("", controller.GetAllQuestionBanks)
		q.GET(":id", controller.GetQuestionBank)
		q.POST("", controller.CreateQuestionBank)
		q.PUT(":id", controller.UpdateQuestionBank)
		q.PATCH(":id", controller.PatchQuestionBank)
		q.DELETE(":id", controller.DeleteQuestionBank)
	}
}
