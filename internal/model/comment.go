package model

import "time"

type Comment struct {
	ID          string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	CommentBody string    `json:"commentBody" bson:"commentBody" validate:"required"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive    bool      `json:"isActive" bson:"isActive"`
}

// NewComment creates a new Comment with default values
func NewComment() *Comment {
	now := time.Now()
	return &Comment{
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
}
