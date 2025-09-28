package controller

import (
	"lumenslate/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var usageTrackingService = service.NewUsageTrackingService()

// TrackQuestionBankUsage godoc
// @Summary Track question bank usage
// @Description Tracks question bank usage for a specific user
// @Tags Usage Tracking
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body map[string]int64 true "Usage count"
// @Success 200 {object} map[string]interface{} "Usage tracked successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/track/question-banks [post]
func TrackQuestionBankUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := usageTrackingService.TrackQuestionBankUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackQuestionUsage godoc
// @Summary Track question usage
// @Description Tracks question usage for a specific user
// @Tags Usage Tracking
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body map[string]int64 true "Usage count"
// @Success 200 {object} map[string]interface{} "Usage tracked successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/track/questions [post]
func TrackQuestionUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := usageTrackingService.TrackQuestionUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackAIGenerationUsage godoc
// @Summary Track AI generation usage
// @Description Tracks AI-powered feature usage for a specific user
// @Tags Usage Tracking
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body map[string]int64 true "Usage count"
// @Success 200 {object} map[string]interface{} "Usage tracked successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/track/ai-generation [post]
func TrackAIGenerationUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := usageTrackingService.TrackAIGenerationUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// Legacy endpoints for backward compatibility
func TrackIAUsage(c *gin.Context) {
	TrackAIGenerationUsage(c)
}

func TrackLumenAgentUsage(c *gin.Context) {
	TrackAIGenerationUsage(c)
}

func TrackRAAgentUsage(c *gin.Context) {
	TrackAIGenerationUsage(c)
}

func TrackRecapClassUsage(c *gin.Context) {
	TrackAIGenerationUsage(c)
}

// TrackAssignmentExportUsage tracks assignment export usage for a user
func TrackAssignmentExportUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := usageTrackingService.TrackAssignmentExportUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackBulkUsage godoc
// @Summary Track bulk usage
// @Description Tracks multiple usage types for a user in a single request
// @Tags Usage Tracking
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body service.BulkUsageRequest true "Bulk usage data"
// @Success 200 {object} map[string]interface{} "Bulk usage tracked successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/track/bulk [post]
func TrackBulkUsage(c *gin.Context) {
	userID := c.Param("id")

	var req service.BulkUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = userID
	if err := usageTrackingService.TrackBulkUsage(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track bulk usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk usage tracked successfully"})
}

// GetCurrentUsageMetrics godoc
// @Summary Get current usage metrics
// @Description Retrieves current usage metrics for a specific user
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.Usage "Current usage metrics"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage metrics"
// @Router /api/v1/usage/user/{id}/current [get]
func GetCurrentUsageMetrics(c *gin.Context) {
	userID := c.Param("id")

	metrics, err := usageTrackingService.GetCurrentUsageMetrics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage metrics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetUsageByUser godoc
// @Summary Get usage by user
// @Description Retrieves usage tracking for a specific user
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.Usage "User usage data"
// @Failure 404 {object} map[string]interface{} "Usage not found"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage"
// @Router /api/v1/usage/user/{id} [get]
func GetUsageByUser(c *gin.Context) {
	userID := c.Param("id")

	usage, err := usageTrackingService.GetUsageByUserID(userID)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usage not found for user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetAllUsage godoc
// @Summary Get all usage records
// @Description Retrieves all usage records with optional filters
// @Tags Usage Tracking
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param lookup_key query string false "Filter by subscription lookup key"
// @Param limit query string false "Pagination limit (default 10)"
// @Param offset query string false "Pagination offset (default 0)"
// @Success 200 {array} model.Usage "Usage records"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage records"
// @Router /api/v1/usage [get]
func GetAllUsage(c *gin.Context) {
	filters := service.UsageFilters{
		UserID:    c.Query("user_id"),
		LookupKey: c.Query("lookup_key"),
		Limit:     c.DefaultQuery("limit", "10"),
		Offset:    c.DefaultQuery("offset", "0"),
	}

	usageRecords, err := usageTrackingService.GetAllUsage(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage records", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usageRecords)
}

// GetUsageSummary godoc
// @Summary Get usage summary
// @Description Retrieves aggregated usage summary for all users
// @Tags Usage Tracking
// @Produce json
// @Param lookup_key query string false "Filter by subscription lookup key"
// @Success 200 {object} map[string]interface{} "Usage summary"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage summary"
// @Router /api/v1/usage/summary [get]
func GetUsageSummary(c *gin.Context) {
	lookupKey := c.Query("lookup_key")

	// For now, we'll use a mock period. In a real implementation,
	// you might want to get the current billing period
	summary, err := usageTrackingService.GetUsageSummaryByPeriod(lookupKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage summary", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ResetUserUsage godoc
// @Summary Reset user usage
// @Description Resets usage counters for a user
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "Usage reset successfully"
// @Failure 500 {object} map[string]interface{} "Failed to reset usage"
// @Router /api/v1/usage/user/{id}/reset [post]
func ResetUserUsage(c *gin.Context) {
	userID := c.Param("id")

	newUsage, err := usageTrackingService.ResetUserUsage(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Usage reset successfully",
		"new_usage": newUsage,
	})
}

// DeleteUsage godoc
// @Summary Delete usage record
// @Description Deletes a usage record by ID
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "Usage ID"
// @Success 200 {object} map[string]interface{} "Usage deleted successfully"
// @Failure 500 {object} map[string]interface{} "Failed to delete usage"
// @Router /api/v1/usage/{id} [delete]
func DeleteUsage(c *gin.Context) {
	id := c.Param("id")

	if err := usageTrackingService.DeleteUsage(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage deleted successfully"})
}

// Simple increment endpoints (useful for quick increments)

// IncrementQuestionBankUsage godoc
// @Summary Increment question bank usage
// @Description Increments question bank usage by 1 or specified count
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Param count query int false "Usage count to increment (default 1)"
// @Success 200 {object} map[string]interface{} "Usage incremented successfully"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/increment/question-banks [post]
func IncrementQuestionBankUsage(c *gin.Context) {
	userID := c.Param("id")

	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackQuestionBankUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// IncrementQuestionUsage godoc
// @Summary Increment question usage
// @Description Increments question usage by 1 or specified count
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Param count query int false "Usage count to increment (default 1)"
// @Success 200 {object} map[string]interface{} "Usage incremented successfully"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/increment/questions [post]
func IncrementQuestionUsage(c *gin.Context) {
	userID := c.Param("id")

	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackQuestionUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// IncrementAIGenerationUsage increments AI generation usage
func IncrementAIGenerationUsage(c *gin.Context) {
	userID := c.Param("id")

	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackAIGenerationUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// Legacy increment methods for backward compatibility
func IncrementIAUsage(c *gin.Context) {
	IncrementAIGenerationUsage(c)
}

func IncrementLumenAgentUsage(c *gin.Context) {
	IncrementAIGenerationUsage(c)
}

func IncrementRAAgentUsage(c *gin.Context) {
	IncrementAIGenerationUsage(c)
}

func IncrementRecapClassUsage(c *gin.Context) {
	IncrementAIGenerationUsage(c)
}

// IncrementAssignmentExportUsage increments assignment export usage
func IncrementAssignmentExportUsage(c *gin.Context) {
	userID := c.Param("id")

	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackAssignmentExportUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// Legacy methods for backward compatibility with old API
func GetAggregatedUsageMetrics(c *gin.Context) {
	GetCurrentUsageMetrics(c)
}

func GetUsageTrackingByPeriod(c *gin.Context) {
	GetCurrentUsageMetrics(c)
}

func GetAllUserUsageHistory(c *gin.Context) {
	userID := c.Param("id")
	filters := service.UsageFilters{
		UserID: userID,
		Limit:  "100",
	}

	usageRecords, err := usageTrackingService.GetAllUsage(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage history", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usageRecords)
}

func GetAllUsageTracking(c *gin.Context) {
	GetAllUsage(c)
}

func GetUsageSummaryByPeriod(c *gin.Context) {
	period := c.Param("period")

	summary, err := usageTrackingService.GetUsageSummaryByPeriod(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage summary", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
