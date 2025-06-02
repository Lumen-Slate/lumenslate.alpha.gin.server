package model

type Classroom struct {
	ID            string   `json:"id" firestore:"id"`
	Subject       string   `json:"subject" firestore:"subject"`
	TeacherIDs    []string `json:"teacherIds" firestore:"teacherIds"`
	AssignmentIDs []string `json:"assignmentIds" firestore:"assignmentIds"`
	Credits       int      `json:"credits" firestore:"credits"`
	Tags          []string `json:"tags" firestore:"tags"` // Added tags field
}
