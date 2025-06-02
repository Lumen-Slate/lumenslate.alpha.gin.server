package model

type Comment struct {
	ID          string `json:"id" firestore:"id"`
	CommentBody string `json:"commentBody" firestore:"commentBody"`
}
