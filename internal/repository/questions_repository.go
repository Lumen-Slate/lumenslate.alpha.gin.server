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
	log.Printf("DEBUG: GetQuestionsBySubject called with subject: '%s'", subject)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	filter := bson.M{
		"subject": subject,
	}

	log.Printf("DEBUG: Database query filter: %+v", filter)
	log.Printf("DEBUG: GetQuestionsBySubject - searching across collections: %v", questionCollections)

	var allResults []model.Questions

	// Search through each collection
	for _, collectionName := range questionCollections {
		log.Printf("DEBUG: Searching collection: %s", collectionName)

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			log.Printf("DEBUG: Database query error in collection %s: %v", collectionName, err)
			continue // Skip this collection and try the next one
		}

		var results []model.Questions
		if err = cursor.All(ctx, &results); err != nil {
			log.Printf("DEBUG: Cursor decode error in collection %s: %v", collectionName, err)
			cursor.Close(ctx)
			continue // Skip this collection and try the next one
		}
		cursor.Close(ctx)

		log.Printf("DEBUG: Found %d questions in collection '%s' for subject '%s'", len(results), collectionName, subject)

		allResults = append(allResults, results...)
	}

	log.Printf("DEBUG: GetQuestionsBySubject - total found %d questions across all collections for subject '%s'", len(allResults), subject)

	if allResults == nil {
		allResults = make([]model.Questions, 0)
	}
	return allResults, nil
}

func GetQuestionsBySubjectAndDifficulty(subject model.Subject, difficulty model.Difficulty) ([]model.Questions, error) {
	log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty called with subject: '%s', difficulty: '%s'", subject, difficulty)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	filter := bson.M{
		"subject":    subject,
		"difficulty": difficulty,
	}

	log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty filter: %+v", filter)
	log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty - searching across collections: %v", questionCollections)

	var allResults []model.Questions

	// Search through each collection
	for _, collectionName := range questionCollections {
		log.Printf("DEBUG: Searching collection: %s", collectionName)

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty - database query error in collection %s: %v", collectionName, err)
			continue // Skip this collection and try the next one
		}

		var results []model.Questions
		if err = cursor.All(ctx, &results); err != nil {
			log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty - cursor decode error in collection %s: %v", collectionName, err)
			cursor.Close(ctx)
			continue // Skip this collection and try the next one
		}
		cursor.Close(ctx)

		log.Printf("DEBUG: Found %d questions in collection '%s' for subject '%s' and difficulty '%s'", len(results), collectionName, subject, difficulty)

		allResults = append(allResults, results...)
	}

	log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty - total found %d questions across all collections for subject '%s' and difficulty '%s'", len(allResults), subject, difficulty)
	for i, q := range allResults {
		if i < 3 { // Only log first 3 for brevity
			log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty - question %d: subject='%s', difficulty='%s', id='%s'", i+1, q.Subject, q.Difficulty, q.ID)
		}
	}

	if allResults == nil {
		allResults = make([]model.Questions, 0)
		log.Printf("DEBUG: GetQuestionsBySubjectAndDifficulty - results was nil, returning empty slice")
	}
	return allResults, nil
}

