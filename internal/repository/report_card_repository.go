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

func SaveReportCard(reportCard model.ReportCard) (*model.ReportCard, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := db.GetCollection(db.ReportCardCollection).InsertOne(ctx, reportCard)
	if err != nil {
		return nil, err
	}

	// Set the inserted ID
	if oid, ok := result.InsertedID.(string); ok {
		reportCard.ID = oid
	}

	return &reportCard, nil
}

func GetReportCardByID(id string) (*model.ReportCard, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var reportCard model.ReportCard
	err := db.GetCollection(db.ReportCardCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&reportCard)
	if err != nil {
		return nil, err
	}
	return &reportCard, nil
}

func GetAllReportCards(filters map[string]string) ([]model.ReportCard, error) {
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
	if studentID, ok := filters["studentId"]; ok && studentID != "" {
		if id, err := strconv.Atoi(studentID); err == nil {
			filter["studentId"] = id
		}
	}
	if academicTerm, ok := filters["academicTerm"]; ok && academicTerm != "" {
		filter["academicTerm"] = academicTerm
	}

	cursor, err := db.GetCollection(db.ReportCardCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.ReportCard
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if results == nil {
		results = make([]model.ReportCard, 0)
	}
	return results, nil
}

func GetSubjectReportsByStudentID(studentID int) ([]model.SubjectReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"studentId": studentID}

	cursor, err := db.GetCollection(db.SubjectReportCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.SubjectReport
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = make([]model.SubjectReport, 0)
	}
	return results, nil
}

func DeleteReportCard(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.ReportCardCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func UpdateReportCard(id string, updates map[string]interface{}) (*model.ReportCard, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// First update the document
	_, err := db.GetCollection(db.ReportCardCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.ReportCard
	err = db.GetCollection(db.ReportCardCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
