package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	service "lumenslate/internal/grpc_service"
	"lumenslate/internal/repository"
	"lumenslate/internal/utils"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// RAGAgentHandler godoc
// @Summary      Process Text with RAG Agent
// @Description  Process text input using Retrieval-Augmented Generation (RAG) agent for intelligent knowledge retrieval and response generation. Creates/verifies teacher-specific corpus automatically.
// @Tags         AI RAG Agent
// @Accept       json
// @Produce      json
// @Param        body  body  ai.RAGAgentRequest  true  "RAG agent request with teacher ID, role, and message"
// @Success      200   {object}  map[string]interface{}  "RAG agent response with message, data, and metadata"
// @Failure      400   {object}  gin.H  "Invalid request body or missing required fields"
// @Failure      500   {object}  gin.H  "Internal server error during RAG processing"
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
	_, err := createVertexAICorpus(req.CorpusName)
	if err != nil {
		log.Printf("WARNING: Could not create/verify corpus for teacher %s: %v", req.CorpusName, err)
		// Continue processing even if corpus creation fails
	}

	// Call the gRPC microservice
	resp, err := service.RAGAgentClient(req.CorpusName, req.Message)
	if err != nil {
		log.Printf("ERROR: Failed to process RAG agent request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process RAG agent request", "error": err.Error()})
		return
	}

	// Process the agent response to determine data content and message
	responseData, responseMessage, err := processRAGAgentResponse(resp.GetAgentResponse(), req.CorpusName)
	if err != nil {
		log.Printf("ERROR: Failed to process RAG agent response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process RAG agent response", "error": err.Error()})
		return
	}

	// Build standardized response format (matching /ai/agent structure)
	standardizedResponse := map[string]interface{}{
		"message":      responseMessage,
		"corpusName":   resp.GetCorpusName(),
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
// @Description  Create a new RAG corpus in Vertex AI for document storage and retrieval. If the corpus already exists, returns the existing corpus information.
// @Tags         AI RAG Management
// @Accept       json
// @Produce      json
// @Param        body  body  ai.CreateCorpusRequest  true  "Corpus creation request containing the corpus name"
// @Success      200   {object}  map[string]interface{}  "Corpus created or retrieved successfully with corpus details"
// @Failure      400   {object}  map[string]interface{}  "Invalid request body or missing corpus name"
// @Failure      500   {object}  map[string]interface{}  "Internal server error during corpus creation"
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
// @Tags         AI RAG Management
// @Accept       json
// @Produce      json
// @Param        body  body  ai.CreateCorpusRequest  true  "Request body with corpus name to list content for"
// @Success      200   {object}  map[string]interface{}  "List of documents in the corpus with metadata"
// @Failure      400   {object}  map[string]interface{}  "Invalid request body or missing corpus name"
// @Failure      500   {object}  map[string]interface{}  "Internal server error during content retrieval"
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

// ListCorpusDocumentsHandler godoc
// @Summary      List Documents in RAG Corpus
// @Description  List all documents in a specific RAG corpus with cross-verification between database and RAG engine. Returns unified document information including storage status.
// @Tags         AI RAG Management
// @Accept       json
// @Produce      json
// @Param        corpusName  path    string  true  "Name of the corpus to list documents for"
// @Success      200         {object}  map[string]interface{}  "List of documents with unified information from database and RAG engine"
// @Failure      400         {object}  map[string]interface{}  "Invalid or missing corpus name"
// @Failure      500         {object}  map[string]interface{}  "Internal server error during document retrieval"
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
		log.Printf("[AI] Warning: RAG engine returned error status")
	}

	// Build unified response with cross-checking
	unifiedDocuments := []map[string]interface{}{}

	for _, doc := range documents {
		docInfo := map[string]interface{}{
			"fileId":        doc.FileID,
			"displayName":   doc.DisplayName,
			"corpusName":    doc.CorpusName,
			"ragFileId":     doc.RAGFileID,
			"gcsObject":     doc.GCSObject,
			"createdAt":     doc.CreatedAt,
			"inDatabase":    true,
			"inRAGEngine":   false,
			"ragEngineInfo": nil,
		}

		// Check if this document exists in RAG engine
		var ragFileInfo map[string]interface{}
		found := false

		// Try to find by displayName first
		if ragFileInfo, found = ragFiles[doc.DisplayName]; found {
			log.Printf("[AI] Found matching RAG file by displayName: %s", doc.DisplayName)
		} else if doc.RAGFileID != "" {
			// Try to find by RAG file ID
			if ragFileInfo, found = ragFilesByID[doc.RAGFileID]; found {
				log.Printf("[AI] Found matching RAG file by RAG file ID: %s", doc.RAGFileID)
			} else {
				// Try to find by RAG file ID in displayName
				for _, ragFile := range ragFiles {
					if ragDisplayName := getStringValue(ragFile, "displayName"); ragDisplayName != "" {
						if ragDisplayName == doc.RAGFileID || ragDisplayName == doc.RAGFileID+".pdf" {
							ragFileInfo = ragFile
							found = true
							log.Printf("[AI] Found matching RAG file by RAG file ID in displayName: %s", ragDisplayName)
							break
						}
					}
				}
			}
		}

		if found {
			docInfo["inRAGEngine"] = true
			docInfo["ragEngineInfo"] = ragFileInfo
		}

		unifiedDocuments = append(unifiedDocuments, docInfo)
	}

	response := map[string]interface{}{
		"corpusName":     corpusName,
		"documents":      unifiedDocuments,
		"totalDocuments": len(unifiedDocuments),
		"databaseCount":  len(documents),
		"ragEngineCount": len(ragFiles),
		"status":         "success",
	}

	c.JSON(http.StatusOK, response)
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

