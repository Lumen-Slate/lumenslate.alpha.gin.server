package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AssignmentResult represents the result of a student's assignment submission
type AssignmentResult struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AssignmentID       string             `bson:"assignmentId" json:"assignmentId"`
	StudentID          string             `bson:"studentId" json:"studentId"`
	TotalPointsAwarded int                `bson:"totalPointsAwarded" json:"totalPointsAwarded"`
	TotalMaxPoints     int                `bson:"totalMaxPoints" json:"totalMaxPoints"`
	PercentageScore    float64            `bson:"percentageScore" json:"percentageScore"`
	MCQResults         []MCQResult        `bson:"mcqResults" json:"mcqResults"`
	MSQResults         []MSQResult        `bson:"msqResults" json:"msqResults"`
	NATResults         []NATResult        `bson:"natResults" json:"natResults"`
	SubjectiveResults  []SubjectiveResult `bson:"subjectiveResults" json:"subjectiveResults"`
	CreatedAt          time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// MCQResult represents the result of a multiple choice question
type MCQResult struct {
	QuestionID    string `bson:"questionId" json:"questionId"`
	StudentAnswer int    `bson:"studentAnswer" json:"studentAnswer"`
	CorrectAnswer int    `bson:"correctAnswer" json:"correctAnswer"`
	PointsAwarded int    `bson:"pointsAwarded" json:"pointsAwarded"`
	MaxPoints     int    `bson:"maxPoints" json:"maxPoints"`
	IsCorrect     bool   `bson:"isCorrect" json:"isCorrect"`
}

// MSQResult represents the result of a multiple select question
type MSQResult struct {
	QuestionID     string `bson:"questionId" json:"questionId"`
	StudentAnswers []int  `bson:"studentAnswers" json:"studentAnswers"`
	CorrectAnswers []int  `bson:"correctAnswers" json:"correctAnswers"`
	PointsAwarded  int    `bson:"pointsAwarded" json:"pointsAwarded"`
	MaxPoints      int    `bson:"maxPoints" json:"maxPoints"`
	IsCorrect      bool   `bson:"isCorrect" json:"isCorrect"`
}

// NATResult represents the result of a numerical answer type question
type NATResult struct {
	QuestionID    string      `bson:"questionId" json:"questionId"`
	StudentAnswer interface{} `bson:"studentAnswer" json:"studentAnswer"` // can be int or float
	CorrectAnswer interface{} `bson:"correctAnswer" json:"correctAnswer"` // can be int or float
	PointsAwarded int         `bson:"pointsAwarded" json:"pointsAwarded"`
	MaxPoints     int         `bson:"maxPoints" json:"maxPoints"`
	IsCorrect     bool        `bson:"isCorrect" json:"isCorrect"`
}

// SubjectiveResult represents the result of a subjective question
type SubjectiveResult struct {
	QuestionID         string   `bson:"questionId" json:"questionId"`
	StudentAnswer      string   `bson:"studentAnswer" json:"studentAnswer"`
	IdealAnswer        string   `bson:"idealAnswer" json:"idealAnswer"`
	GradingCriteria    []string `bson:"gradingCriteria" json:"gradingCriteria"`
	PointsAwarded      int      `bson:"pointsAwarded" json:"pointsAwarded"`
	MaxPoints          int      `bson:"maxPoints" json:"maxPoints"`
	AssessmentFeedback string   `bson:"assessmentFeedback" json:"assessmentFeedback"`
	CriteriaMet        []string `bson:"criteriaMet" json:"criteriaMet"`
	CriteriaMissed     []string `bson:"criteriaMissed" json:"criteriaMissed"`
}
