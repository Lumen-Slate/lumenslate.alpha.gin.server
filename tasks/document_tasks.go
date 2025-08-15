package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"lumenslate/internal/repository"
	"lumenslate/internal/service"
	"lumenslate/internal/utils"
	"os"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

// Task type constants
const (
	TypeAddDocumentToCorpus = "add:document_to_corpus"
)

// Global metrics collector instance
var globalMetricsCollector *service.MetricsCollector

// SetMetricsCollector sets the global metrics collector instance
func SetMetricsCollector(collector *service.MetricsCollector) {
	globalMetricsCollector = collector
}

// GetMetricsCollector returns the global metrics collector instance
func GetMetricsCollector() *service.MetricsCollector {
	return globalMetricsCollector
}

// DocumentTaskPayload represents the payload for document processing tasks
type DocumentTaskPayload struct {
	FileID          string `json:"file_id"`
	TempObjectName  string `json:"temp_object_name"`
	FinalObjectName string `json:"final_object_name"`
	CorpusName      string `json:"corpus_name"`
	DisplayName     string `json:"display_name"`
}

// NewAddDocumentToCorpusTask creates a new Asynq task for adding a document to the RAG corpus
func NewAddDocumentToCorpusTask(payload DocumentTaskPayload) (*asynq.Task, error) {
	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document task payload: %w", err)
	}

	// Create the task with appropriate options
	task := asynq.NewTask(
		TypeAddDocumentToCorpus,
		payloadBytes,
		asynq.MaxRetry(3),
		asynq.Timeout(300), // 5 minutes timeout for RAG processing
	)

	return task, nil
}

