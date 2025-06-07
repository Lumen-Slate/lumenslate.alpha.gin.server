package questions

import (
	"time"
)

type MSQ struct {
	ID            string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	BankID        string    `json:"bankId" bson:"bankId" validate:"required"`
	Question      string    `json:"question" bson:"question" validate:"required,min=3"`
	VariableIDs   []string  `json:"variableIds" bson:"variableIds" validate:"omitempty"`
	Points        int       `json:"points" bson:"points" validate:"required,min=1"`
	Options       []string  `json:"options" bson:"options" validate:"required,min=2"`
	AnswerIndices []int     `json:"answerIndices" bson:"answerIndices" validate:"required,min=1"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
	IsActive      bool      `json:"isActive" bson:"isActive"`
}

// NewMSQ creates a new MSQ with default values
func NewMSQ() *MSQ {
	now := time.Now()
	return &MSQ{
		VariableIDs:   make([]string, 0),
		Options:       make([]string, 0),
		AnswerIndices: make([]int, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
		IsActive:      true,
	}
}
