package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"lumenslate/internal/model"
	pb "lumenslate/internal/proto/ai_service"
	"lumenslate/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuestionRequest represents a question request from the agent
type QuestionRequest struct {
	Type              string `json:"type"`
	Subject           string `json:"subject"`
	NumberOfQuestions int    `json:"number_of_questions"`
	Difficulty        string `json:"difficulty"`
}

// AgentResponse represents the parsed agent response structure
type AgentResponse struct {
	QuestionsRequested []QuestionRequest `json:"questions_requested"`
	AssessmentData     interface{}       `json:"assessment_data"`
	AssignmentResult   interface{}       `json:"assignment_result"`
	ReportCardData     interface{}       `json:"report_card_data"`
}

// AssessmentData represents the assessment data structure
type AssessmentData struct {
	StudentID   interface{} `json:"student_id"`
	StudentName string      `json:"student_name"`
	Subject     string      `json:"subject"`
	Score       interface{} `json:"score"`

	// Optional fields
	GradeLetter    string `json:"grade_letter,omitempty"`
	ClassName      string `json:"class_name,omitempty"`
	InstructorName string `json:"instructor_name,omitempty"`
	Term           string `json:"term,omitempty"`
	Remarks        string `json:"remarks,omitempty"`

	// Assessment component breakdown
	MidtermScore          interface{} `json:"midterm_score,omitempty"`
	FinalExamScore        interface{} `json:"final_exam_score,omitempty"`
	QuizScore             interface{} `json:"quiz_score,omitempty"`
	AssignmentScore       interface{} `json:"assignment_score,omitempty"`
	PracticalScore        interface{} `json:"practical_score,omitempty"`
	OralPresentationScore interface{} `json:"oral_presentation_score,omitempty"`

	// Skill evaluation (0-10 scale or %)
	ConceptualUnderstanding interface{} `json:"conceptual_understanding,omitempty"`
	ProblemSolving          interface{} `json:"problem_solving,omitempty"`
	KnowledgeApplication    interface{} `json:"knowledge_application,omitempty"`
	AnalyticalThinking      interface{} `json:"analytical_thinking,omitempty"`
	Creativity              interface{} `json:"creativity,omitempty"`
	PracticalSkills         interface{} `json:"practical_skills,omitempty"`

	// Behavioral metrics
	Participation interface{} `json:"participation,omitempty"`
	Discipline    interface{} `json:"discipline,omitempty"`
	Punctuality   interface{} `json:"punctuality,omitempty"`
	Teamwork      interface{} `json:"teamwork,omitempty"`
	EffortLevel   interface{} `json:"effort_level,omitempty"`
	Improvement   interface{} `json:"improvement,omitempty"`

	// Advanced insights
	LearningObjectivesMastered string `json:"learning_objectives_mastered,omitempty"`
	AreasForImprovement        string `json:"areas_for_improvement,omitempty"`
	RecommendedResources       string `json:"recommended_resources,omitempty"`
	TargetGoals                string `json:"target_goals,omitempty"`
}

