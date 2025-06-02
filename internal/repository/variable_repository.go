package repository

import (
	"context"
	"lumenslate/internal/firebase"
	"lumenslate/internal/model"
	"strconv"

	"cloud.google.com/go/firestore"
)

func SaveVariable(v model.Variable) error {
	_, err := firebase.Client.Collection("variables").Doc(v.ID).Set(context.Background(), v)
	return err
}

func GetVariableByID(id string) (*model.Variable, error) {
	doc, err := firebase.Client.Collection("variables").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var v model.Variable
	doc.DataTo(&v)
	return &v, nil
}

func DeleteVariable(id string) error {
	_, err := firebase.Client.Collection("variables").Doc(id).Delete(context.Background())
	return err
}

func GetAllVariables(filters map[string]string) ([]model.Variable, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("variables").Query

	limit := 10
	offset := 0
	if l, err := strconv.Atoi(filters["limit"]); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(filters["offset"]); err == nil {
		offset = o
	}

	iter := q.Offset(offset).Limit(limit).Documents(ctx)
	var results []model.Variable
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var v model.Variable
		doc.DataTo(&v)
		results = append(results, v)
	}
	return results, nil
}

func PatchVariable(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("variables").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}

func SaveBulkVariables(variables []model.Variable) error {
	ctx := context.Background()
	bulkWriter := firebase.Client.BulkWriter(ctx)
	collection := firebase.Client.Collection("variables")

	for _, v := range variables {
		docRef := collection.Doc(v.ID)
		bulkWriter.Create(docRef, v)
	}

	// Close the bulk writer and wait for all operations to complete
	bulkWriter.End()

	return nil
}
