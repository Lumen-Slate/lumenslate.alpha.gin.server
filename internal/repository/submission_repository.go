package repository

import (
	"context"
	"lumenslate/internal/firebase"
	"lumenslate/internal/model"

	"cloud.google.com/go/firestore"
)

func SaveSubmission(s model.Submission) error {
	_, err := firebase.Client.Collection("submissions").Doc(s.ID).Set(context.Background(), s)
	return err
}

func GetSubmissionByID(id string) (*model.Submission, error) {
	doc, err := firebase.Client.Collection("submissions").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var s model.Submission
	doc.DataTo(&s)
	return &s, nil
}

func DeleteSubmission(id string) error {
	_, err := firebase.Client.Collection("submissions").Doc(id).Delete(context.Background())
	return err
}

func GetAllSubmissions(filters map[string]string) ([]model.Submission, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("submissions").Query

	if studentId, ok := filters["studentId"]; ok && studentId != "" {
		q = q.Where("studentId", "==", studentId)
	}
	if assignmentId, ok := filters["assignmentId"]; ok && assignmentId != "" {
		q = q.Where("assignmentId", "==", assignmentId)
	}

	iter := q.Documents(ctx)
	var results []model.Submission
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var s model.Submission
		doc.DataTo(&s)
		results = append(results, s)
	}
	return results, nil
}

func PatchSubmission(id string, updates map[string]interface{}) error {
	_, err := firebase.Client.Collection("submissions").Doc(id).Set(context.Background(), updates, firestore.MergeAll)
	return err
}