func Agent(file, fileType, teacherId, role, message, createdAt, updatedAt string) (map[string]interface{}, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.AgentRequest{
		File:      file,
		FileType:  fileType,
		TeacherId: teacherId,
		Role:      role,
		Message:   message,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	res, err := client.Agent(ctx, req)
	if err != nil {
		return createErrorResponse(teacherId, err.Error(), res), nil
	}

	// Get the raw agent response
	rawAgentResponse := res.GetAgentResponse()

	// Try to parse the agent response as JSON to see if it contains database operations
	var agentResponse AgentResponse
	agentName := "root_agent"
	responseMessage := "Agent response processed successfully"
	var responseData interface{}

	if err := json.Unmarshal([]byte(rawAgentResponse), &agentResponse); err == nil {
		// Successfully parsed JSON, check for database operations
		if len(agentResponse.QuestionsRequested) > 0 {
			// Handle question generation
			if questionsData, err := handleQuestionGeneration(agentResponse.QuestionsRequested, teacherId); err == nil {
				responseData = questionsData
				agentName = "assignment_generator_general"
				responseMessage = "Assignment generated successfully"
			} else {
				log.Printf("Error handling question generation: %v", err)
				return createErrorResponse(teacherId, err.Error(), res), nil
			}
		} else if agentResponse.AssignmentResult != nil {
			// Handle assignment result saving (from assessor agent)
			if assignmentResult, err := handleAssignmentResultSaving(agentResponse.AssignmentResult, teacherId); err == nil {
				responseData = assignmentResult
				agentName = "assessor_agent"
				responseMessage = "Assignment assessment completed successfully"
			} else {
				log.Printf("Error handling assignment result saving: %v", err)
				return createErrorResponse(teacherId, err.Error(), res), nil
			}
		} else if agentResponse.AssessmentData != nil {
			// Handle legacy assessment data saving (for backward compatibility)
			if assessmentData, err := handleAssessmentSaving(agentResponse.AssessmentData, teacherId); err == nil {
				responseData = assessmentData
				agentName = "assessor_agent"
				responseMessage = "Subject assessment report processed successfully"
			} else {
				log.Printf("Error handling assessment saving: %v", err)
				return createErrorResponse(teacherId, err.Error(), res), nil
			}
		} else if agentResponse.ReportCardData != nil {
			// Handle report card generation and saving
			if reportCardData, err := handleReportCardGeneration(agentResponse.ReportCardData, teacherId); err == nil {
				responseData = reportCardData
				agentName = "report_card_generator"
				responseMessage = "Report card generated and saved successfully"
			} else {
				log.Printf("Error handling report card generation: %v", err)
				return createErrorResponse(teacherId, err.Error(), res), nil
			}
		} else {
			// Regular agent response without database operations
			responseData = map[string]interface{}{
				"agentResponse": rawAgentResponse,
			}
		}
	} else {
		// Couldn't parse as JSON, treat as regular response
		responseData = map[string]interface{}{
			"agentResponse": rawAgentResponse,
		}
	}

	// Return the standardized response format
	return map[string]interface{}{
		"message":      responseMessage,
		"teacherId":    res.GetTeacherId(),
		"agentName":    agentName,
		"data":         responseData,
		"sessionId":    res.GetSessionId(),
		"createdAt":    res.GetCreatedAt(),
		"updatedAt":    res.GetUpdatedAt(),
		"responseTime": res.GetResponseTime(),
		"role":         res.GetRole(),
		"feedback":     res.GetFeedback(),
	}, nil
}

func createErrorResponse(teacherId, errorMessage string, res *pb.AgentResponse) map[string]interface{} {
	var sessionId, createdAt, updatedAt, responseTime, feedback string
	if res != nil {
		sessionId = res.GetSessionId()
		createdAt = res.GetCreatedAt()
		updatedAt = res.GetUpdatedAt()
		responseTime = res.GetResponseTime()
		feedback = res.GetFeedback()
	}

	return map[string]interface{}{
		"message":      errorMessage,
		"teacherId":    teacherId,
		"agentName":    "root_agent",
		"data":         map[string]interface{}{},
		"sessionId":    sessionId,
		"createdAt":    createdAt,
		"updatedAt":    updatedAt,
		"responseTime": responseTime,
		"role":         "agent",
		"feedback":     feedback,
	}
}

func handleQuestionGeneration(questionsRequested []QuestionRequest, teacherId string) (map[string]interface{}, error) {
	// Debug: Check what's actually in the database
	log.Printf("=== DEBUG: Starting question generation for teacher: %s ===", teacherId)
	log.Printf("DEBUG: Total requests received: %d", len(questionsRequested))

	// Show the raw requests first
	for i, req := range questionsRequested {
		log.Printf("DEBUG: Request %d - Type: '%s', Subject: '%s', NumberOfQuestions: %d, Difficulty: '%s'",
			i+1, req.Type, req.Subject, req.NumberOfQuestions, req.Difficulty)
	}

	log.Printf("DEBUG: Checking database connectivity...")
	if err := repository.DebugTestDatabaseConnection(); err != nil {
		log.Printf("ERROR: Database connectivity test failed: %v", err)
		return nil, fmt.Errorf("database connectivity failed: %v", err)
	}

	log.Printf("DEBUG: Checking database contents...")
	repository.DebugGetAllSubjects()
	repository.DebugGetSampleQuestions(5)

	log.Printf("DEBUG: Running advanced math question search...")
	repository.DebugFindMathQuestions()

	log.Printf("DEBUG: Listing all collections in database...")
	repository.DebugListAllCollections()

	// Group requests by subject to handle multiple difficulty levels for the same subject
	subjectRequests := make(map[string][]QuestionRequest)

	for _, request := range questionsRequested {
		if request.Type == "assignment_generator_general" {
			subject := strings.ToLower(strings.TrimSpace(request.Subject))
			log.Printf("DEBUG: Processing request for subject: '%s' -> normalized: '%s'", request.Subject, subject)
			subjectRequests[subject] = append(subjectRequests[subject], request)
		}
	}

	totalSubjectsRequested := len(subjectRequests)
	totalQuestionsReturned := 0
	subjects := []map[string]interface{}{}

	for subjectKey, requests := range subjectRequests {
		// Convert subject string to Subject enum
		subject, validSubject := model.GetSubjectFromString(subjectKey)
		displaySubject := strings.Title(subjectKey)

		log.Printf("DEBUG: Subject conversion - key: '%s' -> enum: '%s', valid: %v", subjectKey, subject, validSubject)

		if !validSubject {
			// Invalid subject
			totalRequested := 0
			for _, req := range requests {
				totalRequested += req.NumberOfQuestions
			}
			subjectData := map[string]interface{}{
				"subject":                 displaySubject,
				"questionsRequestedCount": totalRequested,
				"questionsAvailableCount": 0,
				"questionsReturnedCount":  0,
				"questions":               []map[string]interface{}{},
				"message":                 fmt.Sprintf("Subject '%s' is not supported. Available subjects: Math, Science, English, History, Geography", displaySubject),
			}
			subjects = append(subjects, subjectData)
			continue
		}

		// Process all requests for this subject
		allQuestionsForSubject := []model.Questions{}
		totalRequested := 0

		for _, req := range requests {
			numQuestions := req.NumberOfQuestions
			difficulty := strings.ToLower(strings.TrimSpace(req.Difficulty))
			totalRequested += numQuestions

			log.Printf("DEBUG: Processing request - subject: '%s', difficulty: '%s', count: %d", subject, difficulty, numQuestions)

			var availableQuestions []model.Questions
			var err error

			// Query database for questions from this subject with specific difficulty
			if difficulty != "" {
				// Convert difficulty string to Difficulty enum
				var difficultyEnum model.Difficulty
				switch difficulty {
				case "easy":
					difficultyEnum = model.DifficultyEasy
				case "medium":
					difficultyEnum = model.DifficultyMedium
				case "hard":
					difficultyEnum = model.DifficultyHard
				default:
					// Invalid difficulty, get all questions for this subject
					log.Printf("DEBUG: Invalid difficulty '%s', getting all questions for subject '%s'", difficulty, subject)
					availableQuestions, err = repository.GetQuestionsBySubject(subject)
				}

				if difficultyEnum != "" {
					log.Printf("DEBUG: Getting questions by subject '%s' and difficulty '%s'", subject, difficultyEnum)
					availableQuestions, err = repository.GetQuestionsBySubjectAndDifficulty(subject, difficultyEnum)
				}
			} else {
				// No difficulty specified, get all questions for this subject
				log.Printf("DEBUG: No difficulty specified, getting all questions for subject '%s'", subject)
				availableQuestions, err = repository.GetQuestionsBySubject(subject)
			}

			if err != nil {
				log.Printf("Error getting questions for subject %s: %v", subject, err)
				continue
			}

			log.Printf("DEBUG: Database query result - subject: '%s', available: %d, requested: %d", subject, len(availableQuestions), numQuestions)

			if len(availableQuestions) == 0 {
				// No questions available for this difficulty
				log.Printf("DEBUG: No questions found for subject '%s' with difficulty '%s'", subject, difficulty)
				continue
			} else if numQuestions > len(availableQuestions) {
				// Not enough questions for this difficulty, skip this request
				log.Printf("DEBUG: Not enough questions - requested: %d, available: %d", numQuestions, len(availableQuestions))
				continue
			} else {
				// Randomly sample the requested number of questions
				rand.Seed(time.Now().UnixNano())
				selectedQuestions := make([]model.Questions, numQuestions)
				perm := rand.Perm(len(availableQuestions))
				for i := 0; i < numQuestions; i++ {
					selectedQuestions[i] = availableQuestions[perm[i]]
				}
				allQuestionsForSubject = append(allQuestionsForSubject, selectedQuestions...)
				log.Printf("DEBUG: Successfully selected %d questions for subject '%s'", numQuestions, subject)
			}
		}

		// Get total available questions for this subject
		totalAvailable, err := repository.CountQuestionsBySubject(subject)
		if err != nil {
			log.Printf("DEBUG: Error counting questions for subject '%s': %v", subject, err)
			totalAvailable = 0
		}

		log.Printf("DEBUG: Final counts - subject: '%s', total available: %d, total requested: %d, total selected: %d", subject, totalAvailable, totalRequested, len(allQuestionsForSubject))

		// Check if we got all requested questions
		if len(allQuestionsForSubject) < totalRequested {
			subjectData := map[string]interface{}{
				"subject":                 displaySubject,
				"questionsRequestedCount": totalRequested,
				"questionsAvailableCount": int(totalAvailable),
				"questionsReturnedCount":  0,
				"questions":               []map[string]interface{}{},
				"message":                 fmt.Sprintf("Could not fulfill all requests for %s. Some difficulty levels may not have enough questions available.", displaySubject),
			}
			subjects = append(subjects, subjectData)
		} else if len(allQuestionsForSubject) == 0 {
			subjectData := map[string]interface{}{
				"subject":                 displaySubject,
				"questionsRequestedCount": totalRequested,
				"questionsAvailableCount": int(totalAvailable),
				"questionsReturnedCount":  0,
				"questions":               []map[string]interface{}{},
				"message":                 fmt.Sprintf("No questions available for %s with the requested difficulty levels", displaySubject),
			}
			subjects = append(subjects, subjectData)
		} else {
			// Successfully got all requested questions
			// Format questions for response
			formattedQuestions := make([]map[string]interface{}, len(allQuestionsForSubject))
			for i, q := range allQuestionsForSubject {
				questionData := map[string]interface{}{
					"questionId": q.ID,
					"question":   q.Question,
					"type":       "MCQ", // Default to MCQ, you can enhance this based on your question structure
					"answer":     q.Answer,
					"difficulty": string(q.Difficulty),
				}

				// Add options only for MCQ and MSQ types
				if len(q.Options) > 0 {
					questionData["options"] = q.Options
				}

				formattedQuestions[i] = questionData
			}

			subjectData := map[string]interface{}{
				"subject":                 displaySubject,
				"questionsRequestedCount": totalRequested,
				"questionsAvailableCount": int(totalAvailable),
				"questionsReturnedCount":  len(formattedQuestions),
				"questions":               formattedQuestions,
			}

			totalQuestionsReturned += len(formattedQuestions)
			subjects = append(subjects, subjectData)
		}
	}

	return map[string]interface{}{
		"totalSubjectsRequested": totalSubjectsRequested,
		"totalQuestionsReturned": totalQuestionsReturned,
		"subjects":               subjects,
	}, nil
}

func handleAssessmentSaving(assessmentDataInterface interface{}, teacherId string) (map[string]interface{}, error) {
	// Parse the assessment data
	assessmentBytes, err := json.Marshal(assessmentDataInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assessment data: %v", err)
	}

	var assessmentData AssessmentData
	if err := json.Unmarshal(assessmentBytes, &assessmentData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assessment data: %v", err)
	}

	// Check for mandatory fields
	var missingFields []string

	if assessmentData.StudentID == nil {
		missingFields = append(missingFields, "student_id")
	}
	if assessmentData.StudentName == "" {
		missingFields = append(missingFields, "student_name")
	}
	if assessmentData.Subject == "" {
		missingFields = append(missingFields, "subject")
	}
	if assessmentData.Score == nil {
		missingFields = append(missingFields, "score")
	}

	if len(missingFields) > 0 {
		return nil, fmt.Errorf("missing mandatory fields: %s", strings.Join(missingFields, ", "))
	}

	// Validate and convert subject to enum
	subject, validSubject := model.GetSubjectFromString(strings.ToLower(strings.TrimSpace(assessmentData.Subject)))
	if !validSubject {
		return nil, fmt.Errorf("invalid subject: %s. Available subjects are: math, science, english, history, geography", assessmentData.Subject)
	}

	// Convert studentID to int
	var studentID int
	switch v := assessmentData.StudentID.(type) {
	case float64:
		studentID = int(v)
	case int:
		studentID = v
	case string:
		if id, err := strconv.Atoi(v); err == nil {
			studentID = id
		} else {
			return nil, fmt.Errorf("invalid student_id format: %v", v)
		}
	default:
		return nil, fmt.Errorf("invalid student_id type: %T", v)
	}

	// Convert score to int
	var score int
	switch v := assessmentData.Score.(type) {
	case float64:
		score = int(v)
	case int:
		score = v
	case string:
		if s, err := strconv.Atoi(v); err == nil {
			score = s
		} else {
			return nil, fmt.Errorf("invalid score format: %v", v)
		}
	default:
		return nil, fmt.Errorf("invalid score type: %T", v)
	}

	// Create SubjectReport object
	now := time.Now()
	subjectReport := model.SubjectReport{
		UserID:      teacherId,
		StudentID:   studentID,
		StudentName: assessmentData.StudentName,
		Subject:     subject,
		Score:       score,
		Timestamp:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Helper function to convert interface{} to string pointer
	toStringPtr := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		if str, ok := v.(string); ok && strings.TrimSpace(str) != "" {
			trimmed := strings.TrimSpace(str)
			return &trimmed
		}
		return nil
	}

	// Helper function to convert interface{} to int pointer
	toIntPtr := func(v interface{}) *int {
		if v == nil {
			return nil
		}
		switch val := v.(type) {
		case float64:
			intVal := int(val)
			return &intVal
		case int:
			return &val
		case string:
			if intVal, err := strconv.Atoi(val); err == nil {
				return &intVal
			}
		}
		return nil
	}

	// Helper function to convert interface{} to float64 pointer
	toFloat64Ptr := func(v interface{}) *float64 {
		if v == nil {
			return nil
		}
		switch val := v.(type) {
		case float64:
			return &val
		case int:
			floatVal := float64(val)
			return &floatVal
		case string:
			if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
				return &floatVal
			}
		}
		return nil
	}

	// Set optional string fields
	subjectReport.GradeLetter = toStringPtr(assessmentData.GradeLetter)
	subjectReport.ClassName = toStringPtr(assessmentData.ClassName)
	subjectReport.InstructorName = toStringPtr(assessmentData.InstructorName)
	subjectReport.Term = toStringPtr(assessmentData.Term)
	subjectReport.Remarks = toStringPtr(assessmentData.Remarks)
	subjectReport.LearningObjectivesMastered = toStringPtr(assessmentData.LearningObjectivesMastered)
	subjectReport.AreasForImprovement = toStringPtr(assessmentData.AreasForImprovement)
	subjectReport.RecommendedResources = toStringPtr(assessmentData.RecommendedResources)
	subjectReport.TargetGoals = toStringPtr(assessmentData.TargetGoals)

	// Set optional integer fields
	subjectReport.MidtermScore = toIntPtr(assessmentData.MidtermScore)
	subjectReport.FinalExamScore = toIntPtr(assessmentData.FinalExamScore)
	subjectReport.QuizScore = toIntPtr(assessmentData.QuizScore)
	subjectReport.AssignmentScore = toIntPtr(assessmentData.AssignmentScore)
	subjectReport.PracticalScore = toIntPtr(assessmentData.PracticalScore)
	subjectReport.OralPresentationScore = toIntPtr(assessmentData.OralPresentationScore)

	// Set optional float fields
	subjectReport.ConceptualUnderstanding = toFloat64Ptr(assessmentData.ConceptualUnderstanding)
	subjectReport.ProblemSolving = toFloat64Ptr(assessmentData.ProblemSolving)
	subjectReport.KnowledgeApplication = toFloat64Ptr(assessmentData.KnowledgeApplication)
	subjectReport.AnalyticalThinking = toFloat64Ptr(assessmentData.AnalyticalThinking)
	subjectReport.Creativity = toFloat64Ptr(assessmentData.Creativity)
	subjectReport.PracticalSkills = toFloat64Ptr(assessmentData.PracticalSkills)
	subjectReport.Participation = toFloat64Ptr(assessmentData.Participation)
	subjectReport.Discipline = toFloat64Ptr(assessmentData.Discipline)
	subjectReport.Punctuality = toFloat64Ptr(assessmentData.Punctuality)
	subjectReport.Teamwork = toFloat64Ptr(assessmentData.Teamwork)
	subjectReport.EffortLevel = toFloat64Ptr(assessmentData.EffortLevel)
	subjectReport.Improvement = toFloat64Ptr(assessmentData.Improvement)

	// Save to database
	savedReport, err := repository.SaveSubjectReport(subjectReport)
	if err != nil {
		return nil, fmt.Errorf("failed to save subject report: %v", err)
	}

	// Build the response data structure
	assessmentDataResponse := map[string]interface{}{
		"studentId":   savedReport.StudentID,
		"studentName": savedReport.StudentName,
		"subject":     string(savedReport.Subject),
		"score":       savedReport.Score,
	}

	// Add optional fields only if they exist
	if savedReport.GradeLetter != nil {
		assessmentDataResponse["gradeLetter"] = *savedReport.GradeLetter
	}
	if savedReport.ClassName != nil {
		assessmentDataResponse["className"] = *savedReport.ClassName
	}
	if savedReport.InstructorName != nil {
		assessmentDataResponse["instructorName"] = *savedReport.InstructorName
	}
	if savedReport.Term != nil {
		assessmentDataResponse["term"] = *savedReport.Term
	}
	if savedReport.Remarks != nil {
		assessmentDataResponse["remarks"] = *savedReport.Remarks
	}
	if savedReport.MidtermScore != nil {
		assessmentDataResponse["midtermScore"] = *savedReport.MidtermScore
	}
	if savedReport.FinalExamScore != nil {
		assessmentDataResponse["finalExamScore"] = *savedReport.FinalExamScore
	}
	if savedReport.QuizScore != nil {
		assessmentDataResponse["quizScore"] = *savedReport.QuizScore
	}
	if savedReport.AssignmentScore != nil {
		assessmentDataResponse["assignmentScore"] = *savedReport.AssignmentScore
	}
	if savedReport.PracticalScore != nil {
		assessmentDataResponse["practicalScore"] = *savedReport.PracticalScore
	}
	if savedReport.OralPresentationScore != nil {
		assessmentDataResponse["oralPresentationScore"] = *savedReport.OralPresentationScore
	}
	if savedReport.ConceptualUnderstanding != nil {
		assessmentDataResponse["conceptualUnderstanding"] = *savedReport.ConceptualUnderstanding
	}
	if savedReport.ProblemSolving != nil {
		assessmentDataResponse["problemSolving"] = *savedReport.ProblemSolving
	}
	if savedReport.KnowledgeApplication != nil {
		assessmentDataResponse["knowledgeApplication"] = *savedReport.KnowledgeApplication
	}
	if savedReport.AnalyticalThinking != nil {
		assessmentDataResponse["analyticalThinking"] = *savedReport.AnalyticalThinking
	}
	if savedReport.Creativity != nil {
		assessmentDataResponse["creativity"] = *savedReport.Creativity
	}
	if savedReport.PracticalSkills != nil {
		assessmentDataResponse["practicalSkills"] = *savedReport.PracticalSkills
	}
	if savedReport.Participation != nil {
		assessmentDataResponse["participation"] = *savedReport.Participation
	}
	if savedReport.Discipline != nil {
		assessmentDataResponse["discipline"] = *savedReport.Discipline
	}
	if savedReport.Punctuality != nil {
		assessmentDataResponse["punctuality"] = *savedReport.Punctuality
	}
	if savedReport.Teamwork != nil {
		assessmentDataResponse["teamwork"] = *savedReport.Teamwork
	}
	if savedReport.EffortLevel != nil {
		assessmentDataResponse["effortLevel"] = *savedReport.EffortLevel
	}
	if savedReport.Improvement != nil {
		assessmentDataResponse["improvement"] = *savedReport.Improvement
	}
	if savedReport.LearningObjectivesMastered != nil {
		assessmentDataResponse["learningObjectivesMastered"] = *savedReport.LearningObjectivesMastered
	}
	if savedReport.AreasForImprovement != nil {
		assessmentDataResponse["areasForImprovement"] = *savedReport.AreasForImprovement
	}
	if savedReport.RecommendedResources != nil {
		assessmentDataResponse["recommendedResources"] = *savedReport.RecommendedResources
	}
	if savedReport.TargetGoals != nil {
		assessmentDataResponse["targetGoals"] = *savedReport.TargetGoals
	}

	return map[string]interface{}{
		"assessmentData": assessmentDataResponse,
	}, nil
}

