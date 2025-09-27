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

const UsageTrackingCollection = "usage_tracking"

// GetOrCreateUsageTracking gets or creates a usage tracking record for a user and period
func GetOrCreateUsageTracking(userID, period string) (*model.UsageTracking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var usage model.UsageTracking
	filter := bson.M{"user_id": userID, "period": period}

	err := db.GetCollection(UsageTrackingCollection).FindOne(ctx, filter).Decode(&usage)
	if err != nil {
		// If not found, create a new one
		if err.Error() == "mongo: no documents in result" {
			newUsage := model.NewUsageTracking(userID, period)
			if saveErr := SaveUsageTracking(*newUsage); saveErr != nil {
				return nil, saveErr
			}
			return newUsage, nil
		}
		return nil, err
	}

	return &usage, nil
}

// SaveUsageTracking saves a new usage tracking record to the database
func SaveUsageTracking(usage model.UsageTracking) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.GetCollection(UsageTrackingCollection).InsertOne(ctx, usage)
	return err
}

// GetUsageTrackingByID retrieves usage tracking by ID
func GetUsageTrackingByID(id string) (*model.UsageTracking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var usage model.UsageTracking
	err := db.GetCollection(UsageTrackingCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&usage)
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

// GetCurrentUsageTracking retrieves the current period's usage tracking for a user
func GetCurrentUsageTracking(userID string) (*model.UsageTracking, error) {
	currentPeriod := model.GetCurrentPeriod()
	return GetOrCreateUsageTracking(userID, currentPeriod)
}

// IncrementQuestionBanks increments the question banks counter for a user
func IncrementQuestionBanks(userID string, count int64) error {
	currentPeriod := model.GetCurrentPeriod()
	return incrementUsageField(userID, currentPeriod, "total_question_banks", count)
}

// IncrementQuestions increments the questions counter for a user
func IncrementQuestions(userID string, count int64) error {
	currentPeriod := model.GetCurrentPeriod()
	return incrementUsageField(userID, currentPeriod, "total_questions", count)
}

// IncrementIAUses increments the IA uses counter for a user
func IncrementIAUses(userID string, count int64) error {
	currentPeriod := model.GetCurrentPeriod()
	return incrementUsageField(userID, currentPeriod, "total_ia_uses", count)
}

// IncrementLumenAgentUses increments the Lumen Agent uses counter for a user
func IncrementLumenAgentUses(userID string, count int64) error {
	currentPeriod := model.GetCurrentPeriod()
	return incrementUsageField(userID, currentPeriod, "total_lumen_agent_uses", count)
}

// IncrementRAAgentUses increments the RA Agent uses counter for a user
func IncrementRAAgentUses(userID string, count int64) error {
	currentPeriod := model.GetCurrentPeriod()
	return incrementUsageField(userID, currentPeriod, "total_ra_agent_uses", count)
}

// IncrementRecapClasses increments the recap classes counter for a user
func IncrementRecapClasses(userID string, count int64) error {
	currentPeriod := model.GetCurrentPeriod()
	return incrementUsageField(userID, currentPeriod, "total_recap_classes", count)
}

// IncrementAssignmentExports increments the assignment exports counter for a user
func IncrementAssignmentExports(userID string, count int64) error {
	currentPeriod := model.GetCurrentPeriod()
	return incrementUsageField(userID, currentPeriod, "total_assignment_exports", count)
}

// incrementUsageField is a helper function to increment a specific usage field
func incrementUsageField(userID, period, field string, count int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID, "period": period}
	update := bson.M{
		"$inc": bson.M{field: count},
		"$set": bson.M{"updated_at": time.Now()},
		"$setOnInsert": bson.M{
			"user_id":    userID,
			"period":     period,
			"created_at": time.Now(),
		},
	}

	// Create indexes to ensure fast lookups
	opts := options.Update().SetUpsert(true)
	_, err := db.GetCollection(UsageTrackingCollection).UpdateOne(ctx, filter, update, opts)
	return err
}

