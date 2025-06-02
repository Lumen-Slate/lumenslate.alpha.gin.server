package questions

type MCQ struct {
	ID          string   `json:"id" firestore:"id"`
	BankID      string   `json:"bankId" firestore:"bankId"`
	Question    string   `json:"question" firestore:"question"`
	VariableIDs []string `json:"variableIds" firestore:"variableIds"`
	Points      int      `json:"points" firestore:"points"`
	Options     []string `json:"options" firestore:"options"`
	AnswerIndex int      `json:"answerIndex" firestore:"answerIndex"`
}
