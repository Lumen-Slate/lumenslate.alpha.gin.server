package model

import (
	"time"
)

type Teacher struct {
	ID        string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Name      string    `json:"name" bson:"name" validate:"required,min=3,max=100"`
	Email     string    `json:"email" bson:"email" validate:"required,email"`
	Phone     string    `json:"phone" bson:"phone" validate:"required"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive  bool      `json:"isActive" bson:"isActive"`
}

// NewTeacher creates a new Teacher with default values
func NewTeacher() *Teacher {
	now := time.Now()
	return &Teacher{
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
}
