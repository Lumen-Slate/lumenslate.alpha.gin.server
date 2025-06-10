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
	"regexp"
	"time"

	service "lumenslate/internal/grpc_service"
	"lumenslate/internal/utils"

	"github.com/gin-gonic/gin"
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
	CorpusName      string `json:"corpusName" binding:"required"`
	FileDisplayName string `json:"fileDisplayName" binding:"required"`
}

type AddCorpusDocumentRequest struct {
	CorpusName string `json:"corpusName" binding:"required"`
	FileLink   string `json:"fileLink" binding:"required"`
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
	log.Println("[AI] /ai/generate-context called")
	var req GenerateContextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)
	content, err := service.GenerateContext(req.Question, req.Keywords, req.Language)
	if err != nil {
		log.Printf("[AI] GenerateContext error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] GenerateContext success")
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
	log.Println("[AI] /ai/detect-variables called")
	var req DetectVariablesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)
	variables, err := service.DetectVariables(req.Question)
	if err != nil {
		log.Printf("[AI] DetectVariables error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] DetectVariables success, count: %d", len(variables))
	c.JSON(http.StatusOK, gin.H{"variables": variables})
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
	log.Println("[AI] /ai/segment-question called")
	var req SegmentQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)
	segmented, err := service.SegmentQuestion(req.Question)
	if err != nil {
		log.Printf("[AI] SegmentQuestion error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] SegmentQuestion success")
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
	log.Println("[AI] /ai/rag-agent called")
	var req RAGAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AI] RAG Agent Request: %+v", req)

	// Create/verify corpus for the teacher before processing the request
	log.Printf("[AI] Creating/verifying corpus for teacher: %s", req.TeacherId)
	corpusResponse, err := createVertexAICorpus(req.TeacherId)
	if err != nil {
		log.Printf("[AI] Warning: Could not create/verify corpus for teacher %s: %v", req.TeacherId, err)
		// Continue processing even if corpus creation fails
	} else {
		log.Printf("[AI] Corpus operation result: %s", corpusResponse["message"])
	}

	// Call the gRPC microservice
	resp, err := service.RAGAgentClient(req.TeacherId, req.Message, "")
	if err != nil {
		log.Printf("[AI] Failed to process RAG agent request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process RAG agent request", "error": err.Error()})
		return
	}

	// Process the agent response to determine data content and message
	responseData, responseMessage, err := processRAGAgentResponse(resp.GetAgentResponse(), req.TeacherId)
	if err != nil {
		log.Printf("[AI] Failed to process RAG agent response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process RAG agent response", "error": err.Error()})
		return
	}

	// Return standardized response format (matching /ai/agent structure)
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

	log.Printf("[AI] RAG agent request processed successfully")
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
		files = append(files, map[string]interface{}{
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
	log.Println("[AI] /ai/rag-agent/delete-corpus-document called")
	var req DeleteCorpusDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)

	deleteResponse, err := deleteVertexAICorpusDocument(req.CorpusName, req.FileDisplayName)
	if err != nil {
		log.Printf("[AI] Delete corpus document error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete document: %v", err)})
		return
	}

	log.Printf("[AI] Delete corpus document success")
	c.JSON(http.StatusOK, deleteResponse)
}

// AddCorpusDocumentHandler godoc
// @Summary      Add Document to RAG Corpus
// @Description  Add a document from Google Drive to a RAG corpus
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.AddCorpusDocumentRequest  true  "Request body"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /ai/rag-agent/add-corpus-document [post]
func AddCorpusDocumentHandler(c *gin.Context) {
	log.Println("[AI] /ai/rag-agent/add-corpus-document called")
	var req AddCorpusDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AI] Request: %+v", req)

	// Check if corpus exists, create if it doesn't
	log.Printf("[AI] Checking if corpus '%s' exists before adding document", req.CorpusName)
	corpusResponse, err := createVertexAICorpus(req.CorpusName)
	if err != nil {
		log.Printf("[AI] Warning: Could not create/verify corpus '%s': %v", req.CorpusName, err)
		// Continue with document addition even if corpus creation fails
	} else {
		log.Printf("[AI] Corpus operation result: %s", corpusResponse["message"])
	}

	addResponse, err := addVertexAICorpusDocument(req.CorpusName, req.FileLink)
	if err != nil {
		log.Printf("[AI] Add corpus document error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to add document: %v", err)})
		return
	}

	log.Printf("[AI] Add corpus document success")
	c.JSON(http.StatusOK, addResponse)
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

