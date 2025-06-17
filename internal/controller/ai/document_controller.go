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
	gcsService "lumenslate/internal/service"
	"lumenslate/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// DeleteCorpusDocumentHandler godoc
// @Summary      Delete Document from RAG Corpus
// @Description  Delete a specific document from a RAG corpus using its display name
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  ai.DeleteCorpusDocumentRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
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
// @Summary      Generate pre-signed URL for document viewing
// @Description  Generate a time-limited pre-signed URL to view a document from GCS
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        id path string true "Document ID"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
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
	gcs, err := gcsService.NewGCSService()
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
// @Summary      Delete Document from RAG Corpus by ID
// @Description  Delete a document from RAG corpus and GCS storage by document ID
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        id path string true "Document ID"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
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
	gcs, err := gcsService.NewGCSService()
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
		log.Printf("[AI] Deleting document from Vertex AI RAG corpus...")
		_, err := deleteVertexAICorpusDocument(document.CorpusName, document.RAGFileID)
		if err != nil {
			log.Printf("[AI] Warning: Failed to delete document from Vertex AI RAG corpus: %v", err)
			// Continue with deletion even if Vertex AI deletion fails
		} else {
			log.Printf("[AI] Successfully deleted document from Vertex AI RAG corpus")
		}
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
// @Summary      Add Document to RAG Corpus
// @Description  Upload a document to GCS, add it to RAG corpus, and store metadata
// @Tags         ai
// @Accept       multipart/form-data
// @Produce      json
// @Param        corpusName formData string true "Corpus name"
// @Param        file formData file true "Document file"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/rag-agent/add-corpus-document [post]
func AddCorpusDocumentHandler(c *gin.Context) {
	log.Println("[AI] /ai/rag-agent/add-corpus-document called")

	// Log current environment configuration for debugging
	log.Printf("[AI] Environment check - GCS_BUCKET_NAME: %s, GOOGLE_CLOUD_PROJECT: %s, GOOGLE_CLOUD_LOCATION: %s",
		func() string {
			if v := os.Getenv("GCS_BUCKET_NAME"); v != "" {
				return v
			} else {
				return "NOT_SET"
			}
		}(),
		func() string {
			if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
				return v
			} else {
				return "NOT_SET"
			}
		}(),
		func() string {
			if v := os.Getenv("GOOGLE_CLOUD_LOCATION"); v != "" {
				return v
			} else {
				return "NOT_SET"
			}
		}())

	// Validate required environment variables
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		log.Printf("[AI] GCS_BUCKET_NAME environment variable not set")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Storage configuration missing: GCS_BUCKET_NAME environment variable is required",
			"details": "Please set the GCS_BUCKET_NAME environment variable to your Google Cloud Storage bucket name",
		})
		return
	}

	projectID := utils.GetProjectID()
	if projectID == "" {
		log.Printf("[AI] Project ID not configured")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Project configuration missing: Google Cloud Project ID is required",
			"details": "Please ensure GOOGLE_CLOUD_PROJECT or project ID is properly configured",
		})
		return
	}

	log.Printf("[AI] Using GCS bucket: %s, Project: %s", bucketName, projectID)

	// Parse form data
	var req AddCorpusDocumentFormRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[AI] Invalid form request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AI] Request: corpusName=%s, file=%s", req.CorpusName, req.File.Filename)

	// Initialize services
	gcs, err := gcsService.NewGCSService()
	if err != nil {
		log.Printf("[AI] Failed to initialize GCS service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize storage service"})
		return
	}
	defer gcs.Close()

	docRepo := repository.NewDocumentRepository()
	ctx := context.Background()

	// Open uploaded file
	file, err := req.File.Open()
	if err != nil {
		log.Printf("[AI] Failed to open uploaded file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read uploaded file"})
		return
	}
	defer file.Close()

	// Generate temporary object name for initial upload
	tempFileID := uuid.New().String()
	fileExtension := filepath.Ext(req.File.Filename)
	tempObjectName := fmt.Sprintf("temp/%s%s", tempFileID, fileExtension)

	// Upload file to GCS with temporary name
	log.Printf("[AI] Uploading file to GCS with temporary name: %s", tempObjectName)
	fileSize, err := gcs.UploadFileWithCustomName(ctx, file, tempObjectName, req.File.Header.Get("Content-Type"), req.File.Filename)
	if err != nil {
		log.Printf("[AI] Failed to upload file to GCS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to storage"})
		return
	}

	log.Printf("[AI] File uploaded to GCS temporarily: %s (size: %d bytes)", tempObjectName, fileSize)

	// Generate GCS URL for Vertex AI to access
	gcsURL := fmt.Sprintf("gs://%s/%s", os.Getenv("GCS_BUCKET_NAME"), tempObjectName)
	log.Printf("[AI] GCS URL for Vertex AI: %s", gcsURL)

	// Validate file type for RAG compatibility
	allowedExtensions := []string{".pdf", ".txt", ".docx", ".doc", ".html", ".md"}
	isValidType := false
	for _, ext := range allowedExtensions {
		if strings.EqualFold(fileExtension, ext) {
			isValidType = true
			break
		}
	}

	if !isValidType {
		log.Printf("[AI] Unsupported file type: %s. Allowed types: %v", fileExtension, allowedExtensions)
		// Clean up uploaded file
		if deleteErr := gcs.DeleteObject(ctx, tempObjectName); deleteErr != nil {
			log.Printf("[AI] Failed to clean up GCS object: %v", deleteErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported file type: %s. Allowed types: %v", fileExtension, allowedExtensions)})
		return
	}

	log.Printf("[AI] File type %s is valid for RAG import", fileExtension)

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
	log.Printf("[AI] GCS file verified: %s", gcsURL)

	// Check if corpus exists, create if it doesn't
	log.Printf("[AI] Checking if corpus '%s' exists before adding document", req.CorpusName)
	corpusResponse, err := createVertexAICorpus(req.CorpusName)
	if err != nil {
		log.Printf("[AI] Warning: Could not create/verify corpus '%s': %v", req.CorpusName, err)
		// Continue with document addition even if corpus creation fails
	} else {
		log.Printf("[AI] Corpus operation result: %s", corpusResponse["message"])
	}

	// Store the count of existing files before adding
	existingFiles, err := listVertexAICorpusContent(req.CorpusName)
	var existingFileCount int
	if err == nil {
		if files, ok := existingFiles["files"].([]interface{}); ok {
			existingFileCount = len(files)
		}
	}
	log.Printf("[AI] Existing files in corpus before addition: %d", existingFileCount)

	// Add document to Vertex AI RAG corpus using GCS URL
	log.Printf("[AI] Starting to add document to RAG corpus...")
	addResult, err := addVertexAICorpusDocument(req.CorpusName, gcsURL)
	if err != nil {
		log.Printf("[AI] Add corpus document error: %v", err)

		// Clean up: delete the uploaded file from GCS
		if deleteErr := gcs.DeleteObject(ctx, tempObjectName); deleteErr != nil {
			log.Printf("[AI] Failed to clean up GCS object after Vertex AI error: %v", deleteErr)
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to add document to corpus: %v", err)})
		return
	}

	log.Printf("[AI] Successfully added document to RAG corpus. Result: %+v", addResult)

	// Wait a bit for the operation to complete and verify
	log.Printf("[AI] Waiting for RAG import operation to complete...")
	time.Sleep(10 * time.Second) // Increased wait time

	// List files again to find the newly added RAG file
	log.Printf("[AI] Verifying document was added to RAG corpus...")
	updatedFiles, err := listVertexAICorpusContent(req.CorpusName)
	var ragFileID string
	var documentAdded bool

	if err == nil {
		if files, ok := updatedFiles["files"].([]interface{}); ok {
			log.Printf("[AI] Found %d files in corpus after addition (was %d before)", len(files), existingFileCount)

			// Find the newly added file by comparing with previous count
			if len(files) > existingFileCount {
				// Look for the file with matching display name
				targetFileName := req.File.Filename
				for _, file := range files {
					if fileMap, ok := file.(map[string]interface{}); ok {
						if displayName, ok := fileMap["displayName"].(string); ok && displayName == targetFileName {
							if fileID, ok := fileMap["id"].(string); ok {
								ragFileID = fileID
								documentAdded = true
								log.Printf("[AI] Found newly added RAG file: %s (ID: %s)", displayName, ragFileID)
								break
							}
						}
					}
				}
			}
		}
	} else {
		log.Printf("[AI] Warning: Could not verify document addition: %v", err)
	}

	if !documentAdded {
		log.Printf("[AI] Warning: Document may not have been successfully added to RAG corpus")
		// Don't fail the request, but log the issue
	}

	// If we couldn't get the RAG file ID, use a fallback
	if ragFileID == "" {
		log.Printf("[AI] Warning: Could not determine RAG file ID, using temp file ID as fallback")
		ragFileID = tempFileID
	}

	// Rename the GCS object to use the RAG file ID
	finalObjectName := fmt.Sprintf("documents/%s%s", ragFileID, fileExtension)
	log.Printf("[AI] Renaming GCS object from %s to %s", tempObjectName, finalObjectName)

	if err := gcs.RenameObject(ctx, tempObjectName, finalObjectName); err != nil {
		log.Printf("[AI] Warning: Failed to rename GCS object: %v", err)
		// Continue with temp name if rename fails
		finalObjectName = tempObjectName
	}

	// Generate unique file ID for API access
	fileID := uuid.New().String()

	// Store document metadata in database
	document := &model.Document{
		FileID:      fileID,
		DisplayName: req.File.Filename,
		GCSBucket:   os.Getenv("GCS_BUCKET_NAME"),
		GCSObject:   finalObjectName,
		ContentType: req.File.Header.Get("Content-Type"),
		Size:        fileSize,
		CorpusName:  req.CorpusName,
		RAGFileID:   ragFileID,
		UploadedBy:  "system", // TODO: Get from authentication context
	}

	if err := docRepo.CreateDocument(ctx, document); err != nil {
		log.Printf("[AI] Failed to store document metadata: %v", err)
		// Don't fail the request, just log the error
	} else {
		log.Printf("[AI] Document metadata stored successfully with fileID: %s", fileID)
	}

	log.Printf("[AI] Add corpus document success")
	c.JSON(http.StatusOK, gin.H{
		"message":     "Document uploaded and added to corpus successfully",
		"fileId":      fileID,
		"displayName": req.File.Filename,
		"size":        fileSize,
		"corpusName":  req.CorpusName,
		"ragFileId":   ragFileID,
		"gcsObject":   finalObjectName,
	})
}

// addVertexAICorpusDocument adds a document from GCS or Google Drive to a RAG corpus
func addVertexAICorpusDocument(corpusName, fileLink string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Project configuration - force RAG-compatible location
	projectID := utils.GetProjectID()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // Default fallback
	}

	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)

	// Create AI Platform service client with regional endpoint (using ADC)
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

	// Create the import request
	importRequest := &aiplatform.GoogleCloudAiplatformV1ImportRagFilesRequest{
		ImportRagFilesConfig: &aiplatform.GoogleCloudAiplatformV1ImportRagFilesConfig{
			GcsSource: &aiplatform.GoogleCloudAiplatformV1GcsSource{
				Uris: []string{fileLink},
			},
		},
	}

	// Import the file to the corpus
	operation, err := service.Projects.Locations.RagCorpora.RagFiles.Import(corpusResourceName, importRequest).Do()
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
	gcs, err := gcsService.NewGCSService()
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
	projectID := utils.GetProjectID()
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

	// Determine search term for RAG engine
	var ragSearchTerm string
	var fileToDelete string

	if documentToDelete != nil && documentToDelete.RAGFileID != "" {
		ragSearchTerm = documentToDelete.RAGFileID
		log.Printf("[AI] Searching for RAG file using RAG file ID from database: '%s'", ragSearchTerm)

		filesResponse, err := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list files in corpus: %v", err)
		}

		log.Printf("[AI] Found %d files in corpus '%s'", len(filesResponse.RagFiles), corpusName)

		for _, file := range filesResponse.RagFiles {
			log.Printf("[AI] Checking file: Name='%s', DisplayName='%s'", file.Name, file.DisplayName)

			// Try multiple matching strategies
			if file.DisplayName == ragSearchTerm ||
				strings.Contains(file.Name, ragSearchTerm) ||
				strings.Contains(file.DisplayName, ragSearchTerm) ||
				strings.Contains(file.DisplayName, documentToDelete.DisplayName) ||
				// Handle case where RAG file ID doesn't have extension but display name does
				strings.HasPrefix(file.DisplayName, ragSearchTerm+".") {
				fileToDelete = file.Name
				log.Printf("[AI] Found matching RAG file: %s (matched display name: %s)", fileToDelete, file.DisplayName)
				break
			}
		}
	} else {
		log.Printf("[AI] No RAG file ID in database, trying to find by display name: '%s'", documentToDelete.DisplayName)
		ragSearchTerm = documentToDelete.DisplayName

		filesResponse, err := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list files in corpus: %v", err)
		}

		for _, file := range filesResponse.RagFiles {
			log.Printf("[AI] Checking file: Name='%s', DisplayName='%s'", file.Name, file.DisplayName)

			if file.DisplayName == ragSearchTerm ||
				strings.Contains(file.DisplayName, ragSearchTerm) ||
				strings.Contains(ragSearchTerm, file.DisplayName) {
				fileToDelete = file.Name
				log.Printf("[AI] Found matching RAG file by display name: %s", fileToDelete)
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
