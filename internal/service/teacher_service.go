package service

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
)

func CreateTeacher(t model.Teacher) error {
	return repo.SaveTeacher(t)
}

func GetTeacher(id string) (*model.Teacher, error) {
	return repo.GetTeacherByID(id)
}

func DeleteTeacher(id string) error {
	return repo.DeleteTeacher(id)
}

func GetAllTeachers(filters map[string]string) ([]model.Teacher, error) {
	return repo.GetAllTeachers(filters)
}

func UpdateTeacher(id string, updated model.Teacher) error {
	updated.ID = id
	return repo.SaveTeacher(updated)
}

func PatchTeacher(id string, updates map[string]interface{}) (*model.Teacher, error) {
	return repo.PatchTeacher(id, updates)
}
