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

func SaveNAT(n questions.NAT) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.NATCollection).InsertOne(ctx, n)
	return err
}

func GetNATByID(id string) (*questions.NAT, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var n questions.NAT
	err := db.GetCollection(db.NATCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func DeleteNAT(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.NATCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllNATs(filters map[string]string) ([]questions.NAT, error) {
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

	cursor, err := db.GetCollection(db.NATCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []questions.NAT
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func PatchNAT(id string, updates map[string]interface{}) (*questions.NAT, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First update the document
	_, err := db.GetCollection(db.NATCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated questions.NAT
	err = db.GetCollection(db.NATCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func SaveBulkNATs(nats []questions.NAT) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert slice to []interface{} for InsertMany
	documents := make([]interface{}, len(nats))
	for i, n := range nats {
		documents[i] = n
	}

	_, err := db.GetCollection(db.NATCollection).InsertMany(ctx, documents)
	return err
}
