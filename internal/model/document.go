package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Document represents a document stored in GCS and referenced in RAG corpus
type Document struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FileID      string             `bson:"fileId" json:"fileId"`           // Unique identifier for the document
	DisplayName string             `bson:"displayName" json:"displayName"` // Original filename
	GCSBucket   string             `bson:"gcsBucket" json:"gcsBucket"`     // GCS bucket name
	GCSObject   string             `bson:"gcsObject" json:"gcsObject"`     // GCS object key/path
	ContentType string             `bson:"contentType" json:"contentType"` // MIME type of the file
	Size        int64              `bson:"size" json:"size"`               // File size in bytes
	CorpusName  string             `bson:"corpusName" json:"corpusName"`   // RAG corpus this document belongs to
	RAGFileID   string             `bson:"ragFileId" json:"ragFileId"`     // Vertex AI RAG file ID
	UploadedBy  string             `bson:"uploadedBy" json:"uploadedBy"`   // User who uploaded the document
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`

	// Async processing fields
	Status   string `bson:"status" json:"status"`                         // "pending", "completed", or "failed"
	ErrorMsg string `bson:"errorMsg,omitempty" json:"errorMsg,omitempty"` // Error message if processing failed
}

// NewDocument creates a new Document with default values
func NewDocument(fileID, displayName, gcsBucket, gcsObject, contentType, corpusName, ragFileID, uploadedBy string, size int64) *Document {
	now := time.Now()
	return &Document{
		FileID:      fileID,
		DisplayName: displayName,
		GCSBucket:   gcsBucket,
		GCSObject:   gcsObject,
		ContentType: contentType,
		Size:        size,
		CorpusName:  corpusName,
		RAGFileID:   ragFileID,
		UploadedBy:  uploadedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      "pending", // Default status for async processing
		ErrorMsg:    "",        // Empty error message initially
	}
}

// DocumentMetadata represents minimal document information for listings
type DocumentMetadata struct {
	FileID      string    `json:"fileId"`
	DisplayName string    `json:"displayName"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	CorpusName  string    `json:"corpusName"`
	CreatedAt   time.Time `json:"createdAt"`
}
