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
