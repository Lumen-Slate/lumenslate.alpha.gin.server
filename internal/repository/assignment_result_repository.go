package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"lumenslate/internal/db"
	"lumenslate/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetAllAssignmentResults retrieves all assignment results with optional filtering
func GetAllAssignmentResults(filters map[string]string) ([]model.AssignmentResult, error) {
	collection := db.GetCollection(db.AssignmentResultCollection)

	// Build MongoDB filter
	filter := bson.M{}

	if studentId, exists := filters["studentId"]; exists && studentId != "" {
		filter["studentId"] = studentId
	}

	if assignmentId, exists := filters["assignmentId"]; exists && assignmentId != "" {
		filter["assignmentId"] = assignmentId
	}

	// Pagination options
	opts := options.Find()
	if limit, exists := filters["limit"]; exists && limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			opts.SetLimit(int64(l))
		}
	}
	if offset, exists := filters["offset"]; exists && offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			opts.SetSkip(int64(o))
		}
	}

	// Sort by creation date (newest first)
	opts.SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Printf("Error finding assignment results: %v", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []model.AssignmentResult
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Printf("Error decoding assignment results: %v", err)
		return nil, err
	}

	return results, nil
}

// GetAssignmentResultByID retrieves a specific assignment result by ID
func GetAssignmentResultByID(id string) (*model.AssignmentResult, error) {
	collection := db.GetCollection(db.AssignmentResultCollection)

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid assignment result ID: %v", err)
	}

	var result model.AssignmentResult
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&result)
	if err != nil {
		log.Printf("Error finding assignment result by ID: %v", err)
		return nil, err
	}

	return &result, nil
}

// CreateAssignmentResult creates a new assignment result
func CreateAssignmentResult(result model.AssignmentResult) (*model.AssignmentResult, error) {
	collection := db.GetCollection(db.AssignmentResultCollection)

	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()

	insertResult, err := collection.InsertOne(context.TODO(), result)
	if err != nil {
		log.Printf("Error creating assignment result: %v", err)
		return nil, err
	}

	result.ID = insertResult.InsertedID.(primitive.ObjectID)
	return &result, nil
}

// UpdateAssignmentResult updates an existing assignment result
func UpdateAssignmentResult(id string, updates map[string]interface{}) (*model.AssignmentResult, error) {
	collection := db.GetCollection(db.AssignmentResultCollection)

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid assignment result ID: %v", err)
	}

	updates["updatedAt"] = time.Now()
	update := bson.M{"$set": updates}

	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objectId}, update)
	if err != nil {
		log.Printf("Error updating assignment result: %v", err)
		return nil, err
	}

	// Return updated document
	return GetAssignmentResultByID(id)
}

// DeleteAssignmentResult deletes an assignment result by ID
func DeleteAssignmentResult(id string) error {
	collection := db.GetCollection(db.AssignmentResultCollection)

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid assignment result ID: %v", err)
	}

	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objectId})
	if err != nil {
		log.Printf("Error deleting assignment result: %v", err)
		return err
	}

	return nil
}
