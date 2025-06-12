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

func SaveAssignment(a model.Assignment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.AssignmentCollection).InsertOne(ctx, a)
	return err
}

// REMOVED: GetAssignmentByID function - Now handled by gRPC microservice tools

func DeleteAssignment(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.AssignmentCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllAssignments() ([]model.Assignment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := db.GetCollection(db.AssignmentCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assignments []model.Assignment
	if err = cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func FilterAssignments(limitStr, offsetStr, points, due string) ([]model.Assignment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()

	// Handle pagination
	limit := int64(10)
	offset := int64(0)
	if l, err := strconv.Atoi(limitStr); err == nil {
		limit = int64(l)
	}
	if o, err := strconv.Atoi(offsetStr); err == nil {
		offset = int64(o)
	}
	findOptions.SetLimit(limit)
	findOptions.SetSkip(offset)

	// Build filter
	filter := bson.M{}
	if points != "" {
		if val, err := strconv.Atoi(points); err == nil {
			filter["points"] = val
		}
	}
	if due != "" {
		if t, err := time.Parse(time.RFC3339, due); err == nil {
			filter["dueDate"] = bson.M{"$gte": t}
		}
	}

	cursor, err := db.GetCollection(db.AssignmentCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Assignment
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = make([]model.Assignment, 0)
	}
	return results, nil
}

func PatchAssignment(id string, updates map[string]interface{}) (*model.Assignment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First update the document
	_, err := db.GetCollection(db.AssignmentCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.Assignment
	err = db.GetCollection(db.AssignmentCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
