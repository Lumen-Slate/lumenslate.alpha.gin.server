package service

import (
	"errors"
	"fmt"

	"lumenslate/internal/model"
	"lumenslate/internal/repository"
)

// UsageLimitsService handles business logic for usage limits
type UsageLimitsService struct{}

// NewUsageLimitsService creates a new UsageLimitsService instance
func NewUsageLimitsService() *UsageLimitsService {
	return &UsageLimitsService{}
}

// CreateUsageLimitsRequest represents the request to create usage limits
type CreateUsageLimitsRequest struct {
	PlanName                string                `json:"plan_name" validate:"required"`
	Teachers                model.UsageLimitValue `json:"teachers" validate:"required"`
	Classrooms              model.UsageLimitValue `json:"classrooms" validate:"required"`
	StudentsPerClassroom    model.UsageLimitValue `json:"students_per_classroom" validate:"required"`
	QuestionBanks           model.UsageLimitValue `json:"question_banks" validate:"required"`
	Questions               model.UsageLimitValue `json:"questions" validate:"required"`
	AssignmentExportsPerDay model.UsageLimitValue `json:"assignment_exports_per_day" validate:"required"`
	AI                      model.AILimits        `json:"ai" validate:"required"`
}

// UpdateUsageLimitsRequest represents the request to update usage limits
type UpdateUsageLimitsRequest struct {
	PlanName                *string                `json:"plan_name,omitempty"`
	Teachers                *model.UsageLimitValue `json:"teachers,omitempty"`
	Classrooms              *model.UsageLimitValue `json:"classrooms,omitempty"`
	StudentsPerClassroom    *model.UsageLimitValue `json:"students_per_classroom,omitempty"`
	QuestionBanks           *model.UsageLimitValue `json:"question_banks,omitempty"`
	Questions               *model.UsageLimitValue `json:"questions,omitempty"`
	AssignmentExportsPerDay *model.UsageLimitValue `json:"assignment_exports_per_day,omitempty"`
	AI                      *model.AILimits        `json:"ai,omitempty"`
	IsActive                *bool                  `json:"is_active,omitempty"`
}

// CreateUsageLimits creates new usage limits
func (s *UsageLimitsService) CreateUsageLimits(req CreateUsageLimitsRequest) (*model.UsageLimits, error) {
	// Validate the request
	if err := s.validateUsageLimitsRequest(req); err != nil {
		return nil, err
	}

	// Check if plan name already exists
	existing, _ := repository.GetUsageLimitsByPlanName(req.PlanName)
	if existing != nil {
		return nil, errors.New("usage limits for this plan already exist")
	}

	// Create new usage limits
	usageLimits := model.NewUsageLimits()
	usageLimits.PlanName = req.PlanName
	usageLimits.Teachers = req.Teachers
	usageLimits.Classrooms = req.Classrooms
	usageLimits.StudentsPerClassroom = req.StudentsPerClassroom
	usageLimits.QuestionBanks = req.QuestionBanks
	usageLimits.Questions = req.Questions
	usageLimits.AssignmentExportsPerDay = req.AssignmentExportsPerDay
	usageLimits.AI = req.AI

	return repository.CreateUsageLimits(*usageLimits)
}

// GetUsageLimitsByID retrieves usage limits by ID
func (s *UsageLimitsService) GetUsageLimitsByID(id string) (*model.UsageLimits, error) {
	if id == "" {
		return nil, errors.New("usage limits ID is required")
	}
	return repository.GetUsageLimitsByID(id)
}

// GetUsageLimitsByPlanName retrieves usage limits by plan name
func (s *UsageLimitsService) GetUsageLimitsByPlanName(planName string) (*model.UsageLimits, error) {
	if planName == "" {
		return nil, errors.New("plan name is required")
	}
	return repository.GetUsageLimitsByPlanName(planName)
}

// GetAllUsageLimits retrieves all usage limits with filters
func (s *UsageLimitsService) GetAllUsageLimits(filters model.UsageLimitsFilter) ([]model.UsageLimits, error) {
	return repository.GetAllUsageLimits(filters)
}

