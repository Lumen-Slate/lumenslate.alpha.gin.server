package questions

import (
	"lumenslate/internal/controller/questions"

	"github.com/gin-gonic/gin"
)

func RegisterMSQRoutes(r *gin.RouterGroup) {
	group := r.Group("/msqs")
	{
		group.GET("", questions.GetAllMSQs)
		group.GET(":id", questions.GetMSQ)
		group.POST("", questions.CreateMSQ)
		group.PUT(":id", questions.UpdateMSQ)
		group.PATCH(":id", questions.PatchMSQ)
		group.DELETE(":id", questions.DeleteMSQ)
		group.POST("/bulk", questions.CreateBulkMSQs)
	}
}
