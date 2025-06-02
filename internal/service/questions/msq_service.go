package questions

import (
	"server/internal/model/questions"
	repo "server/internal/repository/questions"
)

func CreateMSQ(m questions.MSQ) error {
	return repo.SaveMSQ(m)
}

func GetMSQ(id string) (*questions.MSQ, error) {
	return repo.GetMSQByID(id)
}

func DeleteMSQ(id string) error {
	return repo.DeleteMSQ(id)
}

func GetAllMSQs(filters map[string]string) ([]questions.MSQ, error) {
	return repo.GetAllMSQs(filters)
}

func UpdateMSQ(id string, updated questions.MSQ) error {
	updated.ID = id
	return repo.SaveMSQ(updated)
}

func PatchMSQ(id string, updates map[string]interface{}) error {
	return repo.PatchMSQ(id, updates)
}

func CreateBulkMSQs(msqs []questions.MSQ) error {
	return repo.SaveBulkMSQs(msqs)
}
