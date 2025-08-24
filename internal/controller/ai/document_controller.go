package ai

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"lumenslate/internal/model"
	"lumenslate/internal/repository"
	"lumenslate/internal/service"
	"lumenslate/internal/utils"
	"lumenslate/tasks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// DeleteCorpusDocumentHandler godoc
// @Summary      Delete Document from RAG Corpus
// @Description  Delete a specific document from a RAG corpus using its file identifier (fileId, RAG file ID, or display name). This operation removes the document from the RAG corpus, Google Cloud Storage, and local database.
// @Tags         AI Document Management
// @Accept       json
// @Produce      json
// @Param        body  body  ai.DeleteCorpusDocumentRequest  true  "Delete corpus document request containing corpus name and file identifier"
// @Success      200   {object}  map[string]interface{}  "Document deleted successfully with deletion status for each component"
// @Failure      400   {object}  map[string]interface{}  "Invalid request body or missing required fields"
// @Failure      404   {object}  map[string]interface{}  "Document or corpus not found"
// @Failure      500   {object}  map[string]interface{}  "Internal server error during deletion process"
// @Router       /ai/rag-agent/delete-corpus-document [post]
func DeleteCorpusDocumentHandler(c *gin.Context) {
	var req DeleteCorpusDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleteResponse, err := deleteVertexAICorpusDocument(req.CorpusName, req.FileID)
	if err != nil {
		log.Printf("[AI] Delete corpus document error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete document: %v", err)})
		return
	}

	c.JSON(http.StatusOK, deleteResponse)
}

// ViewDocumentHandler godoc
// @Summary      Generate Pre-signed URL for Document Viewing
// @Description  Generate a time-limited pre-signed URL to securely view a document stored in Google Cloud Storage. The URL expires after 30 minutes for security purposes.
// @Tags         AI Document Management
// @Accept       json
// @Produce      json
// @Param        id   path    string  true  "Document ID (unique identifier for the document)"
// @Success      200  {object}  map[string]interface{}  "Pre-signed URL generated successfully with document metadata"
// @Failure      400  {object}  map[string]interface{}  "Invalid or missing document ID"
// @Failure      404  {object}  map[string]interface{}  "Document not found in database or storage"
// @Failure      500  {object}  map[string]interface{}  "Internal server error during URL generation"
// @Router       /ai/documents/view/{id} [get]
func ViewDocumentHandler(c *gin.Context) {
	log.Println("[AI] /ai/documents/view/:id called")

	// Get document ID from URL parameter
	documentID := c.Param("id")
	if documentID == "" {
		log.Printf("[AI] Invalid request: missing document ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	log.Printf("[AI] Request to view document ID: %s", documentID)

	// Initialize services
	gcs, err := service.NewGCSService()
	if err != nil {
		log.Printf("[AI] Failed to initialize GCS service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize storage service"})
		return
	}
	defer gcs.Close()

	docRepo := repository.NewDocumentRepository()
	ctx := context.Background()

	// Get document metadata from database
	document, err := docRepo.GetDocumentByFileID(ctx, documentID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[AI] Document not found: %s", documentID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		} else {
			log.Printf("[AI] Database error retrieving document: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document information"})
		}
		return
	}

	log.Printf("[AI] Found document: %s (GCS object: %s)", document.DisplayName, document.GCSObject)

	// Verify the object exists in GCS
	exists, err := gcs.ObjectExists(ctx, document.GCSObject)
	if err != nil {
		log.Printf("[AI] Error checking GCS object existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify document availability"})
		return
	}

	if !exists {
		log.Printf("[AI] GCS object not found: %s", document.GCSObject)
		c.JSON(http.StatusNotFound, gin.H{"error": "Document file not found in storage"})
		return
	}

	// Generate pre-signed URL (valid for 30 minutes)
	expiration := 30 * time.Minute
	presignedURL, err := gcs.GenerateSignedURL(ctx, document.GCSObject, expiration)
	if err != nil {
		log.Printf("[AI] Failed to generate signed URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate document access URL"})
		return
	}

	log.Printf("[AI] Generated presigned URL successfully for document: %s", documentID)
	c.JSON(http.StatusOK, gin.H{
		"url":         presignedURL,
		"expiresIn":   "30 minutes",
		"documentId":  documentID,
		"displayName": document.DisplayName,
		"contentType": document.ContentType,
		"size":        document.Size,
		"corpusName":  document.CorpusName,
	})
}

