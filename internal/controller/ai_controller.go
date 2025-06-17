package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	service "lumenslate/internal/grpc_service"
	"lumenslate/internal/model"
	"lumenslate/internal/repository"
	gcsService "lumenslate/internal/service"
	"lumenslate/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// --- Request Structs for Swagger ---
type GenerateContextRequest struct {
	Question string   `json:"question"`
	Keywords []string `json:"keywords"`
	Language string   `json:"language"`
}

type DetectVariablesRequest struct {
	Question string `json:"question"`
}

type SegmentQuestionRequest struct {
	Question string `json:"question"`
}

type GenerateMCQVariationsRequest struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AnswerIndex int32    `json:"answerIndex"`
}

type GenerateMSQVariationsRequest struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	AnswerIndices []int32  `json:"answerIndices"`
}

type FilterAndRandomizeRequest struct {
	Question   string `json:"question"`
	UserPrompt string `json:"userPrompt"`
}

type CreateCorpusRequest struct {
	CorpusName string `json:"corpusName" binding:"required"`
}

type DeleteCorpusDocumentRequest struct {
	CorpusName string `json:"corpusName" binding:"required"`
	FileID     string `json:"fileId" binding:"required"` // Can be fileId, RAG file ID, or display name
}

type AddCorpusDocumentFormRequest struct {
	CorpusName string                `form:"corpusName" binding:"required"`
	File       *multipart.FileHeader `form:"file" binding:"required"`
}

type ViewDocumentRequest struct {
	DocumentID string `uri:"id" binding:"required"`
}

type AgentRequest struct {
	TeacherId string                `form:"teacherId" binding:"required"`
	Role      string                `form:"role" binding:"required"`
	Message   string                `form:"message" binding:"required"`
	File      *multipart.FileHeader `form:"file"`
	FileType  string                `form:"fileType"`
	CreatedAt string                `form:"createdAt"`
	UpdatedAt string                `form:"updatedAt"`
}