// HandleAddDocumentToCorpusTask processes the document ingestion task with comprehensive error handling
func HandleAddDocumentToCorpusTask(ctx context.Context, t *asynq.Task) error {
	startTime := time.Now()

	// Parse the task payload
	var payload DocumentTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("ERROR: Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	// Initialize structured logging with correlation ID and task context
	ctx = utils.WithCorrelationID(ctx, "")
	ctx = utils.LogTaskStart(ctx, TypeAddDocumentToCorpus, payload.FileID, map[string]string{
		"corpus_name":       payload.CorpusName,
		"temp_object_name":  payload.TempObjectName,
		"final_object_name": payload.FinalObjectName,
		"display_name":      payload.DisplayName,
	})

	logger := utils.NewLogger("task_processor")

	taskMetadata := map[string]string{
		"file_id":           payload.FileID,
		"corpus_name":       payload.CorpusName,
		"temp_object_name":  payload.TempObjectName,
		"final_object_name": payload.FinalObjectName,
		"display_name":      payload.DisplayName,
	}
	logger.InfoWithMetrics(ctx, "task_start", "Starting document processing task", 0, taskMetadata)

	// Initialize services with error handling
	docRepo := repository.NewDocumentRepository()
	gcsService, err := service.NewGCSService()
	if err != nil {
		logger.ErrorWithOperation(ctx, "service_init", "Failed to initialize GCS service", err)
		// Update document status to failed with detailed error
		errorMsg := fmt.Sprintf("Service initialization failed: %v", err)
		if updateErr := docRepo.UpdateStatus(ctx, payload.FileID, "failed", errorMsg); updateErr != nil {
			logger.ErrorWithOperation(ctx, "status_update", "Failed to update document status after GCS service initialization error", updateErr)
		}
		utils.LogTaskComplete(ctx, TypeAddDocumentToCorpus, payload.FileID, startTime, false, map[string]string{
			"error": "service_initialization_failed",
		})

		// Record metrics for failed task
		if metricsCollector := GetMetricsCollector(); metricsCollector != nil {
			metricsCollector.RecordTaskFailure(ctx, TypeAddDocumentToCorpus, time.Since(startTime))
		}

		return fmt.Errorf("failed to initialize GCS service: %w", err)
	}
	defer func() {
		if closeErr := gcsService.Close(); closeErr != nil {
			logger.ErrorWithOperation(ctx, "service_cleanup", "Failed to close GCS service", closeErr)
		}
	}()

	logger.InfoWithOperation(ctx, "service_init", "Services initialized successfully")

	// Step 1: Add document to Vertex AI RAG corpus with comprehensive error handling
	gcsURL := fmt.Sprintf("gs://%s/%s", os.Getenv("GCS_BUCKET_NAME"), payload.TempObjectName)

	ragMetadata := map[string]string{
		"file_id": payload.FileID,
		"gcs_url": gcsURL,
		"corpus":  payload.CorpusName,
	}
	logger.InfoWithMetrics(ctx, "rag_ingestion_start", "Starting RAG corpus ingestion", 0, ragMetadata)

	vertexAIService := service.NewVertexAIService()
	ragStartTime := time.Now()
	addResult, err := vertexAIService.AddDocumentToCorpus(ctx, payload.CorpusName, gcsURL)
	ragDuration := time.Since(ragStartTime)

	if err != nil {
		logger.ErrorWithMetrics(ctx, "rag_ingestion", "RAG ingestion failed", err, ragDuration, ragMetadata)

		// Comprehensive cleanup and error reporting
		errorMsg := fmt.Sprintf("RAG ingestion failed: %v", err)

		// Update document status to failed
		if updateErr := docRepo.UpdateStatus(ctx, payload.FileID, "failed", errorMsg); updateErr != nil {
			logger.ErrorWithOperation(ctx, "status_update", "Failed to update document status to failed after RAG error", updateErr)
		} else {
			logger.InfoWithOperation(ctx, "status_update", "Updated document status to failed")
		}

		// Clean up: delete the temporary GCS file
		logger.InfoWithOperation(ctx, "cleanup_start", "Starting cleanup of temporary GCS object after RAG failure")
		if deleteErr := gcsService.DeleteObject(ctx, payload.TempObjectName); deleteErr != nil {
			logger.ErrorWithOperation(ctx, "cleanup", "Failed to clean up temporary GCS object after RAG error", deleteErr)
			// Log but don't fail - the main error is more important
		} else {
			logger.InfoWithOperation(ctx, "cleanup", "Successfully cleaned up temporary GCS object")
		}

		utils.LogTaskComplete(ctx, TypeAddDocumentToCorpus, payload.FileID, startTime, false, map[string]string{
			"error":        "rag_ingestion_failed",
			"rag_duration": ragDuration.String(),
		})

		// Record metrics for failed task
		if metricsCollector := GetMetricsCollector(); metricsCollector != nil {
			metricsCollector.RecordTaskFailure(ctx, TypeAddDocumentToCorpus, time.Since(startTime))
		}

		return fmt.Errorf("failed to add document to RAG corpus: %w", err)
	}

	ragMetadata["rag_duration"] = ragDuration.String()
	ragMetadata["result"] = fmt.Sprintf("%v", addResult)
	logger.InfoWithMetrics(ctx, "rag_ingestion", "RAG import initiated successfully", ragDuration, ragMetadata)

	// Step 1.1: Wait for and verify RAG operation completion
	if operationName, ok := addResult["operation"].(string); ok && operationName != "" {
		logger.InfoWithOperation(ctx, "rag_operation_wait", "Waiting for RAG operation to complete")

		// Wait for the operation to complete (with timeout)
		operationStartTime := time.Now()
		maxWaitTime := 4 * time.Minute // Maximum wait time for RAG processing

		for time.Since(operationStartTime) < maxWaitTime {
			operationStatus, err := checkRAGOperationStatus(operationName)
			if err != nil {
				logger.ErrorWithOperation(ctx, "rag_operation_check", "Failed to check RAG operation status", err)
				break
			}

			if done, ok := operationStatus["done"].(bool); ok && done {
				// Operation completed
				if operationError, hasError := operationStatus["error"]; hasError {
					// Operation failed
					errorMsg := fmt.Sprintf("RAG operation failed: %v", operationError)
					logger.ErrorWithOperation(ctx, "rag_operation_failed", errorMsg, nil)

					// Update document status to failed
					if updateErr := docRepo.UpdateStatus(ctx, payload.FileID, "failed", errorMsg); updateErr != nil {
						logger.ErrorWithOperation(ctx, "status_update", "Failed to update document status after RAG operation failure", updateErr)
					}

					// Clean up GCS file
					if deleteErr := gcsService.DeleteObject(ctx, payload.TempObjectName); deleteErr != nil {
						logger.ErrorWithOperation(ctx, "cleanup", "Failed to clean up GCS object after RAG operation failure", deleteErr)
					}

					utils.LogTaskComplete(ctx, TypeAddDocumentToCorpus, payload.FileID, startTime, false, map[string]string{
						"error": "rag_operation_failed",
					})

					if metricsCollector := GetMetricsCollector(); metricsCollector != nil {
						metricsCollector.RecordTaskFailure(ctx, TypeAddDocumentToCorpus, time.Since(startTime))
					}

					return fmt.Errorf("RAG operation failed: %v", operationError)
				}

				// Operation succeeded - extract RAG file information
				operationDuration := time.Since(operationStartTime)

				// Try to extract RAG file ID from the operation response
				var ragFileID string
				if response, hasResponse := operationStatus["response"]; hasResponse {
					if responseMap, ok := response.(map[string]interface{}); ok {
						if ragFiles, hasRagFiles := responseMap["ragFiles"]; hasRagFiles {
							if ragFilesList, ok := ragFiles.([]interface{}); ok && len(ragFilesList) > 0 {
								if ragFile, ok := ragFilesList[0].(map[string]interface{}); ok {
									if name, hasName := ragFile["name"].(string); hasName {
										// Extract the RAG file ID from the full resource name
										// Format: projects/.../ragCorpora/.../ragFiles/{ragFileId}
										parts := strings.Split(name, "/")
										if len(parts) > 0 {
											ragFileID = parts[len(parts)-1]
										}
									}
								}
							}
						}
					}
				}

				// RAG file ID extraction is now compulsory - task fails if we can't get it
				if ragFileID == "" {
					errorMsg := "Could not extract RAG file ID from operation response - this is required for proper document tracking"
					logger.ErrorWithOperation(ctx, "rag_file_id_extract", errorMsg, nil)

					// Update document status to failed
					if updateErr := docRepo.UpdateStatus(ctx, payload.FileID, "failed", errorMsg); updateErr != nil {
						logger.ErrorWithOperation(ctx, "status_update", "Failed to update document status after RAG file ID extraction failure", updateErr)
					}

					// Clean up GCS file since we can't properly track the RAG file
					if deleteErr := gcsService.DeleteObject(ctx, payload.TempObjectName); deleteErr != nil {
						logger.ErrorWithOperation(ctx, "cleanup", "Failed to clean up GCS object after RAG file ID extraction failure", deleteErr)
					}

					utils.LogTaskComplete(ctx, TypeAddDocumentToCorpus, payload.FileID, startTime, false, map[string]string{
						"error": "rag_file_id_extraction_failed",
					})

					if metricsCollector := GetMetricsCollector(); metricsCollector != nil {
						metricsCollector.RecordTaskFailure(ctx, TypeAddDocumentToCorpus, time.Since(startTime))
					}

					return fmt.Errorf("RAG file ID extraction failed: %s", errorMsg)
				}

				// Database update with RAG file ID is now compulsory - task fails if update fails
				if err := docRepo.UpdateFields(ctx, payload.FileID, bson.M{
					"ragFileId": ragFileID,
					"updatedAt": time.Now(),
				}); err != nil {
					errorMsg := fmt.Sprintf("Failed to update RAG file ID in database: %v", err)
					logger.ErrorWithOperation(ctx, "rag_file_id_update", errorMsg, err)

					// Update document status to failed
					if updateErr := docRepo.UpdateStatus(ctx, payload.FileID, "failed", errorMsg); updateErr != nil {
						logger.ErrorWithOperation(ctx, "status_update", "Failed to update document status after RAG file ID database update failure", updateErr)
					}

					utils.LogTaskComplete(ctx, TypeAddDocumentToCorpus, payload.FileID, startTime, false, map[string]string{
						"error": "rag_file_id_database_update_failed",
					})

					if metricsCollector := GetMetricsCollector(); metricsCollector != nil {
						metricsCollector.RecordTaskFailure(ctx, TypeAddDocumentToCorpus, time.Since(startTime))
					}

					return fmt.Errorf("RAG file ID database update failed: %w", err)
				}

				logger.InfoWithOperation(ctx, "rag_file_id_update", "Successfully updated RAG file ID in database")

				logger.InfoWithMetrics(ctx, "rag_operation_complete", "RAG operation completed successfully", operationDuration, map[string]string{
					"operation_name": operationName,
					"total_rag_time": time.Since(ragStartTime).String(),
					"rag_file_id":    ragFileID,
				})
				break
			}

			// Wait before checking again
			time.Sleep(10 * time.Second)
		}

		// Check if we timed out
		if time.Since(operationStartTime) >= maxWaitTime {
			logger.ErrorWithOperation(ctx, "rag_operation_timeout", "RAG operation timed out", nil)
			// Don't fail the task - let it continue and mark as completed
			// The operation might still complete in the background
		}
	} else {
		logger.ErrorWithOperation(ctx, "rag_operation_missing", "No operation name returned from RAG service", nil)
		// Continue processing - this might be a different response format
	}

	// Step 2: Rename GCS object from temporary to final name with graceful error handling
	renameMetadata := map[string]string{
		"file_id":     payload.FileID,
		"from_object": payload.TempObjectName,
		"to_object":   payload.FinalObjectName,
	}
	logger.InfoWithMetrics(ctx, "gcs_rename_start", "Starting GCS object rename", 0, renameMetadata)

	// Track the actual object name that will be used
	actualObjectName := payload.TempObjectName // Default to temp name in case rename fails

	renameStartTime := time.Now()
	if err := gcsService.RenameObject(ctx, payload.TempObjectName, payload.FinalObjectName); err != nil {
		renameDuration := time.Since(renameStartTime)
		renameMetadata["rename_duration"] = renameDuration.String()
		logger.ErrorWithMetrics(ctx, "gcs_rename", "Failed to rename GCS object, but RAG ingestion succeeded. File will remain with temporary name", err, renameDuration, renameMetadata)
		// Don't fail the task - RAG ingestion succeeded, just log the warning
		// The file will remain with the temporary name but still be accessible
		// actualObjectName remains as payload.TempObjectName
	} else {
		renameDuration := time.Since(renameStartTime)
		renameMetadata["rename_duration"] = renameDuration.String()
		logger.InfoWithMetrics(ctx, "gcs_rename", "Successfully renamed GCS object", renameDuration, renameMetadata)
		// Update actual object name to final name since rename succeeded
		actualObjectName = payload.FinalObjectName
	}

	// Step 2.1: Update database with actual GCS object path
	logger.InfoWithOperation(ctx, "db_object_update", "Updating database with actual GCS object path")
	dbUpdateStartTime := time.Now()

	if err := docRepo.UpdateFields(ctx, payload.FileID, bson.M{
		"gcsObject": actualObjectName,
		"updatedAt": time.Now(),
	}); err != nil {
		dbUpdateDuration := time.Since(dbUpdateStartTime)
		logger.ErrorWithMetrics(ctx, "db_object_update", "Failed to update GCS object path in database", err, dbUpdateDuration, map[string]string{
			"file_id":     payload.FileID,
			"object_name": actualObjectName,
		})
		// Log error but don't fail the task - the file processing was successful
	} else {
		dbUpdateDuration := time.Since(dbUpdateStartTime)
		logger.InfoWithMetrics(ctx, "db_object_update", "Successfully updated GCS object path in database", dbUpdateDuration, map[string]string{
			"file_id":     payload.FileID,
			"object_name": actualObjectName,
		})
	}

	// Step 3: Update MongoDB status to "completed" with error handling
	logger.InfoWithOperation(ctx, "status_update_start", "Updating document status to completed")
	statusUpdateStartTime := time.Now()

	if err := docRepo.UpdateStatus(ctx, payload.FileID, "completed", ""); err != nil {
		statusUpdateDuration := time.Since(statusUpdateStartTime)
		logger.ErrorWithMetrics(ctx, "status_update", "Failed to update document status to completed", err, statusUpdateDuration, map[string]string{
			"file_id": payload.FileID,
			"warning": "document_processing_completed_but_status_update_failed",
		})
		// This is concerning but not fatal - the document was processed successfully
		// The RAG ingestion and file operations succeeded, so we don't want to fail the task
		// However, we should log this as an error for monitoring
	} else {
		statusUpdateDuration := time.Since(statusUpdateStartTime)
		logger.InfoWithMetrics(ctx, "status_update", "Successfully updated document status to completed", statusUpdateDuration, map[string]string{
			"file_id": payload.FileID,
		})
	}

	totalDuration := time.Since(startTime)
	completionMetadata := map[string]string{
		"file_id":        payload.FileID,
		"corpus_name":    payload.CorpusName,
		"total_duration": totalDuration.String(),
	}
	logger.InfoWithMetrics(ctx, "task_complete", "Successfully completed document processing task", totalDuration, completionMetadata)

	utils.LogTaskComplete(ctx, TypeAddDocumentToCorpus, payload.FileID, startTime, true, completionMetadata)

	// Record metrics for successful task completion
	if metricsCollector := GetMetricsCollector(); metricsCollector != nil {
		metricsCollector.RecordTaskSuccess(ctx, TypeAddDocumentToCorpus, totalDuration)
	}

	return nil
}

// logTaskError logs task errors with context for debugging and monitoring
func logTaskError(fileID, operation, message string, err error) {
	log.Printf("TASK_ERROR: fileId=%s, operation=%s, message=%s, error=%v", fileID, operation, message, err)
}

// logTaskWarning logs task warnings with context
func logTaskWarning(fileID, operation, message string) {
	log.Printf("TASK_WARNING: fileId=%s, operation=%s, message=%s", fileID, operation, message)
}

// logTaskInfo logs task information with context
func logTaskInfo(fileID, operation, message string) {
	log.Printf("TASK_INFO: fileId=%s, operation=%s, message=%s", fileID, operation, message)
}

// handleCriticalError handles critical errors that require immediate attention
func handleCriticalError(ctx context.Context, docRepo *repository.DocumentRepository, fileID, operation, errorMsg string, err error) {
	logTaskError(fileID, operation, "Critical error occurred", err)

	// Update document status with detailed error information
	fullErrorMsg := fmt.Sprintf("%s: %v", errorMsg, err)
	if updateErr := docRepo.UpdateStatus(ctx, fileID, "failed", fullErrorMsg); updateErr != nil {
		logTaskError(fileID, "status_update", "Failed to update document status after critical error", updateErr)
	}
}

// checkRAGOperationStatus checks the status of a Vertex AI RAG operation
func checkRAGOperationStatus(operationName string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Create Vertex AI service
	vertexService := service.NewVertexAIService()

	// Use the service's method to check operation status
	return vertexService.CheckOperationStatus(ctx, operationName)
}
