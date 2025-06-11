package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AssignmentResult represents the result of a student's assignment submission
type AssignmentResult struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AssignmentID       string             `bson:"assignmentId" json:"assignment_id"`
	StudentID          string             `bson:"studentId" json:"student_id"`
	TotalPointsAwarded int                `bson:"totalPointsAwarded" json:"total_points_awarded"`
	TotalMaxPoints     int                `bson:"totalMaxPoints" json:"total_max_points"`
	PercentageScore    float64            `bson:"percentageScore" json:"percentage_score"`
	MCQResults         []MCQResult        `bson:"mcqResults" json:"mcq_results"`
	MSQResults         []MSQResult        `bson:"msqResults" json:"msq_results"`
	NATResults         []NATResult        `bson:"natResults" json:"nat_results"`
	SubjectiveResults  []SubjectiveResult `bson:"subjectiveResults" json:"subjective_results"`
	CreatedAt          time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// MCQResult represents the result of a multiple choice question
type MCQResult struct {
	QuestionID    string `bson:"questionId" json:"question_id"`
	StudentAnswer int    `bson:"studentAnswer" json:"student_answer"`
	CorrectAnswer int    `bson:"correctAnswer" json:"correct_answer"`
	PointsAwarded int    `bson:"pointsAwarded" json:"points_awarded"`
	MaxPoints     int    `bson:"maxPoints" json:"max_points"`
	IsCorrect     bool   `bson:"isCorrect" json:"is_correct"`
}

// MSQResult represents the result of a multiple select question
type MSQResult struct {
	QuestionID     string `bson:"questionId" json:"question_id"`
	StudentAnswers []int  `bson:"studentAnswers" json:"student_answers"`
	CorrectAnswers []int  `bson:"correctAnswers" json:"correct_answers"`
	PointsAwarded  int    `bson:"pointsAwarded" json:"points_awarded"`
	MaxPoints      int    `bson:"maxPoints" json:"max_points"`
	IsCorrect      bool   `bson:"isCorrect" json:"is_correct"`
}

// NATResult represents the result of a numerical answer type question
type NATResult struct {
	QuestionID    string      `bson:"questionId" json:"question_id"`
	StudentAnswer interface{} `bson:"studentAnswer" json:"student_answer"` // can be int or float
	CorrectAnswer interface{} `bson:"correctAnswer" json:"correct_answer"` // can be int or float
	PointsAwarded int         `bson:"pointsAwarded" json:"points_awarded"`
	MaxPoints     int         `bson:"maxPoints" json:"max_points"`
	IsCorrect     bool        `bson:"isCorrect" json:"is_correct"`
}

// SubjectiveResult represents the result of a subjective question
type SubjectiveResult struct {
	QuestionID         string   `bson:"questionId" json:"question_id"`
	StudentAnswer      string   `bson:"studentAnswer" json:"student_answer"`
	IdealAnswer        string   `bson:"idealAnswer" json:"ideal_answer"`
	GradingCriteria    []string `bson:"gradingCriteria" json:"grading_criteria"`
	PointsAwarded      int      `bson:"pointsAwarded" json:"points_awarded"`
	MaxPoints          int      `bson:"maxPoints" json:"max_points"`
	AssessmentFeedback string   `bson:"assessmentFeedback" json:"assessment_feedback"`
	CriteriaMet        []string `bson:"criteriaMet" json:"criteria_met"`
	CriteriaMissed     []string `bson:"criteriaMissed" json:"criteria_missed"`
}
