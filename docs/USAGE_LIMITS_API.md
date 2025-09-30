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
  "plan_name": "private_tutor",
  "teachers": 5,
  "classrooms": 2,
  "students_per_classroom": 40,
  "question_banks": 10,
  "questions": 500,
  "assignment_exports_per_day": "unlimited",
  "ai": {
    "independent_agent": 100,
    "lumen_agent": 100,
    "rag_agent": 100,
    "rag_document_uploads": 15
  }
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "plan_name": "private_tutor",
  "teachers": 5,
  "classrooms": 2,
  "students_per_classroom": 40,
  "question_banks": 10,
  "questions": 500,
  "assignment_exports_per_day": "unlimited",
  "ai": {
    "independent_agent": 100,
    "lumen_agent": 100,
    "rag_agent": 100,
    "rag_document_uploads": 15
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
  "plan_name": "private_tutor",
  "teachers": 5,
  "classrooms": 2,
  // ... other fields
}
```

### 3. Get Usage Limits by Plan Name

**GET** `/api/v1/usage-limits/plan/{planName}`

Retrieves usage limits by plan name.

**Example:** `GET /api/v1/usage-limits/plan/private_tutor`

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
  "plan_name": "private_tutor-updated",
  "teachers": 8,
  "classrooms": 3,
  "students_per_classroom": 50,
  "question_banks": 15,
  "questions": 750,
  "assignment_exports_per_day": "unlimited",
  "ai": {
    "independent_agent": 150,
    "lumen_agent": 120,
    "rag_agent": 120,
    "rag_document_uploads": 20
  }
}
```

### 6. Patch Usage Limits

**PATCH** `/api/v1/usage-limits/{id}`

Performs partial updates on usage limits.

**Request Body:**
```json
{
  "teachers": 6,
  "ai": {
    "independent_agent": 120
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
  "plan_name": "private_tutor",
  "limits": {
    "teachers": 5,
    "classrooms": 2,
    "students_per_classroom": 40,
    "question_banks": 10,
    "questions": 500,
    "assignment_exports_per_day": "unlimited",
    "ai": {
      "independent_agent": 100,
      "lumen_agent": 100,
      "rag_agent": 100,
      "rag_document_uploads": 15
    }
  },
  "usage": {
    "teachers_used": 3,
    "classrooms_used": 2,
    "question_banks_used": 8,
    "questions_used": 350,
    "assignment_exports_today": 25,
    "ai_independent_agent_used": 75,
    "ai_lumen_agent_used": 80,
    "ai_rag_agent_used": 60,
    "ai_rag_documents_uploaded": 12
  },
  "within_limits": true,
  "exceeded_limits": []
}
```

## Admin Endpoints

### Initialize Default Usage Limits

**POST** `/api/v1/admin/usage-limits/initialize-defaults`

Creates default usage limits for common plans (free, private_tutor, multi_classroom, enterprise_b2b).

**Response (200 OK):**
```json
{
  "message": "Default usage limits initialized successfully"
}
```

## Default Plans

The system comes with four pre-configured plans:

### Free Plan
- **Teachers:** 1
- **Classrooms:** 0
- **Students per Classroom:** 0
- **Question Banks:** 1
- **Questions:** 30
- **Assignment Exports per Day:** 5
- **AI Independent Agent:** 10 calls
- **AI Lumen Agent:** 10 calls
- **AI RAG Agent:** 5 calls
- **AI RAG Document Uploads:** 1

### Private Tutor Plan
- **Teachers:** 5
- **Classrooms:** 2
- **Students per Classroom:** 40
- **Question Banks:** 10
- **Questions:** 500
- **Assignment Exports per Day:** Unlimited
- **AI Independent Agent:** 100 calls
- **AI Lumen Agent:** 100 calls
- **AI RAG Agent:** 100 calls
- **AI RAG Document Uploads:** 15

### Multi-Classroom Plan
- **Teachers:** 10
- **Classrooms:** 5
- **Students per Classroom:** 60
- **Question Banks:** 25
- **Questions:** 1500
- **Assignment Exports per Day:** Unlimited
- **AI Independent Agent:** 250 calls
- **AI Lumen Agent:** 250 calls
- **AI RAG Agent:** 250 calls
- **AI RAG Document Uploads:** 50

### Enterprise B2B Plan
- **Teachers:** Custom
- **Classrooms:** Unlimited
- **Students per Classroom:** Custom
- **Question Banks:** Unlimited
- **Questions:** Unlimited
- **Assignment Exports per Day:** Unlimited
- **AI Independent Agent:** Custom
- **AI Lumen Agent:** Custom
- **AI RAG Agent:** Custom
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
const usageLimits = await fetch('/api/v1/usage-limits/plan/private_tutor');

// Check if user can create more classrooms
if (usageLimits.classrooms !== 'unlimited' && 
    userCurrentClassrooms >= usageLimits.classrooms) {
    alert('Classroom limit reached for your plan');
}

// Check current usage vs limits
const usageCheck = await fetch(`/api/v1/usage-limits/check/${userId}?planName=private_tutor`);
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

- `usage_limits_free.json` - Free plan example
- `usage_limits_private_tutor.json` - Private tutor plan example
- `usage_limits_multi_classroom.json` - Multi-classroom plan example
- `usage_limits_enterprise_b2b.json` - Enterprise B2B plan example
- `usage_limits_update.json` - Update example

```bash
# Create usage limits
curl -X POST http://localhost:8080/api/v1/usage-limits \
  -H "Content-Type: application/json" \
  -d @examples/usage_limits_private_tutor.json

# Get all usage limits
curl http://localhost:8080/api/v1/usage-limits

# Update usage limits
curl -X PATCH http://localhost:8080/api/v1/usage-limits/{id} \
  -H "Content-Type: application/json" \
  -d @examples/usage_limits_update.json
```