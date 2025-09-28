package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterSubscriptionRoutes registers subscription management routes
func RegisterSubscriptionRoutes(router *gin.RouterGroup) {
	subscriptions := router.Group("/subscriptions")
	{
		// Basic CRUD operations
		subscriptions.POST("", controller.CreateSubscription)       // Create new subscription
		subscriptions.GET("/:id", controller.GetSubscription)       // Get subscription by ID
		subscriptions.PUT("/:id", controller.UpdateSubscription)    // Update subscription
		subscriptions.DELETE("/:id", controller.CancelSubscription) // Cancel subscription immediately

		// Subscription management operations
		subscriptions.POST("/:id/schedule-cancellation", controller.ScheduleSubscriptionCancellation) // Schedule cancellation at period end
		subscriptions.POST("/:id/reactivate", controller.ReactivateSubscription)                      // Reactivate scheduled-to-cancel subscription
		subscriptions.POST("/:id/renew", controller.RenewSubscription)                                // Renew subscription for next period

		// Query operations
		subscriptions.GET("", controller.GetSubscriptionsByStatus)   // Get subscriptions by status (query param: status)
		subscriptions.GET("/stats", controller.GetSubscriptionStats) // Get subscription statistics

		// User-specific subscription routes under /subscriptions
		subscriptions.GET("/user/:id", controller.GetUserSubscription)                // Get active subscription for user
		subscriptions.GET("/user/:id/all", controller.GetAllUserSubscriptions)        // Get all subscriptions for user
		subscriptions.GET("/user/:id/status", controller.CheckUserSubscriptionStatus) // Check if user is subscribed
	}

	// Admin operations
	admin := router.Group("/admin/subscriptions")
	{
		admin.POST("/process-expired", controller.ProcessExpiredSubscriptions) // Process expired subscriptions (cron job endpoint)
	}
}
