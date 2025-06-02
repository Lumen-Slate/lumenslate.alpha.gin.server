package model

type Student struct {
	ID       string   `json:"id" firestore:"id"`
	Name     string   `json:"name" firestore:"name"`
	Email    string   `json:"email" firestore:"email"`
	RollNo   string   `json:"rollNo" firestore:"rollNo"`
	ClassIDs []string `json:"classIds" firestore:"classIds"`
}
