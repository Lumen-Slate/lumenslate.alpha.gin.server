package service

import (
	"errors"
	"lumenslate/internal/model"
	"lumenslate/internal/repository"
	"time"

	"github.com/google/uuid"
)

// SubscriptionService provides business logic for subscription management
type SubscriptionService struct{}

// NewSubscriptionService creates a new subscription service instance
func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{}
}

// CreateSubscriptionRequest represents the request to create a new subscription
type CreateSubscriptionRequest struct {
	UserID             string    `json:"user_id" validate:"required"`
	LookupKey          string    `json:"lookup_key" validate:"required"`
	Currency           string    `json:"currency" validate:"required"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
}

// CreateSubscription creates a new subscription for a user
func (s *SubscriptionService) CreateSubscription(req CreateSubscriptionRequest) (*model.Subscription, error) {
	// Check if user already has an active subscription
	existingSubscription, err := repository.GetSubscriptionByUserID(req.UserID)
	if err == nil && existingSubscription != nil && existingSubscription.IsActive() {
		return nil, errors.New("user already has an active subscription")
	}

	// Create new subscription
	subscription := model.NewSubscription(
		req.UserID,
		req.LookupKey,
		req.Currency,
		req.CurrentPeriodStart,
		req.CurrentPeriodEnd,
	)
	subscription.ID = uuid.New().String()

	// Save to database
	if err := repository.SaveSubscription(*subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetSubscriptionByID retrieves a subscription by its ID
func (s *SubscriptionService) GetSubscriptionByID(id string) (*model.Subscription, error) {
	return repository.GetSubscriptionByID(id)
}

// GetUserSubscription retrieves the active subscription for a user
func (s *SubscriptionService) GetUserSubscription(userID string) (*model.Subscription, error) {
	return repository.GetSubscriptionByUserID(userID)
}

// GetAllUserSubscriptions retrieves all subscriptions for a user
func (s *SubscriptionService) GetAllUserSubscriptions(userID string) ([]model.Subscription, error) {
	return repository.GetAllSubscriptionsByUserID(userID)
}

// UpdateSubscriptionRequest represents the request to update a subscription
type UpdateSubscriptionRequest struct {
	Status             *model.SubscriptionStatus `json:"status,omitempty"`
	LookupKey          *string                   `json:"lookup_key,omitempty"`
	Currency           *string                   `json:"currency,omitempty"`
	CurrentPeriodStart *time.Time                `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time                `json:"current_period_end,omitempty"`
}

// UpdateSubscription updates an existing subscription
func (s *SubscriptionService) UpdateSubscription(id string, req UpdateSubscriptionRequest) (*model.Subscription, error) {
	// Check if subscription exists
	existing, err := repository.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("subscription not found")
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.LookupKey != nil {
		updates["lookup_key"] = *req.LookupKey
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.CurrentPeriodStart != nil {
		updates["current_period_start"] = *req.CurrentPeriodStart
	}
	if req.CurrentPeriodEnd != nil {
		updates["current_period_end"] = *req.CurrentPeriodEnd
	}

	return repository.UpdateSubscription(id, updates)
}

// CancelSubscription immediately cancels a subscription
func (s *SubscriptionService) CancelSubscription(id string) (*model.Subscription, error) {
	// Check if subscription exists and is active
	existing, err := repository.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("subscription not found")
	}
	if !existing.IsActive() && existing.Status != model.StatusScheduledToCancel {
		return nil, errors.New("subscription is already cancelled or inactive")
	}

	return repository.CancelSubscription(id)
}

// ScheduleSubscriptionCancellation schedules a subscription for cancellation at the end of the current period
func (s *SubscriptionService) ScheduleSubscriptionCancellation(id string) (*model.Subscription, error) {
	// Check if subscription exists and is active
	existing, err := repository.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("subscription not found")
	}
	if !existing.IsActive() {
		return nil, errors.New("subscription is not active")
	}

	return repository.ScheduleSubscriptionCancellation(id)
}

// ReactivateSubscription reactivates a scheduled-to-cancel subscription
func (s *SubscriptionService) ReactivateSubscription(id string) (*model.Subscription, error) {
	// Check if subscription exists and is scheduled to cancel
	existing, err := repository.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("subscription not found")
	}
	if existing.Status != model.StatusScheduledToCancel {
		return nil, errors.New("subscription is not scheduled for cancellation")
	}

	updates := map[string]interface{}{
		"status":               model.StatusActive,
		"cancel_at_period_end": false,
		"cancel_at":            nil,
	}

	return repository.UpdateSubscription(id, updates)
}

// RenewSubscription renews a subscription for the next period
func (s *SubscriptionService) RenewSubscription(id string, newPeriodEnd time.Time) (*model.Subscription, error) {
	// Check if subscription exists
	existing, err := repository.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("subscription not found")
	}

	updates := map[string]interface{}{
		"current_period_start": existing.CurrentPeriodEnd,
		"current_period_end":   newPeriodEnd,
		"status":               model.StatusActive,
		"cancel_at_period_end": false,
		"cancel_at":            nil,
	}

	return repository.UpdateSubscription(id, updates)
}

// IsUserSubscribed checks if a user has an active subscription
func (s *SubscriptionService) IsUserSubscribed(userID string) (bool, error) {
	subscription, err := repository.GetSubscriptionByUserID(userID)
	if err != nil {
		return false, err
	}
	return subscription != nil && subscription.IsActive(), nil
}

// GetSubscriptionsByStatus retrieves subscriptions by status
func (s *SubscriptionService) GetSubscriptionsByStatus(status model.SubscriptionStatus) ([]model.Subscription, error) {
	filters := map[string]string{
		"status": string(status),
		"limit":  "100", // Default limit for batch operations
	}
	return repository.GetAllSubscriptions(filters)
}

// ProcessExpiredSubscriptions processes subscriptions that are scheduled to be cancelled and are past their cancel date
func (s *SubscriptionService) ProcessExpiredSubscriptions() ([]model.Subscription, error) {
	expiredSubscriptions, err := repository.GetExpiredSubscriptions()
	if err != nil {
		return nil, err
	}

	var processedSubscriptions []model.Subscription

	for _, subscription := range expiredSubscriptions {
		// Cancel the subscription
		cancelled, err := repository.CancelSubscription(subscription.ID)
		if err != nil {
			// Log error but continue processing others
			continue
		}
		processedSubscriptions = append(processedSubscriptions, *cancelled)
	}

	return processedSubscriptions, nil
}

// GetSubscriptionStats returns statistics about subscriptions
func (s *SubscriptionService) GetSubscriptionStats() (map[string]interface{}, error) {
	// This would typically involve aggregation queries
	// For now, returning basic counts by status
	stats := make(map[string]interface{})

	for _, status := range []model.SubscriptionStatus{
		model.StatusActive,
		model.StatusScheduledToCancel,
		model.StatusCancelled,
		model.StatusInactive,
	} {
		subscriptions, err := s.GetSubscriptionsByStatus(status)
		if err != nil {
			return nil, err
		}
		stats[string(status)] = len(subscriptions)
	}

	return stats, nil
}
