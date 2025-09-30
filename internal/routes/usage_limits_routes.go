package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterUsageLimitsRoutes registers usage limits management routes
func RegisterUsageLimitsRoutes(router *gin.RouterGroup) {
	usageLimits := router.Group("/usage-limits")
	{
		// Basic CRUD operations
		usageLimits.POST("", controller.CreateUsageLimits)       // Create new usage limits
		usageLimits.GET("/:id", controller.GetUsageLimits)       // Get usage limits by ID
		usageLimits.PUT("/:id", controller.UpdateUsageLimits)    // Update usage limits
		usageLimits.PATCH("/:id", controller.PatchUsageLimits)   // Patch usage limits
		usageLimits.DELETE("/:id", controller.DeleteUsageLimits) // Delete usage limits

		// Soft delete operation
		usageLimits.POST("/:id/deactivate", controller.SoftDeleteUsageLimits) // Soft delete (deactivate)

		// Plan-specific operations
		usageLimits.GET("/plan/:planName", controller.GetUsageLimitsByPlan) // Get usage limits by plan name

		// Query operations
		usageLimits.GET("", controller.GetAllUsageLimits)         // Get all usage limits with filters
		usageLimits.GET("/stats", controller.GetUsageLimitsStats) // Get usage limits statistics

		// User usage checking
		usageLimits.GET("/check/:userId", controller.CheckUserUsageAgainstLimits) // Check user usage against limits
	}

	// Admin operations
	admin := router.Group("/admin/usage-limits")
	{
		admin.POST("/initialize-defaults", controller.InitializeDefaultUsageLimits) // Initialize default usage limits
	}
}
