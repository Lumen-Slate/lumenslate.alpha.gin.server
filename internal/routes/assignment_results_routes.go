package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// SetupAssignmentResultsRoutes sets up all assignment results related routes
func SetupAssignmentResultsRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Assignment Results routes
		api.GET("/assignment-results", controller.GetAllAssignmentResultsHandler)
		api.GET("/assignment-results/:id", controller.GetAssignmentResultByIDHandler)
		// api.POST("/assignment-results", controller.CreateAssignmentResultHandler)
		api.PUT("/assignment-results/:id", controller.UpdateAssignmentResultHandler)
		api.DELETE("/assignment-results/:id", controller.DeleteAssignmentResultHandler)
	}
}
