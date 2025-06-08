package service

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
)

func CreateThread(t model.Thread) error {
	return repo.SaveThread(t)
}

func GetThread(id string) (*model.Thread, error) {
	return repo.GetThreadByID(id)
}

func DeleteThread(id string) error {
	return repo.DeleteThread(id)
}

func GetAllThreads(filters map[string]string) ([]model.Thread, error) {
	return repo.GetAllThreads(filters)
}

func UpdateThread(id string, updated model.Thread) error {
	updated.ID = id
	return repo.SaveThread(updated)
}

func PatchThread(id string, updates map[string]interface{}) (*model.Thread, error) {
	return repo.PatchThread(id, updates)
}