// listAllVertexAICorpora lists all RAG corpora names
func listAllVertexAICorpora() (map[string]interface{}, error) {
	log.Printf("[AI] listAllVertexAICorpora called")
	ctx := context.Background()

	// Project configuration - force RAG-compatible location
	projectID := utils.GetProjectID()

	log.Printf("🧪 GOOGLE_PROJECT_ID: %s", projectID)

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
func deleteVertexAICorpusDocument(corpusName, fileDisplayName string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Project configuration - force RAG-compatible location
	projectID := utils.GetProjectID()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // Default fallback
	}

	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)

	// Create AI Platform service client with default ADC credentials
	service, err := aiplatform.NewService(ctx,
		option.WithEndpoint(endpoint)) // ⬅️ Removed WithCredentialsFile
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

	// List files in the corpus to find the one to delete
	filesResponse, err := service.Projects.Locations.RagCorpora.RagFiles.
		List(corpusResourceName).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in corpus: %v", err)
	}

	var fileToDelete string
	for _, file := range filesResponse.RagFiles {
		if file.DisplayName == fileDisplayName {
			fileToDelete = file.Name
			break
		}
	}

	if fileToDelete == "" {
		return nil, fmt.Errorf("file with display name '%s' not found in corpus '%s'", fileDisplayName, corpusName)
	}

	// Delete the file
	operation, err := service.Projects.Locations.RagCorpora.RagFiles.
		Delete(fileToDelete).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to delete file: %v", err)
	}

	return map[string]interface{}{
		"status":          "success",
		"message":         fmt.Sprintf("Successfully deleted file '%s' from corpus '%s'", fileDisplayName, corpusName),
		"operationName":   operation.Name,
		"deletedFileName": fileDisplayName,
		"corpusName":      corpusName,
		"documentDeleted": true,
	}, nil
}

