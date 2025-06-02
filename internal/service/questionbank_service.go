package service

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
)

func CreateQuestionBank(q model.QuestionBank) error {
	return repo.SaveQuestionBank(q)
}

func GetQuestionBank(id string) (*model.QuestionBank, error) {
	return repo.GetQuestionBankByID(id)
}

func DeleteQuestionBank(id string) error {
	return repo.DeleteQuestionBank(id)
}

func GetAllQuestionBanks(filters map[string]string) ([]model.QuestionBank, error) {
	return repo.GetAllQuestionBanks(filters)
}

func UpdateQuestionBank(id string, updated model.QuestionBank) error {
	updated.ID = id
	return repo.SaveQuestionBank(updated)
}

func PatchQuestionBank(id string, updates map[string]interface{}) error {
	return repo.PatchQuestionBank(id, updates)
}
