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

func SaveVariable(v model.Variable) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.VariableCollection).InsertOne(ctx, v)
	return err
}

func GetVariableByID(id string) (*model.Variable, error) {
	ctx := context.Background()
	var v model.Variable
	err := db.GetCollection(db.VariableCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func DeleteVariable(id string) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.VariableCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllVariables(filters map[string]string) ([]model.Variable, error) {
	ctx := context.Background()

	// Set up pagination options
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

	cursor, err := db.GetCollection(db.VariableCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Variable
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.Variable, 0), nil
	}
	return results, nil
}

func PatchVariable(id string, updates map[string]interface{}) (*model.Variable, error) {
	ctx := context.Background()
	updates["updatedAt"] = time.Now()

	// Update the document
	_, err := db.GetCollection(db.VariableCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Fetch the updated document
	var updated model.Variable
	err = db.GetCollection(db.VariableCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func SaveBulkVariables(variables []model.Variable) error {
	ctx := context.Background()

	// Convert variables to interface slice for bulk insert
	docs := make([]interface{}, len(variables))
	for i, v := range variables {
		docs[i] = v
	}

	_, err := db.GetCollection(db.VariableCollection).InsertMany(ctx, docs)
	return err
}
