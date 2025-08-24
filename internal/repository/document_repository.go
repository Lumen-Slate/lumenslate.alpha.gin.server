package repository

import (
	"context"
	"fmt"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// DocumentError represents a document repository error with context
type DocumentError struct {
	Op      string // Operation that failed
	FileID  string // Document file ID (if available)
	Message string // Human-readable error message
	Err     error  // Underlying error
}

func (e *DocumentError) Error() string {
	if e.FileID != "" {
		return fmt.Sprintf("document repository %s (fileId: %s): %s: %v", e.Op, e.FileID, e.Message, e.Err)
	}
	return fmt.Sprintf("document repository %s: %s: %v", e.Op, e.Message, e.Err)
}

func (e *DocumentError) Unwrap() error {
	return e.Err
}

type DocumentRepository struct {
	collection *mongo.Collection
}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{
		collection: db.GetCollection(db.DocumentCollection),
	}
}

// CreateDocument creates a new document record in the database with proper error handling
func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *model.Document) error {
	// Set timestamps
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	// Set default status for async processing
	if doc.Status == "" {
		doc.Status = "pending"
	}

	// Insert document with proper error handling
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		// Wrap MongoDB errors with context
		if mongo.IsDuplicateKeyError(err) {
			return &DocumentError{
				Op:      "CreateDocument",
				FileID:  doc.FileID,
				Message: "document with this fileId already exists",
				Err:     err,
			}
		}
		return &DocumentError{
			Op:      "CreateDocument",
			FileID:  doc.FileID,
			Message: "failed to insert document",
			Err:     err,
		}
	}

	// Set the generated ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	return nil
}

// GetDocumentByFileID retrieves a document by its file ID
func (r *DocumentRepository) GetDocumentByFileID(ctx context.Context, fileID string) (*model.Document, error) {
	var doc model.Document
	filter := bson.M{"fileId": fileID}

	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &DocumentError{
				Op:      "GetDocumentByFileID",
				FileID:  fileID,
				Message: "document not found",
				Err:     err,
			}
		}
		return nil, &DocumentError{
			Op:      "GetDocumentByFileID",
			FileID:  fileID,
			Message: "failed to retrieve document",
			Err:     err,
		}
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

// UpdateStatus atomically updates document status, error message, and timestamp
func (r *DocumentRepository) UpdateStatus(ctx context.Context, fileID string, status string, errorMsg string) error {
	filter := bson.M{"fileId": fileID}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	// Only set errorMsg if it's not empty, otherwise unset it
	if errorMsg != "" {
		update["$set"].(bson.M)["errorMsg"] = errorMsg
	} else {
		update["$unset"] = bson.M{"errorMsg": ""}
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return &DocumentError{
			Op:      "UpdateStatus",
			FileID:  fileID,
			Message: "failed to update document status",
			Err:     err,
		}
	}

	// Check if document was found and updated
	if result.MatchedCount == 0 {
		return &DocumentError{
			Op:      "UpdateStatus",
			FileID:  fileID,
			Message: "document not found",
			Err:     mongo.ErrNoDocuments,
		}
	}

	return nil
}

// UpdateFields flexibly updates document fields with validation against immutable fields
func (r *DocumentRepository) UpdateFields(ctx context.Context, fileID string, updates bson.M) error {
	// Define immutable fields that cannot be updated
	immutableFields := map[string]bool{
		"fileId":    true,
		"_id":       true,
		"createdAt": true,
	}

	// Validate that no immutable fields are being updated
	for field := range updates {
		if immutableFields[field] {
			return &DocumentError{
				Op:      "UpdateFields",
				FileID:  fileID,
				Message: fmt.Sprintf("cannot update immutable field: %s", field),
				Err:     nil,
			}
		}
	}

	// Automatically update the updatedAt timestamp
	updates["updatedAt"] = time.Now()

	filter := bson.M{"fileId": fileID}
	updateDoc := bson.M{"$set": updates}

	result, err := r.collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return &DocumentError{
			Op:      "UpdateFields",
			FileID:  fileID,
			Message: "failed to update document fields",
			Err:     err,
		}
	}

	// Check if document was found and updated
	if result.MatchedCount == 0 {
		return &DocumentError{
			Op:      "UpdateFields",
			FileID:  fileID,
			Message: "document not found",
			Err:     mongo.ErrNoDocuments,
		}
	}

	return nil
}

// UpdateDocument updates an existing document (deprecated - use UpdateFields instead)
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
