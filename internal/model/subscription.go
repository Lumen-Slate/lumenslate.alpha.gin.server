package model

import (
	"time"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	StatusActive            SubscriptionStatus = "active"
	StatusScheduledToCancel SubscriptionStatus = "scheduled_to_cancel"
	StatusCancelled         SubscriptionStatus = "cancelled"
	StatusInactive          SubscriptionStatus = "inactive"
)

// Subscription represents a user's subscription plan
type Subscription struct {
	ID                 string             `bson:"_id,omitempty" json:"id"`
	UserID             string             `bson:"user_id" json:"user_id" validate:"required"`
	LookupKey          string             `bson:"lookup_key" json:"lookup_key" validate:"required"`
	Status             SubscriptionStatus `bson:"status" json:"status" validate:"required"`
	Currency           string             `bson:"currency" json:"currency" validate:"required"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
	CurrentPeriodStart time.Time          `bson:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `bson:"current_period_end" json:"current_period_end"`
	CancelAtPeriodEnd  bool               `bson:"cancel_at_period_end" json:"cancel_at_period_end"`
	CancelAt           *time.Time         `bson:"cancel_at,omitempty" json:"cancel_at,omitempty"`
	CancelledAt        *time.Time         `bson:"cancelled_at,omitempty" json:"cancelled_at,omitempty"`
}

// NewSubscription creates a new subscription with default values
func NewSubscription(userID, lookupKey, currency string, currentPeriodStart, currentPeriodEnd time.Time) *Subscription {
	now := time.Now()
	return &Subscription{
		UserID:             userID,
		LookupKey:          lookupKey,
		Status:             StatusActive,
		Currency:           currency,
		CreatedAt:          now,
		UpdatedAt:          now,
		CurrentPeriodStart: currentPeriodStart,
		CurrentPeriodEnd:   currentPeriodEnd,
		CancelAtPeriodEnd:  false,
	}
}

// IsActive returns true if the subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive
}

// IsCancelled returns true if the subscription is cancelled
func (s *Subscription) IsCancelled() bool {
	return s.Status == StatusCancelled
}

// ScheduleForCancellation marks the subscription for cancellation at period end
func (s *Subscription) ScheduleForCancellation() {
	s.Status = StatusScheduledToCancel
	s.CancelAtPeriodEnd = true
	cancelTime := s.CurrentPeriodEnd
	s.CancelAt = &cancelTime
	s.UpdatedAt = time.Now()
}

// Cancel immediately cancels the subscription
func (s *Subscription) Cancel() {
	s.Status = StatusCancelled
	now := time.Now()
	s.CancelledAt = &now
	s.UpdatedAt = now
}
