package repository

import (
	"context"
	"server/internal/firebase"
	"server/internal/model"
	"strconv"

	"cloud.google.com/go/firestore"
)

func SaveTeacher(t model.Teacher) error {
	_, err := firebase.Client.Collection("teachers").Doc(t.ID).Set(context.Background(), t)
	return err
}

func GetTeacherByID(id string) (*model.Teacher, error) {
	doc, err := firebase.Client.Collection("teachers").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var t model.Teacher
	doc.DataTo(&t)
	return &t, nil
}

func DeleteTeacher(id string) error {
	_, err := firebase.Client.Collection("teachers").Doc(id).Delete(context.Background())
	return err
}

func GetAllTeachers(filters map[string]string) ([]model.Teacher, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("teachers").Query

	if email, ok := filters["email"]; ok && email != "" {
		q = q.Where("email", "==", email)
	}
	if phone, ok := filters["phone"]; ok && phone != "" {
		q = q.Where("phone", "==", phone)
	}

	limit := 10
	offset := 0
	if l, err := strconv.Atoi(filters["limit"]); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(filters["offset"]); err == nil {
		offset = o
	}

	iter := q.Offset(offset).Limit(limit).Documents(ctx)
	var results []model.Teacher
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var t model.Teacher
		doc.DataTo(&t)
		results = append(results, t)
	}
	return results, nil
}

func PatchTeacher(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("teachers").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}
