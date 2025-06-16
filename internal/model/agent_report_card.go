package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AgentReportCard represents the structured report card output from the report card generator agent
type AgentReportCard struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    string             `bson:"userId" json:"userId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`

	// The actual report card data from the agent
	ReportCard AgentReportCardData `bson:"reportCard" json:"reportCard"`
}

// AgentReportCardData matches the agent's ReportCard structure exactly
type AgentReportCardData struct {
	StudentID           string                    `bson:"studentId" json:"student_id"`
	StudentName         string                    `bson:"studentName" json:"student_name"`
	ReportPeriod        string                    `bson:"reportPeriod" json:"report_period"`
	GenerationDate      string                    `bson:"generationDate" json:"generation_date"`
	OverallPerformance  AgentOverallPerformance   `bson:"overallPerformance" json:"overall_performance"`
	SubjectPerformance  []AgentSubjectPerformance `bson:"subjectPerformance" json:"subject_performance"`
	AssignmentSummaries []AgentAssignmentSummary  `bson:"assignmentSummaries" json:"assignment_summaries"`
	AIRemarks           string                    `bson:"aiRemarks" json:"ai_remarks"`
	TeacherRemarks      string                    `bson:"teacherRemarks" json:"teacher_remarks"`
	StudentInsights     AgentStudentInsights      `bson:"studentInsights" json:"student_insights"`
}

// AgentOverallPerformance matches the agent's OverallPerformance structure
type AgentOverallPerformance struct {
	TotalAssignmentsCompleted int     `bson:"totalAssignmentsCompleted" json:"total_assignments_completed"`
	OverallPercentage         float64 `bson:"overallPercentage" json:"overall_percentage"`
	ImprovementTrend          string  `bson:"improvementTrend" json:"improvement_trend"`
	StrongestQuestionType     string  `bson:"strongestQuestionType" json:"strongest_question_type"`
	WeakestQuestionType       string  `bson:"weakestQuestionType" json:"weakest_question_type"`
}

// AgentSubjectPerformance matches the agent's SubjectPerformance structure
type AgentSubjectPerformance struct {
	SubjectName        string   `bson:"subjectName" json:"subject_name"`
	PercentageScore    float64  `bson:"percentageScore" json:"percentage_score"`
	AssignmentCount    int      `bson:"assignmentCount" json:"assignment_count"`
	MCQAccuracy        float64  `bson:"mcqAccuracy" json:"mcq_accuracy"`
	MSQAccuracy        float64  `bson:"msqAccuracy" json:"msq_accuracy"`
	NATAccuracy        float64  `bson:"natAccuracy" json:"nat_accuracy"`
	SubjectiveAvgScore float64  `bson:"subjectiveAvgScore" json:"subjective_avg_score"`
	Strengths          []string `bson:"strengths" json:"strengths"`
	Weaknesses         []string `bson:"weaknesses" json:"weaknesses"`
	ImprovementTrend   string   `bson:"improvementTrend" json:"improvement_trend"`
}

// AgentAssignmentSummary matches the agent's AssignmentSummary structure
type AgentAssignmentSummary struct {
	AssignmentID    string  `bson:"assignmentId" json:"assignment_id"`
	AssignmentTitle string  `bson:"assignmentTitle" json:"assignment_title"`
	PercentageScore float64 `bson:"percentageScore" json:"percentage_score"`
	Subject         string  `bson:"subject" json:"subject"`
}

// AgentStudentInsights matches the agent's StudentInsights structure
type AgentStudentInsights struct {
	KeyStrengths        []string `bson:"keyStrengths" json:"key_strengths"`
	AreasForImprovement []string `bson:"areasForImprovement" json:"areas_for_improvement"`
	RecommendedActions  []string `bson:"recommendedActions" json:"recommended_actions"`
}
