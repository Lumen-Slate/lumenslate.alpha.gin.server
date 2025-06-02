package model

type Post struct {
	ID          string   `json:"id" firestore:"id"`
	Title       string   `json:"title" firestore:"title"`
	Body        string   `json:"body" firestore:"body"`
	Attachments []string `json:"attachments" firestore:"attachments"`
	UserID      string   `json:"userId" firestore:"userId"`
	CommentIDs  []string `json:"commentIds" firestore:"commentIds"`
}
