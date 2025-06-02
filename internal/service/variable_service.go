package service

import (
	"server/internal/model"
	repo "server/internal/repository"
)

func CreateVariable(v model.Variable) error {
	return repo.SaveVariable(v)
}

func GetVariable(id string) (*model.Variable, error) {
	return repo.GetVariableByID(id)
}

func DeleteVariable(id string) error {
	return repo.DeleteVariable(id)
}

func GetAllVariables(filters map[string]string) ([]model.Variable, error) {
	return repo.GetAllVariables(filters)
}

func UpdateVariable(id string, updated model.Variable) error {
	updated.ID = id
	return repo.SaveVariable(updated)
}

func PatchVariable(id string, updates map[string]interface{}) error {
	return repo.PatchVariable(id, updates)
}

func CreateBulkVariables(variables []model.Variable) error {
	return repo.SaveBulkVariables(variables)
}
