package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"lumenslate/internal/utils"

	"github.com/hibiken/asynq"
)

// MetricsCollector collects and tracks metrics for async document processing
type MetricsCollector struct {
	mu               sync.RWMutex
	taskSuccessCount map[string]int64
	taskFailureCount map[string]int64
	taskDurations    map[string][]time.Duration
	queueDepth       int64
	processingLag    time.Duration
	lastUpdated      time.Time
	asynqInspector   *asynq.Inspector
	logger           *utils.Logger
}

// TaskMetrics represents metrics for a specific task type
type TaskMetrics struct {
	TaskType        string        `json:"task_type"`
	SuccessCount    int64         `json:"success_count"`
	FailureCount    int64         `json:"failure_count"`
	TotalCount      int64         `json:"total_count"`
	SuccessRate     float64       `json:"success_rate"`
	AverageDuration time.Duration `json:"average_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// SystemMetrics represents overall system metrics
type SystemMetrics struct {
	QueueDepth         int64         `json:"queue_depth"`
	ProcessingLag      time.Duration `json:"processing_lag"`
	ActiveWorkers      int           `json:"active_workers"`
	TotalProcessed     int64         `json:"total_processed"`
	TotalFailed        int64         `json:"total_failed"`
	OverallSuccessRate float64       `json:"overall_success_rate"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// HealthStatus represents the health status of the background processing system
type HealthStatus struct {
	Status        string                 `json:"status"`
	Healthy       bool                   `json:"healthy"`
	Timestamp     time.Time              `json:"timestamp"`
	SystemMetrics SystemMetrics          `json:"system_metrics"`
	TaskMetrics   map[string]TaskMetrics `json:"task_metrics"`
	Alerts        []Alert                `json:"alerts"`
	Uptime        time.Duration          `json:"uptime"`
}

// Alert represents a system alert
type Alert struct {
	Level     string            `json:"level"` // "warning", "error", "critical"
	Type      string            `json:"type"`  // "high_error_rate", "queue_backup", "processing_lag"
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AlertThresholds defines thresholds for generating alerts
type AlertThresholds struct {
	MaxErrorRate     float64       // Maximum acceptable error rate (0.0-1.0)
	MaxQueueDepth    int64         // Maximum acceptable queue depth
	MaxProcessingLag time.Duration // Maximum acceptable processing lag
	MinSuccessRate   float64       // Minimum acceptable success rate
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(redisAddr string) (*MetricsCollector, error) {
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}
	inspector := asynq.NewInspector(redisOpt)

	return &MetricsCollector{
		taskSuccessCount: make(map[string]int64),
		taskFailureCount: make(map[string]int64),
		taskDurations:    make(map[string][]time.Duration),
		asynqInspector:   inspector,
		logger:           utils.NewLogger("metrics_collector"),
		lastUpdated:      time.Now(),
	}, nil
}

// RecordTaskSuccess records a successful task completion
func (mc *MetricsCollector) RecordTaskSuccess(ctx context.Context, taskType string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.taskSuccessCount[taskType]++
	mc.taskDurations[taskType] = append(mc.taskDurations[taskType], duration)
	mc.lastUpdated = time.Now()

	// Keep only last 100 durations to prevent memory growth
	if len(mc.taskDurations[taskType]) > 100 {
		mc.taskDurations[taskType] = mc.taskDurations[taskType][1:]
	}

	mc.logger.InfoWithOperation(ctx, "metrics_record", fmt.Sprintf("Recorded task success: %s", taskType))
}

// RecordTaskFailure records a failed task
func (mc *MetricsCollector) RecordTaskFailure(ctx context.Context, taskType string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.taskFailureCount[taskType]++
	mc.taskDurations[taskType] = append(mc.taskDurations[taskType], duration)
	mc.lastUpdated = time.Now()

	// Keep only last 100 durations to prevent memory growth
	if len(mc.taskDurations[taskType]) > 100 {
		mc.taskDurations[taskType] = mc.taskDurations[taskType][1:]
	}

	mc.logger.InfoWithOperation(ctx, "metrics_record", fmt.Sprintf("Recorded task failure: %s", taskType))
}

// UpdateQueueMetrics updates queue depth and processing lag metrics
func (mc *MetricsCollector) UpdateQueueMetrics(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Get queue statistics from Asynq inspector
	queueStats, err := mc.asynqInspector.GetQueueInfo("default")
	if err != nil {
		mc.logger.ErrorWithOperation(ctx, "queue_metrics", "Failed to get queue statistics", err)
		return fmt.Errorf("failed to get queue statistics: %w", err)
	}

	mc.queueDepth = int64(queueStats.Pending + queueStats.Active)

	// Calculate processing lag (simplified - based on oldest pending task)
	if queueStats.Pending > 0 {
		// This is a simplified calculation - in a real system you might want to
		// look at the timestamp of the oldest pending task
		mc.processingLag = time.Duration(queueStats.Pending) * time.Second
	} else {
		mc.processingLag = 0
	}

	mc.lastUpdated = time.Now()

	mc.logger.InfoWithOperation(ctx, "queue_metrics", fmt.Sprintf("Updated queue metrics - depth: %d, lag: %v", mc.queueDepth, mc.processingLag))
	return nil
}

// GetTaskMetrics returns metrics for a specific task type
func (mc *MetricsCollector) GetTaskMetrics(taskType string) TaskMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	successCount := mc.taskSuccessCount[taskType]
	failureCount := mc.taskFailureCount[taskType]
	totalCount := successCount + failureCount

	var successRate float64
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount)
	}

	var avgDuration, minDuration, maxDuration time.Duration
	durations := mc.taskDurations[taskType]
	if len(durations) > 0 {
		var total time.Duration
		minDuration = durations[0]
		maxDuration = durations[0]

		for _, d := range durations {
			total += d
			if d < minDuration {
				minDuration = d
			}
			if d > maxDuration {
				maxDuration = d
			}
		}
		avgDuration = total / time.Duration(len(durations))
	}

	return TaskMetrics{
		TaskType:        taskType,
		SuccessCount:    successCount,
		FailureCount:    failureCount,
		TotalCount:      totalCount,
		SuccessRate:     successRate,
		AverageDuration: avgDuration,
		MinDuration:     minDuration,
		MaxDuration:     maxDuration,
		LastUpdated:     mc.lastUpdated,
	}
}

