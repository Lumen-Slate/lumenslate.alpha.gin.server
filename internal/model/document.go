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
