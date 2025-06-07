package model

import "time"

type Variable struct {
	ID             string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Name           string    `json:"name" bson:"name" validate:"required"`
	NamePositions  []int     `json:"namePositions" bson:"namePositions" validate:"required"`
	Value          string    `json:"value" bson:"value" validate:"required"`
	ValuePositions []int     `json:"valuePositions" bson:"valuePositions" validate:"required"`
	VariableType   string    `json:"variableType" bson:"variableType" validate:"required"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive       bool      `json:"isActive" bson:"isActive"`
}

// NewVariable creates a new Variable with default values
func NewVariable() *Variable {
	now := time.Now()
	return &Variable{
		NamePositions:  make([]int, 0),
		ValuePositions: make([]int, 0),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsActive:       true,
	}
}
