package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DatabaseName = "lumen_slate"
)

// Collection names
const (
	MCQCollection          = "mcqs"
	MSQCollection          = "msqs"
	QuestionBankCollection = "questionBanks"
	TeacherCollection      = "teachers"
	NATCollection          = "nats"
	SubjectiveCollection   = "subjectives"
	AssignmentCollection   = "assignments"
	ClassroomCollection    = "classrooms"
	CommentCollection      = "comments"
	PostCollection         = "posts"
	StudentCollection      = "students"
	SubmissionCollection   = "submissions"
	VariableCollection     = "variables"
)

// GetCollection returns a reference to the specified collection
func GetCollection(collectionName string) *mongo.Collection {
	return Client.Database(DatabaseName).Collection(collectionName)
}
