package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UsageLimitValue represents a usage limit that can be either an integer or a string (like "unlimited", "custom")
type UsageLimitValue interface{}

// AILimits represents AI-related usage limits
type AILimits struct {
	IndependentAgent   UsageLimitValue `json:"independent_agent" bson:"independent_agent" validate:"required"`
	LumenAgent         UsageLimitValue `json:"lumen_agent" bson:"lumen_agent" validate:"required"`
	RAGAgent           UsageLimitValue `json:"rag_agent" bson:"rag_agent" validate:"required"`
	RAGDocumentUploads UsageLimitValue `json:"rag_document_uploads" bson:"rag_document_uploads" validate:"required"`
}

// UsageLimits represents the usage limits for a subscription plan
type UsageLimits struct {
	ID                      primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	PlanName                string             `json:"plan_name" bson:"plan_name" validate:"required"`
	Teachers                UsageLimitValue    `json:"teachers" bson:"teachers" validate:"required"`
	Classrooms              UsageLimitValue    `json:"classrooms" bson:"classrooms" validate:"required"`
	StudentsPerClassroom    UsageLimitValue    `json:"students_per_classroom" bson:"students_per_classroom" validate:"required"`
	QuestionBanks           UsageLimitValue    `json:"question_banks" bson:"question_banks" validate:"required"`
	Questions               UsageLimitValue    `json:"questions" bson:"questions" validate:"required"`
	AssignmentExportsPerDay UsageLimitValue    `json:"assignment_exports_per_day" bson:"assignment_exports_per_day" validate:"required"`
	AI                      AILimits           `json:"ai" bson:"ai" validate:"required"`
	CreatedAt               time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt               time.Time          `json:"updated_at" bson:"updated_at"`
	IsActive                bool               `json:"is_active" bson:"is_active"`
}

// UsageLimitType represents the different types of usage limits
type UsageLimitType string

const (
	UsageLimitTypeInteger   UsageLimitType = "integer"
	UsageLimitTypeUnlimited UsageLimitType = "unlimited"
	UsageLimitTypeCustom    UsageLimitType = "custom"
)

// Common usage limit values
const (
	UnlimitedValue = "unlimited"
	CustomValue    = "custom"
	UnlimitedInt   = -1
)

// NewUsageLimits creates a new UsageLimits instance with default values
func NewUsageLimits() *UsageLimits {
	now := time.Now()
	return &UsageLimits{
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
		AI: AILimits{
			IndependentAgent:   0,
			LumenAgent:         0,
			RAGAgent:           0,
			RAGDocumentUploads: 0,
		},
	}
}

// IsUnlimited checks if a usage limit value represents unlimited usage
func IsUnlimited(value UsageLimitValue) bool {
	switch v := value.(type) {
	case string:
		return v == UnlimitedValue || v == CustomValue
	case int:
		return v == UnlimitedInt
	case int64:
		return v == int64(UnlimitedInt)
	case float64:
		return v == float64(UnlimitedInt)
	default:
		return false
	}
}

// GetIntValue safely converts a UsageLimitValue to int, returning -1 for unlimited/custom
func GetIntValue(value UsageLimitValue) int {
	switch v := value.(type) {
	case string:
		if v == UnlimitedValue || v == CustomValue {
			return UnlimitedInt
		}
		return 0 // Invalid string value
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// ValidateUsageLimitValue validates that a usage limit value is either a valid integer or allowed string
func ValidateUsageLimitValue(value UsageLimitValue) bool {
	switch v := value.(type) {
	case string:
		return v == UnlimitedValue || v == CustomValue
	case int, int64, float64:
		return true
	default:
		return false
	}
}

// UsageLimitsFilter represents filters for querying usage limits
type UsageLimitsFilter struct {
	PlanName string `json:"plan_name" form:"plan_name"`
	IsActive *bool  `json:"is_active" form:"is_active"`
	Limit    string `json:"limit" form:"limit"`
	Offset   string `json:"offset" form:"offset"`
}
