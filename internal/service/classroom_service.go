package service

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
)

func CreateClassroom(c model.Classroom) error {
	return repo.SaveClassroom(c)
}

func GetClassroom(id string) (*model.Classroom, error) {
	return repo.GetClassroomByID(id)
}

func DeleteClassroom(id string) error {
	return repo.DeleteClassroom(id)
}

func GetAllClassrooms(filters map[string]string) ([]model.Classroom, error) {
	return repo.GetAllClassrooms(filters)
}

func UpdateClassroom(id string, updated model.Classroom) error {
	updated.ID = id
	return repo.SaveClassroom(updated)
}

func PatchClassroom(id string, updates map[string]interface{}) error {
	return repo.PatchClassroom(id, updates)
}
