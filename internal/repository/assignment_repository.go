package repository

import (
	"context"
	"log"
	"server/internal/firebase"
	"server/internal/model"
	"strconv"
	"time"
)

func SaveAssignment(a model.Assignment) error {
	_, err := firebase.Client.Collection("assignments").Doc(a.ID).Set(context.Background(), a)
	return err
}

func GetAssignmentByID(id string) (*model.Assignment, error) {
	doc, err := firebase.Client.Collection("assignments").Doc(id).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var a model.Assignment
	doc.DataTo(&a)
	return &a, nil
}

func DeleteAssignment(id string) error {
	_, err := firebase.Client.Collection("assignments").Doc(id).Delete(context.Background())
	return err
}

func GetAllAssignments() ([]model.Assignment, error) {
	iter := firebase.Client.Collection("assignments").Documents(context.Background())
	var assignments []model.Assignment
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var a model.Assignment
		doc.DataTo(&a)
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func FilterAssignments(limitStr, offsetStr, points, due string) ([]model.Assignment, error) {
	ctx := context.Background()
	q := firebase.Client.Collection("assignments").Query

	// Filters
	if points != "" {
		val, _ := strconv.Atoi(points)
		q = q.Where("points", "==", val)
	}
	if due != "" {
		t, err := time.Parse(time.RFC3339, due)
		if err == nil {
			q = q.Where("dueDate", ">=", t)
		} else {
			log.Printf("⚠️ Invalid dueDate format: %v", due)
		}
	}

	// Pagination
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 0 {
		q = q.Limit(limit)
	}
	iter := q.Documents(ctx)

	// Apply offset manually
	var results []model.Assignment
	i := 0
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		if i < offset {
			i++
			continue
		}
		var a model.Assignment
		doc.DataTo(&a)
		results = append(results, a)
		i++
	}
	return results, nil
}
