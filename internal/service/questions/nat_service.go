package questions

import (
	"server/internal/model/questions"
	repo "server/internal/repository/questions"
)

func CreateNAT(n questions.NAT) error {
	return repo.SaveNAT(n)
}

func GetNAT(id string) (*questions.NAT, error) {
	return repo.GetNATByID(id)
}

func DeleteNAT(id string) error {
	return repo.DeleteNAT(id)
}

func GetAllNATs(filters map[string]string) ([]questions.NAT, error) {
	return repo.GetAllNATs(filters)
}

func UpdateNAT(id string, updated questions.NAT) error {
	updated.ID = id
	return repo.SaveNAT(updated)
}

func PatchNAT(id string, updates map[string]interface{}) error {
	return repo.PatchNAT(id, updates)
}

func CreateBulkNATs(nats []questions.NAT) error {
	return repo.SaveBulkNATs(nats)
}
