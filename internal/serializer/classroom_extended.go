package serializer

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
)

type ClassroomExtended struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Teachers    []*model.Teacher    `json:"teachers"`
	Assignments []*model.Assignment `json:"assignments"`
	Credits     int                 `json:"credits"`
	Tags        []string            `json:"tags"`
	CreatedAt   string              `json:"createdAt"`
	UpdatedAt   string              `json:"updatedAt"`
	IsActive    bool                `json:"isActive"`
}

func NewClassroomExtended(c *model.Classroom) *ClassroomExtended {
	teachers := []*model.Teacher{}
	assignments := []*model.Assignment{}

	for _, tid := range c.TeacherIDs {
		if t, err := repo.GetTeacherByID(tid); err == nil {
			teachers = append(teachers, t)
		}
	}

	for _, aid := range c.AssignmentIDs {
		if a, err := repo.GetAssignmentByID(aid); err == nil {
			assignments = append(assignments, a)
		}
	}

	return &ClassroomExtended{
		ID:          c.ID,
		Name:        c.Name,
		Teachers:    teachers,
		Assignments: assignments,
		Credits:     c.Credits,
		Tags:        c.Tags,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		IsActive:    c.IsActive,
	}
}
