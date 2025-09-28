# Firebase Authentication Middleware

This document explains how to set up and use Firebase Authentication in the Lumen Slate API backend.

## Overview

The Firebase Authentication middleware provides secure authentication for your API endpoints using Firebase ID tokens. It supports both required and optional authentication patterns.

## Setup

### 1. Firebase Configuration

Set up your Firebase project and service account:

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Create or select your project
3. Go to Project Settings > Service Accounts
4. Generate a new private key (JSON file)
5. Place the JSON file in your project directory or set up Application Default Credentials (ADC)

### 2. Environment Variables

Set one of the following environment variables:

```bash
# Option 1: Use service account key file
FIREBASE_SERVICE_ACCOUNT_PATH=/path/to/your/service-account-key.json

# Option 2: Use Application Default Credentials (ADC)
# No environment variable needed - the middleware will use ADC automatically
# This is recommended for production deployments on Google Cloud
```

### 3. Go Dependencies

The following Firebase dependencies are required:

```bash
go get firebase.google.com/go
go get firebase.google.com/go/auth
go get google.golang.org/api/option
```

## Middleware Types

### 1. AuthMiddleware (Required Authentication)

Blocks requests without valid Firebase tokens:

```go
import "lumenslate/internal/middleware"

// Apply to routes that require authentication
protectedRoutes := router.Group("/api/v1/protected")
protectedRoutes.Use(middleware.AuthMiddleware(firebaseAuthClient))
```

### 2. OptionalAuthMiddleware (Optional Authentication)

Allows requests to proceed but sets user context if token is provided:

```go
// Apply to routes where authentication is optional
apiV1 := router.Group("/api/v1")
apiV1.Use(middleware.OptionalAuthMiddleware(firebaseAuthClient))
```

## Frontend Integration

### Sending Firebase Tokens

Include the Firebase ID token in the `Authorization` header:

```javascript
// Example using fetch API
const idToken = await user.getIdToken();

const response = await fetch('/api/v1/some-endpoint', {
    method: 'GET',
    headers: {
        'Authorization': `Bearer ${idToken}`,
        'Content-Type': 'application/json',
    },
});
```

### Example Frontend Implementation (React/Firebase)

```javascript
import { getAuth, onAuthStateChanged } from 'firebase/auth';

const auth = getAuth();

// Set up auth state listener
onAuthStateChanged(auth, async (user) => {
    if (user) {
        // User is signed in
        const idToken = await user.getIdToken();
        
        // Store token for API calls
        localStorage.setItem('firebase_token', idToken);
        
        // Set up API client with token
        setupAPIClient(idToken);
    } else {
        // User is signed out
        localStorage.removeItem('firebase_token');
    }
});

// API client setup
function setupAPIClient(token) {
    // Configure your HTTP client (axios, fetch, etc.)
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
}
```

## Using Authentication in Controllers

### Getting User Information

```go
import "lumenslate/internal/middleware"

func SomeController(c *gin.Context) {
    // Check if user is authenticated
    if middleware.IsAuthenticated(c) {
        userID, _ := middleware.GetUserID(c)
        userEmail, _ := middleware.GetUserEmail(c)
        userClaims, _ := middleware.GetUserClaims(c)
        
        // Use user information...
    }
    
    // Handle both authenticated and non-authenticated cases
}
```

### Requiring Authentication

```go
func ProtectedController(c *gin.Context) {
    // This assumes the route is protected by AuthMiddleware
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(500, gin.H{"error": "Authentication required"})
        return
    }
    
    // User is guaranteed to be authenticated
}
```

## API Endpoints

### Authentication Status

- `GET /api/v1/auth/me` - Get current user info (optional auth)
- `GET /api/v1/protected/auth/profile` - Get user profile (required auth)
- `PUT /api/v1/protected/auth/profile` - Update user profile (required auth)

## Error Responses

### Authentication Errors

```json
{
    "error": "Unauthorized",
    "message": "Authorization header is required"
}
```

```json
{
    "error": "Unauthorized", 
    "message": "Invalid token: <error details>"
}
```

## Security Best Practices

1. **Token Refresh**: Implement token refresh logic in your frontend
2. **HTTPS Only**: Always use HTTPS in production
3. **Token Storage**: Store tokens securely (avoid localStorage for sensitive apps)
4. **Validation**: Always validate tokens server-side
5. **Logging**: Monitor authentication attempts and failures

## Testing

### Manual Testing with curl

```bash
# Get Firebase token from your frontend app, then:
curl -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
     http://localhost:8080/api/v1/auth/me
```

### Unit Testing

```go
// Example test setup
func TestAuthMiddleware(t *testing.T) {
    // Set up test Firebase client
    // Create test requests with mock tokens
    // Verify middleware behavior
}
```

## Troubleshooting

### Common Issues

1. **"Invalid token" errors**:
   - Check that the service account key is correct
   - Ensure the token hasn't expired (Firebase tokens expire after 1 hour)
   - Verify the token is from the correct Firebase project

2. **"Authorization header is required"**:
   - Make sure the frontend is sending the `Authorization` header
   - Check that the header format is `Bearer <token>`

3. **Firebase initialization errors**:
   - Verify the service account key path is correct
   - Check that the service account has the necessary permissions
   - For ADC, ensure you're running in the correct environment

### Debugging

Enable detailed logging by setting the log level:

```go
log.SetLevel(log.DebugLevel)
```

## Migration from Existing Auth

If you have existing authentication, you can migrate gradually:

1. Start with `OptionalAuthMiddleware` on all routes
2. Update frontend to send Firebase tokens
3. Gradually move sensitive endpoints to protected routes
4. Remove old authentication system once migration is complete

This approach ensures backward compatibility during the migration process.