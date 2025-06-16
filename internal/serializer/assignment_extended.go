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
	println("CommentIds:", len(a.CommentIds))
	println("MCQIds:", len(a.MCQIds))
	println("MSQIds:", len(a.MSQIds))
	println("NATIds:", len(a.NATIds))
	println("SubjectiveIds:", len(a.SubjectiveIds))

	for _, cid := range a.CommentIds {
		println("Fetching CommentId:", cid)
		if c, err := repo.GetCommentByID(cid); err == nil && c != nil {
			ext.Comments = append(ext.Comments, *c)
		} else {
			println("Comment fetch error or nil:", err)
		}
	}
	for _, id := range a.MCQIds {
		println("Fetching MCQId:", id)
		if q, err := quest.GetMCQByID(id); err == nil && q != nil {
			ext.MCQs = append(ext.MCQs, *q)
		} else {
			println("MCQ fetch error or nil:", err)
		}
	}
	for _, id := range a.MSQIds {
		println("Fetching MSQId:", id)
		if q, err := quest.GetMSQByID(id); err == nil && q != nil {
			ext.MSQs = append(ext.MSQs, *q)
		} else {
			println("MSQ fetch error or nil:", err)
		}
	}
	for _, id := range a.NATIds {
		println("Fetching NATId:", id)
		if q, err := quest.GetNATByID(id); err == nil && q != nil {
			ext.NATs = append(ext.NATs, *q)
		} else {
			println("NAT fetch error or nil:", err)
		}
	}
	for _, id := range a.SubjectiveIds {
		println("Fetching SubjectiveId:", id)
		if q, err := quest.GetSubjectiveByID(id); err == nil && q != nil {
			ext.Subjectives = append(ext.Subjectives, *q)
		} else {
			println("Subjective fetch error or nil:", err)
		}
	}
	return ext
}
