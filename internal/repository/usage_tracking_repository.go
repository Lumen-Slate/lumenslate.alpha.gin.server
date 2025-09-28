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

const UsageCollection = "usage"

// CreateUsage creates a new usage document for a user
func CreateUsage(usage *model.Usage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usage.CreatedAt = time.Now()
	usage.UpdatedAt = time.Now()

	_, err := db.GetCollection(UsageCollection).InsertOne(ctx, usage)
	return err
}

// GetUsageByUserID retrieves usage tracking for a user
func GetUsageByUserID(userID string) (*model.Usage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var usage model.Usage
	filter := bson.M{"user_id": userID}

	err := db.GetCollection(UsageCollection).FindOne(ctx, filter).Decode(&usage)
	if err != nil {
		return nil, err
	}

	return &usage, nil
}

// GetUsageByID retrieves usage tracking by ID
func GetUsageByID(id string) (*model.Usage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var usage model.Usage
	err := db.GetCollection(UsageCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&usage)
	if err != nil {
		return nil, err
	}

	return &usage, nil
}

// GetOrCreateUsage gets existing usage or creates a new one with default features
func GetOrCreateUsage(userID, lookupKey string, defaultFeatures map[string]*model.FeatureUsage) (*model.Usage, error) {
	usage, err := GetUsageByUserID(userID)
	if err != nil {
		// If not found, create a new one
		if err.Error() == "mongo: no documents in result" {
			newUsage := model.NewUsage(userID, lookupKey, defaultFeatures)
			if createErr := CreateUsage(newUsage); createErr != nil {
				return nil, createErr
			}
			return newUsage, nil
		}
		return nil, err
	}

	return usage, nil
}

// UpdateUsage updates an existing usage document
func UpdateUsage(usage *model.Usage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usage.UpdatedAt = time.Now()

	filter := bson.M{"_id": usage.ID}
	update := bson.M{"$set": usage}

	_, err := db.GetCollection(UsageCollection).UpdateOne(ctx, filter, update)
	return err
}

// UpsertUsage creates or updates a usage document
func UpsertUsage(usage *model.Usage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usage.UpdatedAt = time.Now()

	filter := bson.M{"user_id": usage.UserID}
	update := bson.M{
		"$set": usage,
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := db.GetCollection(UsageCollection).UpdateOne(ctx, filter, update, opts)
	return err
}

// IncrementFeatureUsage increments usage for a specific feature
func IncrementFeatureUsage(userID, featureName string, subscription *model.Subscription) (bool, error) {
	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, err
	}

	success := usage.Increment(featureName, subscription)
	if success {
		err = UpdateUsage(usage)
		if err != nil {
			return false, err
		}
	}

	return success, nil
}

// SetFeatureUsage sets a specific value for a feature
func SetFeatureUsage(userID, featureName string, value int64) (bool, error) {
	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, err
	}

	success := usage.SetUsage(featureName, value)
	if success {
		err = UpdateUsage(usage)
		if err != nil {
			return false, err
		}
	}

	return success, nil
}

// GetAllUsage retrieves all usage records with optional filters
func GetAllUsage(filters map[string]string) ([]model.Usage, error) {
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
	findOptions.SetSort(bson.M{"updated_at": -1})

	// Build filter
	filter := bson.M{}
	if userID, ok := filters["user_id"]; ok && userID != "" {
		filter["user_id"] = userID
	}
	if lookupKey, ok := filters["lookup_key"]; ok && lookupKey != "" {
		filter["lookup_key"] = lookupKey
	}

	cursor, err := db.GetCollection(UsageCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Usage
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.Usage, 0), nil
	}
	return results, nil
}

// DeleteUsage deletes a usage record
func DeleteUsage(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.GetCollection(UsageCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// GetUsageByLookupKey retrieves all usage records for a specific lookup key (subscription plan)
func GetUsageByLookupKey(lookupKey string) ([]model.Usage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"lookup_key": lookupKey}
	cursor, err := db.GetCollection(UsageCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Usage
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// ResetFeatureUsage resets usage for a specific feature across all users with a lookup key
func ResetFeatureUsage(lookupKey, featureName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"lookup_key": lookupKey}
	update := bson.M{
		"$set": bson.M{
			"features." + featureName + ".value":      0,
			"features." + featureName + ".last_reset": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err := db.GetCollection(UsageCollection).UpdateMany(ctx, filter, update)
	return err
}