// GetSystemMetrics returns overall system metrics
func (mc *MetricsCollector) GetSystemMetrics(ctx context.Context) (SystemMetrics, error) {
	// Update queue metrics before acquiring lock (this method needs write lock)
	if err := mc.UpdateQueueMetrics(ctx); err != nil {
		mc.logger.ErrorWithOperation(ctx, "system_metrics", "Failed to update queue metrics", err)
	}

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalProcessed, totalFailed int64
	for _, count := range mc.taskSuccessCount {
		totalProcessed += count
	}
	for _, count := range mc.taskFailureCount {
		totalFailed += count
		totalProcessed += count
	}

	var overallSuccessRate float64
	if totalProcessed > 0 {
		overallSuccessRate = float64(totalProcessed-totalFailed) / float64(totalProcessed)
	}

	// Get active workers from Asynq inspector
	queueStats, err := mc.asynqInspector.GetQueueInfo("default")
	activeWorkers := 0
	if err == nil {
		activeWorkers = queueStats.Active
	}

	return SystemMetrics{
		QueueDepth:         mc.queueDepth,
		ProcessingLag:      mc.processingLag,
		ActiveWorkers:      activeWorkers,
		TotalProcessed:     totalProcessed,
		TotalFailed:        totalFailed,
		OverallSuccessRate: overallSuccessRate,
		LastUpdated:        mc.lastUpdated,
	}, nil
}

// CheckHealth performs health checks and generates alerts
func (mc *MetricsCollector) CheckHealth(ctx context.Context, thresholds AlertThresholds, startTime time.Time) HealthStatus {
	systemMetrics, err := mc.GetSystemMetrics(ctx)
	if err != nil {
		mc.logger.ErrorWithOperation(ctx, "health_check", "Failed to get system metrics", err)
	}

	// Collect task metrics
	taskMetrics := make(map[string]TaskMetrics)
	mc.mu.RLock()
	for taskType := range mc.taskSuccessCount {
		taskMetrics[taskType] = mc.GetTaskMetrics(taskType)
	}
	for taskType := range mc.taskFailureCount {
		if _, exists := taskMetrics[taskType]; !exists {
			taskMetrics[taskType] = mc.GetTaskMetrics(taskType)
		}
	}
	mc.mu.RUnlock()

	// Generate alerts based on thresholds
	alerts := mc.generateAlerts(systemMetrics, taskMetrics, thresholds)

	// Determine overall health status
	healthy := len(alerts) == 0
	status := "healthy"
	if !healthy {
		status = "unhealthy"
		for _, alert := range alerts {
			if alert.Level == "critical" {
				status = "critical"
				break
			}
		}
	}

	return HealthStatus{
		Status:        status,
		Healthy:       healthy,
		Timestamp:     time.Now(),
		SystemMetrics: systemMetrics,
		TaskMetrics:   taskMetrics,
		Alerts:        alerts,
		Uptime:        time.Since(startTime),
	}
}

