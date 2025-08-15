package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// ContextKey is used for storing values in context
type ContextKey string

const (
	CorrelationIDKey ContextKey = "correlation_id"
	RequestIDKey     ContextKey = "request_id"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     time.Time         `json:"timestamp"`
	Level         LogLevel          `json:"level"`
	Message       string            `json:"message"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	Component     string            `json:"component,omitempty"`
	Operation     string            `json:"operation,omitempty"`
	FileID        string            `json:"file_id,omitempty"`
	TaskType      string            `json:"task_type,omitempty"`
	Duration      string            `json:"duration,omitempty"`
	Error         string            `json:"error,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Logger provides structured logging functionality
type Logger struct {
	component string
}

// NewLogger creates a new logger instance for a specific component
func NewLogger(component string) *Logger {
	return &Logger{
		component: component,
	}
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, message string) {
	l.log(ctx, LogLevelDebug, message, "", "", nil, nil)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, message string) {
	l.log(ctx, LogLevelInfo, message, "", "", nil, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, message string) {
	l.log(ctx, LogLevelWarn, message, "", "", nil, nil)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, message string, err error) {
	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}
	l.log(ctx, LogLevelError, message, "", errorStr, nil, nil)
}

// InfoWithOperation logs an info message with operation context
func (l *Logger) InfoWithOperation(ctx context.Context, operation, message string) {
	l.log(ctx, LogLevelInfo, message, operation, "", nil, nil)
}

// ErrorWithOperation logs an error message with operation context
func (l *Logger) ErrorWithOperation(ctx context.Context, operation, message string, err error) {
	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}
	l.log(ctx, LogLevelError, message, operation, errorStr, nil, nil)
}

// InfoWithMetrics logs an info message with performance metrics
func (l *Logger) InfoWithMetrics(ctx context.Context, operation, message string, duration time.Duration, metadata map[string]string) {
	l.log(ctx, LogLevelInfo, message, operation, "", &duration, metadata)
}

// ErrorWithMetrics logs an error message with performance metrics
func (l *Logger) ErrorWithMetrics(ctx context.Context, operation, message string, err error, duration time.Duration, metadata map[string]string) {
	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}
	l.log(ctx, LogLevelError, message, operation, errorStr, &duration, metadata)
}

// TaskLifecycle logs task lifecycle events (enqueue, start, complete, fail)
func (l *Logger) TaskLifecycle(ctx context.Context, event, taskType, fileID string, metadata map[string]string) {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["event"] = event

	message := fmt.Sprintf("Task %s: %s", event, taskType)
	l.log(ctx, LogLevelInfo, message, "task_lifecycle", "", nil, metadata)
}

// TaskError logs task processing errors with context
func (l *Logger) TaskError(ctx context.Context, taskType, fileID, operation, message string, err error, metadata map[string]string) {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["task_type"] = taskType
	metadata["file_id"] = fileID

	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}
	l.log(ctx, LogLevelError, message, operation, errorStr, nil, metadata)
}

// TaskMetrics logs task performance metrics
func (l *Logger) TaskMetrics(ctx context.Context, taskType, fileID string, duration time.Duration, success bool, metadata map[string]string) {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["task_type"] = taskType
	metadata["file_id"] = fileID
	metadata["success"] = fmt.Sprintf("%t", success)

	level := LogLevelInfo
	message := fmt.Sprintf("Task completed: %s", taskType)
	if !success {
		level = LogLevelError
		message = fmt.Sprintf("Task failed: %s", taskType)
	}

	l.log(ctx, level, message, "task_metrics", "", &duration, metadata)
}

// log is the internal logging method that handles structured output
func (l *Logger) log(ctx context.Context, level LogLevel, message, operation, errorStr string, duration *time.Duration, metadata map[string]string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   message,
		Component: l.component,
	}

	// Add context values if available
	if ctx != nil {
		entry.CorrelationID = GetCorrelationID(ctx)
		entry.RequestID = GetRequestID(ctx)
	}

	// Add optional fields
	if operation != "" {
		entry.Operation = operation
	}
	if errorStr != "" {
		entry.Error = errorStr
	}
	if duration != nil {
		entry.Duration = duration.String()
	}
	if metadata != nil {
		entry.Metadata = metadata

		// Extract common fields from metadata
		if fileID, ok := metadata["file_id"]; ok {
			entry.FileID = fileID
		}
		if taskType, ok := metadata["task_type"]; ok {
			entry.TaskType = taskType
		}
	}

	// Output structured JSON log
	if jsonBytes, err := json.Marshal(entry); err == nil {
		log.Println(string(jsonBytes))
	} else {
		// Fallback to simple logging if JSON marshaling fails
		log.Printf("[%s] %s: %s", level, l.component, message)
	}
}

// LogTaskStart logs the start of a task with correlation tracking
func LogTaskStart(ctx context.Context, taskType, fileID string, metadata map[string]string) context.Context {
	logger := NewLogger("task_processor")

	// Ensure we have a correlation ID for task tracking
	correlationID := GetCorrelationID(ctx)
	if correlationID == "" {
		correlationID = uuid.New().String()
		ctx = WithCorrelationID(ctx, correlationID)
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["file_id"] = fileID
	metadata["task_type"] = taskType

	logger.TaskLifecycle(ctx, "start", taskType, fileID, metadata)
	return ctx
}

// LogTaskComplete logs the completion of a task with performance metrics
func LogTaskComplete(ctx context.Context, taskType, fileID string, startTime time.Time, success bool, metadata map[string]string) {
	logger := NewLogger("task_processor")
	duration := time.Since(startTime)

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["file_id"] = fileID

	logger.TaskMetrics(ctx, taskType, fileID, duration, success, metadata)

	event := "complete"
	if !success {
		event = "fail"
	}
	logger.TaskLifecycle(ctx, event, taskType, fileID, metadata)
}

// LogTaskEnqueue logs when a task is enqueued
func LogTaskEnqueue(ctx context.Context, taskType, fileID string, metadata map[string]string) {
	logger := NewLogger("task_enqueue")

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["file_id"] = fileID

	logger.TaskLifecycle(ctx, "enqueue", taskType, fileID, metadata)
}
