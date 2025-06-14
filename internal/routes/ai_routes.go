package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterAIRoutes(r *gin.Engine) {
	ai := r.Group("/ai")
	{
		ai.POST("/generate-context", controller.GenerateContextHandler)
		ai.POST("/detect-variables", controller.DetectVariablesHandler)
		ai.POST("/segment-question", controller.SegmentQuestionHandler)
		ai.POST("/generate-mcq", controller.GenerateMCQVariationsHandler)
		ai.POST("/generate-msq", controller.GenerateMSQVariationsHandler)
		ai.POST("/filter-randomize", controller.FilterAndRandomizeHandler)
	}
}
