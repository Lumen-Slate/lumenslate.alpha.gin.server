package model

import (
	"time"
)

// UsageTracking represents usage metrics for a user
type UsageTracking struct {
	ID     string `bson:"_id,omitempty" json:"id"`
	UserID string `bson:"user_id" json:"user_id" validate:"required"`
	Period string `bson:"period" json:"period"` // Format: YYYY-MM for monthly tracking

	// Question Bank and Question Usage
	TotalQuestionBanks int64 `bson:"total_question_banks" json:"total_question_banks"`
	TotalQuestions     int64 `bson:"total_questions" json:"total_questions"`

	// AI Agent Usage
	TotalIAUses         int64 `bson:"total_ia_uses" json:"total_ia_uses"` // Intelligent Agent Uses
	TotalLumenAgentUses int64 `bson:"total_lumen_agent_uses" json:"total_lumen_agent_uses"`
	TotalRAAgentUses    int64 `bson:"total_ra_agent_uses" json:"total_ra_agent_uses"` // Research Assistant Agent Uses

	// Class and Assignment Usage
	TotalRecapClasses      int64 `bson:"total_recap_classes" json:"total_recap_classes"`
	TotalAssignmentExports int64 `bson:"total_assignment_exports" json:"total_assignment_exports"`

	// Additional tracking fields
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	// Daily breakdown (optional for detailed tracking)
	DailyBreakdown map[string]int64 `bson:"daily_breakdown,omitempty" json:"daily_breakdown,omitempty"`
}

// UsageMetrics represents the current usage totals for a user
type UsageMetrics struct {
	UserID                 string    `json:"user_id"`
	TotalQuestionBanks     int64     `json:"total_question_banks"`
	TotalQuestions         int64     `json:"total_questions"`
	TotalIAUses            int64     `json:"total_ia_uses"`
	TotalLumenAgentUses    int64     `json:"total_lumen_agent_uses"`
	TotalRAAgentUses       int64     `json:"total_ra_agent_uses"`
	TotalRecapClasses      int64     `json:"total_recap_classes"`
	TotalAssignmentExports int64     `json:"total_assignment_exports"`
	CurrentPeriod          string    `json:"current_period"`
	LastUpdated            time.Time `json:"last_updated"`
}

// NewUsageTracking creates a new usage tracking record for a user and period
func NewUsageTracking(userID, period string) *UsageTracking {
	now := time.Now()
	return &UsageTracking{
		UserID:                 userID,
		Period:                 period,
		TotalQuestionBanks:     0,
		TotalQuestions:         0,
		TotalIAUses:            0,
		TotalLumenAgentUses:    0,
		TotalRAAgentUses:       0,
		TotalRecapClasses:      0,
		TotalAssignmentExports: 0,
		CreatedAt:              now,
		UpdatedAt:              now,
		DailyBreakdown:         make(map[string]int64),
	}
}

// IncrementQuestionBanks increments the question banks counter
func (ut *UsageTracking) IncrementQuestionBanks(count int64) {
	ut.TotalQuestionBanks += count
	ut.UpdatedAt = time.Now()
}

// IncrementQuestions increments the questions counter
func (ut *UsageTracking) IncrementQuestions(count int64) {
	ut.TotalQuestions += count
	ut.UpdatedAt = time.Now()
}

// IncrementIAUses increments the IA uses counter
func (ut *UsageTracking) IncrementIAUses(count int64) {
	ut.TotalIAUses += count
	ut.UpdatedAt = time.Now()
}

// IncrementLumenAgentUses increments the Lumen Agent uses counter
func (ut *UsageTracking) IncrementLumenAgentUses(count int64) {
	ut.TotalLumenAgentUses += count
	ut.UpdatedAt = time.Now()
}

// IncrementRAAgentUses increments the RA Agent uses counter
func (ut *UsageTracking) IncrementRAAgentUses(count int64) {
	ut.TotalRAAgentUses += count
	ut.UpdatedAt = time.Now()
}

// IncrementRecapClasses increments the recap classes counter
func (ut *UsageTracking) IncrementRecapClasses(count int64) {
	ut.TotalRecapClasses += count
	ut.UpdatedAt = time.Now()
}

// IncrementAssignmentExports increments the assignment exports counter
func (ut *UsageTracking) IncrementAssignmentExports(count int64) {
	ut.TotalAssignmentExports += count
	ut.UpdatedAt = time.Now()
}

// GetCurrentPeriod returns the current period in YYYY-MM format
func GetCurrentPeriod() string {
	now := time.Now()
	return now.Format("2006-01")
}

// ToMetrics converts UsageTracking to UsageMetrics
func (ut *UsageTracking) ToMetrics() *UsageMetrics {
	return &UsageMetrics{
		UserID:                 ut.UserID,
		TotalQuestionBanks:     ut.TotalQuestionBanks,
		TotalQuestions:         ut.TotalQuestions,
		TotalIAUses:            ut.TotalIAUses,
		TotalLumenAgentUses:    ut.TotalLumenAgentUses,
		TotalRAAgentUses:       ut.TotalRAAgentUses,
		TotalRecapClasses:      ut.TotalRecapClasses,
		TotalAssignmentExports: ut.TotalAssignmentExports,
		CurrentPeriod:          ut.Period,
		LastUpdated:            ut.UpdatedAt,
	}
}
