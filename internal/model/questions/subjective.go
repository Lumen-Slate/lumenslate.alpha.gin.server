package questions

import "server/internal/model"

type Subjective struct {
	ID              string           `json:"id" firestore:"id"`
	BankID          string           `json:"bankId" firestore:"bankId"`
	Question        string           `json:"question" firestore:"question"`
	Variable        []model.Variable `json:"variable" firestore:"variable"`
	Points          int              `json:"points" firestore:"points"`
	IdealAnswer     *string          `json:"idealAnswer,omitempty" firestore:"idealAnswer,omitempty"`
	GradingCriteria []string         `json:"gradingCriteria,omitempty" firestore:"gradingCriteria,omitempty"`
}
