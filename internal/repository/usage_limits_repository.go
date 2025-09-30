package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"lumenslate/internal/db"
	"lumenslate/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const UsageLimitsCollection = "usage_limits"

// CreateUsageLimits creates a new usage limits document
func CreateUsageLimits(usageLimits model.UsageLimits) (*model.UsageLimits, error) {
	collection := db.GetCollection(UsageLimitsCollection)

	usageLimits.ID = primitive.NewObjectID()
	usageLimits.CreatedAt = time.Now()
	usageLimits.UpdatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), usageLimits)
	if err != nil {
		return nil, err
	}

	return &usageLimits, nil
}

// GetUsageLimitsByID retrieves usage limits by ID
func GetUsageLimitsByID(id string) (*model.UsageLimits, error) {
	collection := db.GetCollection(UsageLimitsCollection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid usage limits ID")
	}

	var usageLimits model.UsageLimits
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&usageLimits)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("usage limits not found")
		}
		return nil, err
	}

	return &usageLimits, nil
}

// GetUsageLimitsByPlanName retrieves usage limits by plan name
func GetUsageLimitsByPlanName(planName string) (*model.UsageLimits, error) {
	collection := db.GetCollection(UsageLimitsCollection)

	var usageLimits model.UsageLimits
	err := collection.FindOne(context.Background(), bson.M{
		"plan_name": planName,
		"is_active": true,
	}).Decode(&usageLimits)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("usage limits not found for plan")
		}
		return nil, err
	}

	return &usageLimits, nil
}

// GetAllUsageLimits retrieves all usage limits with optional filters
func GetAllUsageLimits(filters model.UsageLimitsFilter) ([]model.UsageLimits, error) {
	collection := db.GetCollection(UsageLimitsCollection)

	// Build filter
	filter := bson.M{}

	if filters.PlanName != "" {
		filter["plan_name"] = bson.M{"$regex": filters.PlanName, "$options": "i"}
	}

	if filters.IsActive != nil {
		filter["is_active"] = *filters.IsActive
	}

	// Set up pagination
	findOptions := options.Find()

	if filters.Limit != "" {
		if limit, err := strconv.Atoi(filters.Limit); err == nil && limit > 0 {
			findOptions.SetLimit(int64(limit))
		}
	}

	if filters.Offset != "" {
		if offset, err := strconv.Atoi(filters.Offset); err == nil && offset > 0 {
			findOptions.SetSkip(int64(offset))
		}
	}

	// Sort by creation date (newest first)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var usageLimitsList []model.UsageLimits
	if err = cursor.All(context.Background(), &usageLimitsList); err != nil {
		return nil, err
	}

	return usageLimitsList, nil
}

// UpdateUsageLimits updates an existing usage limits document
func UpdateUsageLimits(id string, updates map[string]interface{}) (*model.UsageLimits, error) {
	collection := db.GetCollection(UsageLimitsCollection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid usage limits ID")
	}

	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	updateDoc := bson.M{"$set": updates}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		updateDoc,
	)
	if err != nil {
		return nil, err
	}

	// Return updated document
	return GetUsageLimitsByID(id)
}

// PatchUsageLimits performs a partial update on usage limits
func PatchUsageLimits(id string, updates map[string]interface{}) (*model.UsageLimits, error) {
	return UpdateUsageLimits(id, updates)
}

// DeleteUsageLimits deletes a usage limits document
func DeleteUsageLimits(id string) error {
	collection := db.GetCollection(UsageLimitsCollection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid usage limits ID")
	}

	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("usage limits not found")
	}

	return nil
}

// SoftDeleteUsageLimits marks usage limits as inactive instead of deleting
func SoftDeleteUsageLimits(id string) (*model.UsageLimits, error) {
	return UpdateUsageLimits(id, map[string]interface{}{
		"is_active": false,
	})
}

// CreateDefaultUsageLimits creates default usage limits for common plans
func CreateDefaultUsageLimits() error {
	// Define default plans
	defaultPlans := []model.UsageLimits{
		{
			PlanName:                "basic",
			Teachers:                5,
			Classrooms:              10,
			StudentsPerClassroom:    30,
			QuestionBanks:           50,
			Questions:               1000,
			AssignmentExportsPerDay: 10,
			AI: model.AILimits{
				IndependentAgent:   100,
				LumenAgent:         50,
				RAGAgent:           25,
				RAGDocumentUploads: 10,
			},
			IsActive: true,
		},
		{
			PlanName:                "premium",
			Teachers:                25,
			Classrooms:              "unlimited",
			StudentsPerClassroom:    50,
			QuestionBanks:           "unlimited",
			Questions:               "unlimited",
			AssignmentExportsPerDay: "unlimited",
			AI: model.AILimits{
				IndependentAgent:   500,
				LumenAgent:         300,
				RAGAgent:           150,
				RAGDocumentUploads: 100,
			},
			IsActive: true,
		},
		{
			PlanName:                "enterprise",
			Teachers:                "custom",
			Classrooms:              "unlimited",
			StudentsPerClassroom:    "custom",
			QuestionBanks:           "unlimited",
			Questions:               "unlimited",
			AssignmentExportsPerDay: "unlimited",
			AI: model.AILimits{
				IndependentAgent:   "unlimited",
				LumenAgent:         "unlimited",
				RAGAgent:           "unlimited",
				RAGDocumentUploads: "unlimited",
			},
			IsActive: true,
		},
	}

	for _, plan := range defaultPlans {
		// Check if plan already exists
		existing, _ := GetUsageLimitsByPlanName(plan.PlanName)
		if existing == nil {
			_, err := CreateUsageLimits(plan)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetUsageLimitsStats returns statistics about usage limits
func GetUsageLimitsStats() (map[string]interface{}, error) {
	collection := db.GetCollection(UsageLimitsCollection)

	// Count total usage limits
	totalCount, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	// Count active usage limits
	activeCount, err := collection.CountDocuments(context.Background(), bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}

	// Count inactive usage limits
	inactiveCount, err := collection.CountDocuments(context.Background(), bson.M{"is_active": false})
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_usage_limits":    totalCount,
		"active_usage_limits":   activeCount,
		"inactive_usage_limits": inactiveCount,
	}

	return stats, nil
}
