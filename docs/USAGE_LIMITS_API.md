# Usage Limits API Documentation

## Overview

The Usage Limits API manages subscription plan limits for the Lumen Slate platform. It defines restrictions on various features like teachers, classrooms, questions, and AI agent usage for different subscription tiers.

## Schema

Usage limits support flexible value types:
- **Integer**: Specific numeric limits (e.g., `10`, `100`)
- **String**: Special values like `"unlimited"` or `"custom"`
- **-1**: Represents unlimited usage (equivalent to `"unlimited"`)

### Usage Limit Value Types

| Value Type | Example | Description |
|------------|---------|-------------|
| Integer | `10`, `100`, `1000` | Specific numeric limit |
| "unlimited" | `"unlimited"` | No restrictions |
| "custom" | `"custom"` | Negotiated/custom limits |
| -1 | `-1` | Unlimited (numeric equivalent) |

## API Endpoints

### Base URL
```
/api/v1/usage-limits
```

### 1. Create Usage Limits

**POST** `/api/v1/usage-limits`

Creates new usage limits for a subscription plan.

**Request Body:**
```json
{
  "plan_name": "premium",
  "teachers": 25,
  "classrooms": "unlimited",
  "students_per_classroom": 50,
  "question_banks": "unlimited",
  "questions": "unlimited",
  "assignment_exports_per_day": "unlimited",
  "ai": {
    "independent_agent": 500,
    "lumen_agent": 300,
    "rag_agent": 150,
    "rag_document_uploads": 100
  }
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "plan_name": "premium",
  "teachers": 25,
  "classrooms": "unlimited",
  "students_per_classroom": 50,
  "question_banks": "unlimited",
  "questions": "unlimited",
  "assignment_exports_per_day": "unlimited",
  "ai": {
    "independent_agent": 500,
    "lumen_agent": 300,
    "rag_agent": 150,
    "rag_document_uploads": 100
  },
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "is_active": true
}
```

### 2. Get Usage Limits by ID

**GET** `/api/v1/usage-limits/{id}`

Retrieves usage limits by their ID.

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "plan_name": "premium",
  "teachers": 25,
  "classrooms": "unlimited",
  // ... other fields
}
```

### 3. Get Usage Limits by Plan Name

**GET** `/api/v1/usage-limits/plan/{planName}`

Retrieves usage limits by plan name.

**Example:** `GET /api/v1/usage-limits/plan/premium`

### 4. Get All Usage Limits

**GET** `/api/v1/usage-limits`

Retrieves all usage limits with optional filtering.

**Query Parameters:**
- `plan_name` (string): Filter by plan name
- `is_active` (boolean): Filter by active status
- `limit` (string): Pagination limit
- `offset` (string): Pagination offset

**Example:** `GET /api/v1/usage-limits?is_active=true&limit=10`

### 5. Update Usage Limits

**PUT** `/api/v1/usage-limits/{id}`

Updates existing usage limits completely.

**Request Body:**
```json
{
  "plan_name": "premium-updated",
  "teachers": 30,
  "classrooms": "unlimited",
  "students_per_classroom": 60,
  "question_banks": "unlimited",
  "questions": "unlimited",
  "assignment_exports_per_day": "unlimited",
  "ai": {
    "independent_agent": 600,
    "lumen_agent": 400,
    "rag_agent": 200,
    "rag_document_uploads": 150
  }
}
```

### 6. Patch Usage Limits

**PATCH** `/api/v1/usage-limits/{id}`

Performs partial updates on usage limits.

**Request Body:**
```json
{
  "teachers": 35,
  "ai": {
    "independent_agent": 700
  }
}
```

### 7. Delete Usage Limits

**DELETE** `/api/v1/usage-limits/{id}`

Permanently deletes usage limits.

**Response (200 OK):**
```json
{
  "message": "Usage limits deleted successfully"
}
```

### 8. Soft Delete (Deactivate) Usage Limits

**POST** `/api/v1/usage-limits/{id}/deactivate`

Marks usage limits as inactive instead of deleting them.

### 9. Get Usage Limits Statistics

**GET** `/api/v1/usage-limits/stats`

Returns statistical information about usage limits.

**Response (200 OK):**
```json
{
  "total_usage_limits": 15,
  "active_usage_limits": 12,
  "inactive_usage_limits": 3
}
```

### 10. Check User Usage Against Limits

**GET** `/api/v1/usage-limits/check/{userId}?planName={planName}`

Checks if a user's current usage exceeds their plan limits.

**Response (200 OK):**
```json
{
  "plan_name": "premium",
  "limits": {
    "teachers": 25,
    "classrooms": "unlimited",
    // ... all limits
  },
  "usage": {
    "teachers_used": 15,
    "classrooms_used": 45,
    "question_banks_used": 120,
    "questions_used": 5000,
    "assignment_exports_today": 25,
    "ai_independent_agent_used": 350,
    "ai_lumen_agent_used": 200,
    "ai_rag_agent_used": 100,
    "ai_rag_documents_uploaded": 75
  },
  "within_limits": true,
  "exceeded_limits": []
}
```

## Admin Endpoints

### Initialize Default Usage Limits

**POST** `/api/v1/admin/usage-limits/initialize-defaults`

Creates default usage limits for common plans (basic, premium, enterprise).

**Response (200 OK):**
```json
{
  "message": "Default usage limits initialized successfully"
}
```

## Default Plans

The system comes with three pre-configured plans:

### Basic Plan
- **Teachers:** 5
- **Classrooms:** 10
- **Students per Classroom:** 30
- **Question Banks:** 50
- **Questions:** 1000
- **Assignment Exports per Day:** 10
- **AI Independent Agent:** 100 calls
- **AI Lumen Agent:** 50 calls
- **AI RAG Agent:** 25 calls
- **AI RAG Document Uploads:** 10

### Premium Plan
- **Teachers:** 25
- **Classrooms:** Unlimited
- **Students per Classroom:** 50
- **Question Banks:** Unlimited
- **Questions:** Unlimited
- **Assignment Exports per Day:** Unlimited
- **AI Independent Agent:** 500 calls
- **AI Lumen Agent:** 300 calls
- **AI RAG Agent:** 150 calls
- **AI RAG Document Uploads:** 100

### Enterprise Plan
- **Teachers:** Custom
- **Classrooms:** Unlimited
- **Students per Classroom:** Custom
- **Question Banks:** Unlimited
- **Questions:** Unlimited
- **Assignment Exports per Day:** Unlimited
- **AI Independent Agent:** Unlimited
- **AI Lumen Agent:** Unlimited
- **AI RAG Agent:** Unlimited
- **AI RAG Document Uploads:** Unlimited

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid teachers limit value"
}
```

