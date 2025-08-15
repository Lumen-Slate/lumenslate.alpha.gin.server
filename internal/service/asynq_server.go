package service

import (
	"context"
	"fmt"
	"log"
	"lumenslate/internal/utils"
	"os"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
)

// AsynqServer wraps the Asynq server and multiplexer for background task processing
type AsynqServer struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewAsynqServer creates a new AsynqServer with Redis configuration and configurable settings
func NewAsynqServer(redisAddr string, concurrency int) *AsynqServer {
	// Set default values if not provided
	if redisAddr == "" {
		redisAddr = getEnvWithDefault("REDIS_ADDR", "localhost:6379")
	}
	if concurrency <= 0 {
		concurrency = getIntEnvWithDefault("ASYNQ_CONCURRENCY", 10)
	}

	// Configure Redis connection
	redisOpt := asynq.RedisClientOpt{
		Addr: redisAddr,
	}

	// Configure server options with retry policies
	serverConfig := asynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			"default": 6, // Higher priority for default queue
			"low":     3, // Lower priority for less critical tasks
		},
		// Configure retry policy
		RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
			// Exponential backoff: 1min, 2min, 4min, etc.
			delay := time.Duration(1<<uint(n)) * time.Minute
			maxDelay := getIntEnvWithDefault("ASYNQ_MAX_RETRY_DELAY", 300) // 5 minutes default
			if delay > time.Duration(maxDelay)*time.Second {
				delay = time.Duration(maxDelay) * time.Second
			}
			return delay
		},
		// Configure error handler for monitoring with structured logging
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger := utils.NewLogger("asynq_error_handler")

			errorMessage := fmt.Sprintf("Task processing failed - type=%s, payload=%s",
				task.Type(), string(task.Payload()[:min(100, len(task.Payload()))]))

			logger.ErrorWithOperation(ctx, "task_error", errorMessage, err)
		}),
		// Use default logger settings
	}

	// Create server and multiplexer
	server := asynq.NewServer(redisOpt, serverConfig)
	mux := asynq.NewServeMux()

	return &AsynqServer{
		server: server,
		mux:    mux,
	}
}

// RegisterHandlers maps task types to their corresponding handler functions
func (s *AsynqServer) RegisterHandlers() error {
	log.Printf("INFO: Registering Asynq task handlers")

	// Add middleware for logging and monitoring
	s.mux.Use(loggingMiddleware())
	s.mux.Use(recoveryMiddleware())

	log.Printf("INFO: Successfully registered task handler middleware")
	return nil
}

// RegisterTaskHandler registers a specific task handler with the multiplexer
func (s *AsynqServer) RegisterTaskHandler(taskType string, handler asynq.HandlerFunc) error {
	if taskType == "" {
		return fmt.Errorf("task type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	s.mux.HandleFunc(taskType, handler)
	log.Printf("INFO: Registered handler for task type: %s", taskType)
	return nil
}

// Start begins processing background tasks
func (s *AsynqServer) Start() error {
	log.Printf("INFO: Starting Asynq server for background task processing")

	// Register base handlers and middleware
	if err := s.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register base handlers: %w", err)
	}

	// Start the server in a goroutine to avoid blocking
	go func() {
		if err := s.server.Run(s.mux); err != nil {
			log.Printf("ERROR: Asynq server stopped with error: %v", err)
		}
	}()

	log.Printf("INFO: Asynq server started successfully")
	return nil
}

// Stop gracefully shuts down the Asynq server
func (s *AsynqServer) Stop() {
	log.Printf("INFO: Stopping Asynq server...")

	// Shutdown the server
	s.server.Shutdown()
	log.Printf("INFO: Asynq server stopped successfully")
}

// loggingMiddleware adds structured logging to task processing
func loggingMiddleware() asynq.MiddlewareFunc {
	return asynq.MiddlewareFunc(func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()

			// Add correlation ID to context if not present
			ctx = utils.WithCorrelationID(ctx, "")
			logger := utils.NewLogger("asynq_middleware")

			taskMetadata := map[string]string{
				"task_type": t.Type(),
				"task_id":   string(t.Payload()[:min(50, len(t.Payload()))]), // First 50 chars of payload for identification
			}

			logger.InfoWithMetrics(ctx, "task_start", "Starting task processing", 0, taskMetadata)

			err := h.ProcessTask(ctx, t)

			duration := time.Since(start)
			taskMetadata["duration"] = duration.String()

			if err != nil {
				logger.ErrorWithMetrics(ctx, "task_processing", "Task processing failed", err, duration, taskMetadata)
			} else {
				logger.InfoWithMetrics(ctx, "task_processing", "Task processing completed successfully", duration, taskMetadata)
			}

			return err
		})
	})
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// recoveryMiddleware handles panics in task processing
func recoveryMiddleware() asynq.MiddlewareFunc {
	return asynq.MiddlewareFunc(func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC: Task processing panicked - type=%s, panic=%v",
						t.Type(), r)
					err = fmt.Errorf("task processing panicked: %v", r)
				}
			}()

			return h.ProcessTask(ctx, t)
		})
	})
}

// Helper functions for environment variable handling
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