// DeleteCorpusDocumentByIDHandler godoc
// @Summary      Delete Document by ID
// @Description  Delete a document from RAG corpus, Google Cloud Storage, and database using its unique document ID. This is a comprehensive deletion that removes all traces of the document from the system.
// @Tags         AI Document Management
// @Accept       json
// @Produce      json
// @Param        id   path    string  true  "Document ID (unique identifier for the document to delete)"
// @Success      200  {object}  map[string]interface{}  "Document deleted successfully from all systems"
// @Failure      400  {object}  map[string]interface{}  "Invalid or missing document ID"
// @Failure      404  {object}  map[string]interface{}  "Document not found"
// @Failure      500  {object}  map[string]interface{}  "Internal server error during deletion process"
// @Router       /ai/documents/{id} [delete]
func DeleteCorpusDocumentByIDHandler(c *gin.Context) {
	log.Println("[AI] /ai/documents/:id DELETE called")

	documentID := c.Param("id")
	if documentID == "" {
		log.Printf("[AI] Invalid request: missing document ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	log.Printf("[AI] Request to delete document ID: %s", documentID)

	// Initialize services
	gcs, err := service.NewGCSService()
	if err != nil {
		log.Printf("[AI] Failed to initialize GCS service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize storage service"})
		return
	}
	defer gcs.Close()

	docRepo := repository.NewDocumentRepository()
	ctx := context.Background()

	// Get document metadata from database
	document, err := docRepo.GetDocumentByFileID(ctx, documentID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[AI] Document not found: %s", documentID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		} else {
			log.Printf("[AI] Database error retrieving document: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve document information"})
		}
		return
	}

	log.Printf("[AI] Found document: %s (GCS object: %s)", document.DisplayName, document.GCSObject)

	// Delete from Vertex AI RAG corpus if RAG file ID exists
	if document.RAGFileID != "" {
		log.Printf("[AI] Deleting document from Vertex AI RAG corpus using RAG file ID: %s", document.RAGFileID)
		_, err := deleteVertexAICorpusDocument(document.CorpusName, document.RAGFileID)
		if err != nil {
			log.Printf("[AI] Warning: Failed to delete document from Vertex AI RAG corpus: %v", err)
			// Continue with deletion even if Vertex AI deletion fails
		} else {
			log.Printf("[AI] Successfully deleted document from Vertex AI RAG corpus")
		}
	} else {
		log.Printf("[AI] No RAG file ID found for document %s, skipping RAG engine deletion", documentID)
	}

	// Delete from GCS
	log.Printf("[AI] Deleting document from GCS...")
	if err := gcs.DeleteObject(ctx, document.GCSObject); err != nil {
		log.Printf("[AI] Warning: Failed to delete document from GCS: %v", err)
		// Continue with database deletion even if GCS deletion fails
	} else {
		log.Printf("[AI] Successfully deleted document from GCS")
	}

	// Delete from database
	log.Printf("[AI] Deleting document metadata from database...")
	if err := docRepo.DeleteDocument(ctx, documentID); err != nil {
		log.Printf("[AI] Failed to delete document metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document metadata"})
		return
	}

	log.Printf("[AI] Document deleted successfully: %s", documentID)
	c.JSON(http.StatusOK, gin.H{
		"message":    "Document deleted successfully",
		"documentId": documentID,
	})
}

