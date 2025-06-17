package utils

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/compute/metadata"
)

// GetProjectID attempts to fetch the project ID from the GCE metadata server.
// If unavailable, it falls back to the GOOGLE_PROJECT_ID environment variable.
func GetProjectID() string {
	if metadata.OnGCE() {
		ctx := context.Background()
		projectID, err := metadata.ProjectIDWithContext(ctx)
		if err == nil {
			log.Println("🌐 Project ID fetched from GCE metadata server.")
			return projectID
		}
		log.Printf("⚠️ Metadata server error: %v", err)
	}

	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		log.Println("⚠️ GOOGLE_PROJECT_ID not set in environment.")
	}
	return projectID
}
