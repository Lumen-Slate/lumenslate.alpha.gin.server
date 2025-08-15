package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSService struct {
	client     *storage.Client
	bucketName string
}

// NewGCSService creates a new GCS service instance
func NewGCSService() (*GCSService, error) {
	ctx := context.Background()

	// Get bucket name from environment variable
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME environment variable is required")
	}

	// Create GCS client using default credentials
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %v", err)
	}

	return &GCSService{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// NewGCSServiceWithCredentials creates a new GCS service instance with explicit credentials
func NewGCSServiceWithCredentials(credentialsPath string) (*GCSService, error) {
	ctx := context.Background()

	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME environment variable is required")
	}

	// Create GCS client with explicit credentials
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client with credentials: %v", err)
	}

	return &GCSService{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Close closes the GCS client
func (s *GCSService) Close() error {
	return s.client.Close()
}

// UploadFile uploads a file to GCS and returns the object name
func (s *GCSService) UploadFile(ctx context.Context, file multipart.File, filename, contentType string) (string, int64, error) {
	// Generate unique object name with timestamp prefix
	timestamp := time.Now().Format("20060102-150405")
	objectName := fmt.Sprintf("documents/%s-%s", timestamp, sanitizeFilename(filename))

	// Create GCS object writer
	obj := s.client.Bucket(s.bucketName).Object(objectName)
	writer := obj.NewWriter(ctx)

	// Set metadata
	writer.ContentType = contentType
	writer.Metadata = map[string]string{
		"original-filename": filename,
		"uploaded-at":       time.Now().UTC().Format(time.RFC3339),
	}

	// Copy file content to GCS
	size, err := io.Copy(writer, file)
	if err != nil {
		writer.Close()
		return "", 0, fmt.Errorf("failed to copy file to GCS: %v", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return "", 0, fmt.Errorf("failed to close GCS writer: %v", err)
	}

	return objectName, size, nil
}

// GenerateSignedURL generates a signed URL for reading an object
func (s *GCSService) GenerateSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	// Generate signed URL
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := s.client.Bucket(s.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %v", err)
	}

	return url, nil
}

// DeleteObject deletes an object from GCS with proper error handling and audit logging
func (s *GCSService) DeleteObject(ctx context.Context, objectName string) error {
	obj := s.client.Bucket(s.bucketName).Object(objectName)

	// Log deletion attempt for audit trail
	log.Printf("Attempting to delete GCS object: bucket=%s, object=%s", s.bucketName, objectName)

	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			log.Printf("GCS object not found during deletion: bucket=%s, object=%s", s.bucketName, objectName)
			return fmt.Errorf("object '%s' not found in bucket '%s'", objectName, s.bucketName)
		}
		log.Printf("Failed to delete GCS object: bucket=%s, object=%s, error=%v", s.bucketName, objectName, err)
		return fmt.Errorf("failed to delete object '%s' from bucket '%s': %v", objectName, s.bucketName, err)
	}

	// Log successful deletion for audit trail
	log.Printf("Successfully deleted GCS object: bucket=%s, object=%s", s.bucketName, objectName)
	return nil
}

// ObjectExists checks if an object exists in GCS
func (s *GCSService) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	obj := s.client.Bucket(s.bucketName).Object(objectName)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %v", err)
	}
	return true, nil
}

// GetObjectAttributes retrieves metadata about an object
func (s *GCSService) GetObjectAttributes(ctx context.Context, objectName string) (*storage.ObjectAttrs, error) {
	obj := s.client.Bucket(s.bucketName).Object(objectName)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %v", err)
	}
	return attrs, nil
}

// RenameObject atomically moves an object from oldObjectName to newObjectName using copy-and-delete pattern
func (s *GCSService) RenameObject(ctx context.Context, oldObjectName, newObjectName string) error {
	// Source and destination objects
	src := s.client.Bucket(s.bucketName).Object(oldObjectName)
	dst := s.client.Bucket(s.bucketName).Object(newObjectName)

	// Copy the object to new location
	_, err := dst.CopierFrom(src).Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy object from '%s' to '%s' during rename operation: %v", oldObjectName, newObjectName, err)
	}

	// Delete the original object
	if err := src.Delete(ctx); err != nil {
		// Try to clean up the copied object if original deletion fails
		if cleanupErr := dst.Delete(ctx); cleanupErr != nil {
			return fmt.Errorf("failed to delete original object '%s' during rename and cleanup of copied object '%s' also failed: original error: %v, cleanup error: %v", oldObjectName, newObjectName, err, cleanupErr)
		}
		return fmt.Errorf("failed to delete original object '%s' after copying to '%s': %v", oldObjectName, newObjectName, err)
	}

	return nil
}

// UploadFileWithCustomName uploads a file to GCS with a specific object name
func (s *GCSService) UploadFileWithCustomName(ctx context.Context, file io.Reader, objectName, contentType, originalFilename string) (int64, error) {
	// Create GCS object writer
	obj := s.client.Bucket(s.bucketName).Object(objectName)
	writer := obj.NewWriter(ctx)

	// Set metadata
	writer.ContentType = contentType
	writer.Metadata = map[string]string{
		"original-filename": originalFilename,
		"uploaded-at":       time.Now().UTC().Format(time.RFC3339),
	}

	// Copy file content to GCS
	size, err := io.Copy(writer, file)
	if err != nil {
		writer.Close()
		return 0, fmt.Errorf("failed to copy file to GCS: %v", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return 0, fmt.Errorf("failed to close GCS writer: %v", err)
	}

	return size, nil
}

// sanitizeFilename removes or replaces invalid characters in filenames
func sanitizeFilename(filename string) string {
	// Remove path separators and other potentially problematic characters
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "(", "")
	filename = strings.ReplaceAll(filename, ")", "")
	filename = strings.ReplaceAll(filename, "[", "")
	filename = strings.ReplaceAll(filename, "]", "")
	filename = strings.ReplaceAll(filename, "{", "")
	filename = strings.ReplaceAll(filename, "}", "")
	filename = strings.ReplaceAll(filename, "#", "")
	filename = strings.ReplaceAll(filename, "&", "")
	filename = strings.ReplaceAll(filename, "?", "")
	return filename
}
