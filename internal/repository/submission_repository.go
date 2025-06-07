package repository

import (
	"context"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func SaveSubmission(s model.Submission) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.SubmissionCollection).InsertOne(ctx, s)
	return err
}

func GetSubmissionByID(id string) (*model.Submission, error) {
	ctx := context.Background()
	var s model.Submission
	err := db.GetCollection(db.SubmissionCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func DeleteSubmission(id string) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.SubmissionCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllSubmissions(filters map[string]string) ([]model.Submission, error) {
	ctx := context.Background()
	filter := bson.M{}

	if studentId, ok := filters["studentId"]; ok && studentId != "" {
		filter["studentId"] = studentId
	}
	if assignmentId, ok := filters["assignmentId"]; ok && assignmentId != "" {
		filter["assignmentId"] = assignmentId
	}

	cursor, err := db.GetCollection(db.SubmissionCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Submission
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func PatchSubmission(id string, updates map[string]interface{}) (*model.Submission, error) {
	ctx := context.Background()
	updates["updatedAt"] = time.Now()

	// Update the document
	_, err := db.GetCollection(db.SubmissionCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Fetch the updated document
	var updated model.Submission
	err = db.GetCollection(db.SubmissionCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
