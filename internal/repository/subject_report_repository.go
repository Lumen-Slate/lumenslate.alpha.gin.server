package repository

import (
	"context"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveSubjectReport(report model.SubjectReport) (*model.SubjectReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := db.GetCollection(db.SubjectReportCollection).InsertOne(ctx, report)
	if err != nil {
		return nil, err
	}

	// Set the inserted ID
	if oid, ok := result.InsertedID.(string); ok {
		report.ID = oid
	}

	return &report, nil
}

func GetSubjectReportByID(id string) (*model.SubjectReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var report model.SubjectReport
	err := db.GetCollection(db.SubjectReportCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func GetAllSubjectReports(filters map[string]string) ([]model.SubjectReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()

	// Handle pagination
	limit := int64(10)
	offset := int64(0)
	if l, err := strconv.Atoi(filters["limit"]); err == nil {
		limit = int64(l)
	}
	if o, err := strconv.Atoi(filters["offset"]); err == nil {
		offset = int64(o)
	}
	findOptions.SetLimit(limit)
	findOptions.SetSkip(offset)

	// Build filter
	filter := bson.M{}
	if userID, ok := filters["userId"]; ok && userID != "" {
		filter["userId"] = userID
	}
	if subject, ok := filters["subject"]; ok && subject != "" {
		filter["subject"] = subject
	}
	if studentID, ok := filters["studentId"]; ok && studentID != "" {
		if id, err := strconv.Atoi(studentID); err == nil {
			filter["studentId"] = id
		}
	}

	cursor, err := db.GetCollection(db.SubjectReportCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.SubjectReport
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if results == nil {
		results = make([]model.SubjectReport, 0)
	}
	return results, nil
}

func DeleteSubjectReport(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.SubjectReportCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func UpdateSubjectReport(id string, updates map[string]interface{}) (*model.SubjectReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// First update the document
	_, err := db.GetCollection(db.SubjectReportCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.SubjectReport
	err = db.GetCollection(db.SubjectReportCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
