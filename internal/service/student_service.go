package service

import (
	"lumenslate/internal/model"
	"lumenslate/internal/repository"
)

func CreateStudent(s model.Student) error {
	return repository.SaveStudent(s)
}

func GetStudent(id string) (*model.Student, error) {
	return repository.GetStudentByID(id)
}

func DeleteStudent(id string) error {
	return repository.DeleteStudent(id)
}

func GetAllStudents(filters map[string]string) ([]model.Student, error) {
	return repository.GetAllStudents(filters)
}

func UpdateStudent(id string, updated model.Student) error {
	updated.ID = id
	return repository.SaveStudent(updated)
}

func PatchStudent(id string, updates map[string]interface{}) error {
	return repository.PatchStudent(id, updates)
}
