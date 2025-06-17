package model

import (
	"time"
)

type ReportCard struct {
	ID           string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	UserID       string    `json:"userId" bson:"userId" validate:"required"`
	StudentID    int       `json:"studentId" bson:"studentId" validate:"required"`
	StudentName  string    `json:"studentName" bson:"studentName" validate:"required"`
	AcademicTerm string    `json:"academicTerm" bson:"academicTerm" validate:"required"`
	GeneratedAt  time.Time `json:"generatedAt" bson:"generatedAt"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`

	// Overall Academic Performance
	OverallGPA           *float64 `json:"overallGpa,omitempty" bson:"overallGpa"`
	OverallGrade         *string  `json:"overallGrade,omitempty" bson:"overallGrade"`
	OverallPercentage    *float64 `json:"overallPercentage,omitempty" bson:"overallPercentage"`
	ClassRank            *int     `json:"classRank,omitempty" bson:"classRank"`
	TotalStudentsInClass *int     `json:"totalStudentsInClass,omitempty" bson:"totalStudentsInClass"`

	// Subject-wise Performance Summary
	SubjectsCount         *int     `json:"subjectsCount,omitempty" bson:"subjectsCount"`
	HighestSubjectScore   *int     `json:"highestSubjectScore,omitempty" bson:"highestSubjectScore"`
	LowestSubjectScore    *int     `json:"lowestSubjectScore,omitempty" bson:"lowestSubjectScore"`
	AverageSubjectScore   *float64 `json:"averageSubjectScore,omitempty" bson:"averageSubjectScore"`
	BestPerformingSubject *string  `json:"bestPerformingSubject,omitempty" bson:"bestPerformingSubject"`
	WeakestSubject        *string  `json:"weakestSubject,omitempty" bson:"weakestSubject"`

	// Academic Strengths & Weaknesses
	AcademicStrengths       *string `json:"academicStrengths,omitempty" bson:"academicStrengths"`
	AreasNeedingImprovement *string `json:"areasNeedingImprovement,omitempty" bson:"areasNeedingImprovement"`
	RecommendedActions      *string `json:"recommendedActions,omitempty" bson:"recommendedActions"`
	StudyRecommendations    *string `json:"studyRecommendations,omitempty" bson:"studyRecommendations"`

	// Skill Analysis (Aggregated from all subjects)
	OverallConceptualUnderstanding *float64 `json:"overallConceptualUnderstanding,omitempty" bson:"overallConceptualUnderstanding"`
	OverallProblemSolving          *float64 `json:"overallProblemSolving,omitempty" bson:"overallProblemSolving"`
	OverallKnowledgeApplication    *float64 `json:"overallKnowledgeApplication,omitempty" bson:"overallKnowledgeApplication"`
	OverallAnalyticalThinking      *float64 `json:"overallAnalyticalThinking,omitempty" bson:"overallAnalyticalThinking"`
	OverallCreativity              *float64 `json:"overallCreativity,omitempty" bson:"overallCreativity"`
	OverallPracticalSkills         *float64 `json:"overallPracticalSkills,omitempty" bson:"overallPracticalSkills"`

	// Behavioral Analysis (Aggregated from all subjects)
	OverallParticipation *float64 `json:"overallParticipation,omitempty" bson:"overallParticipation"`
	OverallDiscipline    *float64 `json:"overallDiscipline,omitempty" bson:"overallDiscipline"`
	OverallPunctuality   *float64 `json:"overallPunctuality,omitempty" bson:"overallPunctuality"`
	OverallTeamwork      *float64 `json:"overallTeamwork,omitempty" bson:"overallTeamwork"`
	OverallEffortLevel   *float64 `json:"overallEffortLevel,omitempty" bson:"overallEffortLevel"`
	OverallImprovement   *float64 `json:"overallImprovement,omitempty" bson:"overallImprovement"`

	// Assessment Breakdown (Aggregated)
	AverageMidtermScore          *float64 `json:"averageMidtermScore,omitempty" bson:"averageMidtermScore"`
	AverageFinalExamScore        *float64 `json:"averageFinalExamScore,omitempty" bson:"averageFinalExamScore"`
	AverageQuizScore             *float64 `json:"averageQuizScore,omitempty" bson:"averageQuizScore"`
	AverageAssignmentScore       *float64 `json:"averageAssignmentScore,omitempty" bson:"averageAssignmentScore"`
	AveragePracticalScore        *float64 `json:"averagePracticalScore,omitempty" bson:"averagePracticalScore"`
	AverageOralPresentationScore *float64 `json:"averageOralPresentationScore,omitempty" bson:"averagePresentationScore"`

	// Performance Trends
	ImprovementTrend     *string  `json:"improvementTrend,omitempty" bson:"improvementTrend"`         // "improving", "declining", "stable"
	ConsistencyRating    *float64 `json:"consistencyRating,omitempty" bson:"consistencyRating"`       // 0-10 scale
	PerformanceStability *string  `json:"performanceStability,omitempty" bson:"performanceStability"` // "very stable", "stable", "variable", "inconsistent"

	// Attendance & Engagement
	AttendanceRate     *float64 `json:"attendanceRate,omitempty" bson:"attendanceRate"`
	EngagementLevel    *string  `json:"engagementLevel,omitempty" bson:"engagementLevel"` // "high", "medium", "low"
	ClassParticipation *string  `json:"classParticipation,omitempty" bson:"classParticipation"`

	// Goals & Recommendations
	AcademicGoals                *string `json:"academicGoals,omitempty" bson:"academicGoals"`
	ShortTermObjectives          *string `json:"shortTermObjectives,omitempty" bson:"shortTermObjectives"`
	LongTermObjectives           *string `json:"longTermObjectives,omitempty" bson:"longTermObjectives"`
	ParentTeacherRecommendations *string `json:"parentTeacherRecommendations,omitempty" bson:"parentTeacherRecommendations"`

	// Subject Details (Reference to individual subject reports)
	SubjectReports []SubjectReportSummary `json:"subjectReports,omitempty" bson:"subjectReports"`

	// Overall Comments
	TeacherComments   *string `json:"teacherComments,omitempty" bson:"teacherComments"`
	PrincipalComments *string `json:"principalComments,omitempty" bson:"principalComments"`
	OverallRemarks    *string `json:"overallRemarks,omitempty" bson:"overallRemarks"`

	// Next Steps
	RecommendedResources *string    `json:"recommendedResources,omitempty" bson:"recommendedResources"`
	SuggestedActivities  *string    `json:"suggestedActivities,omitempty" bson:"suggestedActivities"`
	NextReviewDate       *time.Time `json:"nextReviewDate,omitempty" bson:"nextReviewDate"`
}

// SubjectReportSummary provides a condensed view of individual subject performance
type SubjectReportSummary struct {
	Subject                        string   `json:"subject" bson:"subject"`
	Score                          int      `json:"score" bson:"score"`
	Grade                          *string  `json:"grade,omitempty" bson:"grade"`
	ConceptualUnderstanding        *float64 `json:"conceptualUnderstanding,omitempty" bson:"conceptualUnderstanding"`
	ProblemSolving                 *float64 `json:"problemSolving,omitempty" bson:"problemSolving"`
	AnalyticalThinking             *float64 `json:"analyticalThinking,omitempty" bson:"analyticalThinking"`
	AreasForImprovement            *string  `json:"areasForImprovement,omitempty" bson:"areasForImprovement"`
	KeyStrengths                   *string  `json:"keyStrengths,omitempty" bson:"keyStrengths"`
	SubjectSpecificRecommendations *string  `json:"subjectSpecificRecommendations,omitempty" bson:"subjectSpecificRecommendations"`
}

// NewReportCard creates a new ReportCard with default values
func NewReportCard() *ReportCard {
	now := time.Now()
	return &ReportCard{
		GeneratedAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
		SubjectReports: make([]SubjectReportSummary, 0),
	}
}