// addVertexAICorpusDocument adds a document from Google Drive to a RAG corpus
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

	// Extract file ID from Google Drive URL
	re := regexp.MustCompile(`/file/d/([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(fileLink)
	if len(matches) <= 1 {
		return nil, fmt.Errorf("invalid Google Drive URL format: could not extract file ID")
	}
	fileID := matches[1]

	// Fetch file display name using Drive API (using ADC, no credentialsPath)
	fileDisplayName, err := getGoogleDriveFileName(fileID) // Adjusted to match function signature
	if err != nil {
		log.Printf("[AI] Warning: Could not fetch file name from Google Drive: %v. Using fallback name.", err)
		fileDisplayName = fmt.Sprintf("gdrive_file_%s", fileID)
	}

	log.Printf("[AI] Adding file '%s' (ID: %s) to corpus '%s'", fileDisplayName, fileID, corpusName)

	// Build the import request
	importRequest := &aiplatform.GoogleCloudAiplatformV1ImportRagFilesRequest{
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

	// Trigger the import
	operation, err := service.Projects.Locations.RagCorpora.RagFiles.Import(corpusResourceName, importRequest).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to add file to corpus: %v", err)
	}

	return map[string]interface{}{
		"status":          "success",
		"message":         fmt.Sprintf("Successfully added file from Google Drive to corpus '%s'", corpusName),
		"operationName":   operation.Name,
		"fileDisplayName": fileDisplayName,
		"sourceUrl":       fileLink,
		"corpusName":      corpusName,
		"documentAdded":   true,
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
	log.Printf("[AI] Processing RAG agent response for teacherId: %s", teacherId)

	// Try to parse as JSON to see if it contains structured questions
	var questionsData map[string]interface{}
	if err := json.Unmarshal([]byte(agentResponse), &questionsData); err != nil {
		// Not valid JSON, return as regular text response in message field
		log.Printf("[AI] Agent response is not JSON, returning as text message")
		return nil, agentResponse, nil
	}

	// Check if this looks like question data (has mcq/mcqs, msq/msqs, nat/nats, or subjective/subjectives fields)
	hasQuestions := false
	for _, key := range []string{"mcq", "mcqs", "msq", "msqs", "nat", "nats", "subjective", "subjectives"} {
		if _, exists := questionsData[key]; exists {
			hasQuestions = true
			break
		}
	}

	if !hasQuestions {
		// JSON but not question data, return the parsed JSON in data field
		log.Printf("[AI] JSON response doesn't contain questions, returning as data")
		return questionsData, "", nil
	}

	// This is structured question data, save to MongoDB and return IDs
	log.Printf("[AI] Processing structured question data")

	result := map[string]interface{}{
		"mcqs":        []string{},
		"msqs":        []string{},
		"nats":        []string{},
		"subjectives": []string{},
		"questions":   questionsData, // Include the original question data
	}

	// Process MCQ questions (handle both "mcq" and "mcqs")
	var mcqData interface{}
	var exists bool
	if mcqData, exists = questionsData["mcq"]; !exists {
		mcqData, exists = questionsData["mcqs"]
	}
	if exists {
		if mcqArray, ok := mcqData.([]interface{}); ok {
			for _, mcqInterface := range mcqArray {
				if mcqMap, ok := mcqInterface.(map[string]interface{}); ok {
					_ = mcqMap // Suppress unused variable warning
					// id, err := saveMCQQuestion(mcqMap, teacherId)
					// if err != nil {
					// 	log.Printf("[AI] Error saving MCQ question: %v", err)
					// 	continue
					// }
					// result["mcqs"] = append(result["mcqs"].([]string), id)
				}
			}
		}
	}

	// Process MSQ questions (handle both "msq" and "msqs")
	var msqData interface{}
	if msqData, exists = questionsData["msq"]; !exists {
		msqData, exists = questionsData["msqs"]
	}
	if exists {
		if msqArray, ok := msqData.([]interface{}); ok {
			for _, msqInterface := range msqArray {
				if msqMap, ok := msqInterface.(map[string]interface{}); ok {
					_ = msqMap // Suppress unused variable warning
					// id, err := saveMSQQuestion(msqMap, teacherId)
					// if err != nil {
					// 	log.Printf("[AI] Error saving MSQ question: %v", err)
					// 	continue
					// }
					// result["msqs"] = append(result["msqs"].([]string), id)
				}
			}
		}
	}

	// Process NAT questions (handle both "nat" and "nats")
	var natData interface{}
	if natData, exists = questionsData["nat"]; !exists {
		natData, exists = questionsData["nats"]
	}
	if exists {
		if natArray, ok := natData.([]interface{}); ok {
			for _, natInterface := range natArray {
				if natMap, ok := natInterface.(map[string]interface{}); ok {
					_ = natMap // Suppress unused variable warning
					// id, err := saveNATQuestion(natMap, teacherId)
					// if err != nil {
					// 	log.Printf("[AI] Error saving NAT question: %v", err)
					// 	continue
					// }
					// result["nats"] = append(result["nats"].([]string), id)
				}
			}
		}
	}

	// Process Subjective questions (handle both "subjective" and "subjectives")
	var subjectiveData interface{}
	if subjectiveData, exists = questionsData["subjective"]; !exists {
		subjectiveData, exists = questionsData["subjectives"]
	}
	if exists {
		if subjectiveArray, ok := subjectiveData.([]interface{}); ok {
			for _, subjectiveInterface := range subjectiveArray {
				if subjectiveMap, ok := subjectiveInterface.(map[string]interface{}); ok {
					_ = subjectiveMap // Suppress unused variable warning
					// id, err := saveSubjectiveQuestion(subjectiveMap, teacherId)
					// if err != nil {
					// 	log.Printf("[AI] Error saving Subjective question: %v", err)
					// 	continue
					// }
					// result["subjectives"] = append(result["subjectives"].([]string), id)
				}
			}
		}
	}

	log.Printf("[AI] Question processing complete. Saved %d MCQs, %d MSQs, %d NATs, %d Subjectives",
		len(result["mcqs"].([]string)), len(result["msqs"].([]string)), len(result["nats"].([]string)), len(result["subjectives"].([]string)))

	return result, "", nil
}

// // saveMCQQuestion saves an MCQ question to MongoDB and returns its ID
// func saveMCQQuestion(questionData map[string]interface{}, teacherId string) (string, error) {
// 	mcq := questionModels.NewMCQ()
// 	mcq.ID = primitive.NewObjectID().Hex()
// 	mcq.BankID = teacherId // Use teacherId as bankId for now

// 	// Extract question data
// 	if question, ok := questionData["question"].(string); ok {
// 		mcq.Question = question
// 	}

// 	if points, ok := questionData["points"].(float64); ok {
// 		mcq.Points = int(points)
// 	}

// 	if difficulty, ok := questionData["difficulty"].(string); ok {
// 		mcq.Difficulty = difficulty
// 	}

// 	if subject, ok := questionData["subject"].(string); ok {
// 		mcq.Subject = subject
// 	}

// 	if answerIndex, ok := questionData["answerIndex"].(float64); ok {
// 		mcq.AnswerIndex = int(answerIndex)
// 	}

// 	if optionsInterface, ok := questionData["options"].([]interface{}); ok {
// 		options := make([]string, len(optionsInterface))
// 		for i, opt := range optionsInterface {
// 			if optStr, ok := opt.(string); ok {
// 				options[i] = optStr
// 			}
// 		}
// 		mcq.Options = options
// 	}

// 	if err := questionRepo.SaveMCQ(*mcq); err != nil {
// 		return "", fmt.Errorf("failed to save MCQ: %v", err)
// 	}

// 	return mcq.ID, nil
// }

// // saveMSQQuestion saves an MSQ question to MongoDB and returns its ID
// func saveMSQQuestion(questionData map[string]interface{}, teacherId string) (string, error) {
// 	msq := questionModels.NewMSQ()
// 	msq.ID = primitive.NewObjectID().Hex()
// 	msq.BankID = teacherId

// 	if question, ok := questionData["question"].(string); ok {
// 		msq.Question = question
// 	}

// 	if points, ok := questionData["points"].(float64); ok {
// 		msq.Points = int(points)
// 	}

// 	if difficulty, ok := questionData["difficulty"].(string); ok {
// 		msq.Difficulty = difficulty
// 	}

// 	if subject, ok := questionData["subject"].(string); ok {
// 		msq.Subject = subject
// 	}

// 	if answerIndicesInterface, ok := questionData["answerIndices"].([]interface{}); ok {
// 		answerIndices := make([]int, len(answerIndicesInterface))
// 		for i, idx := range answerIndicesInterface {
// 			if idxFloat, ok := idx.(float64); ok {
// 				answerIndices[i] = int(idxFloat)
// 			}
// 		}
// 		msq.AnswerIndices = answerIndices
// 	}

// 	if optionsInterface, ok := questionData["options"].([]interface{}); ok {
// 		options := make([]string, len(optionsInterface))
// 		for i, opt := range optionsInterface {
// 			if optStr, ok := opt.(string); ok {
// 				options[i] = optStr
// 			}
// 		}
// 		msq.Options = options
// 	}

// 	if err := questionRepo.SaveMSQ(*msq); err != nil {
// 		return "", fmt.Errorf("failed to save MSQ: %v", err)
// 	}

// 	return msq.ID, nil
// }

// // saveNATQuestion saves a NAT question to MongoDB and returns its ID
// func saveNATQuestion(questionData map[string]interface{}, teacherId string) (string, error) {
// 	nat := questionModels.NewNAT()
// 	nat.ID = primitive.NewObjectID().Hex()
// 	nat.BankID = teacherId

// 	if question, ok := questionData["question"].(string); ok {
// 		nat.Question = question
// 	}

// 	if points, ok := questionData["points"].(float64); ok {
// 		nat.Points = int(points)
// 	}

// 	if difficulty, ok := questionData["difficulty"].(string); ok {
// 		nat.Difficulty = difficulty
// 	}

// 	if subject, ok := questionData["subject"].(string); ok {
// 		nat.Subject = subject
// 	}

// 	if answer, ok := questionData["answer"].(float64); ok {
// 		nat.Answer = answer
// 	}

// 	if err := questionRepo.SaveNAT(*nat); err != nil {
// 		return "", fmt.Errorf("failed to save NAT: %v", err)
// 	}

// 	return nat.ID, nil
// }

// // saveSubjectiveQuestion saves a Subjective question to MongoDB and returns its ID
// func saveSubjectiveQuestion(questionData map[string]interface{}, teacherId string) (string, error) {
// 	subjective := questionModels.NewSubjective()
// 	subjective.ID = primitive.NewObjectID().Hex()
// 	subjective.BankID = teacherId

// 	if question, ok := questionData["question"].(string); ok {
// 		subjective.Question = question
// 	}

// 	if points, ok := questionData["points"].(float64); ok {
// 		subjective.Points = int(points)
// 	}

// 	if difficulty, ok := questionData["difficulty"].(string); ok {
// 		subjective.Difficulty = difficulty
// 	}

// 	if subject, ok := questionData["subject"].(string); ok {
// 		subjective.Subject = subject
// 	}

// 	if idealAnswer, ok := questionData["idealAnswer"].(string); ok {
// 		subjective.IdealAnswer = &idealAnswer
// 	}

// 	if gradingCriteriaInterface, ok := questionData["gradingCriteria"].([]interface{}); ok {
// 		criteria := make([]string, len(gradingCriteriaInterface))
// 		for i, criterion := range gradingCriteriaInterface {
// 			if criterionStr, ok := criterion.(string); ok {
// 				criteria[i] = criterionStr
// 			}
// 		}
// 		subjective.GradingCriteria = criteria
// 	}

// 	if err := questionRepo.SaveSubjective(*subjective); err != nil {
// 		return "", fmt.Errorf("failed to save Subjective: %v", err)
// 	}

// 	return subjective.ID, nil
// }
