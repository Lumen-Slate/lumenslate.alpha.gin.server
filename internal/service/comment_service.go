package service

import (
	"server/internal/model"
	"server/internal/repository"
)

func CreateComment(c model.Comment) error {
	return repository.SaveComment(c)
}

func GetComment(id string) (*model.Comment, error) {
	return repository.GetCommentByID(id)
}

func DeleteComment(id string) error {
	return repository.DeleteComment(id)
}

func GetAllComments() ([]model.Comment, error) {
	return repository.GetAllComments()
}

func UpdateComment(id string, updated model.Comment) error {
	updated.ID = id
	return repository.SaveComment(updated)
}

func PatchComment(id string, updates map[string]interface{}) error {
	return repository.PatchComment(id, updates)
}
