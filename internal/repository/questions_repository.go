package repository

import (
	"context"
	"fmt"
	"log"
	"lumenslate/internal/db"
	"lumenslate/internal/model"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveQuestion(question model.Questions) (*model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := db.GetCollection(db.QuestionsCollection).InsertOne(ctx, question)
	if err != nil {
		return nil, err
	}

	// Set the inserted ID
	if oid, ok := result.InsertedID.(string); ok {
		question.ID = oid
	}

	return &question, nil
}

func GetQuestionByID(id string) (*model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	var question model.Questions
	err := db.GetCollection(db.QuestionsCollection).FindOne(ctx, filter).Decode(&question)
	if err != nil {
		return nil, err
	}

	return &question, nil
}

func GetAllQuestions(filters map[string]string) ([]model.Questions, error) {
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
	filter := bson.M{} // Get all questions
	if subject, ok := filters["subject"]; ok && subject != "" {
		filter["subject"] = subject
	}
	if difficulty, ok := filters["difficulty"]; ok && difficulty != "" {
		filter["difficulty"] = difficulty
	}

	cursor, err := db.GetCollection(db.QuestionsCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Questions
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if results == nil {
		results = make([]model.Questions, 0)
	}
	return results, nil
}

func GetQuestionsBySubject(subject model.Subject) ([]model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	filter := bson.M{
		"subject": subject,
	}

	var allResults []model.Questions

	// Search through each collection
	for _, collectionName := range questionCollections {

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			log.Printf("Database query error in collection %s: %v", collectionName, err)
			continue // Skip this collection and try the next one
		}

		var results []model.Questions
		if err = cursor.All(ctx, &results); err != nil {
			log.Printf("Cursor decode error in collection %s: %v", collectionName, err)
			cursor.Close(ctx)
			continue // Skip this collection and try the next one
		}
		cursor.Close(ctx)

		allResults = append(allResults, results...)
	}

	if allResults == nil {
		allResults = make([]model.Questions, 0)
	}
	return allResults, nil
}

func GetQuestionsBySubjectAndDifficulty(subject model.Subject, difficulty model.Difficulty) ([]model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	filter := bson.M{
		"subject":    subject,
		"difficulty": difficulty,
	}

	var allResults []model.Questions

	// Search through each collection
	for _, collectionName := range questionCollections {

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			log.Printf("Database query error in collection %s: %v", collectionName, err)
			continue // Skip this collection and try the next one
		}

		var results []model.Questions
		if err = cursor.All(ctx, &results); err != nil {
			log.Printf("Cursor decode error in collection %s: %v", collectionName, err)
			cursor.Close(ctx)
			continue // Skip this collection and try the next one
		}
		cursor.Close(ctx)

		allResults = append(allResults, results...)
	}

	if allResults == nil {
		allResults = make([]model.Questions, 0)
	}
	return allResults, nil
}

func CountQuestionsBySubject(subject model.Subject) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	filter := bson.M{
		"subject": subject,
	}

	var totalCount int64 = 0

	// Count through each collection
	for _, collectionName := range questionCollections {

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			log.Printf("Database count error in collection %s: %v", collectionName, err)
			continue // Skip this collection and try the next one
		}

		totalCount += count
	}

	return totalCount, nil
}

func DeleteQuestion(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.GetCollection(db.QuestionsCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func UpdateQuestion(id string, updates map[string]interface{}) (*model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// First update the document
	_, err := db.GetCollection(db.QuestionsCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}

	// Then fetch the updated document
	var updated model.Questions
	err = db.GetCollection(db.QuestionsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func SaveBulkQuestions(questions []model.Questions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert slice to []interface{} for InsertMany
	documents := make([]interface{}, len(questions))
	for i, q := range questions {
		documents[i] = q
	}

	_, err := db.GetCollection(db.QuestionsCollection).InsertMany(ctx, documents)
	return err
}

// DebugGetAllSubjects returns all unique subjects in the database for debugging
func DebugGetAllSubjects() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$subject", "count": bson.M{"$sum": 1}}},
	}

	// Aggregate subjects from all collections
	subjectCounts := make(map[string]int)

	for _, collectionName := range questionCollections {

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			log.Printf("Aggregation error in collection %s: %v", collectionName, err)
			continue
		}

		var results []struct {
			Subject string `bson:"_id"`
			Count   int    `bson:"count"`
		}

		if err = cursor.All(ctx, &results); err != nil {
			log.Printf("Cursor decode error in collection %s: %v", collectionName, err)
			cursor.Close(ctx)
			continue
		}
		cursor.Close(ctx)

		for _, result := range results {
			subjectCounts[result.Subject] += result.Count
		}
	}

	subjects := make([]string, 0, len(subjectCounts))
	for subject, count := range subjectCounts {
		subjects = append(subjects, fmt.Sprintf("%s (%d questions)", subject, count))
	}

	return subjects, nil
}

// DebugGetSampleQuestions returns a few sample questions for debugging
func DebugGetSampleQuestions(limit int) ([]model.Questions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find().SetLimit(int64(limit))

	cursor, err := db.GetCollection(db.QuestionsCollection).Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Questions
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Cursor decode error: %v", err)
		return nil, err
	}

	return results, nil
}

