package model

import "time"

type Student struct {
	ID        string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Name      string    `json:"name" bson:"name" validate:"required"`
	Email     string    `json:"email" bson:"email" validate:"required,email"`
	RollNo    *string   `json:"rollNo" bson:"rollNo"`
	ClassIDs  []string  `json:"classIds" bson:"classIds"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive  bool      `json:"isActive" bson:"isActive"`
}

// NewStudent creates a new Student with default values
func NewStudent() *Student {
	now := time.Now()
	return &Student{
		ClassIDs:  make([]string, 0),
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
}
