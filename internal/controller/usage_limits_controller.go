package controller

import (
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"lumenslate/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var usageLimitsService = service.NewUsageLimitsService()

// CreateUsageLimits godoc
// @Summary Create new usage limits
// @Description Creates new usage limits for a subscription plan. Supports flexible value types: integers, "unlimited", "custom", or -1 for unlimited.
// @Tags Usage Limits
// @Accept json
// @Produce json
// @Param usageLimits body service.CreateUsageLimitsRequest true "Usage limits data"
// @Success 201 {object} model.UsageLimits "Successfully created usage limits"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/usage-limits [post]
// @Example application/json Basic Plan {"plan_name":"basic","teachers":5,"classrooms":10,"students_per_classroom":30,"question_banks":50,"questions":1000,"assignment_exports_per_day":10,"ai":{"independent_agent":100,"lumen_agent":50,"rag_agent":25,"rag_document_uploads":10}}
// @Example application/json Premium Plan {"plan_name":"premium","teachers":25,"classrooms":"unlimited","students_per_classroom":50,"question_banks":"unlimited","questions":"unlimited","assignment_exports_per_day":"unlimited","ai":{"independent_agent":500,"lumen_agent":300,"rag_agent":150,"rag_document_uploads":100}}
// @Example application/json Enterprise Plan {"plan_name":"enterprise","teachers":"custom","classrooms":"unlimited","students_per_classroom":"custom","question_banks":"unlimited","questions":"unlimited","assignment_exports_per_day":"unlimited","ai":{"independent_agent":"unlimited","lumen_agent":"unlimited","rag_agent":"unlimited","rag_document_uploads":"unlimited"}}
func CreateUsageLimits(c *gin.Context) {
	var req service.CreateUsageLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usageLimits, err := usageLimitsService.CreateUsageLimits(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, usageLimits)
}

// GetUsageLimits godoc
// @Summary Get usage limits by ID
// @Description Retrieves usage limits by their ID
// @Tags Usage Limits
// @Produce json
// @Param id path string true "Usage Limits ID"
// @Success 200 {object} model.UsageLimits "Successfully retrieved usage limits"
// @Failure 404 {object} map[string]interface{} "Usage limits not found"
// @Router /api/v1/usage-limits/{id} [get]
// @Example application/json Success Response {"id":"507f1f77bcf86cd799439011","plan_name":"premium","teachers":25,"classrooms":"unlimited","students_per_classroom":50,"question_banks":"unlimited","questions":"unlimited","assignment_exports_per_day":"unlimited","ai":{"independent_agent":500,"lumen_agent":300,"rag_agent":150,"rag_document_uploads":100},"created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z","is_active":true}
func GetUsageLimits(c *gin.Context) {
	id := c.Param("id")
	usageLimits, err := usageLimitsService.GetUsageLimitsByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usage limits not found"})
		return
	}
	c.JSON(http.StatusOK, usageLimits)
}

// GetUsageLimitsByPlan godoc
// @Summary Get usage limits by plan name
// @Description Retrieves usage limits by plan name (basic, premium, enterprise, etc.)
// @Tags Usage Limits
// @Produce json
// @Param planName path string true "Plan Name" example(premium)
// @Success 200 {object} model.UsageLimits "Successfully retrieved usage limits"
// @Failure 404 {object} map[string]interface{} "Usage limits not found"
// @Router /api/v1/usage-limits/plan/{planName} [get]
// @Example application/json Premium Plan Response {"id":"507f1f77bcf86cd799439011","plan_name":"premium","teachers":25,"classrooms":"unlimited","students_per_classroom":50,"question_banks":"unlimited","questions":"unlimited","assignment_exports_per_day":"unlimited","ai":{"independent_agent":500,"lumen_agent":300,"rag_agent":150,"rag_document_uploads":100},"created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z","is_active":true}
func GetUsageLimitsByPlan(c *gin.Context) {
	planName := c.Param("planName")
	usageLimits, err := usageLimitsService.GetUsageLimitsByPlanName(planName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usage limits not found for plan"})
		return
	}
	c.JSON(http.StatusOK, usageLimits)
}

