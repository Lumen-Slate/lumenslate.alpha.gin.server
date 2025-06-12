package questions

import (
	"time"
)

type MCQ struct {
	ID          string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	BankID      string    `json:"bankId" bson:"bankId" validate:"required"`
	Question    string    `json:"question" bson:"question" validate:"required,min=3"`
	VariableIDs []string  `json:"variableIds" bson:"variableIds" validate:"omitempty"`
	Points      int       `json:"points" bson:"points" validate:"required,min=1"`
	Options     []string  `json:"options" bson:"options" validate:"required,min=2"`
	AnswerIndex int       `json:"answerIndex" bson:"answerIndex" validate:"min=0"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive    bool      `json:"isActive" bson:"isActive"`
}

// NewMCQ creates a new MCQ with default values
func NewMCQ() *MCQ {
	now := time.Now()
	return &MCQ{
		VariableIDs: make([]string, 0),
		Options:     make([]string, 0),
		AnswerIndex: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsActive:    true,
	}
}
