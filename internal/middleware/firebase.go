package middleware

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

// FirebaseAuthClient holds the Firebase Auth client
type FirebaseAuthClient struct {
	Client *auth.Client
}

// InitializeFirebaseAuth initializes Firebase authentication
func InitializeFirebaseAuth() (*FirebaseAuthClient, error) {
	ctx := context.Background()

	// Try to get service account key path from environment variable
	serviceAccountPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	var app *firebase.App
	var err error

	if serviceAccountPath != "" {
		// Use service account key file
		opt := option.WithCredentialsFile(serviceAccountPath)
		app, err = firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Printf("Error initializing Firebase app with service account: %v", err)
			return nil, err
		}
		log.Println("Firebase initialized with service account key")
	} else {
		// Try to use default credentials (ADC - Application Default Credentials)
		app, err = firebase.NewApp(ctx, nil)
		if err != nil {
			log.Printf("Error initializing Firebase app with default credentials: %v", err)
			return nil, err
		}
		log.Println("Firebase initialized with Application Default Credentials")
	} // Get Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Printf("Error getting Firebase Auth client: %v", err)
		return nil, err
	}

	log.Println("Firebase Auth client initialized successfully")
	return &FirebaseAuthClient{Client: authClient}, nil
}

// GetAuthClient returns the Firebase Auth client
func (f *FirebaseAuthClient) GetAuthClient() *auth.Client {
	return f.Client
}
