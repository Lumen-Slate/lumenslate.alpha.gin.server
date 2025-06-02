package model

type Submission struct {
	ID                string              `json:"id" firestore:"id"`
	StudentID         string              `json:"studentId" firestore:"studentId"`
	AssignmentID      string              `json:"assignmentId" firestore:"assignmentId"`
	MCQAnswers        map[string]string   `json:"mcqAnswers,omitempty" firestore:"mcqAnswers,omitempty"`
	MSQAnswers        map[string][]string `json:"msqAnswers,omitempty" firestore:"msqAnswers,omitempty"`
	NATAnswers        map[string]int      `json:"natAnswers,omitempty" firestore:"natAnswers,omitempty"`
	SubjectiveAnswers map[string]string   `json:"subjectiveAnswers" firestore:"subjectiveAnswers"`
}
