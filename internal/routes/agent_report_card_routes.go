package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// SetupAgentReportCardRoutes sets up all routes for agent report cards
func SetupAgentReportCardRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/agent-report-cards", controller.GetAllAgentReportCards)
		api.GET("/agent-report-cards/:id", controller.GetAgentReportCardByID)
		api.GET("/agent-report-cards/student/:studentId", controller.GetAgentReportCardsByStudentID)
		api.POST("/agent-report-cards", controller.CreateAgentReportCard)
		api.PUT("/agent-report-cards/:id", controller.UpdateAgentReportCard)
		api.DELETE("/agent-report-cards/:id", controller.DeleteAgentReportCard)
	}
}
