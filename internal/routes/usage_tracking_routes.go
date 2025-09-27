package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterUsageTrackingRoutes registers usage tracking routes
func RegisterUsageTrackingRoutes(router *gin.Engine) {
	usage := router.Group("/usage")
	{
		// Get all usage tracking records with filters
		usage.GET("", controller.GetAllUsageTracking) // Query params: user_id, period, limit, offset

		// Get usage summary by period
		usage.GET("/summary/:period", controller.GetUsageSummaryByPeriod) // Get aggregated usage for a period
	}

	// User-specific usage tracking routes
	users := router.Group("/users/:userId/usage")
	{
		// Get usage metrics and history
		users.GET("/current", controller.GetCurrentUsageMetrics)          // Get current period usage metrics
		users.GET("/aggregated", controller.GetAggregatedUsageMetrics)    // Get all-time aggregated usage metrics
		users.GET("/history", controller.GetAllUserUsageHistory)          // Get all usage history for user
		users.GET("/period/:period", controller.GetUsageTrackingByPeriod) // Get usage for specific period

		// Reset usage (admin function)
		users.POST("/reset", controller.ResetUserUsage) // Reset usage counters for current period

		// Track usage with JSON payload (detailed tracking)
		users.POST("/track/question-banks", controller.TrackQuestionBankUsage)         // Track question bank usage
		users.POST("/track/questions", controller.TrackQuestionUsage)                  // Track question usage
		users.POST("/track/ia-agent", controller.TrackIAUsage)                         // Track IA usage
		users.POST("/track/lumen-agent", controller.TrackLumenAgentUsage)              // Track Lumen Agent usage
		users.POST("/track/ra-agent", controller.TrackRAAgentUsage)                    // Track RA Agent usage
		users.POST("/track/recap-classes", controller.TrackRecapClassUsage)            // Track recap classes usage
		users.POST("/track/assignment-exports", controller.TrackAssignmentExportUsage) // Track assignment exports usage
		users.POST("/track/bulk", controller.TrackBulkUsage)                           // Track multiple usage types at once

		// Simple increment endpoints (GET requests for easy integration)
		users.POST("/increment/question-banks", controller.IncrementQuestionBankUsage)         // Increment question bank usage
		users.POST("/increment/questions", controller.IncrementQuestionUsage)                  // Increment question usage
		users.POST("/increment/ia-agent", controller.IncrementIAUsage)                         // Increment IA usage
		users.POST("/increment/lumen-agent", controller.IncrementLumenAgentUsage)              // Increment Lumen Agent usage
		users.POST("/increment/ra-agent", controller.IncrementRAAgentUsage)                    // Increment RA Agent usage
		users.POST("/increment/recap-classes", controller.IncrementRecapClassUsage)            // Increment recap classes usage
		users.POST("/increment/assignment-exports", controller.IncrementAssignmentExportUsage) // Increment assignment exports usage
	}
}
