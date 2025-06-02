package questions

import (
	"context"
	"lumenslate/internal/firebase"
	"lumenslate/internal/model/questions"
	"strconv"

	"cloud.google.com/go/firestore"
)

func SaveSubjective(s questions.Subjective) error {
	_, err := firebase.Client.Collection("subjectives").Doc(s.ID).Set(context.Background(), s)
	return err
}

func GetSubjectiveByID(id string) (*questions.Subjective, error) {
	doc, err := firebase.Client.Collection("subjectives").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var s questions.Subjective
	doc.DataTo(&s)
	return &s, nil
}

func DeleteSubjective(id string) error {
	_, err := firebase.Client.Collection("subjectives").Doc(id).Delete(context.Background())
	return err
}

func GetAllSubjectives(filters map[string]string) ([]questions.Subjective, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("subjectives").Query

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
	var results []questions.Subjective
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var s questions.Subjective
		doc.DataTo(&s)
		results = append(results, s)
	}
	return results, nil
}

func PatchSubjective(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("subjectives").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}

func SaveBulkSubjectives(subjectives []questions.Subjective) error {
	ctx := context.Background()
	bw := firebase.Client.BulkWriter(ctx)

	for _, s := range subjectives {
		ref := firebase.Client.Collection("subjectives").Doc(s.ID)
		if _, err := bw.Create(ref, s); err != nil {
			return err
		}
	}

	bw.End()
	return nil
}