// GetAllUsageLimits godoc
// @Summary Get all usage limits
// @Description Retrieves all usage limits with optional filtering by plan name, active status, and pagination
// @Tags Usage Limits
// @Produce json
// @Param plan_name query string false "Filter by plan name" example(premium)
// @Param is_active query bool false "Filter by active status" example(true)
// @Param limit query string false "Pagination limit" example(10)
// @Param offset query string false "Pagination offset" example(0)
// @Success 200 {array} model.UsageLimits "Successfully retrieved usage limits"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/usage-limits [get]
// @Example application/json Success Response [{"id":"507f1f77bcf86cd799439011","plan_name":"basic","teachers":5,"classrooms":10,"students_per_classroom":30,"question_banks":50,"questions":1000,"assignment_exports_per_day":10,"ai":{"independent_agent":100,"lumen_agent":50,"rag_agent":25,"rag_document_uploads":10},"created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z","is_active":true},{"id":"507f1f77bcf86cd799439012","plan_name":"premium","teachers":25,"classrooms":"unlimited","students_per_classroom":50,"question_banks":"unlimited","questions":"unlimited","assignment_exports_per_day":"unlimited","ai":{"independent_agent":500,"lumen_agent":300,"rag_agent":150,"rag_document_uploads":100},"created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z","is_active":true}]
func GetAllUsageLimits(c *gin.Context) {
	var filters struct {
		PlanName string `form:"plan_name"`
		IsActive *bool  `form:"is_active"`
		Limit    string `form:"limit"`
		Offset   string `form:"offset"`
	}

	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to model filter
	modelFilter := model.UsageLimitsFilter{}
	modelFilter.PlanName = filters.PlanName
	modelFilter.IsActive = filters.IsActive

	// If model.UsageLimitsFilter expects Limit and Offset as int, convert them
	// Otherwise, assign directly if they are string
	if filters.Limit != "" {
		modelFilter.Limit = filters.Limit // assign as string if model expects string
	}

	if filters.Offset != "" {
		modelFilter.Offset = filters.Offset // assign as string if model expects string
	}

	usageLimitsList, err := usageLimitsService.GetAllUsageLimits(modelFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage limits"})
		return
	}

	c.JSON(http.StatusOK, usageLimitsList)
}

// UpdateUsageLimits godoc
// @Summary Update usage limits
// @Description Updates existing usage limits completely. All fields will be replaced with new values.
// @Tags Usage Limits
// @Accept json
// @Produce json
// @Param id path string true "Usage Limits ID"
// @Param usageLimits body service.UpdateUsageLimitsRequest true "Updated usage limits data"
// @Success 200 {object} model.UsageLimits "Successfully updated usage limits"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/usage-limits/{id} [put]
// @Example application/json Update Request {"plan_name":"premium-updated","teachers":30,"classrooms":"unlimited","students_per_classroom":60,"question_banks":"unlimited","questions":"unlimited","assignment_exports_per_day":"unlimited","ai":{"independent_agent":600,"lumen_agent":400,"rag_agent":200,"rag_document_uploads":150}}
func UpdateUsageLimits(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateUsageLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usageLimits, err := usageLimitsService.UpdateUsageLimits(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usageLimits)
}

// PatchUsageLimits godoc
// @Summary Patch usage limits
// @Description Performs partial updates on usage limits. Only specified fields will be updated.
// @Tags Usage Limits
// @Accept json
// @Produce json
// @Param id path string true "Usage Limits ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} model.UsageLimits "Successfully updated usage limits"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/usage-limits/{id} [patch]
// @Example application/json Partial Update {"teachers":15,"classrooms":"unlimited","ai":{"independent_agent":200,"lumen_agent":100,"rag_agent":50,"rag_document_uploads":25},"is_active":true}
func PatchUsageLimits(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usageLimits, err := usageLimitsService.PatchUsageLimits(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usageLimits)
}

