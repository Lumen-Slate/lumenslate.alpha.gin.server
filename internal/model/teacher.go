package model

type Teacher struct {
	ID    string `json:"id" firestore:"id"`
	Name  string `json:"name" firestore:"name"`
	Email string `json:"email" firestore:"email"`
	Phone string `json:"phone" firestore:"phone"`
}
