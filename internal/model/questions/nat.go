package questions

import (
	"lumenslate/internal/model"
	"time"
)

type NAT struct {
	ID        string           `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	BankID    string           `json:"bankId" bson:"bankId" validate:"required"`
	Question  string           `json:"question" bson:"question" validate:"required,min=3"`
	Variable  []model.Variable `json:"variable" bson:"variable" validate:"required,min=1"`
	Points    int              `json:"points" bson:"points" validate:"required,min=1"`
	Answer    float64          `json:"answer" bson:"answer" validate:"required"`
	CreatedAt time.Time        `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt" bson:"updatedAt"`
	IsActive  bool             `json:"isActive" bson:"isActive"`
}

// NewNAT creates a new NAT with default values
func NewNAT() *NAT {
	now := time.Now()
	return &NAT{
		Variable:  make([]model.Variable, 0),
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
}
