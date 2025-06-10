package controller

import (
	"net/http"

	"lumenslate/internal/service"

	"github.com/gin-gonic/gin"
)

// POST /generate-context
func GenerateContextHandler(c *gin.Context) {
	var req struct {
		Question string   `json:"question"`
		Keywords []string `json:"keywords"`
		Language string   `json:"language"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content, err := service.GenerateContext(req.Question, req.Keywords, req.Language) // use service
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// POST /detect-variables
func DetectVariablesHandler(c *gin.Context) {
	var req struct {
		Question string `json:"question"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variables, err := service.DetectVariables(req.Question) // use service
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variables": variables})
}

// POST /segment-question
func SegmentQuestionHandler(c *gin.Context) {
	var req struct {
		Question string `json:"question"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	segmented, err := service.SegmentQuestion(req.Question) // use service
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"segmentedQuestion": segmented})
}

// POST /generate-mcq
func GenerateMCQVariationsHandler(c *gin.Context) {
	var req struct {
		Question    string   `json:"question"`
		Options     []string `json:"options"`
		AnswerIndex int32    `json:"answerIndex"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variations, err := service.GenerateMCQVariations(req.Question, req.Options, req.AnswerIndex) // use service
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variations": variations})
}

// POST /generate-msq
func GenerateMSQVariationsHandler(c *gin.Context) {
	var req struct {
		Question      string   `json:"question"`
		Options       []string `json:"options"`
		AnswerIndices []int32  `json:"answerIndices"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variations, err := service.GenerateMSQVariations(req.Question, req.Options, req.AnswerIndices) // use service
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variations": variations})
}

// POST /filter-randomize
func FilterAndRandomizeHandler(c *gin.Context) {
	var req struct {
		Question   string `json:"question"`
		UserPrompt string `json:"userPrompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vars, err := service.FilterAndRandomize(req.Question, req.UserPrompt) // use service
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variables": vars})
}
