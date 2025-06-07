package model

import "time"

type Submission struct {
	ID                string              `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	StudentID         string              `json:"studentId" bson:"studentId" validate:"required"`
	AssignmentID      string              `json:"assignmentId" bson:"assignmentId" validate:"required"`
	MCQAnswers        map[string]string   `json:"mcqAnswers,omitempty" bson:"mcqAnswers,omitempty"`
	MSQAnswers        map[string][]string `json:"msqAnswers,omitempty" bson:"msqAnswers,omitempty"`
	NATAnswers        map[string]int      `json:"natAnswers,omitempty" bson:"natAnswers,omitempty"`
	SubjectiveAnswers map[string]string   `json:"subjectiveAnswers" bson:"subjectiveAnswers"`
	CreatedAt         time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time           `json:"updatedAt" bson:"updatedAt"`
	IsActive          bool                `json:"isActive" bson:"isActive"`
}

// NewSubmission creates a new Submission with default values
func NewSubmission() *Submission {
	now := time.Now()
	return &Submission{
		MCQAnswers:        make(map[string]string),
		MSQAnswers:        make(map[string][]string),
		NATAnswers:        make(map[string]int),
		SubjectiveAnswers: make(map[string]string),
		CreatedAt:         now,
		UpdatedAt:         now,
		IsActive:          true,
	}
}
