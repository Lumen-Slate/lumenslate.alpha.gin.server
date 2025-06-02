package questions

import (
	"context"
	"server/internal/firebase"
	"server/internal/model/questions"
	"strconv"

	"cloud.google.com/go/firestore"
)

func SaveNAT(n questions.NAT) error {
	_, err := firebase.Client.Collection("nats").Doc(n.ID).Set(context.Background(), n)
	return err
}

func GetNATByID(id string) (*questions.NAT, error) {
	doc, err := firebase.Client.Collection("nats").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var n questions.NAT
	doc.DataTo(&n)
	return &n, nil
}

func DeleteNAT(id string) error {
	_, err := firebase.Client.Collection("nats").Doc(id).Delete(context.Background())
	return err
}

func GetAllNATs(filters map[string]string) ([]questions.NAT, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("nats").Query

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
	var results []questions.NAT
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var n questions.NAT
		doc.DataTo(&n)
		results = append(results, n)
	}
	return results, nil
}

func PatchNAT(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("nats").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}

func SaveBulkNATs(nats []questions.NAT) error {
	ctx := context.Background()
	bw := firebase.Client.BulkWriter(ctx)

	for _, n := range nats {
		ref := firebase.Client.Collection("nats").Doc(n.ID)
		if _, err := bw.Create(ref, n); err != nil {
			return err
		}
	}

	bw.End()
	return nil
}