// DebugTestDatabaseConnection tests if we can connect to the database and collection
func DebugTestDatabaseConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetCollection(db.QuestionsCollection)

	// Try to count total documents
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("ERROR counting documents: %v", err)
		return err
	}

	// Try to get database stats
	if count > 0 {
		// Get first document to check structure
		var firstDoc bson.M
		err = collection.FindOne(ctx, bson.M{}).Decode(&firstDoc)
		if err != nil {
			log.Printf("ERROR getting first document: %v", err)
		} else {
			log.Printf("First document structure: %+v", firstDoc)
		}
	}

	return nil
}

// DebugFindMathQuestions tries to find math questions using different search strategies
func DebugFindMathQuestions() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.GetCollection(db.QuestionsCollection)

	// Strategy 1: Case-insensitive search for "math"
	filter1 := bson.M{"subject": bson.M{"$regex": "math", "$options": "i"}}
	count1, err := collection.CountDocuments(ctx, filter1)
	if err != nil {
		log.Printf("Strategy 1 ERROR: %v", err)
	} else {
		log.Printf("Strategy 1 - Found %d documents with case-insensitive 'math'", count1)
		if count1 > 0 {
			var samples []bson.M
			cursor, err := collection.Find(ctx, filter1, options.Find().SetLimit(3))
			if err == nil {
				cursor.All(ctx, &samples)
				cursor.Close(ctx)
				for i, sample := range samples {
					log.Printf("Strategy 1 Sample %d: %+v", i+1, sample)
				}
			}
		}
	}

	// Strategy 2: Search for any subject that contains "math"
	filter2 := bson.M{"$or": []bson.M{
		{"subject": bson.M{"$regex": "math", "$options": "i"}},
		{"topic": bson.M{"$regex": "math", "$options": "i"}},
		{"category": bson.M{"$regex": "math", "$options": "i"}},
	}}
	count2, err := collection.CountDocuments(ctx, filter2)
	if err != nil {
		log.Printf("Strategy 2 ERROR: %v", err)
	} else {
		log.Printf("Strategy 2 - Found %d documents with 'math' in any field", count2)
	}

	// Strategy 3: Get all unique values for the subject field
	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$subject", "count": bson.M{"$sum": 1}, "sample_id": bson.M{"$first": "$_id"}}},
		{"$sort": bson.M{"count": -1}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Strategy 3 ERROR: %v", err)
	} else {
		var results []bson.M
		err = cursor.All(ctx, &results)
		cursor.Close(ctx)
		if err != nil {
			log.Printf("Strategy 3 decode ERROR: %v", err)
		} else {
			log.Printf("Strategy 3 - Found %d unique subject values:", len(results))
			for i, result := range results {
				log.Printf("Subject %d: '%v' (%v documents)", i+1, result["_id"], result["count"])
			}
		}
	}

	// Strategy 4: Get a few random documents to see the actual structure
	pipeline2 := []bson.M{
		{"$sample": bson.M{"size": 5}},
	}
	cursor2, err := collection.Aggregate(ctx, pipeline2)
	if err != nil {
		log.Printf("Strategy 4 ERROR: %v", err)
	} else {
		var randomDocs []bson.M
		err = cursor2.All(ctx, &randomDocs)
		cursor2.Close(ctx)
		if err != nil {
			log.Printf("Strategy 4 decode ERROR: %v", err)
		} else {
			log.Printf("Strategy 4 - Random documents:")
			for i, doc := range randomDocs {
				log.Printf("Random doc %d: %+v", i+1, doc)
			}
		}
	}

	return nil
}

// DebugListAllCollections lists all collections in the current database
func DebugListAllCollections() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the database instance
	database := db.GetCollection(db.QuestionsCollection).Database()

	// List all collections
	collections, err := database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Printf("ERROR listing collections: %v", err)
		return err
	}

	for _, collName := range collections {

		// Count documents in each collection
		coll := database.Collection(collName)
		count, err := coll.CountDocuments(ctx, bson.M{})
		if err != nil {
			log.Printf("Collection %s - ERROR counting documents: %v", collName, err)
		} else {
			log.Printf("Collection %s - document count: %d", collName, count)

			// If this collection has documents, show a sample
			if count > 0 {
				var sample bson.M
				err = coll.FindOne(ctx, bson.M{}).Decode(&sample)
				if err != nil {
					log.Printf("Collection %s - ERROR getting sample: %v", collName, err)
				} else {
					log.Printf("Collection %s - sample document: %+v", collName, sample)
				}
			}
		}
	}

	return nil
}
