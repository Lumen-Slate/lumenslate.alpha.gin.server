package model

import (
	"time"
)

type QuestionBank struct {
	ID          string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Name        string    `json:"name" bson:"name" validate:"required,min=3,max=100"`
	Topic       string    `json:"topic" bson:"topic" validate:"required"`
	TeacherID   string    `json:"teacherId" bson:"teacherId" validate:"required"`
	Tags        []string  `json:"tags" bson:"tags" validate:"required,min=0"`
	Description string    `json:"description" bson:"description" validate:"omitempty,max=500"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive    bool      `json:"isActive" bson:"isActive"`
}

// NewQuestionBank creates a new QuestionBank with default values
func NewQuestionBank() *QuestionBank {
	now := time.Now()
	return &QuestionBank{
		Tags:      make([]string, 0),
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
}
