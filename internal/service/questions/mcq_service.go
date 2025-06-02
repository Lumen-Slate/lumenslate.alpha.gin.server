package questions

import (
	"server/internal/model/questions"
	repo "server/internal/repository/questions"
)

func CreateMCQ(m questions.MCQ) error {
	return repo.SaveMCQ(m)
}

func GetMCQ(id string) (*questions.MCQ, error) {
	return repo.GetMCQByID(id)
}

func DeleteMCQ(id string) error {
	return repo.DeleteMCQ(id)
}

func GetAllMCQs(filters map[string]string) ([]questions.MCQ, error) {
	return repo.GetAllMCQs(filters)
}

func UpdateMCQ(id string, updated questions.MCQ) error {
	updated.ID = id
	return repo.SaveMCQ(updated)
}

func PatchMCQ(id string, updates map[string]interface{}) error {
	return repo.PatchMCQ(id, updates)
}

func CreateBulkMCQs(mcqs []questions.MCQ) error {
	return repo.SaveBulkMCQs(mcqs)
}
