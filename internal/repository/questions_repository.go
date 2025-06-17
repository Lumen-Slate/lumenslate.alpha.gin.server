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

func SaveQuestion(question model.Questions) (*model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := db.GetCollection(db.QuestionsCollection).InsertOne(ctx, question)
	if err != nil {
		return nil, err
	}

	// Set the inserted ID
	if oid, ok := result.InsertedID.(string); ok {
		question.ID = oid
	}

	return &question, nil
}

func GetQuestionByID(id string) (*model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var question model.Questions
	err := db.GetCollection(db.QuestionsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&question)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func GetAllQuestions(filters map[string]string) ([]model.Questions, error) {
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

	// Build filter
	filter := bson.M{"isActive": true} // Only get active questions
	if subject, ok := filters["subject"]; ok && subject != "" {
		filter["subject"] = subject
	}
	if difficulty, ok := filters["difficulty"]; ok && difficulty != "" {
		filter["difficulty"] = difficulty
	}

	cursor, err := db.GetCollection(db.QuestionsCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Questions
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if results == nil {
		results = make([]model.Questions, 0)
	}
	return results, nil
}

func GetQuestionsBySubject(subject model.Subject) ([]model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"subject":  subject,
		"isActive": true,
	}

	cursor, err := db.GetCollection(db.QuestionsCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Questions
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = make([]model.Questions, 0)
	}
	return results, nil
}

func GetQuestionsBySubjectAndDifficulty(subject model.Subject, difficulty model.Difficulty) ([]model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"subject":    subject,
		"difficulty": difficulty,
		"isActive":   true,
	}

	cursor, err := db.GetCollection(db.QuestionsCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Questions
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = make([]model.Questions, 0)
	}
	return results, nil
}

func CountQuestionsBySubject(subject model.Subject) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"subject":  subject,
		"isActive": true,
	}

	return db.GetCollection(db.QuestionsCollection).CountDocuments(ctx, filter)
}

func DeleteQuestion(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.QuestionsCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func UpdateQuestion(id string, updates map[string]interface{}) (*model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// First update the document
	_, err := db.GetCollection(db.QuestionsCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.Questions
	err = db.GetCollection(db.QuestionsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func SaveBulkQuestions(questions []model.Questions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert slice to []interface{} for InsertMany
	documents := make([]interface{}, len(questions))
	for i, q := range questions {
		documents[i] = q
	}

	_, err := db.GetCollection(db.QuestionsCollection).InsertMany(ctx, documents)
	return err
}
