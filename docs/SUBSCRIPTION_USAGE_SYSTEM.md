# Subscription and Usage Tracking System

## Overview
This system provides comprehensive subscription management and usage tracking capabilities for the Lumen Slate application. It tracks various usage metrics including question banks, questions, AI agent uses, recap classes, and assignment exports.

## Components Created

### 1. Models
- **Subscription** (`internal/model/subscription.go`)
  - Manages user subscriptions with status tracking
  - Supports active, scheduled_to_cancel, cancelled, and inactive states
  - Includes billing period management and cancellation scheduling

- **UsageTracking** (`internal/model/usage_tracking.go`)
  - Tracks usage metrics per user per period (monthly)
  - Supports various usage types: question banks, questions, AI agents, etc.
  - Provides aggregation and metrics functionality

### 2. Repositories
- **SubscriptionRepository** (`internal/repository/subscription_repository.go`)
  - CRUD operations for subscriptions
  - User-specific subscription queries
  - Expired subscription processing

- **UsageTrackingRepository** (`internal/repository/usage_tracking_repository.go`)
  - Atomic increment operations for usage counters
  - Period-based usage retrieval
  - Aggregation queries for usage metrics

### 3. Services
- **SubscriptionService** (`internal/service/subscription_service.go`)
  - Business logic for subscription management
  - Subscription lifecycle management (create, update, cancel, renew)
  - Validation and business rules

- **UsageTrackingService** (`internal/service/usage_tracking_service.go`)
  - Usage tracking business logic
  - Bulk usage tracking capabilities
  - Usage metrics and reporting

### 4. Controllers
- **SubscriptionController** (`internal/controller/subscription_controller.go`)
  - HTTP handlers for subscription operations
  - RESTful API endpoints
  - Request validation and error handling

- **UsageTrackingController** (`internal/controller/usage_tracking_controller.go`)
  - HTTP handlers for usage tracking
  - Multiple tracking methods (JSON, query params)
  - Usage reporting endpoints

### 5. Routes
- **SubscriptionRoutes** (`internal/routes/subscription_routes.go`)
  - Subscription management endpoints
  - User-specific subscription queries
  - Admin operations

- **UsageTrackingRoutes** (`internal/routes/usage_tracking_routes.go`)
  - Usage tracking endpoints
  - Flexible tracking options (detailed and simple)
  - Usage reporting and analytics

### 6. Database Collections
- Updated `collections.go` to include:
  - `subscriptions` collection
  - `usage_tracking` collection

## Integration Steps

### 1. Update main.go
Add the new routes to your main application file:

```go
import (
    "lumenslate/internal/routes"
    // ... other imports
)

func main() {
    // ... existing setup code
    
    // Register new routes
    routes.RegisterSubscriptionRoutes(router)
    routes.RegisterUsageTrackingRoutes(router)
    
    // ... rest of your application
}
```

### 2. Create Database Indexes (Optional but Recommended)
For better performance, create indexes on frequently queried fields:

```javascript
// MongoDB shell commands
use lumen_slate

// Subscription indexes
db.subscriptions.createIndex({ "user_id": 1 })
db.subscriptions.createIndex({ "status": 1 })
db.subscriptions.createIndex({ "lookup_key": 1 })
db.subscriptions.createIndex({ "cancel_at": 1 })

// Usage tracking indexes
db.usage_tracking.createIndex({ "user_id": 1, "period": 1 }, { unique: true })
db.usage_tracking.createIndex({ "period": 1 })
db.usage_tracking.createIndex({ "user_id": 1 })
```

## API Endpoints

### Subscription Management

#### Create Subscription
```
POST /api/v1/subscriptions
Content-Type: application/json

{
  "user_id": "user123",
  "lookup_key": "pro_monthly",
  "currency": "USD",
  "current_period_start": "2023-01-01T00:00:00Z",
  "current_period_end": "2023-02-01T00:00:00Z"
}
```

#### Get User's Active Subscription
```
GET /api/v1/subscriptions/user/{id}
```

#### Cancel Subscription
```
DELETE /api/v1/subscriptions/{id}
```

#### Schedule Cancellation
```
POST /api/v1/subscriptions/{id}/schedule-cancellation
```

### Usage Tracking

#### Track Question Bank Usage
```
POST /api/v1/usage/user/{id}/track/question-banks
Content-Type: application/json

{
  "count": 5
}
```

#### Simple Increment (with query param)
```
POST /api/v1/usage/user/{id}/increment/ia-agent?count=1
```

#### Bulk Usage Tracking
```
POST /api/v1/usage/user/{id}/track/bulk
Content-Type: application/json

{
  "question_banks": 2,
  "questions": 10,
  "ia_uses": 5,
  "lumen_agent_uses": 3
}
```

#### Get Current Usage Metrics
```
GET /api/v1/usage/user/{id}/current
```

#### Get Usage History
```
GET /api/v1/usage/user/{id}/history
```

## Usage Examples

### Track AI Agent Usage
When a user uses an AI agent in your application:

```go
// In your AI service
func ProcessAIRequest(userID string) {
    // ... AI processing logic
    
    // Track usage
    usageService := service.NewUsageTrackingService()
    usageService.TrackIAUsage(userID, 1)
}
```

### Check Subscription Status
Before allowing premium features:

```go
func CheckPremiumAccess(userID string) (bool, error) {
    subscriptionService := service.NewSubscriptionService()
    return subscriptionService.IsUserSubscribed(userID)
}
```

### Process Expired Subscriptions (Cron Job)
```go
func ProcessExpiredSubscriptions() {
    subscriptionService := service.NewSubscriptionService()
    processed, err := subscriptionService.ProcessExpiredSubscriptions()
    if err != nil {
        log.Printf("Error processing expired subscriptions: %v", err)
        return
    }
    log.Printf("Processed %d expired subscriptions", len(processed))
}
```

## Features

### Subscription Management
- ✅ Create, read, update, delete subscriptions
- ✅ Multiple subscription statuses
- ✅ Schedule cancellation at period end
- ✅ Reactivate scheduled cancellations
- ✅ Subscription renewal
- ✅ Bulk expired subscription processing
- ✅ User subscription status checking

### Usage Tracking
- ✅ Track multiple usage types
- ✅ Monthly period-based tracking
- ✅ Atomic increment operations
- ✅ Bulk usage tracking
- ✅ Current and historical usage metrics
- ✅ Aggregated usage reporting
- ✅ Usage summary by period
- ✅ Multiple tracking methods (JSON payload and simple increments)

### Admin Features
- ✅ Subscription statistics
- ✅ Usage analytics
- ✅ Expired subscription processing
- ✅ Usage reset functionality

## Notes

1. **Thread Safety**: All database operations use atomic increments for usage tracking to ensure thread safety.

2. **Period Management**: Usage tracking is organized by monthly periods (YYYY-MM format).

3. **Flexible Tracking**: The system supports both detailed JSON-based tracking and simple increment operations for easy integration.

4. **Validation**: All requests are validated using the existing validation framework.

5. **Error Handling**: Comprehensive error handling with appropriate HTTP status codes.

6. **Scalability**: The system is designed to handle high-frequency usage tracking with minimal performance impact.

This system provides a solid foundation for subscription management and usage tracking that can be easily extended as your application grows.