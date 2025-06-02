package repository

import (
	"context"
	"server/internal/firebase"
	"server/internal/model"
	"strconv"

	"cloud.google.com/go/firestore"
)

func SavePost(p model.Post) error {
	_, err := firebase.Client.Collection("posts").Doc(p.ID).Set(context.Background(), p)
	return err
}

func GetPostByID(id string) (*model.Post, error) {
	doc, err := firebase.Client.Collection("posts").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var p model.Post
	doc.DataTo(&p)
	return &p, nil
}

func DeletePost(id string) error {
	_, err := firebase.Client.Collection("posts").Doc(id).Delete(context.Background())
	return err
}

func GetAllPosts(filters map[string]string) ([]model.Post, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("posts").Query

	if userId, ok := filters["userId"]; ok && userId != "" {
		q = q.Where("userId", "==", userId)
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
	var results []model.Post
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var p model.Post
		doc.DataTo(&p)
		results = append(results, p)
	}
	return results, nil
}

func PatchPost(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("posts").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}
