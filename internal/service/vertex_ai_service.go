package service

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

type VertexAIService struct {
	projectID string
	location  string
}

// ListRAGFilesInCorpus lists all RAG files in a given corpus for debugging and fallback matching
func (v *VertexAIService) ListRAGFilesInCorpus(ctx context.Context, corpusName string) ([]map[string]string, error) {
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", v.location)
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	corpusDisplayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")
	parent := fmt.Sprintf("projects/%s/locations/%s", v.projectID, v.location)
	listCall := service.Projects.Locations.RagCorpora.List(parent)
	existingCorpora, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list corpora: %v", err)
	}

	var corpusResourceName string
	for _, corpus := range existingCorpora.RagCorpora {
		if corpus.DisplayName == corpusDisplayName {
			corpusResourceName = corpus.Name
			break
		}
	}
	if corpusResourceName == "" {
		return nil, fmt.Errorf("corpus '%s' not found", corpusName)
	}

	filesResponse, err := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in corpus: %v", err)
	}

	var result []map[string]string
	for _, file := range filesResponse.RagFiles {
		result = append(result, map[string]string{
			"displayName": file.DisplayName,
			"name":        file.Name,
		})
	}
	return result, nil
}

// NewVertexAIService creates a new Vertex AI service instance
func NewVertexAIService() *VertexAIService {
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // Default fallback
	}

	return &VertexAIService{
		projectID: projectID,
		location:  location,
	}
}

// AddDocumentToCorpus adds a document from GCS to a RAG corpus
func (v *VertexAIService) AddDocumentToCorpus(ctx context.Context, corpusName, fileLink string) (map[string]interface{}, error) {
	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", v.location)

	// Create AI Platform service client with regional endpoint (using ADC)
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Clean corpus name for use as display name
	displayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")

	// Find the corpus first
	parent := fmt.Sprintf("projects/%s/locations/%s", v.projectID, v.location)
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

// CheckOperationStatus checks the status of a Vertex AI operation
func (v *VertexAIService) CheckOperationStatus(ctx context.Context, operationName string) (map[string]interface{}, error) {
	// Use regional endpoint
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", v.location)

	// Create AI Platform service client
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Get operation status
	operation, err := service.Projects.Locations.Operations.Get(operationName).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get operation status: %v", err)
	}

	result := map[string]interface{}{
		"name": operation.Name,
		"done": operation.Done,
	}

	if operation.Error != nil {
		result["error"] = map[string]interface{}{
			"code":    operation.Error.Code,
			"message": operation.Error.Message,
		}
	}

	if operation.Response != nil {
		result["response"] = operation.Response
	}

	if operation.Metadata != nil {
		result["metadata"] = operation.Metadata
	}

	return result, nil
}

// ExtractRAGFileIDFromDocument attempts to find and extract RAG file ID for a document
// This is useful for cases where the original processing failed but the file was actually added to RAG
func (v *VertexAIService) ExtractRAGFileIDFromDocument(ctx context.Context, corpusName, displayName string) (string, error) {
	// Use regional endpoint for RAG operations
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", v.location)

	// Create AI Platform service client with regional endpoint (using ADC)
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return "", fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Clean corpus name for use as display name
	corpusDisplayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")

	// Find the corpus first
	parent := fmt.Sprintf("projects/%s/locations/%s", v.projectID, v.location)
	listCall := service.Projects.Locations.RagCorpora.List(parent)

	existingCorpora, err := listCall.Do()
	if err != nil {
		return "", fmt.Errorf("failed to list corpora: %v", err)
	}

	var corpusResourceName string
	for _, corpus := range existingCorpora.RagCorpora {
		if corpus.DisplayName == corpusDisplayName {
			corpusResourceName = corpus.Name
			break
		}
	}

	if corpusResourceName == "" {
		return "", fmt.Errorf("corpus '%s' not found", corpusName)
	}

	// List all files in the corpus
	filesResponse, err := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName).Do()
	if err != nil {
		return "", fmt.Errorf("failed to list files in corpus: %v", err)
	}

	// Try exact match on display name
	for _, file := range filesResponse.RagFiles {
		if file.DisplayName == displayName {
			parts := strings.Split(file.Name, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1], nil
			}
		}
	}

	// Try substring match (case-insensitive, ignore extension)
	displayNameLower := strings.ToLower(displayName)
	displayNameNoExt := strings.TrimSuffix(displayNameLower, getFileExtension(displayNameLower))
	for _, file := range filesResponse.RagFiles {
		fileDisplayLower := strings.ToLower(file.DisplayName)
		fileDisplayNoExt := strings.TrimSuffix(fileDisplayLower, getFileExtension(fileDisplayLower))
		if strings.Contains(fileDisplayLower, displayNameLower) || strings.Contains(fileDisplayNoExt, displayNameNoExt) {
			parts := strings.Split(file.Name, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1], nil
			}
		}
	}

	// Try matching by temp object name (substring)
	// Example: displayName="Agent Development Kit Hackathon _ Project Submission.pdf", tempObjectName="temp/8251d763-69d5-4086-b105-b721e55b7b76.pdf"
	// Try to find a file whose displayName or name contains the temp object name (or its basename)
	for _, file := range filesResponse.RagFiles {
		if strings.Contains(file.DisplayName, displayName) {
			parts := strings.Split(file.Name, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1], nil
			}
		}
	}

	// Try matching by basename of temp object name
	// If displayName is not found, try matching by basename of temp object name
	// This is useful if the file was ingested with the temp object name
	for _, file := range filesResponse.RagFiles {
		fileBase := getFileBase(file.DisplayName)
		if strings.Contains(fileBase, displayName) || strings.Contains(displayName, fileBase) {
			parts := strings.Split(file.Name, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1], nil
			}
		}
	}

	// Log all files in the corpus for debugging
	fmt.Printf("[VertexAIService] No RAG file found for displayName='%s' in corpus='%s'. Corpus files:\n", displayName, corpusName)
	for _, file := range filesResponse.RagFiles {
		fmt.Printf("  - displayName: %s, name: %s\n", file.DisplayName, file.Name)
	}

	return "", fmt.Errorf("no RAG file found with display name or temp object name '%s' in corpus '%s'", displayName, corpusName)
}

// Helper functions for extension and basename
func getFileExtension(filename string) string {
	dot := strings.LastIndex(filename, ".")
	if dot == -1 {
		return ""
	}
	return filename[dot:]
}

func getFileBase(filename string) string {
	slash := strings.LastIndex(filename, "/")
	if slash != -1 {
		filename = filename[slash+1:]
	}
	dot := strings.LastIndex(filename, ".")
	if dot != -1 {
		filename = filename[:dot]
	}
	return filename
}

// ExtractRAGFileIDFromOperation extracts RAG file ID from operation response
func (v *VertexAIService) ExtractRAGFileIDFromOperation(operationResponse map[string]interface{}) string {
	if response, hasResponse := operationResponse["response"]; hasResponse {
		if responseMap, ok := response.(map[string]interface{}); ok {
			if ragFiles, hasRagFiles := responseMap["ragFiles"]; hasRagFiles {
				if ragFilesList, ok := ragFiles.([]interface{}); ok && len(ragFilesList) > 0 {
					if ragFile, ok := ragFilesList[0].(map[string]interface{}); ok {
						if name, hasName := ragFile["name"].(string); hasName {
							// Extract the RAG file ID from the full resource name
							// Format: projects/.../ragCorpora/.../ragFiles/{ragFileId}
							parts := strings.Split(name, "/")
							if len(parts) > 0 {
								return parts[len(parts)-1]
							}
						}
					}
				}
			}
		}
	}
	return ""
}
