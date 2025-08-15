package controller

import (
	"fmt"
	"net/http"
	"time"

	"lumenslate/internal/service"
	"lumenslate/internal/utils"

	"github.com/gin-gonic/gin"
)

// HealthController handles health check endpoints
type HealthController struct {
	metricsCollector *service.MetricsCollector
	startTime        time.Time
	logger           *utils.Logger
}

// NewHealthController creates a new health controller
func NewHealthController(metricsCollector *service.MetricsCollector, startTime time.Time) *HealthController {
	return &HealthController{
		metricsCollector: metricsCollector,
		startTime:        startTime,
		logger:           utils.NewLogger("health_controller"),
	}
}

// BasicHealthHandler godoc
// @Summary      Basic Health Check
// @Description  Returns basic health status of the application
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Application is healthy"
// @Router       /health [get]
func (hc *HealthController) BasicHealthHandler(c *gin.Context) {
	ctx := utils.WithCorrelationID(c.Request.Context(), "")

	hc.logger.InfoWithOperation(ctx, "basic_health", "Basic health check requested")

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(hc.startTime).String(),
	})
}

// BackgroundProcessingHealthHandler godoc
// @Summary      Background Processing Health Check
// @Description  Returns detailed health status of the background processing system including metrics and alerts
// @Tags         Health
// @Produce      json
// @Success      200  {object}  service.HealthStatus  "Background processing system health status"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /health/background-processing [get]
func (hc *HealthController) BackgroundProcessingHealthHandler(c *gin.Context) {
	ctx := utils.WithCorrelationID(c.Request.Context(), "")
	ctx = utils.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	hc.logger.InfoWithOperation(ctx, "bg_health_check", "Background processing health check requested")

	// Define alert thresholds
	thresholds := service.AlertThresholds{
		MaxErrorRate:     0.1,             // 10% max error rate
		MaxQueueDepth:    100,             // Max 100 items in queue
		MaxProcessingLag: 5 * time.Minute, // Max 5 minutes processing lag
		MinSuccessRate:   0.9,             // Min 90% success rate
	}

	// Get health status
	healthStatus := hc.metricsCollector.CheckHealth(ctx, thresholds, hc.startTime)

	// Log health check results
	metadata := map[string]string{
		"status":               healthStatus.Status,
		"healthy":              fmt.Sprintf("%t", healthStatus.Healthy),
		"queue_depth":          fmt.Sprintf("%d", healthStatus.SystemMetrics.QueueDepth),
		"processing_lag":       healthStatus.SystemMetrics.ProcessingLag.String(),
		"overall_success_rate": fmt.Sprintf("%.2f%%", healthStatus.SystemMetrics.OverallSuccessRate*100),
		"alert_count":          fmt.Sprintf("%d", len(healthStatus.Alerts)),
	}
	hc.logger.InfoWithMetrics(ctx, "bg_health_result", "Background processing health check completed", 0, metadata)

	// Return appropriate HTTP status based on health
	statusCode := http.StatusOK
	if !healthStatus.Healthy {
		statusCode = http.StatusServiceUnavailable
		if healthStatus.Status == "critical" {
			statusCode = http.StatusInternalServerError
		}
	}

	c.JSON(statusCode, healthStatus)
}

// MetricsHandler godoc
// @Summary      System Metrics
// @Description  Returns detailed system metrics for monitoring and observability
// @Tags         Health
// @Produce      json
// @Success      200  {object}  service.SystemMetrics  "System metrics"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /health/metrics [get]
func (hc *HealthController) MetricsHandler(c *gin.Context) {
	ctx := utils.WithCorrelationID(c.Request.Context(), "")
	ctx = utils.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	hc.logger.InfoWithOperation(ctx, "metrics_request", "System metrics requested")

	systemMetrics, err := hc.metricsCollector.GetSystemMetrics(ctx)
	if err != nil {
		hc.logger.ErrorWithOperation(ctx, "metrics_error", "Failed to get system metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve system metrics",
			"message": err.Error(),
		})
		return
	}

	hc.logger.InfoWithOperation(ctx, "metrics_success", "System metrics retrieved successfully")
	c.JSON(http.StatusOK, systemMetrics)
}

