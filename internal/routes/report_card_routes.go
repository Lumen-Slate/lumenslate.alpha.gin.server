package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// SetupReportCardRoutes sets up all report card related routes
func SetupReportCardRoutes(router *gin.RouterGroup) {
	api := router.Group("/api")
	{
		// Report Card routes
		api.GET("/report-cards", controller.GetAllReportCardsHandler)
		api.GET("/report-cards/:id", controller.GetReportCardByIDHandler)
		api.PUT("/report-cards/:id", controller.UpdateReportCardHandler)
		api.DELETE("/report-cards/:id", controller.DeleteReportCardHandler)
	}
}
