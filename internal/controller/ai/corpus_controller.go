package ai

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"lumenslate/internal/utils"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// ListAllCorporaHandler godoc
// @Summary      List All RAG Corpora
// @Description  Retrieve a comprehensive list of all RAG corpora available in the Vertex AI project, including their display names, creation times, and update times.
// @Tags         AI RAG Management
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "List of all corpora with their metadata and count"
// @Failure      500  {object}  map[string]interface{}  "Internal server error during corpora retrieval from Vertex AI"
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
