package service

import (
	"lumenslate/internal/model"
	"lumenslate/internal/repository"
)

func CreateSubmission(s model.Submission) error {
	return repository.SaveSubmission(s)
}

func GetSubmission(id string) (*model.Submission, error) {
	return repository.GetSubmissionByID(id)
}

func DeleteSubmission(id string) error {
	return repository.DeleteSubmission(id)
}

func GetAllSubmissions(filters map[string]string) ([]model.Submission, error) {
	return repository.GetAllSubmissions(filters)
}

func UpdateSubmission(id string, updated model.Submission) error {
	updated.ID = id
	return repository.SaveSubmission(updated)
}

func PatchSubmission(id string, updates map[string]interface{}) error {
	return repository.PatchSubmission(id, updates)
}
