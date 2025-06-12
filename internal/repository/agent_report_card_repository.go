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

// GetAllAgentReportCards retrieves all agent report cards with optional filtering
func GetAllAgentReportCards(filters map[string]string) ([]model.AgentReportCard, error) {
	collection := db.GetCollection("report_cards")

	// Build MongoDB filter
	filter := bson.M{}

	if studentId, exists := filters["studentId"]; exists && studentId != "" {
		filter["reportCard.studentId"] = studentId
	}

	if userId, exists := filters["userId"]; exists && userId != "" {
		filter["userId"] = userId
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
		log.Printf("Error finding agent report cards: %v", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []model.AgentReportCard
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Printf("Error decoding agent report cards: %v", err)
		return nil, err
	}

	return results, nil
}

// GetAgentReportCardByID retrieves a specific agent report card by ID
func GetAgentReportCardByID(id string) (*model.AgentReportCard, error) {
	collection := db.GetCollection("report_cards")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid agent report card ID: %v", err)
	}

	var result model.AgentReportCard
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&result)
	if err != nil {
		log.Printf("Error finding agent report card by ID: %v", err)
		return nil, err
	}

	return &result, nil
}

// REMOVED: GetAgentReportCardsByStudentID function - Now handled by gRPC microservice tools

// CreateAgentReportCard creates a new agent report card
func CreateAgentReportCard(reportCard model.AgentReportCard) (*model.AgentReportCard, error) {
	collection := db.GetCollection("report_cards")

	reportCard.CreatedAt = time.Now()
	reportCard.UpdatedAt = time.Now()

	insertResult, err := collection.InsertOne(context.TODO(), reportCard)
	if err != nil {
		log.Printf("Error creating agent report card: %v", err)
		return nil, err
	}

	reportCard.ID = insertResult.InsertedID.(primitive.ObjectID)
	return &reportCard, nil
}

// UpdateAgentReportCard updates an existing agent report card
func UpdateAgentReportCard(id string, updates map[string]interface{}) (*model.AgentReportCard, error) {
	collection := db.GetCollection("report_cards")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid agent report card ID: %v", err)
	}

	updates["updatedAt"] = time.Now()
	update := bson.M{"$set": updates}

	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objectId}, update)
	if err != nil {
		log.Printf("Error updating agent report card: %v", err)
		return nil, err
	}

	// Return updated document
	return GetAgentReportCardByID(id)
}

// DeleteAgentReportCard deletes an agent report card by ID
func DeleteAgentReportCard(id string) error {
	collection := db.GetCollection("report_cards")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid agent report card ID: %v", err)
	}

	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objectId})
	if err != nil {
		log.Printf("Error deleting agent report card: %v", err)
		return err
	}

	return nil
}
