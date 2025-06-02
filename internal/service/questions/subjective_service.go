package questions

import (
	"lumenslate/internal/model/questions"
	repo "lumenslate/internal/repository/questions"
)

func CreateSubjective(s questions.Subjective) error {
	return repo.SaveSubjective(s)
}

func GetSubjective(id string) (*questions.Subjective, error) {
	return repo.GetSubjectiveByID(id)
}

func DeleteSubjective(id string) error {
	return repo.DeleteSubjective(id)
}

func GetAllSubjectives(filters map[string]string) ([]questions.Subjective, error) {
	return repo.GetAllSubjectives(filters)
}

func UpdateSubjective(id string, updated questions.Subjective) error {
	updated.ID = id
	return repo.SaveSubjective(updated)
}

func PatchSubjective(id string, updates map[string]interface{}) error {
	return repo.PatchSubjective(id, updates)
}

func CreateBulkSubjectives(subjectives []questions.Subjective) error {
	return repo.SaveBulkSubjectives(subjectives)
}