// createVertexAICorpus creates a RAG corpus directly in Vertex AI
func createVertexAICorpus(corpusName string) (map[string]interface{}, error) {
	log.Printf("[AI] createVertexAICorpus called with corpusName: %s", corpusName)
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
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
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
		log.Printf("[AI] Failed to list existing corpora: %v", err)
		return nil, fmt.Errorf("failed to list existing corpora: %v", err)
	}

	log.Printf("[AI] Found %d existing corpora", len(existingCorpora.RagCorpora))
	for _, corpus := range existingCorpora.RagCorpora {
		log.Printf("[AI] Existing corpus: %s (displayName: %s)", corpus.Name, corpus.DisplayName)
		if corpus.DisplayName == displayName {
			log.Printf("[AI] Corpus already exists: %s", corpus.Name)
			return map[string]interface{}{
				"status":    "exists",
				"message":   fmt.Sprintf("Corpus '%s' already exists", corpusName),
				"corpus":    corpus,
				"projectID": projectID,
				"location":  location,
			}, nil
		}
	}

	// Create new corpus
	log.Printf("[AI] Creating new corpus with displayName: %s", displayName)
	ragCorpus := &aiplatform.GoogleCloudAiplatformV1RagCorpus{
		DisplayName: displayName,
	}

	createCall := service.Projects.Locations.RagCorpora.Create(parent, ragCorpus)
	operation, err := createCall.Do()
	if err != nil {
		log.Printf("[AI] Failed to create corpus: %v", err)
		return nil, fmt.Errorf("failed to create corpus: %v", err)
	}

	log.Printf("[AI] Corpus creation initiated. Operation: %s", operation.Name)
	return map[string]interface{}{
		"status":    "created",
		"message":   fmt.Sprintf("Corpus '%s' created successfully", corpusName),
		"operation": operation,
		"projectID": projectID,
		"location":  location,
	}, nil
}

// listVertexAICorpusContent lists the content of a RAG corpus
func listVertexAICorpusContent(corpusName string) (map[string]interface{}, error) {
	log.Printf("[AI] listVertexAICorpusContent called with corpusName: %s", corpusName)
	ctx := context.Background()

	// Project configuration
	projectID := utils.GetProjectID()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	displayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)

	// Find the corpus
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
		return map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("Corpus '%s' not found", corpusName),
		}, nil
	}

	// List files in the corpus
	listFilesCall := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName)
	files, err := listFilesCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in corpus: %v", err)
	}

	fileList := make([]map[string]interface{}, 0, len(files.RagFiles))
	for _, file := range files.RagFiles {
		fileList = append(fileList, map[string]interface{}{
			"id":          file.Name,
			"displayName": file.DisplayName,
			"createTime":  file.CreateTime,
			"updateTime":  file.UpdateTime,
		})
	}

	return map[string]interface{}{
		"status":     "success",
		"corpusName": corpusName,
		"files":      fileList,
		"totalFiles": len(fileList),
	}, nil
}

