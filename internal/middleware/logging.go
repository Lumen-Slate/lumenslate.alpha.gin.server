package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string

	// SkipPathRegexps is a regex array which logs are not written if the request path matches.
	// Optional.
	SkipPathRegexps []string
}

// RequestLogger creates a request logging middleware with user information
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get user info if available
		userID, _ := GetUserID(c)
		userEmail, _ := GetUserEmail(c)

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Create log message
		logMsg := "[GIN] " + time.Now().Format("2006/01/02 - 15:04:05") + " | " +
			latency.String() + " | " +
			clientIP + " | " +
			method + " " +
			path + " | " +
			"Status: " + string(rune(statusCode))

		// Add user info if authenticated
		if userID != "" {
			logMsg += " | User: " + userID
			if userEmail != "" {
				logMsg += " (" + userEmail + ")"
			}
		}

		log.Println(logMsg)
	}
}

// RequestLoggerWithConfig creates a request logging middleware with custom config
func RequestLoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path should be skipped
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Start timer
		start := time.Now()
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get user info if available
		userID, _ := GetUserID(c)
		userEmail, _ := GetUserEmail(c)

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Create log message
		logMsg := "[GIN] " + time.Now().Format("2006/01/02 - 15:04:05") + " | " +
			latency.String() + " | " +
			clientIP + " | " +
			method + " " +
			path + " | " +
			"Status: " + string(rune(statusCode))

		// Add user info if authenticated
		if userID != "" {
			logMsg += " | User: " + userID
			if userEmail != "" {
				logMsg += " (" + userEmail + ")"
			}
		}

		log.Println(logMsg)
	}
}
