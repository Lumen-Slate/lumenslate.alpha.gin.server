package controller

import (
	"lumenslate/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCurrentUser returns the current authenticated user's information
func GetCurrentUser(c *gin.Context) {
	// Check if user is authenticated
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User is not authenticated",
		})
		return
	}

	// Get user information from context
	userID, _ := middleware.GetUserID(c)
	userEmail, _ := middleware.GetUserEmail(c)
	userClaims, _ := middleware.GetUserClaims(c)

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   userEmail,
		"claims":  userClaims,
		"message": "User authenticated successfully",
	})
}

// GetProfile returns user profile information (example of protected endpoint)
func GetProfile(c *gin.Context) {
	// This function assumes it's only called on protected routes
	// so authentication is guaranteed by middleware

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID from context",
		})
		return
	}

	userEmail, _ := middleware.GetUserEmail(c)

	// In a real application, you would fetch user profile from database
	// For now, we'll return the basic Firebase user info
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   userEmail,
		"profile": gin.H{
			"created_at": "2023-01-01",
			"status":     "active",
			// Add more profile fields as needed
		},
		"message": "Profile retrieved successfully",
	})
}

// UpdateProfile updates user profile information (example of protected endpoint)
func UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID from context",
		})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// In a real application, you would update the user profile in the database
	// For now, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "Profile updated successfully",
		"updated": updateData,
	})
}
