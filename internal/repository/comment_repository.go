package repository

import (
	"context"
	"server/internal/firebase"
	"server/internal/model"

	"cloud.google.com/go/firestore"
)

func SaveComment(c model.Comment) error {
	_, err := firebase.Client.Collection("comments").Doc(c.ID).Set(context.Background(), c)
	return err
}

func GetCommentByID(id string) (*model.Comment, error) {
	doc, err := firebase.Client.Collection("comments").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var c model.Comment
	doc.DataTo(&c)
	return &c, nil
}

func DeleteComment(id string) error {
	_, err := firebase.Client.Collection("comments").Doc(id).Delete(context.Background())
	return err
}

func GetAllComments() ([]model.Comment, error) {
	iter := firebase.Client.Collection("comments").Documents(context.Background())
	var results []model.Comment
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var c model.Comment
		doc.DataTo(&c)
		results = append(results, c)
	}
	return results, nil
}

func PatchComment(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("comments").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}
