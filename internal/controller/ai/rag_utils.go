package ai

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"lumenslate/internal/model"
	"lumenslate/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// SyncRAGFileIDsHandler godoc
// @Summary      Sync RAG File IDs
// @Description  Find and update missing RAG file IDs in the database by matching with actual RAG engine files
// @Tags         AI Document Management
// @Accept       json
// @Produce      json
// @Param        corpusName  query  string  true  "Name of the RAG corpus to sync"
// @Success      200         {object}  map[string]interface{}  "Sync completed successfully"
// @Failure      400         {object}  map[string]interface{}  "Invalid request parameters"
// @Failure      500         {object}  map[string]interface{}  "Internal server error"
// @Router       /ai/rag-agent/sync-file-ids [post]
func SyncRAGFileIDsHandler(c *gin.Context) {
	corpusName := c.Query("corpusName")
	if corpusName == "" {
		c.JSON(400, gin.H{"error": "corpusName query parameter is required"})
		return
	}

	result, err := syncRAGFileIDs(corpusName)
	if err != nil {
		log.Printf("[AI] Failed to sync RAG file IDs: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to sync RAG file IDs: %v", err)})
		return
	}

	c.JSON(200, result)
}

// syncRAGFileIDs finds and updates missing RAG file IDs in the database
func syncRAGFileIDs(corpusName string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Initialize services
	docRepo := repository.NewDocumentRepository()

	// Get all documents for this corpus from database
	documents, err := docRepo.GetDocumentsByCorpus(ctx, corpusName)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents from database: %w", err)
	}

	// Get all RAG files from Vertex AI
	ragFiles, err := listRAGFiles(corpusName)
	if err != nil {
		return nil, fmt.Errorf("failed to list RAG files: %w", err)
	}

	log.Printf("[AI] Found %d documents in database and %d files in RAG engine for corpus '%s'",
		len(documents), len(ragFiles), corpusName)

	updated := 0
	matched := 0

	// Try to match documents with RAG files
	for _, doc := range documents {
		if doc.RAGFileID != "" {
			matched++
			continue // Already has RAG file ID
		}

		// Try to find matching RAG file
		for _, ragFile := range ragFiles {
			if isDocumentMatch(doc, ragFile) {
				// Extract RAG file ID from the full resource name
				ragFileID := extractRAGFileID(ragFile["name"].(string))

				// Update database with RAG file ID
				if err := docRepo.UpdateFields(ctx, doc.FileID, bson.M{
					"ragFileId": ragFileID,
				}); err != nil {
					log.Printf("[AI] Failed to update RAG file ID for document %s: %v", doc.FileID, err)
					continue
				}

				log.Printf("[AI] Updated document %s with RAG file ID: %s", doc.FileID, ragFileID)
				updated++
				break
			}
		}
	}

	return map[string]interface{}{
		"status":         "success",
		"corpusName":     corpusName,
		"documentsInDB":  len(documents),
		"filesInRAG":     len(ragFiles),
		"alreadyMatched": matched,
		"newlyUpdated":   updated,
		"message":        fmt.Sprintf("Sync completed: %d documents updated with RAG file IDs", updated),
	}, nil
}

// listRAGFiles gets all files from a RAG corpus
func listRAGFiles(corpusName string) ([]map[string]interface{}, error) {
	ctx := context.Background()

	// Project configuration
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	// Use regional endpoint
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/", location)

	// Create AI Platform service client
	service, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %v", err)
	}

	// Clean corpus name for use as display name
	displayName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(corpusName, "_")

	// Find the corpus
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

	// List files in the corpus
	filesResponse, err := service.Projects.Locations.RagCorpora.RagFiles.List(corpusResourceName).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in corpus: %v", err)
	}

	var files []map[string]interface{}
	for _, file := range filesResponse.RagFiles {
		files = append(files, map[string]interface{}{
			"name":        file.Name,
			"displayName": file.DisplayName,
			"createTime":  file.CreateTime,
			"updateTime":  file.UpdateTime,
		})
	}

	return files, nil
}

// isDocumentMatch checks if a database document matches a RAG file
func isDocumentMatch(doc model.Document, ragFile map[string]interface{}) bool {
	ragDisplayName, ok := ragFile["displayName"].(string)
	if !ok {
		return false
	}

	// Try multiple matching strategies
	return doc.DisplayName == ragDisplayName ||
		strings.Contains(ragDisplayName, doc.DisplayName) ||
		strings.Contains(doc.DisplayName, ragDisplayName) ||
		// Handle case where one has extension and other doesn't
		strings.HasPrefix(ragDisplayName, strings.TrimSuffix(doc.DisplayName, filepath.Ext(doc.DisplayName))) ||
		strings.HasPrefix(doc.DisplayName, strings.TrimSuffix(ragDisplayName, filepath.Ext(ragDisplayName)))
}

// extractRAGFileID extracts the RAG file ID from a full resource name
func extractRAGFileID(resourceName string) string {
	// Format: projects/.../ragCorpora/.../ragFiles/{ragFileId}
	parts := strings.Split(resourceName, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
