package repository

import (
	"context"
	"lumenslate/internal/firebase"
	"lumenslate/internal/model"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
)

func SaveClassroom(c model.Classroom) error {
	_, err := firebase.Client.Collection("classrooms").Doc(c.ID).Set(context.Background(), c)
	return err
}

func GetClassroomByID(id string) (*model.Classroom, error) {
	doc, err := firebase.Client.Collection("classrooms").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var c model.Classroom
	doc.DataTo(&c)
	return &c, nil
}

func DeleteClassroom(id string) error {
	_, err := firebase.Client.Collection("classrooms").Doc(id).Delete(context.Background())
	return err
}

func GetAllClassrooms(filters map[string]string) ([]model.Classroom, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("classrooms").Query

	limit := 10
	offset := 0
	if l, err := strconv.Atoi(filters["limit"]); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(filters["offset"]); err == nil {
		offset = o
	}

	// Apply tags filter if provided
	if tags, ok := filters["tags"]; ok && tags != "" {
		tagList := strings.Split(tags, ",")
		for _, tag := range tagList {
			q = q.Where("tags", "array-contains", tag)
		}
	}

	iter := q.Offset(offset).Limit(limit).Documents(ctx)
	var results []model.Classroom
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var c model.Classroom
		doc.DataTo(&c)
		results = append(results, c)
	}
	return results, nil
}

func PatchClassroom(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("classrooms").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}
