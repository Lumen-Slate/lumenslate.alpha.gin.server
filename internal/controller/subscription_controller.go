package controller

import (
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"lumenslate/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var subscriptionService = service.NewSubscriptionService()

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Creates a new subscription for a user
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscription body service.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} model.Subscription "Successfully created subscription"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/subscriptions [post]
func CreateSubscription(c *gin.Context) {
	var req service.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := subscriptionService.CreateSubscription(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription godoc
// @Summary Get subscription by ID
// @Description Retrieves a specific subscription by its ID
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} model.Subscription "Successfully retrieved subscription"
// @Failure 404 {object} map[string]interface{} "Subscription not found"
// @Router /api/v1/subscriptions/{id} [get]
func GetSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.GetSubscriptionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// GetUserSubscription godoc
// @Summary Get active subscription for user
// @Description Retrieves the active subscription for a specific user
// @Tags Subscriptions
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.Subscription "Successfully retrieved user subscription"
// @Failure 404 {object} map[string]interface{} "No active subscription found for user"
// @Router /api/v1/subscriptions/user/{id} [get]
func GetUserSubscription(c *gin.Context) {
	userID := c.Param("id")
	subscription, err := subscriptionService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found for user"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// GetAllUserSubscriptions godoc
// @Summary Get all subscriptions for user
// @Description Retrieves all subscriptions (active and inactive) for a specific user
// @Tags Subscriptions
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {array} model.Subscription "Successfully retrieved user subscriptions"
// @Failure 500 {object} map[string]interface{} "Failed to fetch user subscriptions"
// @Router /api/v1/subscriptions/user/{id}/all [get]
func GetAllUserSubscriptions(c *gin.Context) {
	userID := c.Param("id")
	subscriptions, err := subscriptionService.GetAllUserSubscriptions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user subscriptions"})
		return
	}
	c.JSON(http.StatusOK, subscriptions)
}

// UpdateSubscription godoc
// @Summary Update subscription
// @Description Updates an existing subscription
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body service.UpdateSubscriptionRequest true "Updated subscription data"
// @Success 200 {object} model.Subscription "Successfully updated subscription"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/subscriptions/{id} [put]
func UpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := subscriptionService.UpdateSubscription(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// CancelSubscription godoc
// @Summary Cancel subscription immediately
// @Description Immediately cancels an active subscription
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} model.Subscription "Successfully cancelled subscription"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/subscriptions/{id} [delete]
func CancelSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.CancelSubscription(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// ScheduleSubscriptionCancellation godoc
// @Summary Schedule subscription cancellation
// @Description Schedules a subscription for cancellation at the end of the current billing period
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} model.Subscription "Successfully scheduled cancellation"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/subscriptions/{id}/schedule-cancellation [post]
func ScheduleSubscriptionCancellation(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.ScheduleSubscriptionCancellation(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// ReactivateSubscription godoc
// @Summary Reactivate subscription
// @Description Reactivates a subscription that was scheduled for cancellation
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} model.Subscription "Successfully reactivated subscription"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/subscriptions/{id}/reactivate [post]
func ReactivateSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.ReactivateSubscription(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// RenewSubscription godoc
// @Summary Renew subscription
// @Description Renews a subscription for the next billing period
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param renewal body map[string]string true "Renewal data with new_period_end"
// @Success 200 {object} model.Subscription "Successfully renewed subscription"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/subscriptions/{id}/renew [post]
func RenewSubscription(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		NewPeriodEnd string `json:"new_period_end" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the time string (expected format: RFC3339)
	newPeriodEnd, err := time.Parse(time.RFC3339, req.NewPeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Expected RFC3339 format (e.g., 2023-12-31T23:59:59Z)"})
		return
	}

	subscription, err := subscriptionService.RenewSubscription(id, newPeriodEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// CheckUserSubscriptionStatus godoc
// @Summary Check user subscription status
// @Description Checks if a user has an active subscription
// @Tags Subscriptions
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User subscription status"
// @Failure 500 {object} map[string]interface{} "Failed to check subscription status"
// @Router /api/v1/subscriptions/user/{id}/status [get]
func CheckUserSubscriptionStatus(c *gin.Context) {
	userID := c.Param("id")
	isSubscribed, err := subscriptionService.IsUserSubscribed(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subscription status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"subscribed": isSubscribed,
	})
}

// GetSubscriptionsByStatus godoc
// @Summary Get subscriptions by status
// @Description Retrieves subscriptions filtered by their status
// @Tags Subscriptions
// @Produce json
// @Param status query string true "Subscription status" Enums(active, scheduled_to_cancel, cancelled, inactive)
// @Success 200 {array} model.Subscription "Successfully retrieved subscriptions"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Failed to fetch subscriptions"
// @Router /api/v1/subscriptions [get]
func GetSubscriptionsByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":              true,
		"scheduled_to_cancel": true,
		"cancelled":           true,
		"inactive":            true,
	}

	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	subscriptions, err := subscriptionService.GetSubscriptionsByStatus(model.SubscriptionStatus(status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// ProcessExpiredSubscriptions godoc
// @Summary Process expired subscriptions
// @Description Processes subscriptions that have expired and should be cancelled (admin endpoint)
// @Tags Subscriptions
// @Produce json
// @Success 200 {object} map[string]interface{} "Processing results"
// @Failure 500 {object} map[string]interface{} "Failed to process expired subscriptions"
// @Router /api/v1/admin/subscriptions/process-expired [post]
func ProcessExpiredSubscriptions(c *gin.Context) {
	processedSubscriptions, err := subscriptionService.ProcessExpiredSubscriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process expired subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"processed_count": len(processedSubscriptions),
		"subscriptions":   processedSubscriptions,
	})
}

// GetSubscriptionStats godoc
// @Summary Get subscription statistics
// @Description Returns statistical information about subscriptions
// @Tags Subscriptions
// @Produce json
// @Success 200 {object} map[string]interface{} "Subscription statistics"
// @Failure 500 {object} map[string]interface{} "Failed to fetch subscription statistics"
// @Router /api/v1/subscriptions/stats [get]
func GetSubscriptionStats(c *gin.Context) {
	stats, err := subscriptionService.GetSubscriptionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
