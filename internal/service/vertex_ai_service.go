package service

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"lumenslate/internal/utils"

	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

type VertexAIService struct {
	projectID string
	location  string
}

// NewVertexAIService creates a new Vertex AI service instance
func NewVertexAIService() *VertexAIService {
	projectID := utils.GetProjectID()
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
