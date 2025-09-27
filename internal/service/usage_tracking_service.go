package service

import (
	"lumenslate/internal/model"
	"lumenslate/internal/repository"
)

// UsageTrackingService provides business logic for usage tracking
type UsageTrackingService struct{}

// NewUsageTrackingService creates a new usage tracking service instance
func NewUsageTrackingService() *UsageTrackingService {
	return &UsageTrackingService{}
}

// TrackQuestionBankUsage increments the question bank usage counter for a user
func (s *UsageTrackingService) TrackQuestionBankUsage(userID string, count int64) error {
	if count <= 0 {
		count = 1 // Default to 1 if not specified or invalid
	}
	return repository.IncrementQuestionBanks(userID, count)
}

// TrackQuestionUsage increments the question usage counter for a user
func (s *UsageTrackingService) TrackQuestionUsage(userID string, count int64) error {
	if count <= 0 {
		count = 1 // Default to 1 if not specified or invalid
	}
	return repository.IncrementQuestions(userID, count)
}

// TrackIAUsage increments the Intelligent Agent usage counter for a user
func (s *UsageTrackingService) TrackIAUsage(userID string, count int64) error {
	if count <= 0 {
		count = 1 // Default to 1 if not specified or invalid
	}
	return repository.IncrementIAUses(userID, count)
}

// TrackLumenAgentUsage increments the Lumen Agent usage counter for a user
func (s *UsageTrackingService) TrackLumenAgentUsage(userID string, count int64) error {
	if count <= 0 {
		count = 1 // Default to 1 if not specified or invalid
	}
	return repository.IncrementLumenAgentUses(userID, count)
}

// TrackRAAgentUsage increments the Research Assistant Agent usage counter for a user
func (s *UsageTrackingService) TrackRAAgentUsage(userID string, count int64) error {
	if count <= 0 {
		count = 1 // Default to 1 if not specified or invalid
	}
	return repository.IncrementRAAgentUses(userID, count)
}

// TrackRecapClassUsage increments the recap classes counter for a user
func (s *UsageTrackingService) TrackRecapClassUsage(userID string, count int64) error {
	if count <= 0 {
		count = 1 // Default to 1 if not specified or invalid
	}
	return repository.IncrementRecapClasses(userID, count)
}

// TrackAssignmentExportUsage increments the assignment exports counter for a user
func (s *UsageTrackingService) TrackAssignmentExportUsage(userID string, count int64) error {
	if count <= 0 {
		count = 1 // Default to 1 if not specified or invalid
	}
	return repository.IncrementAssignmentExports(userID, count)
}

// GetCurrentUsageMetrics retrieves the current period's usage metrics for a user
func (s *UsageTrackingService) GetCurrentUsageMetrics(userID string) (*model.UsageMetrics, error) {
	return repository.GetUsageMetricsByUser(userID)
}

// GetUsageTrackingByPeriod retrieves usage tracking for a specific user and period
func (s *UsageTrackingService) GetUsageTrackingByPeriod(userID, period string) (*model.UsageTracking, error) {
	return repository.GetUsageTrackingByUserAndPeriod(userID, period)
}

// GetAllUserUsageHistory retrieves all usage tracking records for a user
func (s *UsageTrackingService) GetAllUserUsageHistory(userID string) ([]model.UsageTracking, error) {
	return repository.GetAllUsageTrackingByUser(userID)
}

// GetAggregatedUserUsage retrieves aggregated usage metrics across all periods for a user
func (s *UsageTrackingService) GetAggregatedUserUsage(userID string) (*model.UsageMetrics, error) {
	return repository.GetAggregatedUsageByUser(userID)
}

// UsageTrackingFilters represents filters for querying usage tracking data
type UsageTrackingFilters struct {
	UserID string `json:"user_id,omitempty"`
	Period string `json:"period,omitempty"`
	Limit  string `json:"limit,omitempty"`
	Offset string `json:"offset,omitempty"`
}

// GetAllUsageTracking retrieves all usage tracking records with optional filters
func (s *UsageTrackingService) GetAllUsageTracking(filters UsageTrackingFilters) ([]model.UsageTracking, error) {
	filterMap := make(map[string]string)

	if filters.UserID != "" {
		filterMap["user_id"] = filters.UserID
	}
	if filters.Period != "" {
		filterMap["period"] = filters.Period
	}
	if filters.Limit != "" {
		filterMap["limit"] = filters.Limit
	} else {
		filterMap["limit"] = "10" // Default limit
	}
	if filters.Offset != "" {
		filterMap["offset"] = filters.Offset
	} else {
		filterMap["offset"] = "0" // Default offset
	}

	return repository.GetAllUsageTracking(filterMap)
}

