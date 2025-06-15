package repository

import (
	"context"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func SaveComment(c model.Comment) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.CommentCollection).InsertOne(ctx, c)
	return err
}

func GetCommentByID(id string) (*model.Comment, error) {
	ctx := context.Background()
	var c model.Comment
	err := db.GetCollection(db.CommentCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func DeleteComment(id string) error {
	ctx := context.Background()
	_, err := db.GetCollection(db.CommentCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetAllComments() ([]model.Comment, error) {
	ctx := context.Background()
	cursor, err := db.GetCollection(db.CommentCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Comment
	if err = cursor.All(ctx, &results); err != nil {
		return make([]model.Comment, 0), nil
	}
	return results, nil
}

func PatchComment(id string, updates map[string]interface{}) (*model.Comment, error) {
	ctx := context.Background()
	updates["updatedAt"] = time.Now()

	// First update the document
	_, err := db.GetCollection(db.CommentCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.Comment
	err = db.GetCollection(db.CommentCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
