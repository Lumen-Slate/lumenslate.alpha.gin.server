package questions

import (
	"time"
)

type NAT struct {
	ID          string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	BankID      string    `json:"bankId" bson:"bankId" validate:"required"`
	Question    string    `json:"question" bson:"question" validate:"required,min=3"`
	VariableIDs []string  `json:"variableIds" bson:"variableIds" validate:"omitempty"`
	Points      int       `json:"points" bson:"points" validate:"required,min=1"`
	Answer      float64   `json:"answer" bson:"answer"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive    bool      `json:"isActive" bson:"isActive"`
}

// NewNAT creates a new NAT with default values
func NewNAT() *NAT {
	now := time.Now()
	return &NAT{
		VariableIDs: make([]string, 0),
		Answer:      0.0,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsActive:    true,
	}
}
