package repository

import (
	"context"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DocumentRepository struct {
	collection *mongo.Collection
}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{
		collection: db.GetCollection(db.DocumentCollection),
	}
}

// CreateDocument creates a new document record in the database
func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *model.Document) error {
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	doc.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetDocumentByFileID retrieves a document by its file ID
func (r *DocumentRepository) GetDocumentByFileID(ctx context.Context, fileID string) (*model.Document, error) {
	var doc model.Document
	filter := bson.M{"fileId": fileID}

	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

// GetDocumentsByCorpus retrieves all documents in a specific corpus
func (r *DocumentRepository) GetDocumentsByCorpus(ctx context.Context, corpusName string) ([]model.Document, error) {
	var documents []model.Document
	filter := bson.M{"corpusName": corpusName}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}

	return documents, nil
}

// UpdateDocument updates an existing document
func (r *DocumentRepository) UpdateDocument(ctx context.Context, fileID string, update bson.M) error {
	update["updatedAt"] = time.Now()

	filter := bson.M{"fileId": fileID}
	updateDoc := bson.M{"$set": update}

	_, err := r.collection.UpdateOne(ctx, filter, updateDoc)
	return err
}

// DeleteDocument deletes a document by file ID
func (r *DocumentRepository) DeleteDocument(ctx context.Context, fileID string) error {
	filter := bson.M{"fileId": fileID}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// ListAllDocuments retrieves all documents with pagination
func (r *DocumentRepository) ListAllDocuments(ctx context.Context, skip, limit int64) ([]model.DocumentMetadata, error) {
	var documents []model.DocumentMetadata

	pipeline := []bson.M{
		{
			"$project": bson.M{
				"fileId":      1,
				"displayName": 1,
				"contentType": 1,
				"size":        1,
				"corpusName":  1,
				"createdAt":   1,
			},
		},
		{"$skip": skip},
		{"$limit": limit},
		{"$sort": bson.M{"createdAt": -1}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}

	return documents, nil
}
