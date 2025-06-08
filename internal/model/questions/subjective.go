package questions

import (
	"time"
)

type Subjective struct {
	ID              string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	BankID          string    `json:"bankId" bson:"bankId" validate:"required"`
	Question        string    `json:"question" bson:"question" validate:"required"`
	VariableIDs     []string  `json:"variableIds" bson:"variableIds" validate:"omitempty"`
	Points          int       `json:"points" bson:"points" validate:"required,min=0"`
	IdealAnswer     *string   `json:"idealAnswer,omitempty" bson:"idealAnswer,omitempty"`
	GradingCriteria []string  `json:"gradingCriteria,omitempty" bson:"gradingCriteria,omitempty"`
	CreatedAt       time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive        bool      `json:"isActive" bson:"isActive"`
}

// NewSubjective creates a new Subjective with default values
func NewSubjective() *Subjective {
	now := time.Now()
	return &Subjective{
		VariableIDs:     make([]string, 0),
		GradingCriteria: make([]string, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
		IsActive:        true,
	}
}
