package model

import (
	"time"
)

type Subject string

const (
	SubjectMath      Subject = "math"
	SubjectScience   Subject = "science"
	SubjectHistory   Subject = "history"
	SubjectGeography Subject = "geography"
	SubjectEnglish   Subject = "english"
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type SubjectReport struct {
	ID          string    `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	UserID      string    `json:"userId" bson:"userId" validate:"required"`
	StudentID   int       `json:"studentId" bson:"studentId" validate:"required"`
	StudentName string    `json:"studentName" bson:"studentName" validate:"required"`
	Subject     Subject   `json:"subject" bson:"subject" validate:"required"`
	Score       int       `json:"score" bson:"score" validate:"required,min=0,max=100"`
	Timestamp   time.Time `json:"timestamp" bson:"timestamp"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`

	// Optional fields
	GradeLetter    *string `json:"gradeLetter,omitempty" bson:"gradeLetter"`
	ClassName      *string `json:"className,omitempty" bson:"className"`
	InstructorName *string `json:"instructorName,omitempty" bson:"instructorName"`
	Term           *string `json:"term,omitempty" bson:"term"`
	Remarks        *string `json:"remarks,omitempty" bson:"remarks"`

	// Assessment component breakdown
	MidtermScore          *int `json:"midtermScore,omitempty" bson:"midtermScore"`
	FinalExamScore        *int `json:"finalExamScore,omitempty" bson:"finalExamScore"`
	QuizScore             *int `json:"quizScore,omitempty" bson:"quizScore"`
	AssignmentScore       *int `json:"assignmentScore,omitempty" bson:"assignmentScore"`
	PracticalScore        *int `json:"practicalScore,omitempty" bson:"practicalScore"`
	OralPresentationScore *int `json:"oralPresentationScore,omitempty" bson:"oralPresentationScore"`

	// Skill evaluation (0-10 scale or %)
	ConceptualUnderstanding *float64 `json:"conceptualUnderstanding,omitempty" bson:"conceptualUnderstanding"`
	ProblemSolving          *float64 `json:"problemSolving,omitempty" bson:"problemSolving"`
	KnowledgeApplication    *float64 `json:"knowledgeApplication,omitempty" bson:"knowledgeApplication"`
	AnalyticalThinking      *float64 `json:"analyticalThinking,omitempty" bson:"analyticalThinking"`
	Creativity              *float64 `json:"creativity,omitempty" bson:"creativity"`
	PracticalSkills         *float64 `json:"practicalSkills,omitempty" bson:"practicalSkills"`

	// Behavioral metrics
	Participation *float64 `json:"participation,omitempty" bson:"participation"`
	Discipline    *float64 `json:"discipline,omitempty" bson:"discipline"`
	Punctuality   *float64 `json:"punctuality,omitempty" bson:"punctuality"`
	Teamwork      *float64 `json:"teamwork,omitempty" bson:"teamwork"`
	EffortLevel   *float64 `json:"effortLevel,omitempty" bson:"effortLevel"`
	Improvement   *float64 `json:"improvement,omitempty" bson:"improvement"`

	// Advanced insights
	LearningObjectivesMastered *string `json:"learningObjectivesMastered,omitempty" bson:"learningObjectivesMastered"`
	AreasForImprovement        *string `json:"areasForImprovement,omitempty" bson:"areasForImprovement"`
	RecommendedResources       *string `json:"recommendedResources,omitempty" bson:"recommendedResources"`
	TargetGoals                *string `json:"targetGoals,omitempty" bson:"targetGoals"`
}

// NewSubjectReport creates a new SubjectReport with default values
func NewSubjectReport() *SubjectReport {
	now := time.Now()
	return &SubjectReport{
		Timestamp: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetSubjectFromString converts subject string to Subject enum, handling various formats
func GetSubjectFromString(subjectString string) (Subject, bool) {
	subjectMapping := map[string]Subject{
		"math":           SubjectMath,
		"mathematics":    SubjectMath,
		"maths":          SubjectMath,
		"science":        SubjectScience,
		"biology":        SubjectScience,
		"chemistry":      SubjectScience,
		"physics":        SubjectScience,
		"english":        SubjectEnglish,
		"language arts":  SubjectEnglish,
		"literature":     SubjectEnglish,
		"reading":        SubjectEnglish,
		"history":        SubjectHistory,
		"social studies": SubjectHistory,
		"world history":  SubjectHistory,
		"geography":      SubjectGeography,
		"geo":            SubjectGeography,
	}

	if subject, exists := subjectMapping[subjectString]; exists {
		return subject, true
	}
	return "", false
}

// Questions model to match the SQLite structure
type Questions struct {
	ID         string     `json:"id,omitempty" bson:"_id" validate:"omitempty"`
	Subject    Subject    `json:"subject" bson:"subject" validate:"required"`
	Question   string     `json:"question" bson:"question" validate:"required"`
	Options    []string   `json:"options" bson:"options" validate:"required"`
	Answer     string     `json:"answer" bson:"answer" validate:"required"`
	Difficulty Difficulty `json:"difficulty" bson:"difficulty" validate:"required"`
	CreatedAt  time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt" bson:"updatedAt"`
	IsActive   bool       `json:"isActive" bson:"isActive"`
}

// NewQuestions creates a new Questions with default values
func NewQuestions() *Questions {
	now := time.Now()
	return &Questions{
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
}
