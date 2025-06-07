package model

import "time"

type Post struct {
	ID          string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Title       string    `json:"title" bson:"title" validate:"required"`
	Body        string    `json:"body" bson:"body" validate:"required"`
	Attachments []string  `json:"attachments" bson:"attachments"`
	UserID      string    `json:"userId" bson:"userId" validate:"required"`
	CommentIDs  []string  `json:"commentIds" bson:"commentIds"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive    bool      `json:"isActive" bson:"isActive"`
}

// NewPost creates a new Post with default values
func NewPost() *Post {
	now := time.Now()
	return &Post{
		Attachments: make([]string, 0),
		CommentIDs:  make([]string, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
		IsActive:    true,
	}
}
