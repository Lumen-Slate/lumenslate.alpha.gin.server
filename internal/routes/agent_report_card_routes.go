package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// SetupAgentReportCardRoutes sets up all routes for agent report cards
func SetupAgentReportCardRoutes(router *gin.Engine) {
	agentReportCardRoutes := router.Group("/api/agent-report-cards")
	{
		agentReportCardRoutes.GET("/", controller.GetAllAgentReportCards)
		agentReportCardRoutes.GET("/:id", controller.GetAgentReportCardByID)
		agentReportCardRoutes.GET("/student/:studentId", controller.GetAgentReportCardsByStudentID)
		// agentReportCardRoutes.POST("/", controller.CreateAgentReportCard)
		agentReportCardRoutes.PUT("/:id", controller.UpdateAgentReportCard)
		agentReportCardRoutes.DELETE("/:id", controller.DeleteAgentReportCard)
	}
}