// AddCorpusDocumentHandler godoc
// @Summary      Upload Document to RAG Corpus (Async)
// @Description  Upload a document file to Google Cloud Storage and enqueue it for asynchronous processing with Vertex AI RAG corpus. Returns immediately with pending status. Supports PDF, TXT, DOCX, DOC, HTML, and MD file formats.
// @Tags         AI Document Management
// @Accept       multipart/form-data
// @Produce      json
// @Param        corpusName  formData  string  true   "Name of the RAG corpus to add the document to"
// @Param        file        formData  file    true   "Document file to upload (supported formats: PDF, TXT, DOCX, DOC, HTML, MD)"
// @Success      200         {object}  map[string]interface{}  "Document uploaded successfully and queued for processing with pending status"
// @Failure      400         {object}  map[string]interface{}  "Invalid request, unsupported file type, or missing required fields"
// @Failure      500         {object}  map[string]interface{}  "Internal server error during upload or task enqueue process"
// @Router       /ai/rag-agent/add-corpus-document [post]
func AddCorpusDocumentHandler(c *gin.Context) {
	startTime := time.Now()

	// Initialize structured logging with correlation ID
	ctx := utils.WithCorrelationID(c.Request.Context(), "")
	ctx = utils.WithRequestID(ctx, c.GetHeader("X-Request-ID"))
	logger := utils.NewLogger("document_controller")

	logger.InfoWithOperation(ctx, "upload_start", "Document upload request received")

	// Validate required environment variables
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		logger.ErrorWithOperation(ctx, "config_validation", "GCS_BUCKET_NAME environment variable not set", nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Storage configuration missing: GCS_BUCKET_NAME environment variable is required",
			"details": "Please set the GCS_BUCKET_NAME environment variable to your Google Cloud Storage bucket name",
		})
		return
	}

	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		logger.ErrorWithOperation(ctx, "config_validation", "Project ID not configured", nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Project configuration missing: Google Cloud Project ID is required",
			"details": "Please ensure GOOGLE_CLOUD_PROJECT or project ID is properly configured",
		})
		return
	}

	metadata := map[string]string{
		"bucket_name": bucketName,
		"project_id":  projectID,
	}
	logger.InfoWithMetrics(ctx, "config_validation", "Environment configuration validated", 0, metadata)

	// Parse form data
	var req AddCorpusDocumentFormRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.ErrorWithOperation(ctx, "request_parsing", "Invalid form request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestMetadata := map[string]string{
		"corpus_name": req.CorpusName,
		"filename":    req.File.Filename,
		"file_size":   fmt.Sprintf("%d", req.File.Size),
	}
	logger.InfoWithMetrics(ctx, "request_parsing", "Form request parsed successfully", 0, requestMetadata)

	// Initialize services
	gcs, err := service.NewGCSService()
	if err != nil {
		logger.ErrorWithOperation(ctx, "service_init", "Failed to initialize GCS service", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize storage service"})
		return
	}
	defer gcs.Close()

	docRepo := repository.NewDocumentRepository()
	logger.InfoWithOperation(ctx, "service_init", "Services initialized successfully")

	// Open uploaded file
	file, err := req.File.Open()
	if err != nil {
		logger.ErrorWithOperation(ctx, "file_open", "Failed to open uploaded file", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read uploaded file"})
		return
	}
	defer file.Close()

	// Generate temporary object name for initial upload
	tempFileID := uuid.New().String()
	fileExtension := filepath.Ext(req.File.Filename)
	tempObjectName := fmt.Sprintf("temp/%s%s", tempFileID, fileExtension)

	fileMetadata := map[string]string{
		"temp_file_id":     tempFileID,
		"file_extension":   fileExtension,
		"temp_object_name": tempObjectName,
	}
	logger.InfoWithMetrics(ctx, "file_processing", "File opened and temporary names generated", 0, fileMetadata)

	// Validate file type for RAG compatibility before upload
	allowedExtensions := []string{".pdf", ".txt", ".docx", ".doc", ".html", ".md"}
	isValidType := false
	for _, ext := range allowedExtensions {
		if strings.EqualFold(fileExtension, ext) {
			isValidType = true
			break
		}
	}

	if !isValidType {
		logger.ErrorWithOperation(ctx, "file_validation", fmt.Sprintf("Unsupported file type: %s. Allowed types: %v", fileExtension, allowedExtensions), nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported file type: %s. Allowed types: %v", fileExtension, allowedExtensions)})
		return
	}

	logger.InfoWithMetrics(ctx, "file_validation", "File type validation passed", 0, map[string]string{"file_extension": fileExtension})

	// Upload file to GCS with temporary name
	log.Printf("[AI] Uploading file to GCS with temporary name: %s", tempObjectName)
	fileSize, err := gcs.UploadFileWithCustomName(ctx, file, tempObjectName, req.File.Header.Get("Content-Type"), req.File.Filename)
	if err != nil {
		log.Printf("[AI] Failed to upload file to GCS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to storage"})
		return
	}

	log.Printf("[AI] File uploaded to GCS temporarily: %s (size: %d bytes)", tempObjectName, fileSize)

	// Verify GCS file accessibility
	log.Printf("[AI] Verifying GCS file accessibility...")
	exists, err := gcs.ObjectExists(ctx, tempObjectName)
	if err != nil {
		log.Printf("[AI] Error checking GCS file existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify file upload"})
		return
	}
	if !exists {
		log.Printf("[AI] GCS file does not exist: %s", tempObjectName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File upload verification failed"})
		return
	}
	log.Printf("[AI] GCS file verified successfully")

	// Generate unique file ID for API access
	fileID := uuid.New().String()

	// Update context with file ID for correlation tracking
	ctx = context.WithValue(ctx, "file_id", fileID)

	// Generate final object name for when processing completes
	finalObjectName := fmt.Sprintf("documents/%s%s", fileID, fileExtension)

	idMetadata := map[string]string{
		"file_id":           fileID,
		"final_object_name": finalObjectName,
	}
	logger.InfoWithMetrics(ctx, "id_generation", "File ID and final object name generated", 0, idMetadata)

	// Store document metadata in database with pending status
	document := model.NewDocument(
		fileID,
		req.File.Filename,
		bucketName,
		tempObjectName, // Initially store with temp name
		req.File.Header.Get("Content-Type"),
		req.CorpusName,
		"",       // RAG file ID will be set during background processing
		"system", // TODO: Get from authentication context
		fileSize,
	)

	if err := docRepo.CreateDocument(ctx, document); err != nil {
		logger.ErrorWithOperation(ctx, "database_store", "Failed to store document metadata", err)
		// Clean up uploaded file on database error
		if deleteErr := gcs.DeleteObject(ctx, tempObjectName); deleteErr != nil {
			logger.ErrorWithOperation(ctx, "cleanup", "Failed to clean up GCS object after database error", deleteErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store document metadata"})
		return
	}

	dbMetadata := map[string]string{
		"file_id": fileID,
		"status":  document.Status,
	}
	logger.InfoWithMetrics(ctx, "database_store", "Document metadata stored successfully", 0, dbMetadata)

	// Initialize Asynq client for task enqueuing
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()

	// Create task payload
	taskPayload := tasks.DocumentTaskPayload{
		FileID:          fileID,
		TempObjectName:  tempObjectName,
		FinalObjectName: finalObjectName,
		CorpusName:      req.CorpusName,
		DisplayName:     req.File.Filename,
	}

	// Create and enqueue the background task
	task, err := tasks.NewAddDocumentToCorpusTask(taskPayload)
	if err != nil {
		logger.ErrorWithOperation(ctx, "task_creation", "Failed to create background task", err)
		// Update document status to failed
		if updateErr := docRepo.UpdateStatus(ctx, fileID, "failed", fmt.Sprintf("Task creation failed: %v", err)); updateErr != nil {
			logger.ErrorWithOperation(ctx, "status_update", "Failed to update document status after task creation error", updateErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create background processing task"})
		return
	}

	// Enqueue the task with logging
	utils.LogTaskEnqueue(ctx, tasks.TypeAddDocumentToCorpus, fileID, map[string]string{
		"corpus_name":       taskPayload.CorpusName,
		"temp_object_name":  taskPayload.TempObjectName,
		"final_object_name": taskPayload.FinalObjectName,
	})

	info, err := asynqClient.Enqueue(task)
	if err != nil {
		logger.ErrorWithOperation(ctx, "task_enqueue", "Failed to enqueue background task", err)
		// Update document status to failed
		if updateErr := docRepo.UpdateStatus(ctx, fileID, "failed", fmt.Sprintf("Task enqueue failed: %v", err)); updateErr != nil {
			logger.ErrorWithOperation(ctx, "status_update", "Failed to update document status after task enqueue error", updateErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue background processing task"})
		return
	}

	taskMetadata := map[string]string{
		"task_id":   info.ID,
		"task_type": tasks.TypeAddDocumentToCorpus,
		"file_id":   fileID,
	}
	logger.InfoWithMetrics(ctx, "task_enqueue", "Background task enqueued successfully", 0, taskMetadata)

	// Calculate response time for performance monitoring
	responseTime := time.Since(startTime)

	responseMetadata := map[string]string{
		"file_id":       fileID,
		"status":        "pending",
		"response_time": responseTime.String(),
		"corpus_name":   req.CorpusName,
		"filename":      req.File.Filename,
	}
	logger.InfoWithMetrics(ctx, "upload_complete", "Document upload request completed successfully", responseTime, responseMetadata)

	// Return immediate response with fileId and status="pending" for async processing
	c.JSON(http.StatusOK, gin.H{
		"fileId":       fileID,
		"status":       "pending",
		"message":      "Document uploaded successfully and queued for processing",
		"responseTime": responseTime.String(),
	})
}

// GetDocumentStatusHandler godoc
// @Summary      Get Document Processing Status
// @Description  Retrieve the current processing status of a document by its file ID. Returns status information including processing state, error messages (if any), and last update timestamp.
// @Tags         AI Document Management
// @Accept       json
// @Produce      json
// @Param        fileId   path    string  true  "Document file ID (unique identifier for the document)"
// @Success      200      {object}  map[string]interface{}  "Document status retrieved successfully with fileId, status, errorMsg, and updatedAt"
// @Failure      400      {object}  map[string]interface{}  "Invalid or missing file ID parameter"
// @Failure      404      {object}  map[string]interface{}  "Document not found"
// @Failure      500      {object}  map[string]interface{}  "Internal server error during status retrieval"
// @Router       /ai/rag-agent/document-status/{fileId} [get]
func GetDocumentStatusHandler(c *gin.Context) {
	// Initialize structured logging with correlation ID
	ctx := utils.WithCorrelationID(c.Request.Context(), "")
	ctx = utils.WithRequestID(ctx, c.GetHeader("X-Request-ID"))
	logger := utils.NewLogger("document_status")

	logger.InfoWithOperation(ctx, "status_request", "Document status request received")

	// Extract fileId from URL parameters with validation
	fileID := c.Param("fileId")
	if fileID == "" {
		logger.ErrorWithOperation(ctx, "parameter_validation", "Missing fileId parameter", nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId parameter is required"})
		return
	}

	// Add file ID to context for correlation tracking
	ctx = context.WithValue(ctx, "file_id", fileID)

	statusMetadata := map[string]string{
		"file_id": fileID,
	}
	logger.InfoWithMetrics(ctx, "parameter_validation", "File ID parameter validated", 0, statusMetadata)

	// Initialize document repository
	docRepo := repository.NewDocumentRepository()

	// Query MongoDB using GetDocumentByFileID repository method
	document, err := docRepo.GetDocumentByFileID(ctx, fileID)
	if err != nil {
		// Handle document not found cases with HTTP 404 responses
		if err == mongo.ErrNoDocuments {
			logger.ErrorWithOperation(ctx, "document_query", "Document not found", err)
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Document not found",
				"fileId":  fileID,
				"message": "No document found with the specified fileId",
			})
			return
		}

		// Add proper error handling for database query failures
		logger.ErrorWithOperation(ctx, "document_query", "Database error retrieving document status", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve document status",
			"fileId":  fileID,
			"message": "Database query failed",
		})
		return
	}

	resultMetadata := map[string]string{
		"file_id": fileID,
		"status":  document.Status,
	}
	if document.ErrorMsg != "" {
		resultMetadata["error_msg"] = document.ErrorMsg
	}
	logger.InfoWithMetrics(ctx, "document_query", "Document status retrieved successfully", 0, resultMetadata)

	// Return JSON response with fileId, status, errorMsg, and updatedAt
	response := gin.H{
		"fileId":    document.FileID,
		"status":    document.Status,
		"updatedAt": document.UpdatedAt,
	}

	// Include errorMsg only if it's not empty
	if document.ErrorMsg != "" {
		response["errorMsg"] = document.ErrorMsg
	}

	c.JSON(http.StatusOK, response)
}

// addVertexAICorpusDocument adds a document from GCS or Google Drive to a RAG corpus
func addVertexAICorpusDocument(corpusName, fileLink string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Check GCS object existence before ingestion
	gcsBucket, gcsObject := "", ""
	if strings.HasPrefix(fileLink, "gs://") {
		parts := strings.SplitN(strings.TrimPrefix(fileLink, "gs://"), "/", 2)
		if len(parts) == 2 {
			gcsBucket = parts[0]
			gcsObject = parts[1]
		}
	}
	if gcsBucket == "" || gcsObject == "" {
		return nil, fmt.Errorf("invalid GCS URI: %s", fileLink)
	}
	gcsService, err := service.NewGCSService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS service: %v", err)
	}
	exists, err := gcsService.ObjectExists(ctx, gcsObject)
	if err != nil {
		return nil, fmt.Errorf("error checking GCS object existence: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("GCS object does not exist: gs://%s/%s", gcsBucket, gcsObject)
	}

	// Project configuration - force RAG-compatible location
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // Default fallback
	}

	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)

	// Create AI Platform service client with regional endpoint (using ADC)
	serviceClient, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Clean corpus name for use as display name
	displayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")

	// Find the corpus first
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	listCall := serviceClient.Projects.Locations.RagCorpora.List(parent)

	existingCorpora, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list corpora: %v", err)
	}

	var corpusResourceName string
	for _, corpus := range existingCorpora.RagCorpora {
		if corpus.DisplayName == displayName {
			corpusResourceName = corpus.Name
			break
		}
	}

	if corpusResourceName == "" {
		return nil, fmt.Errorf("corpus '%s' not found", corpusName)
	}

	// Create the import request
	importRequest := &aiplatform.GoogleCloudAiplatformV1ImportRagFilesRequest{
		ImportRagFilesConfig: &aiplatform.GoogleCloudAiplatformV1ImportRagFilesConfig{
			GcsSource: &aiplatform.GoogleCloudAiplatformV1GcsSource{
				Uris: []string{fileLink},
			},
		},
	}

	// Import the file to the corpus
	operation, err := serviceClient.Projects.Locations.RagCorpora.RagFiles.Import(corpusResourceName, importRequest).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to import file to corpus: %v", err)
	}

	return map[string]interface{}{
		"status":    "success",
		"message":   "File import initiated",
		"operation": operation.Name,
		"fileLink":  fileLink,
		"corpus":    corpusResourceName,
	}, nil
}

// deleteVertexAICorpusDocument deletes a specific document from a RAG corpus
// The fileIdentifier can be either a fileId (database ID), RAG file ID, or display name
func deleteVertexAICorpusDocument(corpusName, fileIdentifier string) (map[string]interface{}, error) {
	ctx := context.Background()
	log.Printf("[AI] Starting comprehensive document deletion for '%s' in corpus '%s'", fileIdentifier, corpusName)

	// Initialize services
	gcs, err := service.NewGCSService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS service: %v", err)
	}
	defer gcs.Close()

	docRepo := repository.NewDocumentRepository()

	// First, try to find document by fileIdentifier in database
	var documentToDelete *model.Document
	document, err := docRepo.GetDocumentByFileID(ctx, fileIdentifier)
	if err == nil {
		documentToDelete = document
		log.Printf("[AI] Found document by file ID match: %s", documentToDelete.FileID)
	} else {
		// If not found by file ID, try to find by other criteria
		documents, err := docRepo.GetDocumentsByCorpus(ctx, corpusName)
		if err != nil {
			log.Printf("[AI] Warning: Failed to get documents from database: %v", err)
		} else {
			for _, doc := range documents {
				if doc.RAGFileID == fileIdentifier || doc.DisplayName == fileIdentifier {
					documentToDelete = &doc
					log.Printf("[AI] Found document by RAG file ID or display name match: %s", documentToDelete.FileID)
					break
				}
			}
		}
	}

	// Project configuration for RAG operations
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // Default fallback
	}

	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)

	// Create AI Platform service client with default ADC credentials
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Clean corpus name for use as display name
	displayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")

	// Find the corpus first
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	listCall := service.Projects.Locations.RagCorpora.List(parent)

	existingCorpora, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list corpora: %v", err)
	}

	var corpusResourceName string
	for _, corpus := range existingCorpora.RagCorpora {
		if corpus.DisplayName == displayName {
			corpusResourceName = corpus.Name
			break
		}
	}

	if corpusResourceName == "" {
		return nil, fmt.Errorf("corpus '%s' not found", corpusName)
	}

	// Determine the RAG file to delete
	var fileToDelete string

	// Case 1: If fileIdentifier looks like a RAG file ID (numeric), use it directly
	if isNumeric(fileIdentifier) {
		// Construct the full RAG file resource name
		fileToDelete = fmt.Sprintf("%s/ragFiles/%s", corpusResourceName, fileIdentifier)
		log.Printf("[AI] Using fileIdentifier as RAG file ID: %s -> %s", fileIdentifier, fileToDelete)
	} else {
		// Always resolve display name/UUID to numeric RAG file ID by listing corpus files
		log.Printf("[AI] Resolving fileIdentifier '%s' to numeric RAG file ID by listing corpus files", fileIdentifier)
		filesResponse, err := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list files in corpus: %v", err)
		}
		log.Printf("[AI] Found %d files in corpus '%s'", len(filesResponse.RagFiles), corpusName)

		// Search term priority: fileIdentifier, then document display name if available, then RAGFileID from DB
		searchTerms := []string{fileIdentifier}
		if documentToDelete != nil {
			if documentToDelete.DisplayName != "" {
				searchTerms = append(searchTerms, documentToDelete.DisplayName)
			}
			if documentToDelete.RAGFileID != "" {
				searchTerms = append(searchTerms, documentToDelete.RAGFileID)
			}
		}

		for _, file := range filesResponse.RagFiles {
			log.Printf("[AI] Checking file: Name='%s', DisplayName='%s'", file.Name, file.DisplayName)
			for _, searchTerm := range searchTerms {
				// Match by display name (with or without extension), or by UUID substring
				if file.DisplayName == searchTerm ||
					strings.TrimSuffix(file.DisplayName, filepath.Ext(file.DisplayName)) == searchTerm ||
					strings.Contains(file.DisplayName, searchTerm) ||
					strings.Contains(searchTerm, file.DisplayName) {
					fileToDelete = file.Name
					log.Printf("[AI] Found matching RAG file: %s (matched with search term: %s)", fileToDelete, searchTerm)
					break
				}
			}
			if fileToDelete != "" {
				break
			}
		}
	}

	// Track deletion results
	deletionResults := map[string]interface{}{
		"ragEngineDeleted": false,
		"gcsDeleted":       false,
		"databaseDeleted":  false,
		"errors":           []string{},
	}

	// 1. Delete from RAG engine
	if fileToDelete != "" {
		log.Printf("[AI] Attempting to delete file from RAG engine: %s", fileToDelete)

		deleteCall := service.Projects.Locations.RagCorpora.RagFiles.Delete(fileToDelete)
		_, err := deleteCall.Do()
		if err != nil {
			errorMsg := fmt.Sprintf("failed to delete from RAG engine: %v", err)
			deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
			log.Printf("[AI] %s", errorMsg)
			log.Printf("[AI] Full error details: %+v", err)
		} else {
			deletionResults["ragEngineDeleted"] = true
			log.Printf("[AI] Successfully deleted file from RAG engine: %s", fileToDelete)
		}
	} else {
		errorMsg := fmt.Sprintf("file '%s' not found in RAG engine corpus '%s'", fileIdentifier, corpusName)
		deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
		log.Printf("[AI] %s", errorMsg)
	}

	// 2. Delete from GCS
	if documentToDelete != nil && documentToDelete.GCSObject != "" {
		log.Printf("[AI] Deleting file from GCS: %s", documentToDelete.GCSObject)
		err := gcs.DeleteObject(ctx, documentToDelete.GCSObject)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to delete from GCS: %v", err)
			deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
			log.Printf("[AI] %s", errorMsg)
		} else {
			deletionResults["gcsDeleted"] = true
			log.Printf("[AI] Successfully deleted from GCS")
		}
	} else {
		errorMsg := "GCS object path not found in database"
		deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
		log.Printf("[AI] %s", errorMsg)
	}

	// 3. Delete from database
	if documentToDelete != nil {
		log.Printf("[AI] Deleting document from database: %s", documentToDelete.FileID)
		err := docRepo.DeleteDocument(ctx, documentToDelete.FileID)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to delete from database: %v", err)
			deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
			log.Printf("[AI] %s", errorMsg)
		} else {
			deletionResults["databaseDeleted"] = true
			log.Printf("[AI] Successfully deleted from database")
		}
	} else {
		errorMsg := "document not found in database"
		deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
		log.Printf("[AI] %s", errorMsg)
	}

	// Determine overall status
	allDeleted := deletionResults["ragEngineDeleted"].(bool) &&
		deletionResults["gcsDeleted"].(bool) &&
		deletionResults["databaseDeleted"].(bool)

	status := "partial_success"
	message := fmt.Sprintf("Document '%s' deletion completed with some issues", fileIdentifier)

	if allDeleted {
		status = "success"
		message = fmt.Sprintf("Successfully deleted document '%s' from all locations", fileIdentifier)
	} else if len(deletionResults["errors"].([]string)) == 3 {
		status = "error"
		message = fmt.Sprintf("Failed to delete document '%s' from all locations", fileIdentifier)
	}

	log.Printf("[AI] Document deletion completed. Status: %s", status)

	return map[string]interface{}{
		"status":          status,
		"message":         message,
		"deletedFileName": fileIdentifier,
		"corpusName":      corpusName,
		"deletionResults": deletionResults,
	}, nil
}

// isNumeric checks if a string contains only numeric characters
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
