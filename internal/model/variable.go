package model

type Variable struct {
	ID             string `json:"id" firestore:"id"`
	Name           string `json:"name" firestore:"name"`
	NamePositions  []int  `json:"namePositions" firestore:"namePositions"`
	Value          string `json:"value" firestore:"value"`
	ValuePositions []int  `json:"valuePositions" firestore:"valuePositions"`
	VariableType   string `json:"variableType" firestore:"variableType"`
}
