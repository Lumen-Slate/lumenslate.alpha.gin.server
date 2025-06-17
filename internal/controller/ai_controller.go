package controller

import (
	"log"
	"net/http"

	service "lumenslate/internal/grpc_service"

	"github.com/gin-gonic/gin"
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

type AgentRequest struct {
	UserId    string `form:"userId" binding:"required"`
	Role      string `form:"role" binding:"required"`
	Message   string `form:"message" binding:"required"`
	File      string `form:"file"`
	FileType  string `form:"fileType"`
	CreatedAt string `form:"createdAt"`
	UpdatedAt string `form:"updatedAt"`
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
// @Description  Calls the Agent gRPC method with the provided data
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        body  body  controller.AgentRequest  true  "Request body"
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
	resp, err := service.Agent(
		req.File,
		req.FileType,
		req.UserId,
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
