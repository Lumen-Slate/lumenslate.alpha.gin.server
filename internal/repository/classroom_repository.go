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

func SaveClassroom(c model.Classroom) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.ClassroomCollection).InsertOne(ctx, c)
	return err
}

func GetClassroomByID(id string) (*model.Classroom, error) {
	ctx := context.Background()
	var c model.Classroom
	err := db.GetCollection(db.ClassroomCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func DeleteClassroom(id string) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.ClassroomCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllClassrooms(filters map[string]string) ([]model.Classroom, error) {
	ctx := context.Background()

	// Set up pagination
	limit := 10
	offset := 0
	if l, err := strconv.Atoi(filters["limit"]); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(filters["offset"]); err == nil {
		offset = o
	}

	// Check if search query is provided
	if q, ok := filters["q"]; ok && q != "" {
		// Use aggregation pipeline for search functionality
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"subject": bson.M{"$regex": q, "$options": "i"},
				},
			},
			{
				"$sort": bson.M{
					"subject":   1,  // Sort by subject alphabetically
					"createdAt": -1, // Then by creation date
				},
			},
			{"$skip": int64(offset)},
			{"$limit": int64(limit)},
		}

		// Apply other filters to the match stage
		matchStage := pipeline[0]["$match"].(bson.M)

		if tags, exists := filters["tags"]; exists && tags != "" {
			tagList := strings.Split(tags, ",")
			matchStage["tags"] = bson.M{"$all": tagList}
		}

		if teacherID, exists := filters["teacherId"]; exists && teacherID != "" {
			matchStage["teacherIds"] = teacherID
		}

		if subject, exists := filters["subject"]; exists && subject != "" {
			// If both q and subject filter exist, apply exact subject match in addition to search
			delete(matchStage, "subject") // Remove the regex match
			matchStage["$and"] = []bson.M{
				{"subject": bson.M{"$regex": q, "$options": "i"}},
				{"subject": subject},
			}
		}

		cursor, err := db.GetCollection(db.ClassroomCollection).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var results []model.Classroom
		if err = cursor.All(ctx, &results); err != nil {
			return make([]model.Classroom, 0), nil
		}
		return results, nil
	}

	// Regular filtering when no search query
	filter := bson.M{}

	// Apply tags filter if provided
	if tags, ok := filters["tags"]; ok && tags != "" {
		tagList := strings.Split(tags, ",")
		filter["tags"] = bson.M{"$all": tagList}
	}

	// Apply teacher filter if provided
	if teacherID, ok := filters["teacherId"]; ok && teacherID != "" {
		filter["teacherIds"] = teacherID
	}

	// Apply subject filter if provided
	if subject, ok := filters["subject"]; ok && subject != "" {
		filter["subject"] = subject
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := db.GetCollection(db.ClassroomCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Classroom
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.Classroom, 0), nil
	}
	return results, nil
}

func PatchClassroom(id string, updates map[string]interface{}) (*model.Classroom, error) {
	ctx := context.Background()
	updates["updatedAt"] = time.Now()

	// First update the document
	_, err := db.GetCollection(db.ClassroomCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.Classroom
	err = db.GetCollection(db.ClassroomCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
