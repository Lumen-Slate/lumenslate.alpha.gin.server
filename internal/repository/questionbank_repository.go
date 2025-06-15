package repository

import (
	"context"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveQuestionBank(q model.QuestionBank) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.QuestionBankCollection).InsertOne(ctx, q)
	return err
}

func GetQuestionBankByID(id string) (*model.QuestionBank, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var q model.QuestionBank
	err := db.GetCollection(db.QuestionBankCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&q)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func DeleteQuestionBank(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.QuestionBankCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllQuestionBanks(filters map[string]string) ([]model.QuestionBank, error) {
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
	filter := bson.M{}
	if topic, ok := filters["topic"]; ok && topic != "" {
		filter["topic"] = topic
	}
	if name, ok := filters["name"]; ok && name != "" {
		filter["name"] = name
	}
	if teacherId, ok := filters["teacherId"]; ok && teacherId != "" {
		filter["teacherId"] = teacherId
	}
	if tags, ok := filters["tags"]; ok && tags != "" {
		tagList := strings.Split(tags, ",")
		filter["tags"] = bson.M{"$all": tagList}
	}

	cursor, err := db.GetCollection(db.QuestionBankCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.QuestionBank
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if results == nil {
		results = make([]model.QuestionBank, 0)
	}
	return results, nil
}

func PatchQuestionBank(id string, updates map[string]interface{}) (*model.QuestionBank, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First update the document
	_, err := db.GetCollection(db.QuestionBankCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.QuestionBank
	err = db.GetCollection(db.QuestionBankCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