// UpdateUsageLimits updates existing usage limits
func (s *UsageLimitsService) UpdateUsageLimits(id string, req UpdateUsageLimitsRequest) (*model.UsageLimits, error) {
	if id == "" {
		return nil, errors.New("usage limits ID is required")
	}

	// Check if usage limits exist
	existing, err := repository.GetUsageLimitsByID(id)
	if err != nil {
		return nil, err
	}

	// If plan name is being changed, check for duplicates
	if req.PlanName != nil && *req.PlanName != existing.PlanName {
		existingPlan, _ := repository.GetUsageLimitsByPlanName(*req.PlanName)
		if existingPlan != nil {
			return nil, errors.New("usage limits for this plan already exist")
		}
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.PlanName != nil {
		updates["plan_name"] = *req.PlanName
	}
	if req.Teachers != nil {
		if !model.ValidateUsageLimitValue(*req.Teachers) {
			return nil, errors.New("invalid teachers limit value")
		}
		updates["teachers"] = *req.Teachers
	}
	if req.Classrooms != nil {
		if !model.ValidateUsageLimitValue(*req.Classrooms) {
			return nil, errors.New("invalid classrooms limit value")
		}
		updates["classrooms"] = *req.Classrooms
	}
	if req.StudentsPerClassroom != nil {
		if !model.ValidateUsageLimitValue(*req.StudentsPerClassroom) {
			return nil, errors.New("invalid students per classroom limit value")
		}
		updates["students_per_classroom"] = *req.StudentsPerClassroom
	}
	if req.QuestionBanks != nil {
		if !model.ValidateUsageLimitValue(*req.QuestionBanks) {
			return nil, errors.New("invalid question banks limit value")
		}
		updates["question_banks"] = *req.QuestionBanks
	}
	if req.Questions != nil {
		if !model.ValidateUsageLimitValue(*req.Questions) {
			return nil, errors.New("invalid questions limit value")
		}
		updates["questions"] = *req.Questions
	}
	if req.AssignmentExportsPerDay != nil {
		if !model.ValidateUsageLimitValue(*req.AssignmentExportsPerDay) {
			return nil, errors.New("invalid assignment exports per day limit value")
		}
		updates["assignment_exports_per_day"] = *req.AssignmentExportsPerDay
	}
	if req.AI != nil {
		if err := s.validateAILimits(*req.AI); err != nil {
			return nil, err
		}
		updates["ai"] = *req.AI
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	return repository.UpdateUsageLimits(id, updates)
}

// PatchUsageLimits performs partial updates on usage limits
func (s *UsageLimitsService) PatchUsageLimits(id string, updates map[string]interface{}) (*model.UsageLimits, error) {
	if id == "" {
		return nil, errors.New("usage limits ID is required")
	}

	// Validate usage limit values if they're being updated
	for key, value := range updates {
		switch key {
		case "teachers", "classrooms", "students_per_classroom", "question_banks", "questions", "assignment_exports_per_day":
			if !model.ValidateUsageLimitValue(value) {
				return nil, fmt.Errorf("invalid %s limit value", key)
			}
		case "ai":
			if aiLimits, ok := value.(model.AILimits); ok {
				if err := s.validateAILimits(aiLimits); err != nil {
					return nil, err
				}
			}
		}
	}

	return repository.PatchUsageLimits(id, updates)
}

// DeleteUsageLimits deletes usage limits
func (s *UsageLimitsService) DeleteUsageLimits(id string) error {
	if id == "" {
		return errors.New("usage limits ID is required")
	}
	return repository.DeleteUsageLimits(id)
}

// SoftDeleteUsageLimits marks usage limits as inactive
func (s *UsageLimitsService) SoftDeleteUsageLimits(id string) (*model.UsageLimits, error) {
	if id == "" {
		return nil, errors.New("usage limits ID is required")
	}
	return repository.SoftDeleteUsageLimits(id)
}

// GetUsageLimitsStats returns statistics about usage limits
func (s *UsageLimitsService) GetUsageLimitsStats() (map[string]interface{}, error) {
	return repository.GetUsageLimitsStats()
}

// InitializeDefaultUsageLimits creates default usage limits if they don't exist
func (s *UsageLimitsService) InitializeDefaultUsageLimits() error {
	return repository.CreateDefaultUsageLimits()
}

// CheckUserUsageAgainstLimits checks if user's current usage exceeds their plan limits
func (s *UsageLimitsService) CheckUserUsageAgainstLimits(userID string, planName string) (map[string]interface{}, error) {
	// Get user's usage limits
	limits, err := s.GetUsageLimitsByPlanName(planName)
	if err != nil {
		return nil, err
	}

	// Get user's current usage (you'll need to implement this based on your usage tracking)
	// For now, returning a placeholder structure
	result := map[string]interface{}{
		"plan_name": planName,
		"limits":    limits,
		"usage": map[string]interface{}{
			// This would be populated with actual usage data
			"teachers_used":             0,
			"classrooms_used":           0,
			"question_banks_used":       0,
			"questions_used":            0,
			"assignment_exports_today":  0,
			"ai_independent_agent_used": 0,
			"ai_lumen_agent_used":       0,
			"ai_rag_agent_used":         0,
			"ai_rag_documents_uploaded": 0,
		},
		"within_limits":   true,
		"exceeded_limits": []string{},
	}

	return result, nil
}

// validateUsageLimitsRequest validates the create usage limits request
func (s *UsageLimitsService) validateUsageLimitsRequest(req CreateUsageLimitsRequest) error {
	if req.PlanName == "" {
		return errors.New("plan name is required")
	}

	// Validate usage limit values
	if !model.ValidateUsageLimitValue(req.Teachers) {
		return errors.New("invalid teachers limit value")
	}
	if !model.ValidateUsageLimitValue(req.Classrooms) {
		return errors.New("invalid classrooms limit value")
	}
	if !model.ValidateUsageLimitValue(req.StudentsPerClassroom) {
		return errors.New("invalid students per classroom limit value")
	}
	if !model.ValidateUsageLimitValue(req.QuestionBanks) {
		return errors.New("invalid question banks limit value")
	}
	if !model.ValidateUsageLimitValue(req.Questions) {
		return errors.New("invalid questions limit value")
	}
	if !model.ValidateUsageLimitValue(req.AssignmentExportsPerDay) {
		return errors.New("invalid assignment exports per day limit value")
	}

	// Validate AI limits
	return s.validateAILimits(req.AI)
}

// validateAILimits validates AI limits structure
func (s *UsageLimitsService) validateAILimits(ai model.AILimits) error {
	if !model.ValidateUsageLimitValue(ai.IndependentAgent) {
		return errors.New("invalid independent agent limit value")
	}
	if !model.ValidateUsageLimitValue(ai.LumenAgent) {
		return errors.New("invalid lumen agent limit value")
	}
	if !model.ValidateUsageLimitValue(ai.RAGAgent) {
		return errors.New("invalid RAG agent limit value")
	}
	if !model.ValidateUsageLimitValue(ai.RAGDocumentUploads) {
		return errors.New("invalid RAG document uploads limit value")
	}
	return nil
}