// DeleteUsageLimits godoc
// @Summary Delete usage limits
// @Description Deletes usage limits by ID
// @Tags Usage Limits
// @Param id path string true "Usage Limits ID"
// @Success 200 {object} map[string]interface{} "Successfully deleted usage limits"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/usage-limits/{id} [delete]
func DeleteUsageLimits(c *gin.Context) {
	id := c.Param("id")
	err := usageLimitsService.DeleteUsageLimits(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage limits deleted successfully"})
}

// SoftDeleteUsageLimits godoc
// @Summary Soft delete usage limits
// @Description Marks usage limits as inactive instead of deleting
// @Tags Usage Limits
// @Param id path string true "Usage Limits ID"
// @Success 200 {object} model.UsageLimits "Successfully deactivated usage limits"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/usage-limits/{id}/deactivate [post]
func SoftDeleteUsageLimits(c *gin.Context) {
	id := c.Param("id")
	usageLimits, err := usageLimitsService.SoftDeleteUsageLimits(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usageLimits)
}

// GetUsageLimitsStats godoc
// @Summary Get usage limits statistics
// @Description Returns statistical information about usage limits including total counts and active/inactive breakdown
// @Tags Usage Limits
// @Produce json
// @Success 200 {object} map[string]interface{} "Usage limits statistics"
// @Failure 500 {object} map[string]interface{} "Failed to fetch statistics"
// @Router /api/v1/usage-limits/stats [get]
// @Example application/json Statistics Response {"total_usage_limits":15,"active_usage_limits":12,"inactive_usage_limits":3,"plans_breakdown":{"basic":5,"premium":7,"enterprise":3}}
func GetUsageLimitsStats(c *gin.Context) {
	stats, err := usageLimitsService.GetUsageLimitsStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage limits statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// InitializeDefaultUsageLimits godoc
// @Summary Initialize default usage limits
// @Description Creates default usage limits for common plans (basic, premium, enterprise). Admin endpoint.
// @Tags Usage Limits
// @Produce json
// @Success 200 {object} map[string]interface{} "Default usage limits initialized"
// @Failure 500 {object} map[string]interface{} "Failed to initialize default usage limits"
// @Router /api/v1/admin/usage-limits/initialize-defaults [post]
// @Example application/json Success Response {"message":"Default usage limits initialized successfully","created_plans":["basic","premium","enterprise"]}
func InitializeDefaultUsageLimits(c *gin.Context) {
	err := usageLimitsService.InitializeDefaultUsageLimits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize default usage limits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default usage limits initialized successfully"})
}

// CheckUserUsageAgainstLimits godoc
// @Summary Check user usage against limits
// @Description Checks if user's current usage exceeds their plan limits and returns detailed comparison
// @Tags Usage Limits
// @Produce json
// @Param userId path string true "User ID" example(507f1f77bcf86cd799439013)
// @Param planName query string true "Plan Name" example(premium)
// @Success 200 {object} map[string]interface{} "Usage comparison results"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/usage-limits/check/{userId} [get]
// @Example application/json Success Response {"plan_name":"premium","limits":{"teachers":25,"classrooms":"unlimited","students_per_classroom":50,"question_banks":"unlimited","questions":"unlimited","assignment_exports_per_day":"unlimited","ai":{"independent_agent":500,"lumen_agent":300,"rag_agent":150,"rag_document_uploads":100}},"usage":{"teachers_used":15,"classrooms_used":45,"question_banks_used":120,"questions_used":5000,"assignment_exports_today":25,"ai_independent_agent_used":350,"ai_lumen_agent_used":200,"ai_rag_agent_used":100,"ai_rag_documents_uploaded":75},"within_limits":true,"exceeded_limits":[]}
func CheckUserUsageAgainstLimits(c *gin.Context) {
	userID := c.Param("userId")
	planName := c.Query("planName")

	if planName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plan name is required"})
		return
	}

	result, err := usageLimitsService.CheckUserUsageAgainstLimits(userID, planName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
