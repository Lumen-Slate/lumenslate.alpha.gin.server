package model

import "time"

type MCQ struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answer   int      `json:"answer"`
}

type MSQ struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answers  []int    `json:"answers"`
}

type NAT struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type Subjective struct {
	Question string `json:"question"`
}

type Assignment struct {
	ID          string       `json:"id" firestore:"id"`
	Title       string       `json:"title" firestore:"title"`
	Body        string       `json:"body" firestore:"body"`
	DueDate     time.Time    `json:"dueDate" firestore:"dueDate"`
	CreatedAt   time.Time    `json:"createdAt" firestore:"createdAt"`
	Points      int          `json:"points" firestore:"points"`
	CommentIds  []string     `json:"commentIds" firestore:"commentIds"`
	MCQs        []MCQ        `json:"mcqs,omitempty" firestore:"mcqs,omitempty"`
	MSQs        []MSQ        `json:"msqs,omitempty" firestore:"msqs,omitempty"`
	NATs        []NAT        `json:"nats,omitempty" firestore:"nats,omitempty"`
	Subjectives []Subjective `json:"subjectives,omitempty" firestore:"subjectives,omitempty"`
}
