package model

import "time"

// UsageUnit defines the time period for a feature's usage limit.
type UsageUnit string

const (
	// UnitDaily indicates a limit that resets every day at midnight.
	UnitDaily UsageUnit = "daily"
	// UnitMonthly indicates a limit that resets with the subscription billing cycle.
	UnitMonthly UsageUnit = "monthly"
	// UnitAbsolute indicates a total limit that does not reset.
	UnitAbsolute UsageUnit = "absolute"
)

// FeatureUsage tracks the consumption of a single feature.
// This struct is intended to be embedded within the main Usage document.
type FeatureUsage struct {
	Value     int64     `bson:"value" json:"value"`
	Limit     int64     `bson:"limit" json:"limit"` // A value of -1 can represent "unlimited"
	Unit      UsageUnit `bson:"unit" json:"unit"`
	LastReset time.Time `bson:"last_reset" json:"last_reset"`
}

// Usage tracks a user's feature consumption against their subscription plan.
// It's recommended to have one document of this type per user.
type Usage struct {
	ID        string                   `bson:"_id,omitempty" json:"id"`
	UserID    string                   `bson:"user_id" json:"user_id" validate:"required"`
	LookupKey string                   `bson:"lookup_key" json:"lookup_key" validate:"required"`
	Features  map[string]*FeatureUsage `bson:"features" json:"features"`
	CreatedAt time.Time                `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time                `bson:"updated_at" json:"updated_at"`
}

// NewUsage creates a new usage tracker for a user with an initial set of features.
func NewUsage(userID, lookupKey string, features map[string]*FeatureUsage) *Usage {
	now := time.Now()
	return &Usage{
		UserID:    userID,
		LookupKey: lookupKey,
		Features:  features,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ShouldReset checks if the usage for a feature should be reset.
// It compares the feature's unit and last reset time against the current time
// or the subscription's cycle start date.
func (fu *FeatureUsage) ShouldReset(subscriptionPeriodStart time.Time) bool {
	now := time.Now()
	switch fu.Unit {
	case UnitDaily:
		// Check if the last reset was before the start of today.
		year, month, day := now.Date()
		startOfToday := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		return fu.LastReset.Before(startOfToday)
	case UnitMonthly:
		// Check if the last reset was before the current subscription period started.
		return fu.LastReset.Before(subscriptionPeriodStart)
	case UnitAbsolute:
		// Absolute values never reset.
		return false
	default:
		return false
	}
}

// Increment attempts to increment the usage value for a given feature.
// It returns true if the increment was successful (i.e., the user is within their limit).
// It automatically handles resetting the value if a new usage cycle (day/month) has begun.
func (u *Usage) Increment(featureName string, subscription *Subscription) bool {
	feature, ok := u.Features[featureName]
	if !ok {
		// If the feature is not being tracked for this user, deny usage.
		// Alternatively, you could dynamically add it here if you have a central
		// way to look up its default limit based on the user's lookupKey.
		return false
	}

	// Check if the usage cycle needs to be reset before incrementing.
	if feature.ShouldReset(subscription.CurrentPeriodStart) {
		feature.Value = 0
		feature.LastReset = time.Now()
	}

	// Check if the user is within their limit (-1 means unlimited).
	if feature.Limit != -1 && feature.Value >= feature.Limit {
		return false // Limit reached.
	}

	feature.Value++
	u.UpdatedAt = time.Now()
	return true
}

// SetUsage sets a specific value for a feature. This is useful for features
// with absolute counts, like "total_question_banks" or "classroom_count".
func (u *Usage) SetUsage(featureName string, value int64) bool {
	feature, ok := u.Features[featureName]
	if !ok {
		return false // Feature not found.
	}

	// For absolute values, we just check against the limit.
	if feature.Unit == UnitAbsolute {
		if feature.Limit != -1 && value > feature.Limit {
			return false // Cannot set value over the limit.
		}
	}

	feature.Value = value
	u.UpdatedAt = time.Now()
	return true
}
