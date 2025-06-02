package service

import (
	"server/internal/model"
	"server/internal/repository"
)

func CreateTeacher(t model.Teacher) error {
	return repository.SaveTeacher(t)
}

func GetTeacher(id string) (*model.Teacher, error) {
	return repository.GetTeacherByID(id)
}

func DeleteTeacher(id string) error {
	return repository.DeleteTeacher(id)
}

func GetAllTeachers(filters map[string]string) ([]model.Teacher, error) {
	return repository.GetAllTeachers(filters)
}

func UpdateTeacher(id string, updated model.Teacher) error {
	updated.ID = id
	return repository.SaveTeacher(updated)
}

func PatchTeacher(id string, updates map[string]interface{}) error {
	return repository.PatchTeacher(id, updates)
}
