package service

import (
	"lumenslate/internal/model"
	"lumenslate/internal/repository"
	"time"
)

// UsageTrackingService provides business logic for usage tracking
type UsageTrackingService struct{}

// NewUsageTrackingService creates a new usage tracking service instance
func NewUsageTrackingService() *UsageTrackingService {
	return &UsageTrackingService{}
}

// GetDefaultFeatures returns the default feature set for a subscription plan
func (s *UsageTrackingService) GetDefaultFeatures(lookupKey string) map[string]*model.FeatureUsage {
	// Define default features based on subscription plan
	// This should ideally come from a configuration or database
	features := make(map[string]*model.FeatureUsage)
	
	now := time.Now()
	
	switch lookupKey {
	case "basic_plan":
		features["question_banks"] = &model.FeatureUsage{
			Value:     0,
			Limit:     5,
			Unit:      model.UnitAbsolute,
			LastReset: now,
		}
		features["questions"] = &model.FeatureUsage{
			Value:     0,
			Limit:     100,
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
		features["ai_generations"] = &model.FeatureUsage{
			Value:     0,
			Limit:     20,
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
		features["assignment_exports"] = &model.FeatureUsage{
			Value:     0,
			Limit:     10,
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
	case "premium_plan":
		features["question_banks"] = &model.FeatureUsage{
			Value:     0,
			Limit:     -1, // Unlimited
			Unit:      model.UnitAbsolute,
			LastReset: now,
		}
		features["questions"] = &model.FeatureUsage{
			Value:     0,
			Limit:     -1, // Unlimited
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
		features["ai_generations"] = &model.FeatureUsage{
			Value:     0,
			Limit:     -1, // Unlimited
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
		features["assignment_exports"] = &model.FeatureUsage{
			Value:     0,
			Limit:     -1, // Unlimited
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
	default:
		// Free plan or default
		features["question_banks"] = &model.FeatureUsage{
			Value:     0,
			Limit:     2,
			Unit:      model.UnitAbsolute,
			LastReset: now,
		}
		features["questions"] = &model.FeatureUsage{
			Value:     0,
			Limit:     20,
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
		features["ai_generations"] = &model.FeatureUsage{
			Value:     0,
			Limit:     5,
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
		features["assignment_exports"] = &model.FeatureUsage{
			Value:     0,
			Limit:     3,
			Unit:      model.UnitMonthly,
			LastReset: now,
		}
	}
	
	return features
}

// GetOrCreateUsage gets or creates usage tracking for a user
func (s *UsageTrackingService) GetOrCreateUsage(userID, lookupKey string) (*model.Usage, error) {
	defaultFeatures := s.GetDefaultFeatures(lookupKey)
	return repository.GetOrCreateUsage(userID, lookupKey, defaultFeatures)
}

// IncrementFeature increments usage for a specific feature
func (s *UsageTrackingService) IncrementFeature(userID, featureName string, subscription *model.Subscription) (bool, error) {
	return repository.IncrementFeatureUsage(userID, featureName, subscription)
}

// SetFeatureValue sets a specific value for a feature (useful for absolute counts)
func (s *UsageTrackingService) SetFeatureValue(userID, featureName string, value int64) (bool, error) {
	return repository.SetFeatureUsage(userID, featureName, value)
}

// TrackQuestionBankUsage tracks question bank creation/deletion
func (s *UsageTrackingService) TrackQuestionBankUsage(userID string, count int64) error {
	subscription, err := s.getUserSubscription(userID)
	if err != nil {
		return err
	}

	for i := int64(0); i < count; i++ {
		success, err := s.IncrementFeature(userID, "question_banks", subscription)
		if err != nil {
			return err
		}
		if !success {
			return &UsageError{
				Feature: "question_banks",
				Message: "Question bank limit exceeded",
			}
		}
	}
	
	return nil
}

// TrackQuestionUsage tracks question generation/creation
func (s *UsageTrackingService) TrackQuestionUsage(userID string, count int64) error {
	subscription, err := s.getUserSubscription(userID)
	if err != nil {
		return err
	}

	for i := int64(0); i < count; i++ {
		success, err := s.IncrementFeature(userID, "questions", subscription)
		if err != nil {
			return err
		}
		if !success {
			return &UsageError{
				Feature: "questions",
				Message: "Question generation limit exceeded",
			}
		}
	}
	
	return nil
}

// TrackAIGenerationUsage tracks AI-powered feature usage
func (s *UsageTrackingService) TrackAIGenerationUsage(userID string, count int64) error {
	subscription, err := s.getUserSubscription(userID)
	if err != nil {
		return err
	}

	for i := int64(0); i < count; i++ {
		success, err := s.IncrementFeature(userID, "ai_generations", subscription)
		if err != nil {
			return err
		}
		if !success {
			return &UsageError{
				Feature: "ai_generations",
				Message: "AI generation limit exceeded",
			}
		}
	}
	
	return nil
}

// TrackAssignmentExportUsage tracks assignment exports
func (s *UsageTrackingService) TrackAssignmentExportUsage(userID string, count int64) error {
	subscription, err := s.getUserSubscription(userID)
	if err != nil {
		return err
	}

	for i := int64(0); i < count; i++ {
		success, err := s.IncrementFeature(userID, "assignment_exports", subscription)
		if err != nil {
			return err
		}
		if !success {
			return &UsageError{
				Feature: "assignment_exports",
				Message: "Assignment export limit exceeded",
			}
		}
	}
	
	return nil
}

// Legacy methods for backward compatibility
func (s *UsageTrackingService) TrackIAUsage(userID string, count int64) error {
	return s.TrackAIGenerationUsage(userID, count)
}

func (s *UsageTrackingService) TrackLumenAgentUsage(userID string, count int64) error {
	return s.TrackAIGenerationUsage(userID, count)
}

func (s *UsageTrackingService) TrackRAAgentUsage(userID string, count int64) error {
	return s.TrackAIGenerationUsage(userID, count)
}

func (s *UsageTrackingService) TrackRecapClassUsage(userID string, count int64) error {
	return s.TrackAIGenerationUsage(userID, count)
}

// GetUsageByUserID retrieves usage for a user
func (s *UsageTrackingService) GetUsageByUserID(userID string) (*model.Usage, error) {
	return repository.GetUsageByUserID(userID)
}

// UsageFilters represents filters for querying usage data
type UsageFilters struct {
	UserID    string `json:"user_id,omitempty"`
	LookupKey string `json:"lookup_key,omitempty"`
	Limit     string `json:"limit,omitempty"`
	Offset    string `json:"offset,omitempty"`
}

// GetAllUsage retrieves all usage records with filters
func (s *UsageTrackingService) GetAllUsage(filters UsageFilters) ([]model.Usage, error) {
	filterMap := make(map[string]string)

	if filters.UserID != "" {
		filterMap["user_id"] = filters.UserID
	}
	if filters.LookupKey != "" {
		filterMap["lookup_key"] = filters.LookupKey
	}
	if filters.Limit != "" {
		filterMap["limit"] = filters.Limit
	} else {
		filterMap["limit"] = "10"
	}
	if filters.Offset != "" {
		filterMap["offset"] = filters.Offset
	} else {
		filterMap["offset"] = "0"
	}

	return repository.GetAllUsage(filterMap)
}

// BulkUsageRequest represents a request to track multiple features at once
type BulkUsageRequest struct {
	UserID            string `json:"user_id" validate:"required"`
	QuestionBanks     *int64 `json:"question_banks,omitempty"`
	Questions         *int64 `json:"questions,omitempty"`
	AIGenerations     *int64 `json:"ai_generations,omitempty"`
	AssignmentExports *int64 `json:"assignment_exports,omitempty"`
}

// TrackBulkUsage tracks multiple features in a single operation
func (s *UsageTrackingService) TrackBulkUsage(req BulkUsageRequest) error {
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

	if req.AIGenerations != nil && *req.AIGenerations > 0 {
		if err := s.TrackAIGenerationUsage(req.UserID, *req.AIGenerations); err != nil {
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

// DeleteUsage deletes usage tracking for a user
func (s *UsageTrackingService) DeleteUsage(id string) error {
	return repository.DeleteUsage(id)
}

// ResetFeatureUsage resets usage for a specific feature across users with a lookup key
func (s *UsageTrackingService) ResetFeatureUsage(lookupKey, featureName string) error {
	return repository.ResetFeatureUsage(lookupKey, featureName)
}

// getUserSubscription is a helper method to get user's subscription
// In a real implementation, this would call the subscription repository
func (s *UsageTrackingService) getUserSubscription(userID string) (*model.Subscription, error) {
	// For now, return a mock subscription with current period
	now := time.Now()
	return &model.Subscription{
		UserID:             userID,
		LookupKey:          "basic_plan",
		Status:             model.StatusActive,
		CurrentPeriodStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		CurrentPeriodEnd:   time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second),
	}, nil
}

// UsageError represents a usage-related error
type UsageError struct {
	Feature string
	Message string
}

func (e *UsageError) Error() string {
	return e.Message
}

// Legacy structures for backward compatibility
type BulkUsageTrackingRequest = BulkUsageRequest

type UsageTrackingFilters struct {
	UserID string `json:"user_id,omitempty"`
	Period string `json:"period,omitempty"`
	Limit  string `json:"limit,omitempty"`
	Offset string `json:"offset,omitempty"`
}

// Legacy methods that map to new structure
func (s *UsageTrackingService) GetAllUsageTracking(filters UsageTrackingFilters) ([]model.Usage, error) {
	newFilters := UsageFilters{
		UserID: filters.UserID,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}
	return s.GetAllUsage(newFilters)
}

func (s *UsageTrackingService) GetCurrentUsageMetrics(userID string) (*model.Usage, error) {
	return s.GetUsageByUserID(userID)
}

func (s *UsageTrackingService) GetUsageTrackingByPeriod(userID, period string) (*model.Usage, error) {
	return s.GetUsageByUserID(userID)
}

func (s *UsageTrackingService) GetAllUserUsageHistory(userID string) ([]model.Usage, error) {
	filters := UsageFilters{
		UserID: userID,
		Limit:  "100",
	}
	return s.GetAllUsage(filters)
}

func (s *UsageTrackingService) GetAggregatedUserUsage(userID string) (*model.Usage, error) {
	return s.GetUsageByUserID(userID)
}

func (s *UsageTrackingService) GetUsageSummaryByPeriod(period string) (map[string]interface{}, error) {
	filters := UsageFilters{
		Limit: "1000",
	}

	usageRecords, err := s.GetAllUsage(filters)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"period":                 period,
		"total_users":            len(usageRecords),
		"total_question_banks":   int64(0),
		"total_questions":        int64(0),
		"total_ai_generations":   int64(0),
		"total_assignment_exports": int64(0),
	}

	for _, record := range usageRecords {
		if feature, ok := record.Features["question_banks"]; ok {
			summary["total_question_banks"] = summary["total_question_banks"].(int64) + feature.Value
		}
		if feature, ok := record.Features["questions"]; ok {
			summary["total_questions"] = summary["total_questions"].(int64) + feature.Value
		}
		if feature, ok := record.Features["ai_generations"]; ok {
			summary["total_ai_generations"] = summary["total_ai_generations"].(int64) + feature.Value
		}
		if feature, ok := record.Features["assignment_exports"]; ok {
			summary["total_assignment_exports"] = summary["total_assignment_exports"].(int64) + feature.Value
		}
	}

	return summary, nil
}

func (s *UsageTrackingService) ResetUserUsage(userID string) (*model.Usage, error) {
	usage, err := s.GetUsageByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Reset all features
	now := time.Now()
	for _, feature := range usage.Features {
		feature.Value = 0
		feature.LastReset = now
	}

	err = repository.UpdateUsage(usage)
	if err != nil {
		return nil, err
	}

	return usage, nil
}