func CountQuestionsBySubject(subject model.Subject) (int64, error) {
	log.Printf("DEBUG: CountQuestionsBySubject called with subject: '%s'", subject)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	filter := bson.M{
		"subject": subject,
	}

	log.Printf("DEBUG: CountQuestionsBySubject filter: %+v", filter)
	log.Printf("DEBUG: CountQuestionsBySubject - counting across collections: %v", questionCollections)

	var totalCount int64 = 0

	// Count through each collection
	for _, collectionName := range questionCollections {
		log.Printf("DEBUG: Counting in collection: %s", collectionName)

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			log.Printf("DEBUG: CountQuestionsBySubject - database count error in collection %s: %v", collectionName, err)
			continue // Skip this collection and try the next one
		}

		log.Printf("DEBUG: Found %d questions in collection '%s' for subject '%s'", count, collectionName, subject)
		totalCount += count
	}

	log.Printf("DEBUG: CountQuestionsBySubject - total found %d questions across all collections for subject '%s'", totalCount, subject)
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
	log.Printf("DEBUG: DebugGetAllSubjects called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define all question collections
	questionCollections := []string{"mcqs", "msqs", "nats", "subjectives"}

	log.Printf("DEBUG: DebugGetAllSubjects - searching across collections: %v", questionCollections)

	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$subject", "count": bson.M{"$sum": 1}}},
	}

	log.Printf("DEBUG: DebugGetAllSubjects - using aggregation pipeline: %+v", pipeline)

	// Aggregate subjects from all collections
	subjectCounts := make(map[string]int)

	for _, collectionName := range questionCollections {
		log.Printf("DEBUG: DebugGetAllSubjects - checking collection: %s", collectionName)

		database := db.GetCollection(db.QuestionsCollection).Database()
		collection := database.Collection(collectionName)

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			log.Printf("DEBUG: DebugGetAllSubjects - aggregation error in collection %s: %v", collectionName, err)
			continue
		}

		var results []struct {
			Subject string `bson:"_id"`
			Count   int    `bson:"count"`
		}

		if err = cursor.All(ctx, &results); err != nil {
			log.Printf("DEBUG: DebugGetAllSubjects - cursor decode error in collection %s: %v", collectionName, err)
			cursor.Close(ctx)
			continue
		}
		cursor.Close(ctx)

		log.Printf("DEBUG: DebugGetAllSubjects - collection %s returned %d subject groups", collectionName, len(results))
		for i, result := range results {
			log.Printf("DEBUG: DebugGetAllSubjects - collection %s, result %d: subject='%s', count=%d", collectionName, i+1, result.Subject, result.Count)
			subjectCounts[result.Subject] += result.Count
		}
	}

	log.Printf("DEBUG: DebugGetAllSubjects - total unique subjects found: %d", len(subjectCounts))
	for subject, count := range subjectCounts {
		log.Printf("DEBUG: DebugGetAllSubjects - subject='%s', total_count=%d", subject, count)
	}

	subjects := make([]string, 0, len(subjectCounts))
	for subject, count := range subjectCounts {
		subjects = append(subjects, fmt.Sprintf("%s (%d questions)", subject, count))
	}

	log.Printf("DEBUG: DebugGetAllSubjects - final subjects list: %v", subjects)
	return subjects, nil
}

// DebugGetSampleQuestions returns a few sample questions for debugging
func DebugGetSampleQuestions(limit int) ([]model.Questions, error) {
	log.Printf("DEBUG: DebugGetSampleQuestions called with limit: %d", limit)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find().SetLimit(int64(limit))
	log.Printf("DEBUG: DebugGetSampleQuestions - using empty filter with limit: %d", limit)

	cursor, err := db.GetCollection(db.QuestionsCollection).Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("DEBUG: DebugGetSampleQuestions - database query error: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Questions
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("DEBUG: DebugGetSampleQuestions - cursor decode error: %v", err)
		return nil, err
	}

	log.Printf("DEBUG: DebugGetSampleQuestions - successfully retrieved %d sample questions", len(results))
	log.Printf("DEBUG: Sample questions from database:")
	for i, q := range results {
		log.Printf("DEBUG: Question %d - ID: '%s', Subject: '%s', Difficulty: '%s', IsActive: %v, Question: '%.50s...'",
			i+1, q.ID, q.Subject, q.Difficulty, q.IsActive, q.Question)
	}

	return results, nil
}

// DebugTestDatabaseConnection tests if we can connect to the database and collection
func DebugTestDatabaseConnection() error {
	log.Printf("DEBUG: DebugTestDatabaseConnection - testing database connectivity")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetCollection(db.QuestionsCollection)
	log.Printf("DEBUG: DebugTestDatabaseConnection - collection name: %s", db.QuestionsCollection)

	// Try to count total documents
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("DEBUG: DebugTestDatabaseConnection - ERROR counting documents: %v", err)
		return err
	}

	log.Printf("DEBUG: DebugTestDatabaseConnection - SUCCESS! Total documents in collection: %d", count)

	// Try to get database stats
	if count > 0 {
		// Get first document to check structure
		var firstDoc bson.M
		err = collection.FindOne(ctx, bson.M{}).Decode(&firstDoc)
		if err != nil {
			log.Printf("DEBUG: DebugTestDatabaseConnection - ERROR getting first document: %v", err)
		} else {
			log.Printf("DEBUG: DebugTestDatabaseConnection - First document structure: %+v", firstDoc)
		}
	}

	return nil
}

