package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a Firebase authentication middleware
func AuthMiddleware(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// Extract the token (remove "Bearer " prefix)
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token is required",
			})
			c.Abort()
			return
		}

		// Verify the Firebase ID token
		token, err := authClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": fmt.Sprintf("Invalid token: %v", err),
			})
			c.Abort()
			return
		}

		// Set user information in the context for use in handlers
		c.Set("user_id", token.UID)
		c.Set("user_email", token.Claims["email"])
		c.Set("user_claims", token.Claims)
		c.Set("firebase_token", token)

		// Continue to the next handler
		c.Next()
	}
}

// OptionalAuthMiddleware creates a Firebase authentication middleware that doesn't block requests
// but sets user info if a valid token is provided
func OptionalAuthMiddleware(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header provided, continue without setting user info
			c.Next()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, continue without setting user info
			c.Next()
			return
		}

		// Extract the token (remove "Bearer " prefix)
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == "" {
			// Empty token, continue without setting user info
			c.Next()
			return
		}

		// Verify the Firebase ID token
		token, err := authClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			// Invalid token, continue without setting user info
			c.Next()
			return
		}

		// Set user information in the context for use in handlers
		c.Set("user_id", token.UID)
		c.Set("user_email", token.Claims["email"])
		c.Set("user_claims", token.Claims)
		c.Set("firebase_token", token)
		c.Set("authenticated", true)

		// Continue to the next handler
		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	uid, ok := userID.(string)
	return uid, ok
}

// GetUserEmail extracts the user email from the Gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	email, ok := userEmail.(string)
	return email, ok
}

// GetUserClaims extracts the user claims from the Gin context
func GetUserClaims(c *gin.Context) (map[string]interface{}, bool) {
	userClaims, exists := c.Get("user_claims")
	if !exists {
		return nil, false
	}
	claims, ok := userClaims.(map[string]interface{})
	return claims, ok
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	authenticated, exists := c.Get("authenticated")
	if !exists {
		// If not explicitly set, check if user_id exists
		_, exists := c.Get("user_id")
		return exists
	}
	auth, ok := authenticated.(bool)
	return ok && auth
}

// RequireAuth is a helper middleware that can be used to protect specific routes
func RequireAuth(authClient *auth.Client) gin.HandlerFunc {
	return AuthMiddleware(authClient)
}