// GetUsageTrackingByUserAndPeriod retrieves usage tracking for a specific user and period
func GetUsageTrackingByUserAndPeriod(userID, period string) (*model.UsageTracking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var usage model.UsageTracking
	filter := bson.M{"user_id": userID, "period": period}
	err := db.GetCollection(UsageTrackingCollection).FindOne(ctx, filter).Decode(&usage)
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

// GetAllUsageTrackingByUser retrieves all usage tracking records for a user
func GetAllUsageTrackingByUser(userID string) ([]model.UsageTracking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.M{"period": -1})

	cursor, err := db.GetCollection(UsageTrackingCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.UsageTracking
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetAllUsageTracking retrieves all usage tracking records with optional filters
func GetAllUsageTracking(filters map[string]string) ([]model.UsageTracking, error) {
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
	findOptions.SetSort(bson.M{"period": -1, "created_at": -1})

	// Build filter
	filter := bson.M{}
	if userID, ok := filters["user_id"]; ok && userID != "" {
		filter["user_id"] = userID
	}
	if period, ok := filters["period"]; ok && period != "" {
		filter["period"] = period
	}

	cursor, err := db.GetCollection(UsageTrackingCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.UsageTracking
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.UsageTracking, 0), nil
	}
	return results, nil
}

// GetUsageMetricsByUser retrieves current usage metrics for a user
func GetUsageMetricsByUser(userID string) (*model.UsageMetrics, error) {
	usage, err := GetCurrentUsageTracking(userID)
	if err != nil {
		return nil, err
	}
	return usage.ToMetrics(), nil
}

// GetAggregatedUsageByUser retrieves aggregated usage across all periods for a user
func GetAggregatedUsageByUser(userID string) (*model.UsageMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$group": bson.M{
			"_id":                      "$user_id",
			"total_question_banks":     bson.M{"$sum": "$total_question_banks"},
			"total_questions":          bson.M{"$sum": "$total_questions"},
			"total_ia_uses":            bson.M{"$sum": "$total_ia_uses"},
			"total_lumen_agent_uses":   bson.M{"$sum": "$total_lumen_agent_uses"},
			"total_ra_agent_uses":      bson.M{"$sum": "$total_ra_agent_uses"},
			"total_recap_classes":      bson.M{"$sum": "$total_recap_classes"},
			"total_assignment_exports": bson.M{"$sum": "$total_assignment_exports"},
			"last_updated":             bson.M{"$max": "$updated_at"},
		}},
	}

	cursor, err := db.GetCollection(UsageTrackingCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		UserID                 string    `bson:"_id"`
		TotalQuestionBanks     int64     `bson:"total_question_banks"`
		TotalQuestions         int64     `bson:"total_questions"`
		TotalIAUses            int64     `bson:"total_ia_uses"`
		TotalLumenAgentUses    int64     `bson:"total_lumen_agent_uses"`
		TotalRAAgentUses       int64     `bson:"total_ra_agent_uses"`
		TotalRecapClasses      int64     `bson:"total_recap_classes"`
		TotalAssignmentExports int64     `bson:"total_assignment_exports"`
		LastUpdated            time.Time `bson:"last_updated"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		return &model.UsageMetrics{
			UserID:                 result.UserID,
			TotalQuestionBanks:     result.TotalQuestionBanks,
			TotalQuestions:         result.TotalQuestions,
			TotalIAUses:            result.TotalIAUses,
			TotalLumenAgentUses:    result.TotalLumenAgentUses,
			TotalRAAgentUses:       result.TotalRAAgentUses,
			TotalRecapClasses:      result.TotalRecapClasses,
			TotalAssignmentExports: result.TotalAssignmentExports,
			CurrentPeriod:          model.GetCurrentPeriod(),
			LastUpdated:            result.LastUpdated,
		}, nil
	}

	// If no data found, return zero values
	return &model.UsageMetrics{
		UserID:                 userID,
		TotalQuestionBanks:     0,
		TotalQuestions:         0,
		TotalIAUses:            0,
		TotalLumenAgentUses:    0,
		TotalRAAgentUses:       0,
		TotalRecapClasses:      0,
		TotalAssignmentExports: 0,
		CurrentPeriod:          model.GetCurrentPeriod(),
		LastUpdated:            time.Now(),
	}, nil
}

// DeleteUsageTracking deletes a usage tracking record from the database
func DeleteUsageTracking(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.GetCollection(UsageTrackingCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}
