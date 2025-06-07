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

func SaveStudent(s model.Student) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.StudentCollection).InsertOne(ctx, s)
	return err
}

func GetStudentByID(id string) (*model.Student, error) {
	ctx := context.Background()
	var s model.Student
	err := db.GetCollection(db.StudentCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func DeleteStudent(id string) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.StudentCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllStudents(filters map[string]string) ([]model.Student, error) {
	ctx := context.Background()
	filter := bson.M{}

	// Apply email filter if provided
	if email, ok := filters["email"]; ok && email != "" {
		filter["email"] = email
	}

	// Apply roll number filter if provided
	if rollNo, ok := filters["rollNo"]; ok && rollNo != "" {
		filter["rollNo"] = rollNo
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

	cursor, err := db.GetCollection(db.StudentCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Student
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func PatchStudent(id string, updates map[string]interface{}) (*model.Student, error) {
	ctx := context.Background()
	updates["updatedAt"] = time.Now()

	// Update the document
	_, err := db.GetCollection(db.StudentCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Fetch the updated document
	var updated model.Student
	err = db.GetCollection(db.StudentCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
