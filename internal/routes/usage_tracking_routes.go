package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterUsageTrackingRoutes registers usage tracking routes
func RegisterUsageTrackingRoutes(router *gin.RouterGroup) {
	usage := router.Group("/usage")
	{
		// Get all usage records with filters
		usage.GET("", controller.GetAllUsage) // Query params: user_id, lookup_key, limit, offset

		// Get usage summary
		usage.GET("/summary", controller.GetUsageSummary) // Query params: lookup_key

		// Get usage by user ID
		usage.GET("/user/:id", controller.GetUsageByUser) // Get usage for specific user

		// User-specific usage tracking routes
		usage.GET("/user/:id/current", controller.GetCurrentUsageMetrics)          // Get current usage metrics
		usage.GET("/user/:id/aggregated", controller.GetAggregatedUsageMetrics)    // Get aggregated usage metrics (legacy)
		usage.GET("/user/:id/history", controller.GetAllUserUsageHistory)          // Get usage history for user
		usage.GET("/user/:id/period/:period", controller.GetUsageTrackingByPeriod) // Get usage for specific period (legacy)

		// Reset usage
		usage.POST("/user/:id/reset", controller.ResetUserUsage) // Reset usage counters

		// Track usage with JSON payload (detailed tracking)
		usage.POST("/user/:id/track/question-banks", controller.TrackQuestionBankUsage)         // Track question bank usage
		usage.POST("/user/:id/track/questions", controller.TrackQuestionUsage)                  // Track question usage
		usage.POST("/user/:id/track/ai-generation", controller.TrackAIGenerationUsage)          // Track AI generation usage
		usage.POST("/user/:id/track/assignment-exports", controller.TrackAssignmentExportUsage) // Track assignment exports usage
		usage.POST("/user/:id/track/bulk", controller.TrackBulkUsage)                           // Track multiple usage types at once

		// Legacy AI agent tracking endpoints (for backward compatibility)
		usage.POST("/user/:id/track/ia-agent", controller.TrackIAUsage)              // Track IA usage (legacy)
		usage.POST("/user/:id/track/lumen-agent", controller.TrackLumenAgentUsage)   // Track Lumen Agent usage (legacy)
		usage.POST("/user/:id/track/ra-agent", controller.TrackRAAgentUsage)         // Track RA Agent usage (legacy)
		usage.POST("/user/:id/track/recap-classes", controller.TrackRecapClassUsage) // Track recap classes usage (legacy)

		// Simple increment endpoints (useful for quick increments)
		usage.POST("/user/:id/increment/question-banks", controller.IncrementQuestionBankUsage)         // Increment question bank usage
		usage.POST("/user/:id/increment/questions", controller.IncrementQuestionUsage)                  // Increment question usage
		usage.POST("/user/:id/increment/ai-generation", controller.IncrementAIGenerationUsage)          // Increment AI generation usage
		usage.POST("/user/:id/increment/assignment-exports", controller.IncrementAssignmentExportUsage) // Increment assignment exports usage

		// Legacy increment endpoints (for backward compatibility)
		usage.POST("/user/:id/increment/ia-agent", controller.IncrementIAUsage)              // Increment IA usage (legacy)
		usage.POST("/user/:id/increment/lumen-agent", controller.IncrementLumenAgentUsage)   // Increment Lumen Agent usage (legacy)
		usage.POST("/user/:id/increment/ra-agent", controller.IncrementRAAgentUsage)         // Increment RA Agent usage (legacy)
		usage.POST("/user/:id/increment/recap-classes", controller.IncrementRecapClassUsage) // Increment recap classes usage (legacy)

		// Legacy routes for backward compatibility
		usage.GET("/summary/:period", controller.GetUsageSummaryByPeriod) // Get usage summary by period (legacy)

		// Delete usage (admin function)
		usage.DELETE("/:id", controller.DeleteUsage) // Delete usage record by ID
	}
}
