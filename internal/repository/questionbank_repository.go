package repository

import (
	"context"
	"lumenslate/internal/firebase"
	"lumenslate/internal/model"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
)

func SaveQuestionBank(q model.QuestionBank) error {
	_, err := firebase.Client.Collection("questionBanks").Doc(q.ID).Set(context.Background(), q)
	return err
}

func GetQuestionBankByID(id string) (*model.QuestionBank, error) {
	doc, err := firebase.Client.Collection("questionBanks").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var q model.QuestionBank
	doc.DataTo(&q)
	return &q, nil
}

func DeleteQuestionBank(id string) error {
	_, err := firebase.Client.Collection("questionBanks").Doc(id).Delete(context.Background())
	return err
}

func GetAllQuestionBanks(filters map[string]string) ([]model.QuestionBank, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("questionBanks").Query

	if topic, ok := filters["topic"]; ok && topic != "" {
		q = q.Where("topic", "==", topic)
	}
	if name, ok := filters["name"]; ok && name != "" {
		q = q.Where("name", "==", name)
	}
	if teacherId, ok := filters["teacherId"]; ok && teacherId != "" {
		q = q.Where("teacherId", "==", teacherId)
	}
	if tags, ok := filters["tags"]; ok && tags != "" {
		tagList := strings.Split(tags, ",")
		for _, tag := range tagList {
			q = q.Where("tags", "array-contains", tag)
		}
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
	var results []model.QuestionBank
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var q model.QuestionBank
		doc.DataTo(&q)
		results = append(results, q)
	}
	return results, nil
}

func PatchQuestionBank(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("questionBanks").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}