// DebugFindMathQuestions tries to find math questions using different search strategies
func DebugFindMathQuestions() error {
	log.Printf("DEBUG: DebugFindMathQuestions - searching for math questions using different strategies")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.GetCollection(db.QuestionsCollection)

	// Strategy 1: Case-insensitive search for "math"
	log.Printf("DEBUG: Strategy 1 - Case-insensitive search for 'math'")
	filter1 := bson.M{"subject": bson.M{"$regex": "math", "$options": "i"}}
	count1, err := collection.CountDocuments(ctx, filter1)
	if err != nil {
		log.Printf("DEBUG: Strategy 1 ERROR: %v", err)
	} else {
		log.Printf("DEBUG: Strategy 1 - Found %d documents with case-insensitive 'math'", count1)
		if count1 > 0 {
			var samples []bson.M
			cursor, err := collection.Find(ctx, filter1, options.Find().SetLimit(3))
			if err == nil {
				cursor.All(ctx, &samples)
				cursor.Close(ctx)
				for i, sample := range samples {
					log.Printf("DEBUG: Strategy 1 Sample %d: %+v", i+1, sample)
				}
			}
		}
	}

	// Strategy 2: Search for any subject that contains "math"
	log.Printf("DEBUG: Strategy 2 - Search for any field containing 'math'")
	filter2 := bson.M{"$or": []bson.M{
		{"subject": bson.M{"$regex": "math", "$options": "i"}},
		{"topic": bson.M{"$regex": "math", "$options": "i"}},
		{"category": bson.M{"$regex": "math", "$options": "i"}},
	}}
	count2, err := collection.CountDocuments(ctx, filter2)
	if err != nil {
		log.Printf("DEBUG: Strategy 2 ERROR: %v", err)
	} else {
		log.Printf("DEBUG: Strategy 2 - Found %d documents with 'math' in any field", count2)
	}

	// Strategy 3: Get all unique values for the subject field
	log.Printf("DEBUG: Strategy 3 - Get all unique subject values")
	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$subject", "count": bson.M{"$sum": 1}, "sample_id": bson.M{"$first": "$_id"}}},
		{"$sort": bson.M{"count": -1}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("DEBUG: Strategy 3 ERROR: %v", err)
	} else {
		var results []bson.M
		err = cursor.All(ctx, &results)
		cursor.Close(ctx)
		if err != nil {
			log.Printf("DEBUG: Strategy 3 decode ERROR: %v", err)
		} else {
			log.Printf("DEBUG: Strategy 3 - Found %d unique subject values:", len(results))
			for i, result := range results {
				log.Printf("DEBUG: Subject %d: '%v' (%v documents)", i+1, result["_id"], result["count"])
			}
		}
	}

	// Strategy 4: Get a few random documents to see the actual structure
	log.Printf("DEBUG: Strategy 4 - Get random documents to check structure")
	pipeline2 := []bson.M{
		{"$sample": bson.M{"size": 5}},
	}
	cursor2, err := collection.Aggregate(ctx, pipeline2)
	if err != nil {
		log.Printf("DEBUG: Strategy 4 ERROR: %v", err)
	} else {
		var randomDocs []bson.M
		err = cursor2.All(ctx, &randomDocs)
		cursor2.Close(ctx)
		if err != nil {
			log.Printf("DEBUG: Strategy 4 decode ERROR: %v", err)
		} else {
			log.Printf("DEBUG: Strategy 4 - Random documents:")
			for i, doc := range randomDocs {
				log.Printf("DEBUG: Random doc %d: %+v", i+1, doc)
			}
		}
	}

	return nil
}

// DebugListAllCollections lists all collections in the current database
func DebugListAllCollections() error {
	log.Printf("DEBUG: DebugListAllCollections - listing all collections in current database")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the database instance
	database := db.GetCollection(db.QuestionsCollection).Database()
	log.Printf("DEBUG: DebugListAllCollections - database name: %s", database.Name())

	// List all collections
	collections, err := database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Printf("DEBUG: DebugListAllCollections - ERROR listing collections: %v", err)
		return err
	}

	log.Printf("DEBUG: DebugListAllCollections - found %d collections:", len(collections))
	for i, collName := range collections {
		log.Printf("DEBUG: Collection %d: %s", i+1, collName)

		// Count documents in each collection
		coll := database.Collection(collName)
		count, err := coll.CountDocuments(ctx, bson.M{})
		if err != nil {
			log.Printf("DEBUG: Collection %s - ERROR counting documents: %v", collName, err)
		} else {
			log.Printf("DEBUG: Collection %s - document count: %d", collName, count)

			// If this collection has documents, show a sample
			if count > 0 {
				var sample bson.M
				err = coll.FindOne(ctx, bson.M{}).Decode(&sample)
				if err != nil {
					log.Printf("DEBUG: Collection %s - ERROR getting sample: %v", collName, err)
				} else {
					log.Printf("DEBUG: Collection %s - sample document: %+v", collName, sample)
				}
			}
		}
		log.Printf("DEBUG: ---")
	}

	return nil
}
