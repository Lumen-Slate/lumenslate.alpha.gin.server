package repository

import (
	"context"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const SubscriptionCollection = "subscriptions"

// SaveSubscription saves a new subscription to the database
func SaveSubscription(subscription model.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.GetCollection(SubscriptionCollection).InsertOne(ctx, subscription)
	return err
}

// GetSubscriptionByID retrieves a subscription by its ID
func GetSubscriptionByID(id string) (*model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var subscription model.Subscription
	err := db.GetCollection(SubscriptionCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// GetSubscriptionByUserID retrieves the active subscription for a user
func GetSubscriptionByUserID(userID string) (*model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var subscription model.Subscription

	// Find the most recent active subscription for the user
	filter := bson.M{
		"user_id": userID,
		"status": bson.M{
			"$in": []model.SubscriptionStatus{
				model.StatusActive,
				model.StatusScheduledToCancel,
			},
		},
	}

	opts := options.FindOne().SetSort(bson.M{"created_at": -1})
	err := db.GetCollection(SubscriptionCollection).FindOne(ctx, filter, opts).Decode(&subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// GetAllSubscriptionsByUserID retrieves all subscriptions for a user
func GetAllSubscriptionsByUserID(userID string) ([]model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := db.GetCollection(SubscriptionCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscriptions []model.Subscription
	if err = cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}
	return subscriptions, nil
}

// UpdateSubscription updates an existing subscription
func UpdateSubscription(id string, updates map[string]interface{}) (*model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	// Update the document
	_, err := db.GetCollection(SubscriptionCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Fetch the updated document
	var updated model.Subscription
	err = db.GetCollection(SubscriptionCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// CancelSubscription cancels a subscription immediately
func CancelSubscription(id string) (*model.Subscription, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       model.StatusCancelled,
		"cancelled_at": now,
		"updated_at":   now,
	}
	return UpdateSubscription(id, updates)
}

// ScheduleSubscriptionCancellation schedules a subscription for cancellation at period end
func ScheduleSubscriptionCancellation(id string) (*model.Subscription, error) {
	// First get the subscription to access current_period_end
	subscription, err := GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"status":               model.StatusScheduledToCancel,
		"cancel_at_period_end": true,
		"cancel_at":            subscription.CurrentPeriodEnd,
		"updated_at":           time.Now(),
	}
	return UpdateSubscription(id, updates)
}

// GetAllSubscriptions retrieves all subscriptions with optional filters
func GetAllSubscriptions(filters map[string]string) ([]model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	// Handle pagination
	limit := int64(10)
	offset := int64(0)
	if l, err := strconv.Atoi(filters["limit"]); err == nil {
		limit = int64(l)
	}
	if o, err := strconv.Atoi(filters["offset"]); err == nil {
		offset = int64(o)
	}
	findOptions.SetLimit(limit)
	findOptions.SetSkip(offset)
	findOptions.SetSort(bson.M{"created_at": -1})

	// Build filter
	filter := bson.M{}
	if userID, ok := filters["user_id"]; ok && userID != "" {
		filter["user_id"] = userID
	}
	if status, ok := filters["status"]; ok && status != "" {
		filter["status"] = status
	}
	if lookupKey, ok := filters["lookup_key"]; ok && lookupKey != "" {
		filter["lookup_key"] = lookupKey
	}

	cursor, err := db.GetCollection(SubscriptionCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Subscription
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.Subscription, 0), nil
	}
	return results, nil
}

// GetExpiredSubscriptions retrieves subscriptions that should be cancelled
func GetExpiredSubscriptions() ([]model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	filter := bson.M{
		"status":    model.StatusScheduledToCancel,
		"cancel_at": bson.M{"$lte": now},
	}

	cursor, err := db.GetCollection(SubscriptionCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscriptions []model.Subscription
	if err = cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}
	return subscriptions, nil
}

// DeleteSubscription deletes a subscription from the database
func DeleteSubscription(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.GetCollection(SubscriptionCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}