// generateAlerts generates alerts based on metrics and thresholds
func (mc *MetricsCollector) generateAlerts(systemMetrics SystemMetrics, taskMetrics map[string]TaskMetrics, thresholds AlertThresholds) []Alert {
	var alerts []Alert
	now := time.Now()

	// Check overall error rate
	if systemMetrics.OverallSuccessRate < thresholds.MinSuccessRate && systemMetrics.TotalProcessed > 0 {
		alerts = append(alerts, Alert{
			Level:     "error",
			Type:      "high_error_rate",
			Message:   fmt.Sprintf("Overall success rate (%.2f%%) is below threshold (%.2f%%)", systemMetrics.OverallSuccessRate*100, thresholds.MinSuccessRate*100),
			Timestamp: now,
			Metadata: map[string]string{
				"current_rate": fmt.Sprintf("%.2f", systemMetrics.OverallSuccessRate),
				"threshold":    fmt.Sprintf("%.2f", thresholds.MinSuccessRate),
			},
		})
	}

	// Check queue depth
	if systemMetrics.QueueDepth > thresholds.MaxQueueDepth {
		level := "warning"
		if systemMetrics.QueueDepth > thresholds.MaxQueueDepth*2 {
			level = "critical"
		}
		alerts = append(alerts, Alert{
			Level:     level,
			Type:      "queue_backup",
			Message:   fmt.Sprintf("Queue depth (%d) exceeds threshold (%d)", systemMetrics.QueueDepth, thresholds.MaxQueueDepth),
			Timestamp: now,
			Metadata: map[string]string{
				"current_depth": fmt.Sprintf("%d", systemMetrics.QueueDepth),
				"threshold":     fmt.Sprintf("%d", thresholds.MaxQueueDepth),
			},
		})
	}

	// Check processing lag
	if systemMetrics.ProcessingLag > thresholds.MaxProcessingLag {
		level := "warning"
		if systemMetrics.ProcessingLag > thresholds.MaxProcessingLag*2 {
			level = "critical"
		}
		alerts = append(alerts, Alert{
			Level:     level,
			Type:      "processing_lag",
			Message:   fmt.Sprintf("Processing lag (%v) exceeds threshold (%v)", systemMetrics.ProcessingLag, thresholds.MaxProcessingLag),
			Timestamp: now,
			Metadata: map[string]string{
				"current_lag": systemMetrics.ProcessingLag.String(),
				"threshold":   thresholds.MaxProcessingLag.String(),
			},
		})
	}

	// Check individual task error rates
	for taskType, metrics := range taskMetrics {
		if metrics.TotalCount > 0 && metrics.SuccessRate < thresholds.MinSuccessRate {
			alerts = append(alerts, Alert{
				Level:     "warning",
				Type:      "task_error_rate",
				Message:   fmt.Sprintf("Task %s success rate (%.2f%%) is below threshold (%.2f%%)", taskType, metrics.SuccessRate*100, thresholds.MinSuccessRate*100),
				Timestamp: now,
				Metadata: map[string]string{
					"task_type":    taskType,
					"current_rate": fmt.Sprintf("%.2f", metrics.SuccessRate),
					"threshold":    fmt.Sprintf("%.2f", thresholds.MinSuccessRate),
				},
			})
		}
	}

	return alerts
}

// Close closes the metrics collector and its resources
func (mc *MetricsCollector) Close() error {
	return mc.asynqInspector.Close()
}

// GetMetricsJSON returns metrics as JSON string for easy logging/monitoring
func (mc *MetricsCollector) GetMetricsJSON(ctx context.Context, thresholds AlertThresholds, startTime time.Time) (string, error) {
	health := mc.CheckHealth(ctx, thresholds, startTime)

	jsonBytes, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metrics to JSON: %w", err)
	}

	return string(jsonBytes), nil
}
