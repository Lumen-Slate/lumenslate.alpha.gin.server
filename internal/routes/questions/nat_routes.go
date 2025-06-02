package questions

import (
	"lumenslate/internal/controller/questions"

	"github.com/gin-gonic/gin"
)

func RegisterNATRoutes(r *gin.Engine) {
	n := r.Group("/nats")
	{
		n.GET("", questions.GetAllNATs)
		n.GET(":id", questions.GetNAT)
		n.POST("", questions.CreateNAT)
		n.PUT(":id", questions.UpdateNAT)
		n.PATCH(":id", questions.PatchNAT)
		n.DELETE(":id", questions.DeleteNAT)
		n.POST("/bulk", questions.CreateBulkNATs)

	}
}
