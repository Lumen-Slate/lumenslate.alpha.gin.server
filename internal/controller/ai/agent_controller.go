package ai

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"

	service "lumenslate/internal/grpc_service"

	"github.com/gin-gonic/gin"
)

// AgentHandler godoc
// @Summary      Process Request with AI Agent
// @Description  Process requests using the AI Agent gRPC service with support for file uploads. Handles text processing, analysis, and generation tasks for educational content.
// @Tags         AI Agent
// @Accept       multipart/form-data
// @Produce      json
// @Param        teacherId  formData  string  true   "Teacher ID for context and personalization"
// @Param        role       formData  string  true   "Role/context for the AI agent processing"
// @Param        message    formData  string  true   "Message or prompt for the AI agent"
// @Param        file       formData  file    false  "Optional file upload for processing"
// @Param        fileType   formData  string  false  "Type of the uploaded file (if file is provided)"
// @Param        createdAt  formData  string  false  "Creation timestamp (ISO format)"
// @Param        updatedAt  formData  string  false  "Update timestamp (ISO format)"
// @Success      200        {object}  map[string]interface{}  "AI agent response with processed data and metadata"
// @Failure      400        {object}  map[string]interface{}  "Invalid request body, missing required fields, or file processing error"
// @Failure      500        {object}  map[string]interface{}  "Internal server error during AI processing"
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file content"})
			return
		}

		// Convert to base64 string for service layer
		fileContent = base64.StdEncoding.EncodeToString(fileBytes)
		log.Printf("[AI] File processed: %s, size: %d bytes", req.File.Filename, len(fileBytes))
	}

	resp, err := service.LumenAgent(
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
