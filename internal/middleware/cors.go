package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "Accept", "Accept-Encoding", "Accept-Language",
			"Connection", "Host", "Referer", "User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length", "Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 hours
	}
}

// CORSMiddleware creates a CORS middleware with custom configuration
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if len(config.AllowOrigins) > 0 {
			allowed := false
			for _, allowedOrigin := range config.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if origin != "" {
					c.Header("Access-Control-Allow-Origin", origin)
				} else if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] != "*" {
					c.Header("Access-Control-Allow-Origin", config.AllowOrigins[0])
				} else {
					c.Header("Access-Control-Allow-Origin", "*")
				}
			}
		}

		// Set allowed methods
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		// Set allowed headers
		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		// Set exposed headers
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// Set credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Set max age for preflight requests
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// DefaultCORSMiddleware creates a CORS middleware with default configuration
func DefaultCORSMiddleware() gin.HandlerFunc {
	return CORSMiddleware(DefaultCORSConfig())
}