type RAGAgentRequest struct {
	TeacherId string `json:"teacherId" binding:"required"`
	Role      string `json:"role" binding:"required"`
	Message   string `json:"message" binding:"required"`
	File      string `json:"file"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// GenerateContextHandler godoc
// @Summary      Generate context for a question
// @Description  Generates context using AI for the given question, keywords, and language
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.GenerateContextRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/generate-context [post]
func GenerateContextHandler(c *gin.Context) {
	var req GenerateContextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	content, err := service.GenerateContext(req.Question, req.Keywords, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": content})
}

// DetectVariablesHandler godoc
// @Summary      Detect variables in a question
// @Description  Detects variables in the provided question using AI
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.DetectVariablesRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/detect-variables [post]
func DetectVariablesHandler(c *gin.Context) {
	var req DetectVariablesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variables, err := service.DetectVariables(req.Question)
	if err != nil {
		log.Printf("[AI] DetectVariables error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"variables": variables,
	})
}

// SegmentQuestionHandler godoc
// @Summary      Segment a question
// @Description  Segments the provided question using AI
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.SegmentQuestionRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/segment-question [post]
func SegmentQuestionHandler(c *gin.Context) {
	var req SegmentQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	segmented, err := service.SegmentQuestion(req.Question)
	if err != nil {
		log.Printf("[AI] SegmentQuestion error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"segmentedQuestion": segmented})
}

// GenerateMCQVariationsHandler godoc
// @Summary      Generate MCQ variations
// @Description  Generates MCQ variations for a question using AI
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.GenerateMCQVariationsRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/generate-mcq [post]
func GenerateMCQVariationsHandler(c *gin.Context) {
	log.Println("[AI] /ai/generate-mcq called")
	var req GenerateMCQVariationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)
	variations, err := service.GenerateMCQVariations(req.Question, req.Options, req.AnswerIndex)
	if err != nil {
		log.Printf("[AI] GenerateMCQVariations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] GenerateMCQVariations success, count: %d", len(variations))
	c.JSON(http.StatusOK, gin.H{"variations": variations})
}

// GenerateMSQVariationsHandler godoc
// @Summary      Generate MSQ variations
// @Description  Generates MSQ variations for a question using AI
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.GenerateMSQVariationsRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/generate-msq [post]
func GenerateMSQVariationsHandler(c *gin.Context) {
	log.Println("[AI] /ai/generate-msq called")
	var req GenerateMSQVariationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)
	variations, err := service.GenerateMSQVariations(req.Question, req.Options, req.AnswerIndices)
	if err != nil {
		log.Printf("[AI] GenerateMSQVariations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] GenerateMSQVariations success, count: %d", len(variations))
	c.JSON(http.StatusOK, gin.H{"variations": variations})
}

// FilterAndRandomizeHandler godoc
// @Summary      Filter and randomize variables
// @Description  Filters and randomizes variables in a question using AI
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.FilterAndRandomizeRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/filter-randomize [post]
func FilterAndRandomizeHandler(c *gin.Context) {
	log.Println("[AI] /ai/filter-randomize called")
	var req FilterAndRandomizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)
	vars, err := service.FilterAndRandomize(req.Question, req.UserPrompt)
	if err != nil {
		log.Printf("[AI] FilterAndRandomize error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] FilterAndRandomize success, count: %d", len(vars))
	c.JSON(http.StatusOK, gin.H{"variables": vars})
}

// AgentHandler godoc
// @Summary      Call Agent AI method
// @Description  Calls the Agent gRPC method with the provided data including file upload support
// @Tags         ai
// @Accept       multipart/form-data
// @Produce      json
// @Param        teacherId formData string true "Teacher ID"
// @Param        role formData string true "Role"
// @Param        message formData string true "Message"
// @Param        file formData file false "File upload"
// @Param        fileType formData string false "File type"
// @Param        createdAt formData string false "Created at timestamp"
// @Param        updatedAt formData string false "Updated at timestamp"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/agent [post]
func AgentHandler(c *gin.Context) {
	log.Println("[AI] /ai/agent called")
	var req AgentRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Agent Request: %+v", req)

	// Process file upload if present
	var fileContent string
	if req.File != nil {
		// Open the uploaded file
		file, err := req.File.Open()
		if err != nil {
			log.Printf("[AI] Error opening uploaded file: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open uploaded file"})
			return
		}
		defer file.Close()

		// Read file content
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf("[AI] Error reading file content: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file content"})
			return
		}

		// Convert to base64 string for service layer
		fileContent = base64.StdEncoding.EncodeToString(fileBytes)
		log.Printf("[AI] File processed: %s, size: %d bytes", req.File.Filename, len(fileBytes))
	}

	resp, err := service.Agent(
		fileContent,
		req.FileType,
		req.TeacherId,
		req.Role,
		req.Message,
		req.CreatedAt,
		req.UpdatedAt,
	)
	if err != nil {
		log.Printf("[AI] Agent error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Agent success")
	c.JSON(http.StatusOK, resp)
}

// RAGAgentHandler godoc
// @Summary      Call RAG Agent AI method
// @Description  Process text input using RAG agent for knowledge retrieval
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body controller.RAGAgentRequest true "RAG agent request"
// @Success      200 {object} map[string]interface{} "RAG agent response"
// @Failure      400 {object} gin.H "Invalid request"
// @Failure      500 {object} gin.H "Internal server error"
// @Router       /ai/rag-agent [post]
func RAGAgentHandler(c *gin.Context) {
	// Parse and validate request
	var req RAGAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERROR: Failed to bind JSON request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create/verify corpus for the teacher before processing the request
	_, err := createVertexAICorpus(req.TeacherId)
	if err != nil {
		log.Printf("WARNING: Could not create/verify corpus for teacher %s: %v", req.TeacherId, err)
		// Continue processing even if corpus creation fails
	}

	// Call the gRPC microservice
	resp, err := service.RAGAgentClient(req.TeacherId, req.Message, "")
	if err != nil {
		log.Printf("ERROR: Failed to process RAG agent request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process RAG agent request", "error": err.Error()})
		return
	}

	// Process the agent response to determine data content and message
	responseData, responseMessage, err := processRAGAgentResponse(resp.GetAgentResponse(), req.TeacherId)
	if err != nil {
		log.Printf("ERROR: Failed to process RAG agent response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process RAG agent response", "error": err.Error()})
		return
	}

	// Build standardized response format (matching /ai/agent structure)
	standardizedResponse := map[string]interface{}{
		"message":      responseMessage,
		"teacherId":    resp.GetTeacherId(),
		"agentName":    resp.GetAgentName(),
		"data":         responseData,
		"sessionId":    resp.GetSessionId(),
		"createdAt":    resp.GetCreatedAt(),
		"updatedAt":    resp.GetUpdatedAt(),
		"responseTime": resp.GetResponseTime(),
		"role":         resp.GetRole(),
		"feedback":     resp.GetFeedback(),
	}

	c.JSON(http.StatusOK, standardizedResponse)
}

// CreateCorpusHandler godoc
// @Summary      Create RAG Corpus
// @Description  Create a new RAG corpus directly in Vertex AI
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.CreateCorpusRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/rag-agent/create-corpus [post]
func CreateCorpusHandler(c *gin.Context) {
	log.Println("[AI] /ai/rag-agent/create-corpus called")
	var req CreateCorpusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)

	// Create corpus using Vertex AI
	corpusResponse, err := createVertexAICorpus(req.CorpusName)
	if err != nil {
		log.Printf("[AI] CreateCorpus error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AI] CreateCorpus success")
	c.JSON(http.StatusOK, corpusResponse)
}

// ListCorpusContentHandler godoc
// @Summary      List RAG Corpus Content
// @Description  List all documents/files inside a RAG corpus
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.CreateCorpusRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/rag-agent/list-corpus-content [post]
func ListCorpusContentHandler(c *gin.Context) {
	log.Println("[AI] /ai/rag-agent/list-corpus-content called")
	var req CreateCorpusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)

	// List corpus content using Vertex AI
	contentResponse, err := listVertexAICorpusContent(req.CorpusName)
	if err != nil {
		log.Printf("[AI] ListCorpusContent error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AI] ListCorpusContent success")
	c.JSON(http.StatusOK, contentResponse)
}

// createVertexAICorpus creates a RAG corpus directly in Vertex AI
func createVertexAICorpus(corpusName string) (map[string]interface{}, error) {
	log.Printf("[AI] createVertexAICorpus called with corpusName: %s", corpusName)
	ctx := context.Background()
	// comment
	// Project configuration - force RAG-compatible location
	projectID := utils.GetProjectID()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	log.Printf("[AI] Using projectID: %s, location: %s", projectID, location)
	if location == "" {
		location = "us-central1" // Default fallback
		log.Printf("[AI] GOOGLE_CLOUD_LOCATION not set, defaulting to us-central1")
	}

	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)
	log.Printf("[AI] Using endpoint: %s", endpoint)

	// Create AI Platform service client with default credentials
	service, err := aiplatform.NewService(ctx,
		option.WithEndpoint(endpoint)) // ⬅️ Removed WithCredentialsFile
	if err != nil {
		log.Printf("[AI] Failed to create AI Platform service: %v", err)
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Clean corpus name for use as display name
	displayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")
	log.Printf("[AI] Cleaned displayName: %s", displayName)

	// Check if corpus already exists
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	log.Printf("[AI] Checking for existing corpora under parent: %s", parent)
	listCall := service.Projects.Locations.RagCorpora.List(parent)

	existingCorpora, err := listCall.Do()
	if err != nil {
		log.Printf("[AI] Warning: Could not check existing corpora: %v", err)
	} else {
		for _, corpus := range existingCorpora.RagCorpora {
			log.Printf("[AI] Found existing corpus: %s (displayName: %s)", corpus.Name, corpus.DisplayName)
			if corpus.DisplayName == displayName {
				log.Printf("[AI] Corpus '%s' already exists", corpusName)
				return map[string]interface{}{
					"status":        "info",
					"message":       fmt.Sprintf("Corpus '%s' already exists", corpusName),
					"corpusName":    corpus.Name,
					"displayName":   corpus.DisplayName,
					"corpusCreated": false,
				}, nil
			}
		}
	}

	// Create the corpus
	ragCorpus := &aiplatform.GoogleCloudAiplatformV1RagCorpus{
		DisplayName: displayName,
	}
	log.Printf("[AI] Creating new corpus with displayName: %s", displayName)
	createCall := service.Projects.Locations.RagCorpora.Create(parent, ragCorpus)
	operation, err := createCall.Do()
	if err != nil {
		log.Printf("[AI] Failed to create corpus: %v", err)
		return nil, fmt.Errorf("failed to create corpus: %v", err)
	}

	log.Printf("[AI] Successfully created corpus '%s' (operation: %s)", corpusName, operation.Name)
	return map[string]interface{}{
		"status":        "success",
		"message":       fmt.Sprintf("Successfully created corpus '%s'", corpusName),
		"operationName": operation.Name,
		"displayName":   displayName,
		"corpusCreated": true,
	}, nil
}

// listVertexAICorpusContent lists all documents/files inside a RAG corpus
func listVertexAICorpusContent(corpusName string) (map[string]interface{}, error) {
	log.Printf("[AI] listVertexAICorpusContent called with corpusName: %s", corpusName)
	ctx := context.Background()

	// Project configuration - force RAG-compatible location
	projectID := utils.GetProjectID()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	log.Printf("[AI] Using projectID: %s, location: %s", projectID, location)
	if location == "" {
		location = "us-central1" // Default fallback
		log.Printf("[AI] GOOGLE_CLOUD_LOCATION not set, defaulting to us-central1")
	}

	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)
	log.Printf("[AI] Using endpoint: %s", endpoint)

	// Create AI Platform service client with default credentials
	service, err := aiplatform.NewService(ctx,
		option.WithEndpoint(endpoint)) // ⬅️ removed WithCredentialsFile
	if err != nil {
		log.Printf("[AI] Failed to create AI Platform service: %v", err)
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Clean corpus name for use as display name
	displayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")
	log.Printf("[AI] Cleaned displayName: %s", displayName)

	// Find the corpus first
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	log.Printf("[AI] Looking for corpus under parent: %s", parent)
	listCall := service.Projects.Locations.RagCorpora.List(parent)

	existingCorpora, err := listCall.Do()
	if err != nil {
		log.Printf("[AI] Failed to list corpora: %v", err)
		return nil, fmt.Errorf("failed to list corpora: %v", err)
	}

	var corpusResourceName string
	var foundCorpus *aiplatform.GoogleCloudAiplatformV1RagCorpus

	// Match by displayName
	for _, corpus := range existingCorpora.RagCorpora {
		log.Printf("[AI] Found corpus: %s (displayName: %s)", corpus.Name, corpus.DisplayName)
		if corpus.DisplayName == displayName {
			corpusResourceName = corpus.Name
			foundCorpus = corpus
			break
		}
	}

	if corpusResourceName == "" {
		log.Printf("[AI] Corpus '%s' not found", corpusName)
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Corpus '%s' not found", corpusName),
			"files":   []interface{}{},
		}, nil
	}

	// List files inside the corpus
	log.Printf("[AI] Listing files in corpus: %s", corpusResourceName)
	listFilesCall := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName)
	filesResponse, err := listFilesCall.Do()
	if err != nil {
		log.Printf("[AI] Failed to list files in corpus: %v", err)
		return nil, fmt.Errorf("failed to list files in corpus: %v", err)
	}

	// Format files response
	var files []map[string]interface{}
	for _, file := range filesResponse.RagFiles {
		log.Printf("[AI] Found file: %s (displayName: %s)", file.Name, file.DisplayName)

		// Extract file ID from the resource name (format: projects/.../ragFiles/{fileId})
		fileID := ""
		if file.Name != "" {
			// Split by "/" and get the last part which is the file ID
			parts := strings.Split(file.Name, "/")
			if len(parts) > 0 {
				fileID = parts[len(parts)-1]
			}
		}

		files = append(files, map[string]interface{}{
			"id":          fileID,
			"name":        file.Name,
			"displayName": file.DisplayName,
			"description": file.Description,
			"createTime":  file.CreateTime,
			"updateTime":  file.UpdateTime,
		})
	}

	log.Printf("[AI] Successfully listed %d files for corpus '%s'", len(files), corpusName)
	return map[string]interface{}{
		"status":      "success",
		"message":     fmt.Sprintf("Successfully listed content for corpus '%s'", corpusName),
		"corpusName":  foundCorpus.Name,
		"displayName": foundCorpus.DisplayName,
		"filesCount":  len(files),
		"files":       files,
	}, nil
}

// DeleteCorpusDocumentHandler godoc
// @Summary      Delete Document from RAG Corpus
// @Description  Delete a specific document from a RAG corpus using its display name
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.DeleteCorpusDocumentRequest  true  "Request body"
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

// ListAllCorporaHandler godoc
// @Summary      List All RAG Corpora
// @Description  List the names of all RAG corpora
// @Tags         ai
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/rag-agent/list-all-corpora [post]
func ListAllCorporaHandler(c *gin.Context) {
	log.Println("[AI] /ai/rag-agent/list-all-corpora called")

	corporaResponse, err := listAllVertexAICorpora()
	if err != nil {
		log.Printf("[AI] List all corpora error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list corpora: %v", err)})
		return
	}

	log.Printf("[AI] List all corpora success")
	c.JSON(http.StatusOK, corporaResponse)
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

// listAllVertexAICorpora lists all RAG corpora names
func listAllVertexAICorpora() (map[string]interface{}, error) {
	log.Printf("[AI] listAllVertexAICorpora called")
	ctx := context.Background()

	// Project configuration - force RAG-compatible location
	projectID := utils.GetProjectID()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	log.Printf("[AI] Using projectID: %s, location: %s", projectID, location)
	if location == "" {
		location = "us-central1" // Default fallback
		log.Printf("[AI] GOOGLE_CLOUD_LOCATION not set, defaulting to us-central1")
	}

	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)
	log.Printf("[AI] Using endpoint: %s", endpoint)

	// Create AI Platform service client with default credentials
	service, err := aiplatform.NewService(ctx,
		option.WithEndpoint(endpoint))
	if err != nil {
		log.Printf("[AI] Failed to create AI Platform service: %v", err)
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// List all corpora
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	log.Printf("[AI] Listing all corpora under parent: %s", parent)
	listCall := service.Projects.Locations.RagCorpora.List(parent)

	existingCorpora, err := listCall.Do()
	if err != nil {
		log.Printf("[AI] Failed to list corpora: %v", err)
		return nil, fmt.Errorf("failed to list corpora: %v", err)
	}

	// Format the corpora data in camelCase
	var corpora []map[string]interface{}
	for _, corpus := range existingCorpora.RagCorpora {
		log.Printf("[AI] Found corpus: %s (displayName: %s)", corpus.Name, corpus.DisplayName)
		corpora = append(corpora, map[string]interface{}{
			"name":        corpus.Name,
			"displayName": corpus.DisplayName,
			"createTime":  corpus.CreateTime,
			"updateTime":  corpus.UpdateTime,
		})
	}

	log.Printf("[AI] Successfully listed %d corpora", len(corpora))
	return map[string]interface{}{
		"status":       "success",
		"message":      "Successfully listed all corpora",
		"corporaCount": len(corpora),
		"corpora":      corpora,
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

	// Try to find the document in database using multiple strategies
	var documentToDelete *model.Document

	// Strategy 1: Try as fileId (direct database lookup)
	if doc, err := docRepo.GetDocumentByFileID(ctx, fileIdentifier); err == nil {
		documentToDelete = doc
		log.Printf("[AI] Found document by fileId: %s (DisplayName: %s, RAGFileID: %s)",
			doc.FileID, doc.DisplayName, doc.RAGFileID)
	} else {
		// Strategy 2: Search by corpus and match by display name or RAG file ID
		documents, err := docRepo.GetDocumentsByCorpus(ctx, corpusName)
		if err != nil {
			log.Printf("[AI] Warning: Failed to get documents from database: %v", err)
		} else {
			for _, doc := range documents {
				if doc.DisplayName == fileIdentifier || doc.RAGFileID == fileIdentifier {
					documentToDelete = &doc
					log.Printf("[AI] Found document by %s match: %s",
						map[bool]string{true: "display name", false: "RAG file ID"}[doc.DisplayName == fileIdentifier],
						doc.FileID)
					break
				}
			}
		}
	}

	if documentToDelete == nil {
		return nil, fmt.Errorf("document '%s' not found in database for corpus '%s'", fileIdentifier, corpusName)
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

	// List files in the corpus to find the one to delete using the RAG file ID from database
	var fileToDelete string
	var ragSearchTerm string

	if documentToDelete.RAGFileID != "" {
		// We have a RAG file ID from database, use it to find the file
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

			if file.DisplayName == documentToDelete.DisplayName ||
				strings.Contains(file.DisplayName, documentToDelete.DisplayName) ||
				strings.Contains(documentToDelete.DisplayName, file.DisplayName) {
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
		log.Printf("[AI] Deleting file from RAG engine: %s", fileToDelete)
		_, err := service.Projects.Locations.RagCorpora.RagFiles.Delete(fileToDelete).Do()
		if err != nil {
			errorMsg := fmt.Sprintf("failed to delete from RAG engine: %v", err)
			deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
			log.Printf("[AI] %s", errorMsg)
		} else {
			deletionResults["ragEngineDeleted"] = true
			log.Printf("[AI] Successfully deleted from RAG engine")
		}
	} else {
		errorMsg := fmt.Sprintf("file '%s' not found in RAG engine", ragSearchTerm)
		deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
		log.Printf("[AI] %s", errorMsg)
	}

	// 2. Delete from GCS
	if documentToDelete.GCSObject != "" {
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
	log.Printf("[AI] Deleting document from database: %s", documentToDelete.FileID)
	err = docRepo.DeleteDocument(ctx, documentToDelete.FileID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to delete from database: %v", err)
		deletionResults["errors"] = append(deletionResults["errors"].([]string), errorMsg)
		log.Printf("[AI] %s", errorMsg)
	} else {
		deletionResults["databaseDeleted"] = true
		log.Printf("[AI] Successfully deleted from database")
	}

	// Determine overall status
	allDeleted := deletionResults["ragEngineDeleted"].(bool) &&
		deletionResults["gcsDeleted"].(bool) &&
		deletionResults["databaseDeleted"].(bool)

	status := "partial_success"
	message := fmt.Sprintf("Document '%s' deletion completed with some issues", documentToDelete.DisplayName)

	if allDeleted {
		status = "success"
		message = fmt.Sprintf("Successfully deleted document '%s' from all locations", documentToDelete.DisplayName)
	} else if len(deletionResults["errors"].([]string)) == 3 {
		status = "error"
		message = fmt.Sprintf("Failed to delete document '%s' from all locations", documentToDelete.DisplayName)
	}

	log.Printf("[AI] Document deletion completed. Status: %s", status)

	return map[string]interface{}{
		"status":          status,
		"message":         message,
		"deletedFileName": documentToDelete.DisplayName,
		"fileId":          documentToDelete.FileID,
		"corpusName":      corpusName,
		"deletionResults": deletionResults,
	}, nil
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

	// Find the corpus with retry mechanism (for newly created corpora)
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	var corpusResourceName string

	// Retry up to 3 times with increasing delays to allow corpus propagation
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[AI] Attempting to find corpus '%s' (attempt %d/%d)", corpusName, attempt, maxRetries)

		existingCorpora, err := service.Projects.Locations.RagCorpora.List(parent).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list corpora: %v", err)
		}

		for _, corpus := range existingCorpora.RagCorpora {
			if corpus.DisplayName == displayName {
				corpusResourceName = corpus.Name
				log.Printf("[AI] Found corpus '%s' with resource name: %s", corpusName, corpusResourceName)
				break
			}
		}

		if corpusResourceName != "" {
			break // Found the corpus, exit retry loop
		}

		if attempt < maxRetries {
			sleepDuration := time.Duration(attempt*2) * time.Second // 2s, 4s delays
			log.Printf("[AI] Corpus '%s' not found, retrying in %v...", corpusName, sleepDuration)
			time.Sleep(sleepDuration)
		}
	}

	if corpusResourceName == "" {
		return nil, fmt.Errorf("corpus '%s' not found after %d attempts", corpusName, maxRetries)
	}

	var importRequest *aiplatform.GoogleCloudAiplatformV1ImportRagFilesRequest
	var fileDisplayName string

	// Check if it's a GCS URL or Google Drive URL
	if strings.HasPrefix(fileLink, "gs://") {
		// GCS URL format: gs://bucket/path/to/file
		log.Printf("[AI] Adding GCS file to corpus: %s", fileLink)

		// Extract filename from GCS path
		parts := strings.Split(fileLink, "/")
		if len(parts) > 0 {
			fileDisplayName = parts[len(parts)-1]
		} else {
			fileDisplayName = "gcs_file"
		}

		// Build the import request for GCS
		importRequest = &aiplatform.GoogleCloudAiplatformV1ImportRagFilesRequest{
			ImportRagFilesConfig: &aiplatform.GoogleCloudAiplatformV1ImportRagFilesConfig{
				GcsSource: &aiplatform.GoogleCloudAiplatformV1GcsSource{
					Uris: []string{fileLink},
				},
			},
		}
		log.Printf("[AI] Created GCS import request for file: %s", fileLink)
	} else {
		// Assume Google Drive URL
		log.Printf("[AI] Adding Google Drive file to corpus: %s", fileLink)

		// Extract file ID from Google Drive URL
		re := regexp.MustCompile(`/file/d/([a-zA-Z0-9_-]+)`)
		matches := re.FindStringSubmatch(fileLink)
		if len(matches) <= 1 {
			return nil, fmt.Errorf("invalid Google Drive URL format: could not extract file ID")
		}
		fileID := matches[1]

		// Fetch file display name using Drive API (using ADC, no credentialsPath)
		var driveErr error
		fileDisplayName, driveErr = getGoogleDriveFileName(fileID)
		if driveErr != nil {
			log.Printf("[AI] Warning: Could not fetch file name from Google Drive: %v. Using fallback name.", driveErr)
			fileDisplayName = fmt.Sprintf("gdrive_file_%s", fileID)
		}

		// Build the import request for Google Drive
		importRequest = &aiplatform.GoogleCloudAiplatformV1ImportRagFilesRequest{
			ImportRagFilesConfig: &aiplatform.GoogleCloudAiplatformV1ImportRagFilesConfig{
				GoogleDriveSource: &aiplatform.GoogleCloudAiplatformV1GoogleDriveSource{
					ResourceIds: []*aiplatform.GoogleCloudAiplatformV1GoogleDriveSourceResourceId{
						{
							ResourceId:   fileID,
							ResourceType: "RESOURCE_TYPE_FILE",
						},
					},
				},
			},
		}
	}

	log.Printf("[AI] Adding file '%s' to corpus '%s'", fileDisplayName, corpusName)
	log.Printf("[AI] Import request: %+v", importRequest)

	// Trigger the import
	operation, err := service.Projects.Locations.RagCorpora.RagFiles.Import(corpusResourceName, importRequest).Do()
	if err != nil {
		log.Printf("[AI] Failed to trigger import operation: %v", err)
		return nil, fmt.Errorf("failed to add file to corpus: %v", err)
	}

	log.Printf("[AI] Import operation started: %s", operation.Name)

	// Wait for the operation to complete with timeout
	operationsService := service.Projects.Locations.Operations
	timeout := time.Now().Add(5 * time.Minute) // 5 minute timeout

	for time.Now().Before(timeout) {
		op, err := operationsService.Get(operation.Name).Do()
		if err != nil {
			log.Printf("[AI] Failed to check operation status: %v", err)
			break
		}

		log.Printf("[AI] Operation status: done=%v", op.Done)

		if op.Done {
			if op.Error != nil {
				log.Printf("[AI] Import operation failed: %v", op.Error)
				return nil, fmt.Errorf("import operation failed: %s", op.Error.Message)
			}
			log.Printf("[AI] Import operation completed successfully")
			break
		}

		// Wait before checking again
		time.Sleep(10 * time.Second)
	}

	// Check if operation timed out
	if time.Now().After(timeout) {
		log.Printf("[AI] Import operation timed out")
		return nil, fmt.Errorf("import operation timed out after 5 minutes")
	}

	return map[string]interface{}{
		"status":          "success",
		"message":         fmt.Sprintf("Successfully added file to corpus '%s'", corpusName),
		"operationName":   operation.Name,
		"fileDisplayName": fileDisplayName,
		"sourceUrl":       fileLink,
		"corpusName":      corpusName,
		"documentAdded":   true,
		"ragFile":         map[string]interface{}{"name": operation.Name}, // Include operation name for RAG file ID extraction
	}, nil
}

// getGoogleDriveFileName fetches the actual filename from Google Drive API using ADC
func getGoogleDriveFileName(fileID string) (string, error) {
	ctx := context.Background()

	// Create Google Drive service client using Application Default Credentials
	driveService, err := drive.NewService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create Drive service: %v", err)
	}

	// Get file metadata
	file, err := driveService.Files.Get(fileID).Fields("name").Do()
	if err != nil {
		return "", fmt.Errorf("failed to get file metadata: %v", err)
	}

	return file.Name, nil
}

// processRAGAgentResponse processes the agent response from Python microservice
// If it's structured JSON with questions, saves them to MongoDB and returns question IDs
// If it's regular text, returns it as-is
func processRAGAgentResponse(agentResponse, teacherId string) (interface{}, string, error) {
	log.Printf("=== PROCESS RAG AGENT RESPONSE START ===")
	log.Printf("[AI] Processing RAG agent response for teacherId: %s", teacherId)
	log.Printf("[AI] Agent response length: %d characters", len(agentResponse))
	log.Printf("[AI] Agent response preview (first 500 chars): %.500s%s", agentResponse, func() string {
		if len(agentResponse) > 500 {
			return "..."
		}
		return ""
	}())

	// Try to parse as JSON to see if it contains structured questions
	log.Printf("[AI] Attempting to parse agent response as JSON...")
	var questionsData map[string]interface{}
	if err := json.Unmarshal([]byte(agentResponse), &questionsData); err != nil {
		// Not valid JSON, return as regular text response in message field
		log.Printf("✗ Agent response is not valid JSON: %v", err)
		log.Printf("[AI] Treating response as plain text message")
		log.Printf("[AI] Returning text response in message field")
		log.Printf("=== PROCESS RAG AGENT RESPONSE END (TEXT) ===")
		return nil, agentResponse, nil
	}
	log.Printf("✓ Successfully parsed agent response as JSON")

	log.Printf("=== JSON STRUCTURE ANALYSIS ===")
	log.Printf("[AI] JSON keys found: %v", func() []string {
		keys := make([]string, 0, len(questionsData))
		for k := range questionsData {
			keys = append(keys, k)
		}
		return keys
	}())

	// Check if this looks like question data (has mcq/mcqs, msq/msqs, nat/nats, or subjective/subjectives fields)
	questionFields := []string{"mcq", "mcqs", "msq", "msqs", "nat", "nats", "subjective", "subjectives"}
	hasQuestions := false
	foundQuestionFields := []string{}

	for _, key := range questionFields {
		if _, exists := questionsData[key]; exists {
			hasQuestions = true
			foundQuestionFields = append(foundQuestionFields, key)
			log.Printf("[AI] Found question field: %s", key)
		}
	}

	if !hasQuestions {
		// JSON but not question data, return the parsed JSON in data field
		log.Printf("[AI] JSON response doesn't contain question fields")
		log.Printf("[AI] Expected fields: %v", questionFields)
		log.Printf("[AI] Returning parsed JSON as structured data")
		log.Printf("=== PROCESS RAG AGENT RESPONSE END (STRUCTURED DATA) ===")
		return questionsData, "", nil
	}

	// This is structured question data, save to MongoDB and return IDs
	log.Printf("✓ Detected structured question data with fields: %v", foundQuestionFields)
	log.Printf("[AI] Processing structured question data for database storage...")

	// Create result structure matching the desired format
	result := map[string]interface{}{
		"corpusUsed":              teacherId, // Use teacherId as corpusUsed
		"mcqs":                    make([]interface{}, 0),
		"msqs":                    make([]interface{}, 0),
		"nats":                    make([]interface{}, 0),
		"subjectives":             make([]interface{}, 0),
		"totalQuestionsGenerated": 0,
	}

	// Process MCQ questions (handle both "mcq" and "mcqs")
	log.Printf("=== PROCESSING MCQ QUESTIONS ===")
	var mcqData interface{}
	var exists bool
	if mcqData, exists = questionsData["mcq"]; !exists {
		mcqData, exists = questionsData["mcqs"]
	}
	if exists {
		log.Printf("[AI] Found MCQ data, processing...")
		if mcqArray, ok := mcqData.([]interface{}); ok {
			log.Printf("[AI] MCQ array contains %d items", len(mcqArray))
			result["mcqs"] = mcqArray // Include the actual MCQ data
			log.Printf("[AI] MCQ data included in response")
		} else {
			log.Printf("WARNING: MCQ data is not an array: %T", mcqData)
		}
	} else {
		log.Printf("[AI] No MCQ data found")
	}

	// Process MSQ questions (handle both "msq" and "msqs")
	log.Printf("=== PROCESSING MSQ QUESTIONS ===")
	var msqData interface{}
	if msqData, exists = questionsData["msq"]; !exists {
		msqData, exists = questionsData["msqs"]
	}
	if exists {
		log.Printf("[AI] Found MSQ data, processing...")
		if msqArray, ok := msqData.([]interface{}); ok {
			log.Printf("[AI] MSQ array contains %d items", len(msqArray))
			result["msqs"] = msqArray // Include the actual MSQ data
			log.Printf("[AI] MSQ data included in response")
		} else {
			log.Printf("WARNING: MSQ data is not an array: %T", msqData)
		}
	} else {
		log.Printf("[AI] No MSQ data found")
	}

	// Process NAT questions (handle both "nat" and "nats")
	log.Printf("=== PROCESSING NAT QUESTIONS ===")
	var natData interface{}
	if natData, exists = questionsData["nat"]; !exists {
		natData, exists = questionsData["nats"]
	}
	if exists {
		log.Printf("[AI] Found NAT data, processing...")
		if natArray, ok := natData.([]interface{}); ok {
			log.Printf("[AI] NAT array contains %d items", len(natArray))
			result["nats"] = natArray // Include the actual NAT data
			log.Printf("[AI] NAT data included in response")
		} else {
			log.Printf("WARNING: NAT data is not an array: %T", natData)
		}
	} else {
		log.Printf("[AI] No NAT data found")
	}

	// Process Subjective questions (handle both "subjective" and "subjectives")
	log.Printf("=== PROCESSING SUBJECTIVE QUESTIONS ===")
	var subjectiveData interface{}
	if subjectiveData, exists = questionsData["subjective"]; !exists {
		subjectiveData, exists = questionsData["subjectives"]
	}
	if exists {
		log.Printf("[AI] Found Subjective data, processing...")
		if subjectiveArray, ok := subjectiveData.([]interface{}); ok {
			log.Printf("[AI] Subjective array contains %d items", len(subjectiveArray))
			result["subjectives"] = subjectiveArray // Include the actual Subjective data
			log.Printf("[AI] Subjective data included in response")
		} else {
			log.Printf("WARNING: Subjective data is not an array: %T", subjectiveData)
		}
	} else {
		log.Printf("[AI] No Subjective data found")
	}

	log.Printf("=== QUESTION PROCESSING SUMMARY ===")
	mcqCount := 0
	msqCount := 0
	natCount := 0
	subjectiveCount := 0

	if mcqs, ok := result["mcqs"].([]interface{}); ok {
		mcqCount = len(mcqs)
	}
	if msqs, ok := result["msqs"].([]interface{}); ok {
		msqCount = len(msqs)
	}
	if nats, ok := result["nats"].([]interface{}); ok {
		natCount = len(nats)
	}
	if subjectives, ok := result["subjectives"].([]interface{}); ok {
		subjectiveCount = len(subjectives)
	}

	totalQuestions := mcqCount + msqCount + natCount + subjectiveCount
	result["totalQuestionsGenerated"] = totalQuestions

	log.Printf("[AI] Question processing complete:")
	log.Printf("  - MCQs: %d", mcqCount)
	log.Printf("  - MSQs: %d", msqCount)
	log.Printf("  - NATs: %d", natCount)
	log.Printf("  - Subjectives: %d", subjectiveCount)
	log.Printf("  - Total: %d questions", totalQuestions)
	log.Printf("=== PROCESS RAG AGENT RESPONSE END (QUESTIONS) ===")

	return result, "Agent response processed successfully", nil
}

// Helper function to safely extract string values from map[string]interface{}
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ListCorpusDocumentsHandler godoc
// @Summary      List Documents in RAG Corpus
// @Description  List all documents in a specific RAG corpus
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        corpusName path string true "Corpus name"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/rag-agent/corpus/{corpusName}/documents [get]
func ListCorpusDocumentsHandler(c *gin.Context) {
	corpusName := c.Param("corpusName")
	if corpusName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpus name is required"})
		return
	}

	log.Printf("[AI] Listing documents for corpus: %s", corpusName)

	docRepo := repository.NewDocumentRepository()
	ctx := context.Background()

	// Get documents from database
	documents, err := docRepo.GetDocumentsByCorpus(ctx, corpusName)
	if err != nil {
		log.Printf("[AI] Failed to retrieve documents from database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve documents"})
		return
	}

	log.Printf("[AI] Found %d documents in database", len(documents))

	// Get documents from RAG engine to verify consistency
	ragContentResponse, err := listVertexAICorpusContent(corpusName)
	if err != nil {
		log.Printf("[AI] Failed to retrieve documents from RAG engine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve documents from RAG engine"})
		return
	}

	// Extract RAG files and create a map for quick lookup
	ragFiles := make(map[string]map[string]interface{})
	ragFilesByID := make(map[string]map[string]interface{})

	log.Printf("[AI] RAG response status: %v", ragContentResponse["status"])

	if ragContentResponse["status"] != "error" {
		if files, ok := ragContentResponse["files"].([]map[string]interface{}); ok {
			log.Printf("[AI] Found %d documents in RAG engine", len(files))
			for i, file := range files {
				displayName := getStringValue(file, "displayName")
				fileID := getStringValue(file, "id")
				log.Printf("[AI] RAG file %d: displayName='%s', id='%s'", i+1, displayName, fileID)

				if displayName != "" {
					ragFiles[displayName] = file
				}
				if fileID != "" {
					ragFilesByID[fileID] = file
				}
			}
		}
	} else {
		log.Printf("[AI] RAG engine returned error: %v", ragContentResponse["message"])
	}

	// Filter database documents to only include those present in RAG engine
	var documentList []model.DocumentMetadata
	var inconsistentCount int

	for i, doc := range documents {
		log.Printf("[AI] DB document %d: displayName='%s', ragFileID='%s', fileID='%s'",
			i+1, doc.DisplayName, doc.RAGFileID, doc.FileID)

		// Try multiple matching strategies
		var exists bool
		var matchMethod string
		var ragDisplayName string

		// Strategy 1: Match by display name
		if _, exists = ragFiles[doc.DisplayName]; exists {
			matchMethod = "displayName"
			ragDisplayName = doc.DisplayName
		} else if doc.RAGFileID != "" {
			// Strategy 2: Match by RAG file ID
			if ragFileData, ragExists := ragFilesByID[doc.RAGFileID]; ragExists {
				exists = true
				matchMethod = "ragFileID"
				ragDisplayName = getStringValue(ragFileData, "displayName")
			} else {
				// Strategy 3: Check if RAG file ID appears in any RAG display name
				// (for cases where the file was renamed with UUID)
				for ragName := range ragFiles {
					if strings.Contains(ragName, doc.RAGFileID) {
						exists = true
						matchMethod = "ragFileID_in_displayName"
						ragDisplayName = ragName
						break
					}
				}
			}
		}

		if exists {
			log.Printf("[AI] Document matched using %s: DB='%s' -> RAG='%s'", matchMethod, doc.DisplayName, ragDisplayName)
			// Document exists in both database and RAG engine
			documentList = append(documentList, model.DocumentMetadata{
				FileID:      doc.FileID,
				DisplayName: doc.DisplayName,
				ContentType: doc.ContentType,
				Size:        doc.Size,
				CorpusName:  doc.CorpusName,
				CreatedAt:   doc.CreatedAt,
			})
		} else {
			log.Printf("[AI] Document NOT found in RAG engine: %s (ragFileID: %s)", doc.DisplayName, doc.RAGFileID)
			// Document exists in database but not in RAG engine - inconsistency
			inconsistentCount++
		}
	}

	response := gin.H{
		"corpusName": corpusName,
		"documents":  documentList,
		"count":      len(documentList),
	}

	// Add inconsistency warning if any documents are missing from RAG engine
	if inconsistentCount > 0 {
		response["warning"] = fmt.Sprintf("%d documents exist in database but not in RAG engine", inconsistentCount)
		response["inconsistentCount"] = inconsistentCount
	}

	log.Printf("[AI] Returning %d documents, %d inconsistent", len(documentList), inconsistentCount)
	c.JSON(http.StatusOK, response)
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
		"message":     "Document deleted successfully",
		"documentId":  documentID,
		"displayName": document.DisplayName,
	})
}
