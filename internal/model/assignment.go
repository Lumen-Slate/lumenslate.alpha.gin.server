package model

import (
	"time"
)

type Assignment struct {
	ID            string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Title         string    `json:"title" bson:"title" validate:"required"`
	Body          string    `json:"body" bson:"body" validate:"required"`
	DueDate       time.Time `json:"dueDate" bson:"dueDate" validate:"required"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
	Points        int       `json:"points" bson:"points" validate:"required,min=0"`
	CommentIds    []string  `json:"commentIds" bson:"commentIds"`
	MCQIds        []string  `json:"mcqIds" bson:"mcqIds"`
	MSQIds        []string  `json:"msqIds" bson:"msqIds"`
	NATIds        []string  `json:"natIds" bson:"natIds"`
	SubjectiveIds []string  `json:"subjectiveIds" bson:"subjectiveIds"`
	IsActive      bool      `json:"isActive" bson:"isActive"`
}

// NewAssignment creates a new Assignment with default values
func NewAssignment() *Assignment {
	now := time.Now()
	return &Assignment{
		CommentIds:    make([]string, 0),
		MCQIds:        make([]string, 0),
		MSQIds:        make([]string, 0),
		NATIds:        make([]string, 0),
		SubjectiveIds: make([]string, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
		IsActive:      true,
	}
}
