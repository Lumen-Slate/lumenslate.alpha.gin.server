package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterUsageTrackingRoutes registers usage tracking routes
func RegisterUsageTrackingRoutes(router *gin.RouterGroup) {
	usage := router.Group("/usage")
	{
		// Get all usage tracking records with filters
		usage.GET("", controller.GetAllUsageTracking) // Query params: user_id, period, limit, offset

		// Get usage summary by period
		usage.GET("/summary/:period", controller.GetUsageSummaryByPeriod) // Get aggregated usage for a period

		// User-specific usage tracking routes under /usage
		usage.GET("/user/:id/current", controller.GetCurrentUsageMetrics)          // Get current period usage metrics
		usage.GET("/user/:id/aggregated", controller.GetAggregatedUsageMetrics)    // Get all-time aggregated usage metrics
		usage.GET("/user/:id/history", controller.GetAllUserUsageHistory)          // Get all usage history for user
		usage.GET("/user/:id/period/:period", controller.GetUsageTrackingByPeriod) // Get usage for specific period

		// Reset usage (admin function)
		usage.POST("/user/:id/reset", controller.ResetUserUsage) // Reset usage counters for current period

		// Track usage with JSON payload (detailed tracking)
		usage.POST("/user/:id/track/question-banks", controller.TrackQuestionBankUsage)         // Track question bank usage
		usage.POST("/user/:id/track/questions", controller.TrackQuestionUsage)                  // Track question usage
		usage.POST("/user/:id/track/ia-agent", controller.TrackIAUsage)                         // Track IA usage
		usage.POST("/user/:id/track/lumen-agent", controller.TrackLumenAgentUsage)              // Track Lumen Agent usage
		usage.POST("/user/:id/track/ra-agent", controller.TrackRAAgentUsage)                    // Track RA Agent usage
		usage.POST("/user/:id/track/recap-classes", controller.TrackRecapClassUsage)            // Track recap classes usage
		usage.POST("/user/:id/track/assignment-exports", controller.TrackAssignmentExportUsage) // Track assignment exports usage
		usage.POST("/user/:id/track/bulk", controller.TrackBulkUsage)                           // Track multiple usage types at once

		// Simple increment endpoints
		usage.POST("/user/:id/increment/question-banks", controller.IncrementQuestionBankUsage)         // Increment question bank usage
		usage.POST("/user/:id/increment/questions", controller.IncrementQuestionUsage)                  // Increment question usage
		usage.POST("/user/:id/increment/ia-agent", controller.IncrementIAUsage)                         // Increment IA usage
		usage.POST("/user/:id/increment/lumen-agent", controller.IncrementLumenAgentUsage)              // Increment Lumen Agent usage
		usage.POST("/user/:id/increment/ra-agent", controller.IncrementRAAgentUsage)                    // Increment RA Agent usage
		usage.POST("/user/:id/increment/recap-classes", controller.IncrementRecapClassUsage)            // Increment recap classes usage
		usage.POST("/user/:id/increment/assignment-exports", controller.IncrementAssignmentExportUsage) // Increment assignment exports usage
	}
}