func handleReportCardGeneration(reportCardDataInterface interface{}, teacherId string) (map[string]interface{}, error) {
	// The report card agent returns structured data that we should save to database
	// and pass through to the frontend

	reportCardMap, ok := reportCardDataInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid report card data format")
	}

	// Extract the report_card from the agent's response
	reportCardData, exists := reportCardMap["report_card"]
	if !exists {
		return nil, fmt.Errorf("no report_card field found in agent response")
	}

	// Convert the report card data to our model structure
	reportCardBytes, err := json.Marshal(reportCardData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report card data: %v", err)
	}

	var agentReportCardData model.AgentReportCardData
	if err := json.Unmarshal(reportCardBytes, &agentReportCardData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report card data: %v", err)
	}

	// Create the AgentReportCard with metadata
	agentReportCard := model.AgentReportCard{
		UserID:     teacherId,
		ReportCard: agentReportCardData,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to database
	savedReportCard, err := repository.CreateAgentReportCard(agentReportCard)
	if err != nil {
		log.Printf("Warning: Failed to save agent report card to database: %v", err)
		// Continue and return the data even if saving fails
	}

	// Prepare response data
	responseData := map[string]interface{}{
		"reportCard": reportCardData,
	}

	// Add database metadata if save was successful
	if savedReportCard != nil {
		responseData["databaseId"] = savedReportCard.ID.Hex()
		responseData["savedAt"] = savedReportCard.CreatedAt
	}

	return responseData, nil
}

