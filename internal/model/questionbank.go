package model

type QuestionBank struct {
	ID        string   `json:"id" firestore:"id"`
	Name      string   `json:"name" firestore:"name"`
	Topic     string   `json:"topic" firestore:"topic"`
	TeacherID string   `json:"teacherId" firestore:"teacherId"`
	Tags      []string `json:"tags" firestore:"tags"` // Added tags field
}
