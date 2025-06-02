package questions

import (
	"context"
	"lumenslate/internal/firebase"
	"lumenslate/internal/model/questions"
	"strconv"

	"cloud.google.com/go/firestore"
)

func SaveMCQ(m questions.MCQ) error {
	_, err := firebase.Client.Collection("mcqs").Doc(m.ID).Set(context.Background(), m)
	return err
}

func GetMCQByID(id string) (*questions.MCQ, error) {
	doc, err := firebase.Client.Collection("mcqs").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var m questions.MCQ
	doc.DataTo(&m)
	return &m, nil
}

func DeleteMCQ(id string) error {
	_, err := firebase.Client.Collection("mcqs").Doc(id).Delete(context.Background())
	return err
}

func GetAllMCQs(filters map[string]string) ([]questions.MCQ, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("mcqs").Query

	if bankID, ok := filters["bankId"]; ok && bankID != "" {
		q = q.Where("bankId", "==", bankID)
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
	var results []questions.MCQ
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var m questions.MCQ
		doc.DataTo(&m)
		results = append(results, m)
	}
	return results, nil
}

func PatchMCQ(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("mcqs").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}

func SaveBulkMCQs(mcqs []questions.MCQ) error {
	ctx := context.Background()
	bw := firebase.Client.BulkWriter(ctx)

	for _, m := range mcqs {
		ref := firebase.Client.Collection("mcqs").Doc(m.ID)
		if _, err := bw.Create(ref, m); err != nil {
			return err
		}
	}

	bw.End()
	return nil
}
