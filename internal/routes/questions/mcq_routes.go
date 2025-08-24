// routes/questions/mcq_routes.go
package questions

import (
	"lumenslate/internal/controller/questions"

	"github.com/gin-gonic/gin"
)

func RegisterMCQRoutes(r *gin.RouterGroup) {
	group := r.Group("/mcqs")
	{
		group.GET("", questions.GetAllMCQs)
		group.GET("/:id", questions.GetMCQ)
		group.POST("", questions.CreateMCQ)
		group.PUT("/:id", questions.UpdateMCQ)
		group.PATCH("/:id", questions.PatchMCQ)
		group.DELETE("/:id", questions.DeleteMCQ)
		group.POST("/bulk", questions.CreateBulkMCQs)
	}
}
