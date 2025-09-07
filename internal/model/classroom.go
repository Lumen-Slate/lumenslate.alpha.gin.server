package model

import "time"

type Classroom struct {
	ID               string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Name             string    `json:"name" bson:"name" validate:"required"`
	TeacherIDs       []string  `json:"teacherIds" bson:"teacherIds" validate:"required,dive,required"`
	AssignmentIDs    []string  `json:"assignmentIds" bson:"assignmentIds"`
	Credits          int       `json:"credits" bson:"credits"`
	Tags             []string  `json:"tags" bson:"tags"`
	CreatedAt        time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive         bool      `json:"isActive" bson:"isActive"`
	ClassroomCode    string    `json:"classroomCode" bson:"classroomCode"`
	ClassroomSubject *string   `json:"classroomSubject,omitempty" bson:"classroomSubject,omitempty"`
}

// NewClassroom creates a new Classroom with default values
func NewClassroom() *Classroom {
	now := time.Now()
	return &Classroom{
		TeacherIDs:       make([]string, 0),
		AssignmentIDs:    make([]string, 0),
		Tags:             make([]string, 0),
		CreatedAt:        now,
		UpdatedAt:        now,
		IsActive:         true,
		ClassroomCode:    "", // Set code when creating/inviting
		ClassroomSubject: nil,
	}
}
