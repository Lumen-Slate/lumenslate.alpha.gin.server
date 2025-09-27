package controller

import (
	"lumenslate/internal/service"
	"lumenslate/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var usageTrackingService = service.NewUsageTrackingService()

// TrackQuestionBankUsage tracks question bank usage for a user
func TrackQuestionBankUsage(c *gin.Context) {
	userID := c.Param("userId")

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

// TrackQuestionUsage tracks question usage for a user
func TrackQuestionUsage(c *gin.Context) {
	userID := c.Param("userId")

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

// TrackIAUsage tracks Intelligent Agent usage for a user
func TrackIAUsage(c *gin.Context) {
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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

// TrackBulkUsage tracks multiple usage types for a user in a single request
func TrackBulkUsage(c *gin.Context) {
	userID := c.Param("userId")

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

// GetCurrentUsageMetrics retrieves current usage metrics for a user
func GetCurrentUsageMetrics(c *gin.Context) {
	userID := c.Param("userId")

	metrics, err := usageTrackingService.GetCurrentUsageMetrics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetAggregatedUsageMetrics retrieves aggregated usage metrics for a user
func GetAggregatedUsageMetrics(c *gin.Context) {
	userID := c.Param("userId")

	metrics, err := usageTrackingService.GetAggregatedUserUsage(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch aggregated usage metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetUsageTrackingByPeriod retrieves usage tracking for a specific user and period
func GetUsageTrackingByPeriod(c *gin.Context) {
	userID := c.Param("userId")
	period := c.Param("period")

	usage, err := usageTrackingService.GetUsageTrackingByPeriod(userID, period)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usage tracking not found for the specified period"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetAllUserUsageHistory retrieves all usage tracking records for a user
func GetAllUserUsageHistory(c *gin.Context) {
	userID := c.Param("userId")

	usageHistory, err := usageTrackingService.GetAllUserUsageHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage history"})
		return
	}

	c.JSON(http.StatusOK, usageHistory)
}

// GetAllUsageTracking retrieves all usage tracking records with optional filters
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

// GetUsageSummaryByPeriod retrieves usage summary for a specific period
func GetUsageSummaryByPeriod(c *gin.Context) {
	period := c.Param("period")

	summary, err := usageTrackingService.GetUsageSummaryByPeriod(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ResetUserUsage resets usage counters for a user (creates new tracking record for current period)
func ResetUserUsage(c *gin.Context) {
	userID := c.Param("userId")

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

// IncrementQuestionBankUsage increments question bank usage by 1
func IncrementQuestionBankUsage(c *gin.Context) {
	userID := c.Param("userId")

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

// IncrementQuestionUsage increments question usage by 1
func IncrementQuestionUsage(c *gin.Context) {
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
	userID := c.Param("userId")

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