// TaskMetricsHandler godoc
// @Summary      Task-specific Metrics
// @Description  Returns metrics for a specific task type
// @Tags         Health
// @Produce      json
// @Param        taskType  path    string  true  "Task type to get metrics for"
// @Success      200       {object}  service.TaskMetrics  "Task metrics"
// @Failure      400       {object}  map[string]interface{}  "Invalid task type"
// @Failure      404       {object}  map[string]interface{}  "Task type not found"
// @Router       /health/metrics/task/{taskType} [get]
func (hc *HealthController) TaskMetricsHandler(c *gin.Context) {
	ctx := utils.WithCorrelationID(c.Request.Context(), "")
	ctx = utils.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	taskType := c.Param("taskType")
	if taskType == "" {
		hc.logger.ErrorWithOperation(ctx, "task_metrics_validation", "Missing task type parameter", nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task type parameter is required",
		})
		return
	}

	hc.logger.InfoWithOperation(ctx, "task_metrics_request", "Task metrics requested for: "+taskType)

	taskMetrics := hc.metricsCollector.GetTaskMetrics(taskType)

	// Check if task type has any recorded metrics
	if taskMetrics.TotalCount == 0 {
		hc.logger.ErrorWithOperation(ctx, "task_metrics_not_found", "No metrics found for task type: "+taskType, nil)
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Task type not found",
			"task_type": taskType,
			"message":   "No metrics recorded for this task type",
		})
		return
	}

	hc.logger.InfoWithOperation(ctx, "task_metrics_success", "Task metrics retrieved successfully for: "+taskType)
	c.JSON(http.StatusOK, taskMetrics)
}

// ReadinessHandler godoc
// @Summary      Readiness Check
// @Description  Returns readiness status indicating if the application is ready to serve requests
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Application is ready"
// @Failure      503  {object}  map[string]interface{}  "Application is not ready"
// @Router       /health/ready [get]
func (hc *HealthController) ReadinessHandler(c *gin.Context) {
	ctx := utils.WithCorrelationID(c.Request.Context(), "")

	hc.logger.InfoWithOperation(ctx, "readiness_check", "Readiness check requested")

	// Check if background processing is healthy
	thresholds := service.AlertThresholds{
		MaxErrorRate:     0.2,              // More lenient for readiness
		MaxQueueDepth:    200,              // More lenient for readiness
		MaxProcessingLag: 10 * time.Minute, // More lenient for readiness
		MinSuccessRate:   0.8,              // More lenient for readiness
	}

	healthStatus := hc.metricsCollector.CheckHealth(ctx, thresholds, hc.startTime)

	// Consider ready if not critical
	ready := healthStatus.Status != "critical"

	if ready {
		hc.logger.InfoWithOperation(ctx, "readiness_success", "Application is ready")
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC(),
			"uptime":    time.Since(hc.startTime).String(),
			"background_processing": gin.H{
				"status":      healthStatus.Status,
				"queue_depth": healthStatus.SystemMetrics.QueueDepth,
			},
		})
	} else {
		hc.logger.ErrorWithOperation(ctx, "readiness_failure", "Application is not ready", nil)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"timestamp": time.Now().UTC(),
			"reason":    "Background processing is in critical state",
			"background_processing": gin.H{
				"status":      healthStatus.Status,
				"queue_depth": healthStatus.SystemMetrics.QueueDepth,
				"alerts":      len(healthStatus.Alerts),
			},
		})
	}
}

// LivenessHandler godoc
// @Summary      Liveness Check
// @Description  Returns liveness status indicating if the application is alive and should not be restarted
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Application is alive"
// @Router       /health/live [get]
func (hc *HealthController) LivenessHandler(c *gin.Context) {
	ctx := utils.WithCorrelationID(c.Request.Context(), "")

	hc.logger.InfoWithOperation(ctx, "liveness_check", "Liveness check requested")

	// Simple liveness check - if we can respond, we're alive
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(hc.startTime).String(),
	})
}
