package ai

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// CheckOperationStatusRequest represents the request to check operation status
type CheckOperationStatusRequest struct {
	OperationName string `json:"operation_name" binding:"required"`
}

// CheckOperationStatusHandler godoc
// @Summary      Check Vertex AI Operation Status
// @Description  Check the status of a Vertex AI operation (like RAG file import)
// @Tags         AI Operations
// @Accept       json
// @Produce      json
// @Param        body  body  ai.CheckOperationStatusRequest  true  "Operation status request"
// @Success      200   {object}  map[string]interface{}  "Operation status retrieved successfully"
// @Failure      400   {object}  map[string]interface{}  "Invalid request body"
// @Failure      500   {object}  map[string]interface{}  "Internal server error"
// @Router       /ai/operations/status [post]
func CheckOperationStatusHandler(c *gin.Context) {
	var req CheckOperationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := checkVertexAIOperationStatus(req.OperationName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to check operation status: %v", err)})
		return
	}

	c.JSON(http.StatusOK, status)
}

// checkVertexAIOperationStatus checks the status of a Vertex AI operation
func checkVertexAIOperationStatus(operationName string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Project configuration
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
