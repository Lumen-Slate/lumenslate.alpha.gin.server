package routes

import (
	"lumenslate/internal/controller"
	"lumenslate/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers health check and monitoring routes
func RegisterHealthRoutes(router *gin.Engine, metricsCollector *service.MetricsCollector, startTime time.Time) {
	healthController := controller.NewHealthController(metricsCollector, startTime)

	// Health check routes
	health := router.Group("/health")
	{
		health.GET("", healthController.BasicHealthHandler)                                      // Basic health check
		health.GET("/live", healthController.LivenessHandler)                                    // Kubernetes liveness probe
		health.GET("/ready", healthController.ReadinessHandler)                                  // Kubernetes readiness probe
		health.GET("/background-processing", healthController.BackgroundProcessingHealthHandler) // Detailed background processing health
		health.GET("/metrics", healthController.MetricsHandler)                                  // System metrics
		health.GET("/metrics/task/:taskType", healthController.TaskMetricsHandler)               // Task-specific metrics
	}
}
