package service

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
)

func CreatePost(p model.Post) error {
	return repo.SavePost(p)
}

func GetPost(id string) (*model.Post, error) {
	return repo.GetPostByID(id)
}

func DeletePost(id string) error {
	return repo.DeletePost(id)
}

func GetAllPosts(filters map[string]string) ([]model.Post, error) {
	return repo.GetAllPosts(filters)
}

func UpdatePost(id string, updated model.Post) error {
	updated.ID = id
	return repo.SavePost(updated)
}

func PatchPost(id string, updates map[string]interface{}) error {
	return repo.PatchPost(id, updates)
}
