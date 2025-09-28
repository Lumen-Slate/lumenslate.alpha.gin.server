package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		// Public auth routes (with optional authentication)
		auth.GET("/me", controller.GetCurrentUser) // Get current user info if authenticated
	}
}

// RegisterProtectedAuthRoutes registers routes that require authentication
func RegisterProtectedAuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		// Protected auth routes (require authentication)
		auth.GET("/profile", controller.GetProfile)    // Get user profile (protected)
		auth.PUT("/profile", controller.UpdateProfile) // Update user profile (protected)
	}
}
