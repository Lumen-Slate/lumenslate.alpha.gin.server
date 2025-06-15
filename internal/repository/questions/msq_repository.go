package questions

import (
	"context"
	"lumenslate/internal/db"
	"lumenslate/internal/model/questions"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveMSQ(m questions.MSQ) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.MSQCollection).InsertOne(ctx, m)
	return err
}

func GetMSQByID(id string) (*questions.MSQ, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var m questions.MSQ
	err := db.GetCollection(db.MSQCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func DeleteMSQ(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.MSQCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllMSQs(filters map[string]string) ([]questions.MSQ, error) {
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
	if bankID, ok := filters["bankId"]; ok && bankID != "" {
		filter["bankId"] = bankID
	}

	cursor, err := db.GetCollection(db.MSQCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []questions.MSQ
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if results == nil {
		results = make([]questions.MSQ, 0)
	}
	return results, nil
}

func PatchMSQ(id string, updates map[string]interface{}) (*questions.MSQ, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First update the document
	_, err := db.GetCollection(db.MSQCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated questions.MSQ
	err = db.GetCollection(db.MSQCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func SaveBulkMSQs(msqs []questions.MSQ) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert slice to []interface{} for InsertMany
	documents := make([]interface{}, len(msqs))
	for i, m := range msqs {
		documents[i] = m
	}

	_, err := db.GetCollection(db.MSQCollection).InsertMany(ctx, documents)
	return err
}