### 404 Not Found
```json
{
  "error": "Usage limits not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to create usage limits"
}
```

## Integration with Subscriptions

Usage limits are designed to work with the subscription system:

1. When a user subscribes to a plan, fetch the corresponding usage limits
2. Use the usage limits to enforce restrictions in your application
3. Check current usage against limits before allowing actions
4. Update usage tracking when users perform actions

## Example Usage

### Frontend Integration

```javascript
// Fetch usage limits for a plan
const usageLimits = await fetch('/api/v1/usage-limits/plan/premium');

// Check if user can create more classrooms
if (usageLimits.classrooms !== 'unlimited' && 
    userCurrentClassrooms >= usageLimits.classrooms) {
    alert('Classroom limit reached for your plan');
}

// Check current usage vs limits
const usageCheck = await fetch(`/api/v1/usage-limits/check/${userId}?planName=premium`);
if (!usageCheck.within_limits) {
    console.log('Exceeded limits:', usageCheck.exceeded_limits);
}
```

### Backend Integration

```go
// Enforce limits in your controllers
func CreateClassroom(c *gin.Context) {
    userID := getUserID(c)
    userPlan := getUserPlan(userID)
    
    // Check usage limits
    usageCheck, err := usageLimitsService.CheckUserUsageAgainstLimits(userID, userPlan)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to check limits"})
        return
    }
    
    if !usageCheck["within_limits"].(bool) {
        c.JSON(403, gin.H{"error": "Plan limit exceeded"})
        return
    }
    
    // Proceed with classroom creation
    // ...
}
```

## Database Schema

The usage limits are stored in MongoDB with the following structure:

```javascript
{
  _id: ObjectId,
  plan_name: String,
  teachers: Mixed,              // Integer or String
  classrooms: Mixed,            // Integer or String
  students_per_classroom: Mixed, // Integer or String
  question_banks: Mixed,        // Integer or String
  questions: Mixed,             // Integer or String
  assignment_exports_per_day: Mixed, // Integer or String
  ai: {
    independent_agent: Mixed,    // Integer or String
    lumen_agent: Mixed,         // Integer or String
    rag_agent: Mixed,           // Integer or String
    rag_document_uploads: Mixed  // Integer or String
  },
  created_at: Date,
  updated_at: Date,
  is_active: Boolean
}
```

## Testing

Use the provided test files in `/examples/` to test the API:

- `usage_limits_basic.json` - Basic plan example
- `usage_limits_premium.json` - Premium plan example
- `usage_limits_enterprise.json` - Enterprise plan example
- `usage_limits_update.json` - Update example

```bash
# Create usage limits
curl -X POST http://localhost:8080/api/v1/usage-limits \
  -H "Content-Type: application/json" \
  -d @examples/usage_limits_premium.json

# Get all usage limits
curl http://localhost:8080/api/v1/usage-limits

# Update usage limits
curl -X PATCH http://localhost:8080/api/v1/usage-limits/{id} \
  -H "Content-Type: application/json" \
  -d @examples/usage_limits_update.json
```