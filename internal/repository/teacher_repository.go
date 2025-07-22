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

func SaveTeacher(t model.Teacher) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.TeacherCollection).InsertOne(ctx, t)
	return err
}

func GetTeacherByID(id string) (*model.Teacher, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var t model.Teacher
	err := db.GetCollection(db.TeacherCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func DeleteTeacher(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.TeacherCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllTeachers(filters map[string]string) ([]model.Teacher, error) {
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
	if email, ok := filters["email"]; ok && email != "" {
		filter["email"] = email
	}
	if phone, ok := filters["phone"]; ok && phone != "" {
		filter["phone"] = phone
	}
	if name, ok := filters["name"]; ok && name != "" {
		filter["name"] = name
	}

	cursor, err := db.GetCollection(db.TeacherCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Teacher
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.Teacher, 0), nil
	}
	return results, nil
}

func PatchTeacher(id string, updates map[string]interface{}) (*model.Teacher, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First update the document
	_, err := db.GetCollection(db.TeacherCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.Teacher
	err = db.GetCollection(db.TeacherCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
