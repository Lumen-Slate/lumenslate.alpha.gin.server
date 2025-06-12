package serializer

import (
	"lumenslate/internal/model"
	questions "lumenslate/internal/model/questions"
	repo "lumenslate/internal/repository"
	quest "lumenslate/internal/repository/questions"
	"time"
)

type AssignmentExtended struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	DueDate     string                 `json:"dueDate"`
	CreatedAt   string                 `json:"createdAt"`
	Points      int                    `json:"points"`
	Comments    []model.Comment        `json:"comments,omitempty"`
	MCQs        []questions.MCQ        `json:"mcqs,omitempty"`
	MSQs        []questions.MSQ        `json:"msqs,omitempty"`
	NATs        []questions.NAT        `json:"nats,omitempty"`
	Subjectives []questions.Subjective `json:"subjectives,omitempty"`
}

func NewAssignmentExtended(a *model.Assignment) *AssignmentExtended {
	ext := &AssignmentExtended{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		DueDate:   a.DueDate.Format(time.RFC3339),
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		Points:    a.Points,
	}

	// Debug: Print CommentIds, MCQIds, MSQIds, NATIds, SubjectiveIds

	for _, cid := range a.CommentIds {
		if c, err := repo.GetCommentByID(cid); err == nil && c != nil {
			ext.Comments = append(ext.Comments, *c)
		} else {
		}
	}
	for _, id := range a.MCQIds {
		if q, err := quest.GetMCQByID(id); err == nil && q != nil {
			ext.MCQs = append(ext.MCQs, *q)
		} else {
		}
	}
	for _, id := range a.MSQIds {
		if q, err := quest.GetMSQByID(id); err == nil && q != nil {
			ext.MSQs = append(ext.MSQs, *q)
		} else {
		}
	}
	for _, id := range a.NATIds {
		if q, err := quest.GetNATByID(id); err == nil && q != nil {
			ext.NATs = append(ext.NATs, *q)
		} else {
		}
	}
	for _, id := range a.SubjectiveIds {
		if q, err := quest.GetSubjectiveByID(id); err == nil && q != nil {
			ext.Subjectives = append(ext.Subjectives, *q)
		} else {
		}
	}
	return ext
}
