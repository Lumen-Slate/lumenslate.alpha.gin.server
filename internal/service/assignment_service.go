package service

import (
	"server/internal/model"
	"server/internal/repository"
	"time"
)

func CreateAssignment(a model.Assignment) error {
	// Set default creation time if not provided
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	return repository.SaveAssignment(a)
}

func GetAssignment(id string) (*model.Assignment, error) {
	return repository.GetAssignmentByID(id)
}

func DeleteAssignment(id string) error {
	return repository.DeleteAssignment(id)
}

func GetAllAssignments() ([]model.Assignment, error) {
	return repository.GetAllAssignments()
}

func UpdateAssignment(id string, updated model.Assignment) error {
	return repository.SaveAssignment(updated) // overwrite existing
}

func FilterAssignments(limitStr, offsetStr, points, due string) ([]model.Assignment, error) {
	return repository.FilterAssignments(limitStr, offsetStr, points, due)
}
