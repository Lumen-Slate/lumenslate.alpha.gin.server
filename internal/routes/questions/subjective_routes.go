package questions

import (
	"lumenslate/internal/controller/questions"

	"github.com/gin-gonic/gin"
)

func RegisterSubjectiveRoutes(r *gin.RouterGroup) {
	s := r.Group("/subjectives")
	{
		s.GET("", questions.GetAllSubjectives)
		s.GET(":id", questions.GetSubjective)
		s.POST("", questions.CreateSubjective)
		s.PUT(":id", questions.UpdateSubjective)
		s.PATCH(":id", questions.PatchSubjective)
		s.DELETE(":id", questions.DeleteSubjective)
		s.POST("/bulk", questions.CreateBulkSubjectives)
	}
}
