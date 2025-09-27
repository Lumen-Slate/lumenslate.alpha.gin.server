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

// CreateSubscription creates a new subscription
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

// GetSubscription retrieves a subscription by ID
func GetSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.GetSubscriptionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// GetUserSubscription retrieves the active subscription for a user
func GetUserSubscription(c *gin.Context) {
	userID := c.Param("userId")
	subscription, err := subscriptionService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found for user"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// GetAllUserSubscriptions retrieves all subscriptions for a user
func GetAllUserSubscriptions(c *gin.Context) {
	userID := c.Param("userId")
	subscriptions, err := subscriptionService.GetAllUserSubscriptions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user subscriptions"})
		return
	}
	c.JSON(http.StatusOK, subscriptions)
}

// UpdateSubscription updates an existing subscription
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

// CancelSubscription immediately cancels a subscription
func CancelSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.CancelSubscription(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// ScheduleSubscriptionCancellation schedules a subscription for cancellation at period end
func ScheduleSubscriptionCancellation(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.ScheduleSubscriptionCancellation(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// ReactivateSubscription reactivates a scheduled-to-cancel subscription
func ReactivateSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := subscriptionService.ReactivateSubscription(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// RenewSubscription renews a subscription for the next period
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

// CheckUserSubscriptionStatus checks if a user has an active subscription
func CheckUserSubscriptionStatus(c *gin.Context) {
	userID := c.Param("userId")
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

// GetSubscriptionsByStatus retrieves subscriptions by status
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

// ProcessExpiredSubscriptions processes subscriptions that should be cancelled
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

// GetSubscriptionStats returns statistics about subscriptions
func GetSubscriptionStats(c *gin.Context) {
	stats, err := subscriptionService.GetSubscriptionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