// processRAGAgentResponse processes the agent response from Python microservice
// If it's structured JSON with questions, saves them to MongoDB and returns question IDs
// If it's regular text, returns it as-is
func processRAGAgentResponse(agentResponse, CorpusName string) (interface{}, string, error) {
	log.Printf("=== PROCESS RAG AGENT RESPONSE START ===")
	log.Printf("[AI] Processing RAG agent response for corpusName: %s", CorpusName)
	log.Printf("[AI] Agent response: %s", agentResponse)

	// Try to parse as JSON to see if it contains structured questions
	log.Printf("[AI] Attempting to parse agent response as JSON...")
	var questionsData map[string]interface{}
	if err := json.Unmarshal([]byte(agentResponse), &questionsData); err != nil {
		// Not valid JSON, return as regular text response in message field
		log.Printf("✗ Agent response is not valid JSON: %v", err)
		log.Printf("[AI] Treating response as plain text message")
		return nil, agentResponse, nil
	}
	log.Printf("✓ Successfully parsed agent response as JSON")

	// Check if this looks like question data
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
		return questionsData, "", nil
	}

	// This is structured question data, process and return
	log.Printf("✓ Detected structured question data with fields: %v", foundQuestionFields)

	// Create result structure
	result := map[string]interface{}{
		"corpusUsed":              CorpusName,
		"mcqs":                    make([]interface{}, 0),
		"msqs":                    make([]interface{}, 0),
		"nats":                    make([]interface{}, 0),
		"subjectives":             make([]interface{}, 0),
		"totalQuestionsGenerated": 0,
	}

	// Process each question type
	if mcqData, exists := questionsData["mcq"]; !exists {
		if mcqData, exists = questionsData["mcqs"]; exists {
			if mcqArray, ok := mcqData.([]interface{}); ok {
				result["mcqs"] = mcqArray
			}
		}
	} else {
		if mcqArray, ok := mcqData.([]interface{}); ok {
			result["mcqs"] = mcqArray
		}
	}

	if msqData, exists := questionsData["msq"]; !exists {
		if msqData, exists = questionsData["msqs"]; exists {
			if msqArray, ok := msqData.([]interface{}); ok {
				result["msqs"] = msqArray
			}
		}
	} else {
		if msqArray, ok := msqData.([]interface{}); ok {
			result["msqs"] = msqArray
		}
	}

	if natData, exists := questionsData["nat"]; !exists {
		if natData, exists = questionsData["nats"]; exists {
			if natArray, ok := natData.([]interface{}); ok {
				result["nats"] = natArray
			}
		}
	} else {
		if natArray, ok := natData.([]interface{}); ok {
			result["nats"] = natArray
		}
	}

	if subjectiveData, exists := questionsData["subjective"]; !exists {
		if subjectiveData, exists = questionsData["subjectives"]; exists {
			if subjectiveArray, ok := subjectiveData.([]interface{}); ok {
				result["subjectives"] = subjectiveArray
			}
		}
	} else {
		if subjectiveArray, ok := subjectiveData.([]interface{}); ok {
			result["subjectives"] = subjectiveArray
		}
	}

	// Calculate total questions
	totalQuestions := 0
	if mcqs, ok := result["mcqs"].([]interface{}); ok {
		totalQuestions += len(mcqs)
	}
	if msqs, ok := result["msqs"].([]interface{}); ok {
		totalQuestions += len(msqs)
	}
	if nats, ok := result["nats"].([]interface{}); ok {
		totalQuestions += len(nats)
	}
	if subjectives, ok := result["subjectives"].([]interface{}); ok {
		totalQuestions += len(subjectives)
	}

	result["totalQuestionsGenerated"] = totalQuestions
	log.Printf("[AI] Total questions processed: %d", totalQuestions)

	return result, "Agent response processed successfully", nil
}
