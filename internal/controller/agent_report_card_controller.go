package controller

import (
	"net/http"

	"lumenslate/internal/model"
	"lumenslate/internal/repository"
	"lumenslate/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetAllAgentReportCards handles GET /api/agent-report-cards
func GetAllAgentReportCards(c *gin.Context) {
	// Extract query parameters for filtering
	filters := make(map[string]string)

	if studentId := c.Query("studentId"); studentId != "" {
		filters["studentId"] = studentId
	}

	if userId := c.Query("userId"); userId != "" {
		filters["userId"] = userId
	}

	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}

	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	reportCards, err := repository.GetAllAgentReportCards(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve agent report cards",
			"details": err.Error(),
		})
		return
	}

	// Convert to camelCase before returning
	camelCaseReportCards, err := utils.ConvertStructToMap(reportCards)
	if err != nil {
		camelCaseReportCards = reportCards // fallback to original if conversion fails
	}
	c.JSON(http.StatusOK, gin.H{
		"agentReportCards": camelCaseReportCards,
		"totalCount":       len(reportCards),
	})
}

// GetAgentReportCardByID handles GET /api/agent-report-cards/:id
func GetAgentReportCardByID(c *gin.Context) {
	id := c.Param("id")

	reportCard, err := repository.GetAgentReportCardByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Agent report card not found",
			"details": err.Error(),
		})
		return
	}

	// Convert to camelCase before returning
	camelCaseReportCard, err := utils.ConvertStructToMap(reportCard)
	if err != nil {
		camelCaseReportCard = reportCard // fallback to original if conversion fails
	}
	c.JSON(http.StatusOK, camelCaseReportCard)
}

// GetAgentReportCardsByStudentID handles GET /api/agent-report-cards/student/:studentId
func GetAgentReportCardsByStudentID(c *gin.Context) {
	studentId := c.Param("studentId")

	reportCards, err := repository.GetAgentReportCardsByStudentID(studentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve agent report cards for student",
			"details": err.Error(),
		})
		return
	}

	if len(reportCards) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "No agent report cards found for student",
			"studentId": studentId,
		})
		return
	}

	// Convert to camelCase before returning
	camelCaseReportCards, err := utils.ConvertStructToMap(reportCards)
	if err != nil {
		camelCaseReportCards = reportCards // fallback to original if conversion fails
	}
	c.JSON(http.StatusOK, gin.H{
		"agentReportCards": camelCaseReportCards,
		"studentId":        studentId,
		"totalCount":       len(reportCards),
	})
}

// CreateAgentReportCard handles POST /api/agent-report-cards
func CreateAgentReportCard(c *gin.Context) {
	var reportCard model.AgentReportCard

	if err := c.ShouldBindJSON(&reportCard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	createdReportCard, err := repository.CreateAgentReportCard(reportCard)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create agent report card",
			"details": err.Error(),
		})
		return
	}

	// Convert to camelCase before returning
	camelCaseReportCard, err := utils.ConvertStructToMap(createdReportCard)
	if err != nil {
		camelCaseReportCard = createdReportCard // fallback to original if conversion fails
	}
	c.JSON(http.StatusCreated, camelCaseReportCard)
}

// UpdateAgentReportCard handles PUT /api/agent-report-cards/:id
func UpdateAgentReportCard(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	updatedReportCard, err := repository.UpdateAgentReportCard(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update agent report card",
			"details": err.Error(),
		})
		return
	}

	// Convert to camelCase before returning
	camelCaseReportCard, err := utils.ConvertStructToMap(updatedReportCard)
	if err != nil {
		camelCaseReportCard = updatedReportCard // fallback to original if conversion fails
	}
	c.JSON(http.StatusOK, camelCaseReportCard)
}

// DeleteAgentReportCard handles DELETE /api/agent-report-cards/:id
func DeleteAgentReportCard(c *gin.Context) {
	id := c.Param("id")

	err := repository.DeleteAgentReportCard(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete agent report card",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent report card deleted successfully",
		"id":      id,
	})
}
