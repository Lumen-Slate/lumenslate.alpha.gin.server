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

func SaveThread(t model.Thread) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.PostCollection).InsertOne(ctx, t)
	return err
}

func GetThreadByID(id string) (*model.Thread, error) {
	ctx := context.Background()
	var t model.Thread
	err := db.GetCollection(db.PostCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func DeleteThread(id string) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.PostCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllThreads(filters map[string]string) ([]model.Thread, error) {
	ctx := context.Background()
	filter := bson.M{}

	// Apply user filter if provided
	if userId, ok := filters["userId"]; ok && userId != "" {
		filter["userId"] = userId
	}

	// Set up pagination
	limit := 10
	offset := 0
	if l, err := strconv.Atoi(filters["limit"]); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(filters["offset"]); err == nil {
		offset = o
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection(db.PostCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Thread
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func PatchThread(id string, updates map[string]interface{}) (*model.Thread, error) {
	ctx := context.Background()
	updates["updatedAt"] = time.Now()

	// First update the document
	_, err := db.GetCollection(db.PostCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.Thread
	err = db.GetCollection(db.PostCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