// handleAssignmentResultSaving processes assignment result data from assessor agent
func handleAssignmentResultSaving(assignmentResultData interface{}, teacherId string) (map[string]interface{}, error) {
	// Convert interface{} to map[string]interface{}
	resultMap, ok := assignmentResultData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid assignment result data format")
	}

	// Convert to AssignmentResult model
	var assignmentResult model.AssignmentResult

	// Map the fields from the response
	if assignmentId, ok := resultMap["assignment_id"].(string); ok {
		assignmentResult.AssignmentID = assignmentId
	} else {
		return nil, fmt.Errorf("assignment_id is required")
	}

	if studentId, ok := resultMap["student_id"].(string); ok {
		assignmentResult.StudentID = studentId
	} else {
		return nil, fmt.Errorf("student_id is required")
	}

	if totalPoints, ok := resultMap["total_points_awarded"].(float64); ok {
		assignmentResult.TotalPointsAwarded = int(totalPoints)
	}

	if maxPoints, ok := resultMap["total_max_points"].(float64); ok {
		assignmentResult.TotalMaxPoints = int(maxPoints)
	}

	if percentage, ok := resultMap["percentage_score"].(float64); ok {
		assignmentResult.PercentageScore = percentage
	}

	// Convert the results arrays into the proper structs
	if mcqResults, ok := resultMap["mcq_results"]; ok {
		mcqResultsBytes, _ := json.Marshal(mcqResults)
		var mcqResultsList []model.MCQResult
		if err := json.Unmarshal(mcqResultsBytes, &mcqResultsList); err == nil {
			assignmentResult.MCQResults = mcqResultsList
		}
	}

	if msqResults, ok := resultMap["msq_results"]; ok {
		msqResultsBytes, _ := json.Marshal(msqResults)
		var msqResultsList []model.MSQResult
		if err := json.Unmarshal(msqResultsBytes, &msqResultsList); err == nil {
			assignmentResult.MSQResults = msqResultsList
		}
	}

	if natResults, ok := resultMap["nat_results"]; ok {
		natResultsBytes, _ := json.Marshal(natResults)
		var natResultsList []model.NATResult
		if err := json.Unmarshal(natResultsBytes, &natResultsList); err == nil {
			assignmentResult.NATResults = natResultsList
		}
	}

	if subjectiveResults, ok := resultMap["subjective_results"]; ok {
		subjectiveResultsBytes, _ := json.Marshal(subjectiveResults)
		var subjectiveResultsList []model.SubjectiveResult
		if err := json.Unmarshal(subjectiveResultsBytes, &subjectiveResultsList); err == nil {
			assignmentResult.SubjectiveResults = subjectiveResultsList
		}
	}

	// Set metadata
	assignmentResult.ID = primitive.NewObjectID()
	assignmentResult.CreatedAt = time.Now()
	assignmentResult.UpdatedAt = time.Now()

	// Save to database using repository
	savedResult, err := repository.CreateAssignmentResult(assignmentResult)
	if err != nil {
		return nil, fmt.Errorf("failed to save assignment result: %v", err)
	}
	assignmentResult = *savedResult

	// Return the saved assignment result data
	return map[string]interface{}{
		"assignmentResultData": map[string]interface{}{
			"id":                 assignmentResult.ID.Hex(),
			"assignmentId":       assignmentResult.AssignmentID,
			"studentId":          assignmentResult.StudentID,
			"totalPointsAwarded": assignmentResult.TotalPointsAwarded,
			"totalMaxPoints":     assignmentResult.TotalMaxPoints,
			"percentageScore":    assignmentResult.PercentageScore,
			"mcqResults":         assignmentResult.MCQResults,
			"msqResults":         assignmentResult.MSQResults,
			"natResults":         assignmentResult.NATResults,
			"subjectiveResults":  assignmentResult.SubjectiveResults,
			"createdAt":          assignmentResult.CreatedAt,
			"updatedAt":          assignmentResult.UpdatedAt,
		},
	}, nil
}