// BulkUsageTrackingRequest represents a request to track multiple usage types at once
type BulkUsageTrackingRequest struct {
	UserID            string `json:"user_id" validate:"required"`
	QuestionBanks     *int64 `json:"question_banks,omitempty"`
	Questions         *int64 `json:"questions,omitempty"`
	IAUses            *int64 `json:"ia_uses,omitempty"`
	LumenAgentUses    *int64 `json:"lumen_agent_uses,omitempty"`
	RAAgentUses       *int64 `json:"ra_agent_uses,omitempty"`
	RecapClasses      *int64 `json:"recap_classes,omitempty"`
	AssignmentExports *int64 `json:"assignment_exports,omitempty"`
}

// TrackBulkUsage tracks multiple usage types for a user in a single operation
func (s *UsageTrackingService) TrackBulkUsage(req BulkUsageTrackingRequest) error {
	// Track each usage type if specified
	if req.QuestionBanks != nil && *req.QuestionBanks > 0 {
		if err := s.TrackQuestionBankUsage(req.UserID, *req.QuestionBanks); err != nil {
			return err
		}
	}

	if req.Questions != nil && *req.Questions > 0 {
		if err := s.TrackQuestionUsage(req.UserID, *req.Questions); err != nil {
			return err
		}
	}

	if req.IAUses != nil && *req.IAUses > 0 {
		if err := s.TrackIAUsage(req.UserID, *req.IAUses); err != nil {
			return err
		}
	}

	if req.LumenAgentUses != nil && *req.LumenAgentUses > 0 {
		if err := s.TrackLumenAgentUsage(req.UserID, *req.LumenAgentUses); err != nil {
			return err
		}
	}

	if req.RAAgentUses != nil && *req.RAAgentUses > 0 {
		if err := s.TrackRAAgentUsage(req.UserID, *req.RAAgentUses); err != nil {
			return err
		}
	}

	if req.RecapClasses != nil && *req.RecapClasses > 0 {
		if err := s.TrackRecapClassUsage(req.UserID, *req.RecapClasses); err != nil {
			return err
		}
	}

	if req.AssignmentExports != nil && *req.AssignmentExports > 0 {
		if err := s.TrackAssignmentExportUsage(req.UserID, *req.AssignmentExports); err != nil {
			return err
		}
	}

	return nil
}

// GetUsageSummaryByPeriod retrieves usage summary for a specific period across all users
func (s *UsageTrackingService) GetUsageSummaryByPeriod(period string) (map[string]interface{}, error) {
	filters := UsageTrackingFilters{
		Period: period,
		Limit:  "1000", // Large limit for summary
	}

	usageRecords, err := s.GetAllUsageTracking(filters)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"period":                   period,
		"total_users":              len(usageRecords),
		"total_question_banks":     int64(0),
		"total_questions":          int64(0),
		"total_ia_uses":            int64(0),
		"total_lumen_agent_uses":   int64(0),
		"total_ra_agent_uses":      int64(0),
		"total_recap_classes":      int64(0),
		"total_assignment_exports": int64(0),
	}

	for _, record := range usageRecords {
		summary["total_question_banks"] = summary["total_question_banks"].(int64) + record.TotalQuestionBanks
		summary["total_questions"] = summary["total_questions"].(int64) + record.TotalQuestions
		summary["total_ia_uses"] = summary["total_ia_uses"].(int64) + record.TotalIAUses
		summary["total_lumen_agent_uses"] = summary["total_lumen_agent_uses"].(int64) + record.TotalLumenAgentUses
		summary["total_ra_agent_uses"] = summary["total_ra_agent_uses"].(int64) + record.TotalRAAgentUses
		summary["total_recap_classes"] = summary["total_recap_classes"].(int64) + record.TotalRecapClasses
		summary["total_assignment_exports"] = summary["total_assignment_exports"].(int64) + record.TotalAssignmentExports
	}

	return summary, nil
}

// ResetUserUsage resets usage counters for a user (creates new tracking record for current period)
func (s *UsageTrackingService) ResetUserUsage(userID string) (*model.UsageTracking, error) {
	currentPeriod := model.GetCurrentPeriod()
	newUsage := model.NewUsageTracking(userID, currentPeriod)

	if err := repository.SaveUsageTracking(*newUsage); err != nil {
		return nil, err
	}

	return newUsage, nil
}
