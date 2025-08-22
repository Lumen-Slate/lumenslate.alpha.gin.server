package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// SetupSubjectReportRoutes sets up all subject report related routes
func SetupSubjectReportRoutes(router *gin.RouterGroup) {
	api := router.Group("/api")
	{
		// Subject Report routes
		api.GET("/subject-reports", controller.GetAllSubjectReportsHandler)
		api.GET("/subject-reports/:id", controller.GetSubjectReportByIDHandler)
		api.PUT("/subject-reports/:id", controller.UpdateSubjectReportHandler)
		api.DELETE("/subject-reports/:id", controller.DeleteSubjectReportHandler)

		// Student-specific subject reports
		api.GET("/students/:studentId/subject-reports", controller.GetSubjectReportsByStudentIDHandler)
	}
}
