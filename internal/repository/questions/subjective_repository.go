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

func SaveSubjective(s questions.Subjective) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.SubjectiveCollection).InsertOne(ctx, s)
	return err
}

func GetSubjectiveByID(id string) (*questions.Subjective, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var s questions.Subjective
	err := db.GetCollection(db.SubjectiveCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func DeleteSubjective(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.SubjectiveCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllSubjectives(filters map[string]string) ([]questions.Subjective, error) {
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

	cursor, err := db.GetCollection(db.SubjectiveCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []questions.Subjective
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if results == nil {
		results = make([]questions.Subjective, 0)
	}
	return results, nil
}

func PatchSubjective(id string, updates map[string]interface{}) (*questions.Subjective, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First update the document
	_, err := db.GetCollection(db.SubjectiveCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated questions.Subjective
	err = db.GetCollection(db.SubjectiveCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func SaveBulkSubjectives(subjectives []questions.Subjective) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert slice to []interface{} for InsertMany
	documents := make([]interface{}, len(subjectives))
	for i, s := range subjectives {
		documents[i] = s
	}

	_, err := db.GetCollection(db.SubjectiveCollection).InsertMany(ctx, documents)
	return err
}
