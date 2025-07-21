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
		// Use aggregation pipeline to prioritize name matches over email matches
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"$or": []bson.M{
						{"name": bson.M{"$regex": q, "$options": "i"}},
						{"email": bson.M{"$regex": q, "$options": "i"}},
					},
				},
			},
			{
				"$addFields": bson.M{
					"nameMatch": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$regexMatch": bson.M{"input": "$name", "regex": q, "options": "i"}},
							"then": 1,
							"else": 0,
						},
					},
				},
			},
			{
				"$sort": bson.M{
					"nameMatch": -1, // Name matches first
					"createdAt": -1, // Then by creation date
				},
			},
			{"$skip": int64(offset)},
			{"$limit": int64(limit)},
		}

		// Apply other filters to the match stage
		matchStage := pipeline[0]["$match"].(bson.M)
		if email, exists := filters["email"]; exists && email != "" {
			matchStage["email"] = email
		}
		if rollNo, exists := filters["rollNo"]; exists && rollNo != "" {
			matchStage["rollNo"] = rollNo
		}
		if classIds, exists := filters["classIds"]; exists && classIds != "" {
			classIdList := strings.Split(classIds, ",")
			// Trim whitespace from each classId
			for i, id := range classIdList {
				classIdList[i] = strings.TrimSpace(id)
			}
			matchStage["classIds"] = bson.M{"$in": classIdList}
		}

		cursor, err := db.GetCollection(db.StudentCollection).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var results []model.Student
		if err = cursor.All(ctx, &results); err != nil {
			return make([]model.Student, 0), nil
		}
		return results, nil
	}

	// Regular filtering when no search query
	filter := bson.M{}

	// Apply email filter if provided
	if email, ok := filters["email"]; ok && email != "" {
		filter["email"] = email
	}

	// Apply roll number filter if provided
	if rollNo, ok := filters["rollNo"]; ok && rollNo != "" {
		filter["rollNo"] = rollNo
	}

	// Apply classIds filter if provided
	if classIds, ok := filters["classIds"]; ok && classIds != "" {
		classIdList := strings.Split(classIds, ",")
		// Trim whitespace from each classId
		for i, id := range classIdList {
			classIdList[i] = strings.TrimSpace(id)
		}
		filter["classIds"] = bson.M{"$in": classIdList}
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := db.GetCollection(db.StudentCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Student
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.Student, 0), nil
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
