package questions

import (
	"context"
	"server/internal/firebase"
	"server/internal/model/questions"
	"strconv"

	"cloud.google.com/go/firestore"
)

func SaveMSQ(m questions.MSQ) error {
	_, err := firebase.Client.Collection("msqs").Doc(m.ID).Set(context.Background(), m)
	return err
}

func GetMSQByID(id string) (*questions.MSQ, error) {
	doc, err := firebase.Client.Collection("msqs").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var m questions.MSQ
	doc.DataTo(&m)
	return &m, nil
}

func DeleteMSQ(id string) error {
	_, err := firebase.Client.Collection("msqs").Doc(id).Delete(context.Background())
	return err
}

func GetAllMSQs(filters map[string]string) ([]questions.MSQ, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("msqs").Query

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
	var results []questions.MSQ
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var m questions.MSQ
		doc.DataTo(&m)
		results = append(results, m)
	}
	return results, nil
}

func PatchMSQ(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("msqs").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}

func SaveBulkMSQs(msqs []questions.MSQ) error {
	ctx := context.Background()
	bw := firebase.Client.BulkWriter(ctx)

	for _, m := range msqs {
		ref := firebase.Client.Collection("msqs").Doc(m.ID)
		if _, err := bw.Create(ref, m); err != nil {
			return err
		}
	}

	bw.End()
	return nil
}