// TODO: Uncomment this function after regenerating proto files with:
// protoc --go_out=. --go-grpc_out=. internal/proto/ai_service.proto
/*
func RAGAgent(teacherId, role, message, file, createdAt, updatedAt string) (map[string]interface{}, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.RAGAgentRequest{
		TeacherId: teacherId,
		Role:      role,
		Message:   message,
		File:      file,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	res, err := client.RAGAgent(ctx, req)
	if err != nil {
		return createRAGErrorResponse(teacherId, err.Error(), res), nil
	}

	// Return the standardized response format for RAG Agent
	return map[string]interface{}{
		"message":       res.GetMessage(),
		"teacherId":     res.GetTeacherId(),
		"agentName":     res.GetAgentName(),
		"agentResponse": res.GetAgentResponse(),
		"sessionId":     res.GetSessionId(),
		"createdAt":     res.GetCreatedAt(),
		"updatedAt":     res.GetUpdatedAt(),
		"responseTime":  res.GetResponseTime(),
		"role":          res.GetRole(),
		"feedback":      res.GetFeedback(),
	}, nil
}

func createRAGErrorResponse(teacherId, errorMessage string, res *pb.RAGAgentResponse) map[string]interface{} {
	var sessionId, createdAt, updatedAt, responseTime, feedback string
	if res != nil {
		sessionId = res.GetSessionId()
		createdAt = res.GetCreatedAt()
		updatedAt = res.GetUpdatedAt()
		responseTime = res.GetResponseTime()
		feedback = res.GetFeedback()
	}

	return map[string]interface{}{
		"message":       errorMessage,
		"teacherId":     teacherId,
		"agentName":     "rag_agent",
		"agentResponse": errorMessage,
		"sessionId":     sessionId,
		"createdAt":     createdAt,
		"updatedAt":     updatedAt,
		"responseTime":  responseTime,
		"role":          "agent",
		"feedback":      feedback,
	}
}
*/

func RAGAgentClient(teacherId, message, file string) (*pb.RAGAgentResponse, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.RAGAgentRequest{
		TeacherId: teacherId,
		Message:   message,
		File:      file,
	}

	resp, err := client.RAGAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call RAG agent: %v", err)
	}

	return resp, nil
}
