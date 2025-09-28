package controller

import (
	"lumenslate/internal/service"
	"lumenslate/internal/utils"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackQuestionBankUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackQuestionUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackIAUsage godoc
// @Summary Track IA usage
// @Description Tracks Intelligent Agent usage for a specific user
// @Tags Usage Tracking
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body map[string]int64 true "Usage count"
// @Success 200 {object} map[string]interface{} "Usage tracked successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/track/ia-agent [post]
func TrackIAUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackIAUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackLumenAgentUsage tracks Lumen Agent usage for a user
func TrackLumenAgentUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackLumenAgentUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackRAAgentUsage tracks Research Assistant Agent usage for a user
func TrackRAAgentUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackRAAgentUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackRecapClassUsage tracks recap class usage for a user
func TrackRecapClassUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackRecapClassUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage tracked successfully"})
}

// TrackAssignmentExportUsage tracks assignment export usage for a user
func TrackAssignmentExportUsage(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Count int64 `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackAssignmentExportUsage(userID, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
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
// @Param body body service.BulkUsageTrackingRequest true "Bulk usage data"
// @Success 200 {object} map[string]interface{} "Bulk usage tracked successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Failed to track usage"
// @Router /api/v1/usage/user/{id}/track/bulk [post]
func TrackBulkUsage(c *gin.Context) {
	userID := c.Param("id")

	var req service.BulkUsageTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserID = userID // Set the user ID from the URL parameter

	if err := utils.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := usageTrackingService.TrackBulkUsage(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk usage tracked successfully"})
}

// GetCurrentUsageMetrics godoc
// @Summary Get current usage metrics
// @Description Retrieves current period usage metrics for a specific user
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.UsageTracking "Current usage metrics"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage metrics"
// @Router /api/v1/usage/user/{id}/current [get]
func GetCurrentUsageMetrics(c *gin.Context) {
	userID := c.Param("id")

	metrics, err := usageTrackingService.GetCurrentUsageMetrics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetAggregatedUsageMetrics godoc
// @Summary Get aggregated usage metrics
// @Description Retrieves all-time aggregated usage metrics for a specific user
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "Aggregated usage metrics"
// @Failure 500 {object} map[string]interface{} "Failed to fetch aggregated usage metrics"
// @Router /api/v1/usage/user/{id}/aggregated [get]
func GetAggregatedUsageMetrics(c *gin.Context) {
	userID := c.Param("id")

	metrics, err := usageTrackingService.GetAggregatedUserUsage(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch aggregated usage metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetUsageTrackingByPeriod godoc
// @Summary Get usage by period
// @Description Retrieves usage tracking for a specific user and period
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Param period path string true "Period (e.g., 2023-12)"
// @Success 200 {object} model.UsageTracking "Usage tracking for period"
// @Failure 404 {object} map[string]interface{} "Usage tracking not found for the specified period"
// @Router /api/v1/usage/user/{id}/period/{period} [get]
func GetUsageTrackingByPeriod(c *gin.Context) {
	userID := c.Param("id")
	period := c.Param("period")

	usage, err := usageTrackingService.GetUsageTrackingByPeriod(userID, period)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usage tracking not found for the specified period"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetAllUserUsageHistory godoc
// @Summary Get user usage history
// @Description Retrieves all usage tracking records for a specific user
// @Tags Usage Tracking
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {array} model.UsageTracking "User usage history"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage history"
// @Router /api/v1/usage/user/{id}/history [get]
func GetAllUserUsageHistory(c *gin.Context) {
	userID := c.Param("id")

	usageHistory, err := usageTrackingService.GetAllUserUsageHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage history"})
		return
	}

	c.JSON(http.StatusOK, usageHistory)
}

// GetAllUsageTracking godoc
// @Summary Get all usage tracking records
// @Description Retrieves all usage tracking records with optional filters
// @Tags Usage Tracking
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param period query string false "Filter by period"
// @Param limit query string false "Pagination limit (default 10)"
// @Param offset query string false "Pagination offset (default 0)"
// @Success 200 {array} model.UsageTracking "Usage tracking records"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage tracking records"
// @Router /api/v1/usage [get]
func GetAllUsageTracking(c *gin.Context) {
	filters := service.UsageTrackingFilters{
		UserID: c.Query("user_id"),
		Period: c.Query("period"),
		Limit:  c.DefaultQuery("limit", "10"),
		Offset: c.DefaultQuery("offset", "0"),
	}

	usageRecords, err := usageTrackingService.GetAllUsageTracking(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage tracking records"})
		return
	}

	c.JSON(http.StatusOK, usageRecords)
}

// GetUsageSummaryByPeriod godoc
// @Summary Get usage summary by period
// @Description Retrieves aggregated usage summary for a specific period
// @Tags Usage Tracking
// @Produce json
// @Param period path string true "Period (e.g., 2023-12)"
// @Success 200 {object} map[string]interface{} "Usage summary for period"
// @Failure 500 {object} map[string]interface{} "Failed to fetch usage summary"
// @Router /api/v1/usage/summary/{period} [get]
func GetUsageSummaryByPeriod(c *gin.Context) {
	period := c.Param("period")

	summary, err := usageTrackingService.GetUsageSummaryByPeriod(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ResetUserUsage godoc
// @Summary Reset user usage
// @Description Resets usage counters for a user (creates new tracking record for current period)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Usage reset successfully",
		"new_usage": newUsage,
	})
}

// Simple tracking endpoints that don't require a JSON body (useful for quick increments)

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

	// Get count from query parameter, default to 1
	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackQuestionBankUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
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

	// Get count from query parameter, default to 1
	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackQuestionUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// IncrementIAUsage increments IA usage by 1
func IncrementIAUsage(c *gin.Context) {
	userID := c.Param("id")

	// Get count from query parameter, default to 1
	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackIAUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// IncrementLumenAgentUsage increments Lumen Agent usage by 1
func IncrementLumenAgentUsage(c *gin.Context) {
	userID := c.Param("id")

	// Get count from query parameter, default to 1
	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackLumenAgentUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// IncrementRAAgentUsage increments RA Agent usage by 1
func IncrementRAAgentUsage(c *gin.Context) {
	userID := c.Param("id")

	// Get count from query parameter, default to 1
	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackRAAgentUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// IncrementRecapClassUsage increments recap class usage by 1
func IncrementRecapClassUsage(c *gin.Context) {
	userID := c.Param("id")

	// Get count from query parameter, default to 1
	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackRecapClassUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}

// IncrementAssignmentExportUsage increments assignment export usage by 1
func IncrementAssignmentExportUsage(c *gin.Context) {
	userID := c.Param("id")

	// Get count from query parameter, default to 1
	count := int64(1)
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			count = parsedCount
		}
	}

	if err := usageTrackingService.TrackAssignmentExportUsage(userID, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage incremented successfully",
		"count":   count,
	})
}
