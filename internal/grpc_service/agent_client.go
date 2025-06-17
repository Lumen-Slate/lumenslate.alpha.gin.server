package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"lumenslate/internal/model"
	pb "lumenslate/internal/proto/ai_service"
	"lumenslate/internal/repository"
	"math/rand"
	"strconv"
	"strings"
	"time"
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

func Agent(file, fileType, userId, role, message, createdAt, updatedAt string) (map[string]interface{}, error) {
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
		UserId:    userId,
		Role:      role,
		Message:   message,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	res, err := client.Agent(ctx, req)
	if err != nil {
		return createErrorResponse(userId, err.Error(), res), nil
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
			if questionsData, err := handleQuestionGeneration(agentResponse.QuestionsRequested, userId); err == nil {
				responseData = questionsData
				agentName = "assignment_generator_general"
				responseMessage = "Assignment generated successfully"
			} else {
				log.Printf("Error handling question generation: %v", err)
				return createErrorResponse(userId, err.Error(), res), nil
			}
		} else if agentResponse.AssessmentData != nil {
			// Handle assessment data saving
			if assessmentData, err := handleAssessmentSaving(agentResponse.AssessmentData, userId); err == nil {
				responseData = assessmentData
				agentName = "assessor_agent"
				responseMessage = "Subject assessment report processed successfully"
			} else {
				log.Printf("Error handling assessment saving: %v", err)
				return createErrorResponse(userId, err.Error(), res), nil
			}
		} else if agentResponse.ReportCardData != nil {
			// Handle report card generation and saving
			if reportCardData, err := handleReportCardGeneration(agentResponse.ReportCardData, userId); err == nil {
				responseData = reportCardData
				agentName = "report_card_generator"
				responseMessage = "Report card generated and saved successfully"
			} else {
				log.Printf("Error handling report card generation: %v", err)
				return createErrorResponse(userId, err.Error(), res), nil
			}
		} else {
			// Regular agent response without database operations
			responseData = map[string]interface{}{
				"agent_response": rawAgentResponse,
			}
		}
	} else {
		// Couldn't parse as JSON, treat as regular response
		responseData = map[string]interface{}{
			"agent_response": rawAgentResponse,
		}
	}

	// Return the standardized response format
	return map[string]interface{}{
		"message":       responseMessage,
		"user_id":       res.GetUserId(),
		"agent_name":    agentName,
		"data":          responseData,
		"session_id":    res.GetSessionId(),
		"createdAt":     res.GetCreatedAt(),
		"updatedAt":     res.GetUpdatedAt(),
		"response_time": res.GetResponseTime(),
		"role":          res.GetRole(),
		"feedback":      res.GetFeedback(),
	}, nil
}

func createErrorResponse(userId, errorMessage string, res *pb.AgentResponse) map[string]interface{} {
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
		"user_id":       userId,
		"agent_name":    "root_agent",
		"data":          map[string]interface{}{},
		"session_id":    sessionId,
		"createdAt":     createdAt,
		"updatedAt":     updatedAt,
		"response_time": responseTime,
		"role":          "agent",
		"feedback":      feedback,
	}
}

func handleQuestionGeneration(questionsRequested []QuestionRequest, userId string) (map[string]interface{}, error) {
	// Debug: Check what's actually in the database
	log.Printf("=== DEBUG: Starting question generation for user: %s ===", userId)
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
				"subject":                   displaySubject,
				"questions_requested_count": totalRequested,
				"questions_available_count": 0,
				"questions_returned_count":  0,
				"questions":                 []map[string]interface{}{},
				"message":                   fmt.Sprintf("Subject '%s' is not supported. Available subjects: Math, Science, English, History, Geography", displaySubject),
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
				"subject":                   displaySubject,
				"questions_requested_count": totalRequested,
				"questions_available_count": int(totalAvailable),
				"questions_returned_count":  0,
				"questions":                 []map[string]interface{}{},
				"message":                   fmt.Sprintf("Could not fulfill all requests for %s. Some difficulty levels may not have enough questions available.", displaySubject),
			}
			subjects = append(subjects, subjectData)
		} else if len(allQuestionsForSubject) == 0 {
			subjectData := map[string]interface{}{
				"subject":                   displaySubject,
				"questions_requested_count": totalRequested,
				"questions_available_count": int(totalAvailable),
				"questions_returned_count":  0,
				"questions":                 []map[string]interface{}{},
				"message":                   fmt.Sprintf("No questions available for %s with the requested difficulty levels", displaySubject),
			}
			subjects = append(subjects, subjectData)
		} else {
			// Successfully got all requested questions
			// Format questions for response
			formattedQuestions := make([]map[string]interface{}, len(allQuestionsForSubject))
			for i, q := range allQuestionsForSubject {
				questionData := map[string]interface{}{
					"question_id": q.ID,
					"question":    q.Question,
					"type":        "MCQ", // Default to MCQ, you can enhance this based on your question structure
					"answer":      q.Answer,
					"difficulty":  string(q.Difficulty),
				}

				// Add options only for MCQ and MSQ types
				if len(q.Options) > 0 {
					questionData["options"] = q.Options
				}

				formattedQuestions[i] = questionData
			}

			subjectData := map[string]interface{}{
				"subject":                   displaySubject,
				"questions_requested_count": totalRequested,
				"questions_available_count": int(totalAvailable),
				"questions_returned_count":  len(formattedQuestions),
				"questions":                 formattedQuestions,
			}

			totalQuestionsReturned += len(formattedQuestions)
			subjects = append(subjects, subjectData)
		}
	}

	return map[string]interface{}{
		"total_subjects_requested": totalSubjectsRequested,
		"total_questions_returned": totalQuestionsReturned,
		"subjects":                 subjects,
	}, nil
}

func handleAssessmentSaving(assessmentDataInterface interface{}, userId string) (map[string]interface{}, error) {
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
		UserID:      userId,
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
		"student_id":   savedReport.StudentID,
		"student_name": savedReport.StudentName,
		"subject":      string(savedReport.Subject),
		"score":        savedReport.Score,
	}

	// Add optional fields only if they exist
	if savedReport.GradeLetter != nil {
		assessmentDataResponse["grade_letter"] = *savedReport.GradeLetter
	}
	if savedReport.ClassName != nil {
		assessmentDataResponse["class_name"] = *savedReport.ClassName
	}
	if savedReport.InstructorName != nil {
		assessmentDataResponse["instructor_name"] = *savedReport.InstructorName
	}
	if savedReport.Term != nil {
		assessmentDataResponse["term"] = *savedReport.Term
	}
	if savedReport.Remarks != nil {
		assessmentDataResponse["remarks"] = *savedReport.Remarks
	}
	if savedReport.MidtermScore != nil {
		assessmentDataResponse["midterm_score"] = *savedReport.MidtermScore
	}
	if savedReport.FinalExamScore != nil {
		assessmentDataResponse["final_exam_score"] = *savedReport.FinalExamScore
	}
	if savedReport.QuizScore != nil {
		assessmentDataResponse["quiz_score"] = *savedReport.QuizScore
	}
	if savedReport.AssignmentScore != nil {
		assessmentDataResponse["assignment_score"] = *savedReport.AssignmentScore
	}
	if savedReport.PracticalScore != nil {
		assessmentDataResponse["practical_score"] = *savedReport.PracticalScore
	}
	if savedReport.OralPresentationScore != nil {
		assessmentDataResponse["oral_presentation_score"] = *savedReport.OralPresentationScore
	}
	if savedReport.ConceptualUnderstanding != nil {
		assessmentDataResponse["conceptual_understanding"] = *savedReport.ConceptualUnderstanding
	}
	if savedReport.ProblemSolving != nil {
		assessmentDataResponse["problem_solving"] = *savedReport.ProblemSolving
	}
	if savedReport.KnowledgeApplication != nil {
		assessmentDataResponse["knowledge_application"] = *savedReport.KnowledgeApplication
	}
	if savedReport.AnalyticalThinking != nil {
		assessmentDataResponse["analytical_thinking"] = *savedReport.AnalyticalThinking
	}
	if savedReport.Creativity != nil {
		assessmentDataResponse["creativity"] = *savedReport.Creativity
	}
	if savedReport.PracticalSkills != nil {
		assessmentDataResponse["practical_skills"] = *savedReport.PracticalSkills
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
		assessmentDataResponse["effort_level"] = *savedReport.EffortLevel
	}
	if savedReport.Improvement != nil {
		assessmentDataResponse["improvement"] = *savedReport.Improvement
	}
	if savedReport.LearningObjectivesMastered != nil {
		assessmentDataResponse["learning_objectives_mastered"] = *savedReport.LearningObjectivesMastered
	}
	if savedReport.AreasForImprovement != nil {
		assessmentDataResponse["areas_for_improvement"] = *savedReport.AreasForImprovement
	}
	if savedReport.RecommendedResources != nil {
		assessmentDataResponse["recommended_resources"] = *savedReport.RecommendedResources
	}
	if savedReport.TargetGoals != nil {
		assessmentDataResponse["target_goals"] = *savedReport.TargetGoals
	}

	return map[string]interface{}{
		"assessment_data": assessmentDataResponse,
	}, nil
}

func handleReportCardGeneration(reportCardDataInterface interface{}, userId string) (map[string]interface{}, error) {
	// Parse the report card data
	reportCardBytes, err := json.Marshal(reportCardDataInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report card data: %v", err)
	}

	// First, try to parse as a map[string]interface{} to handle flexible structure
	var reportCardMap map[string]interface{}
	if err := json.Unmarshal(reportCardBytes, &reportCardMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report card data: %v", err)
	}

	// Check for mandatory fields
	var missingFields []string

	studentIDInterface, hasStudentID := reportCardMap["student_id"]
	if !hasStudentID || studentIDInterface == nil {
		missingFields = append(missingFields, "student_id")
	}

	studentNameInterface, hasStudentName := reportCardMap["student_name"]
	if !hasStudentName || studentNameInterface == nil {
		missingFields = append(missingFields, "student_name")
	}

	academicTermInterface, hasAcademicTerm := reportCardMap["academic_term"]
	if !hasAcademicTerm || academicTermInterface == nil {
		missingFields = append(missingFields, "academic_term")
	}

	if len(missingFields) > 0 {
		return nil, fmt.Errorf("missing mandatory fields: %s", strings.Join(missingFields, ", "))
	}

	// Convert studentID to int
	var studentID int
	switch v := studentIDInterface.(type) {
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

	// Extract student name and academic term
	studentName, ok := studentNameInterface.(string)
	if !ok {
		return nil, fmt.Errorf("student_name must be a string")
	}

	academicTerm, ok := academicTermInterface.(string)
	if !ok {
		return nil, fmt.Errorf("academic_term must be a string")
	}

	// Create ReportCard object
	now := time.Now()
	reportCard := model.NewReportCard()
	reportCard.UserID = userId
	reportCard.StudentID = studentID
	reportCard.StudentName = studentName
	reportCard.AcademicTerm = academicTerm
	reportCard.GeneratedAt = now
	reportCard.CreatedAt = now
	reportCard.UpdatedAt = now

	// Helper function to safely extract string pointers
	getStringPtr := func(key string) *string {
		if val, exists := reportCardMap[key]; exists && val != nil {
			if str, ok := val.(string); ok && strings.TrimSpace(str) != "" {
				trimmed := strings.TrimSpace(str)
				return &trimmed
			}
		}
		return nil
	}

	// Helper function to safely extract int pointers
	getIntPtr := func(key string) *int {
		if val, exists := reportCardMap[key]; exists && val != nil {
			switch v := val.(type) {
			case float64:
				intVal := int(v)
				return &intVal
			case int:
				return &v
			case string:
				if intVal, err := strconv.Atoi(v); err == nil {
					return &intVal
				}
			}
		}
		return nil
	}

	// Helper function to safely extract float64 pointers
	getFloat64Ptr := func(key string) *float64 {
		if val, exists := reportCardMap[key]; exists && val != nil {
			switch v := val.(type) {
			case float64:
				return &v
			case int:
				floatVal := float64(v)
				return &floatVal
			case string:
				if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
					return &floatVal
				}
			}
		}
		return nil
	}

	// Helper function to safely extract time pointers
	getTimePtr := func(key string) *time.Time {
		if val, exists := reportCardMap[key]; exists && val != nil {
			if str, ok := val.(string); ok {
				if t, err := time.Parse(time.RFC3339, str); err == nil {
					return &t
				}
			}
		}
		return nil
	}

	// Set all optional fields using helper functions
	reportCard.OverallGPA = getFloat64Ptr("overall_gpa")
	reportCard.OverallGrade = getStringPtr("overall_grade")
	reportCard.OverallPercentage = getFloat64Ptr("overall_percentage")
	reportCard.ClassRank = getIntPtr("class_rank")
	reportCard.TotalStudentsInClass = getIntPtr("total_students_in_class")
	reportCard.SubjectsCount = getIntPtr("subjects_count")
	reportCard.HighestSubjectScore = getIntPtr("highest_subject_score")
	reportCard.LowestSubjectScore = getIntPtr("lowest_subject_score")
	reportCard.AverageSubjectScore = getFloat64Ptr("average_subject_score")
	reportCard.BestPerformingSubject = getStringPtr("best_performing_subject")
	reportCard.WeakestSubject = getStringPtr("weakest_subject")
	reportCard.AcademicStrengths = getStringPtr("academic_strengths")
	reportCard.AreasNeedingImprovement = getStringPtr("areas_needing_improvement")
	reportCard.RecommendedActions = getStringPtr("recommended_actions")
	reportCard.StudyRecommendations = getStringPtr("study_recommendations")
	reportCard.OverallConceptualUnderstanding = getFloat64Ptr("overall_conceptual_understanding")
	reportCard.OverallProblemSolving = getFloat64Ptr("overall_problem_solving")
	reportCard.OverallKnowledgeApplication = getFloat64Ptr("overall_knowledge_application")
	reportCard.OverallAnalyticalThinking = getFloat64Ptr("overall_analytical_thinking")
	reportCard.OverallCreativity = getFloat64Ptr("overall_creativity")
	reportCard.OverallPracticalSkills = getFloat64Ptr("overall_practical_skills")
	reportCard.OverallParticipation = getFloat64Ptr("overall_participation")
	reportCard.OverallDiscipline = getFloat64Ptr("overall_discipline")
	reportCard.OverallPunctuality = getFloat64Ptr("overall_punctuality")
	reportCard.OverallTeamwork = getFloat64Ptr("overall_teamwork")
	reportCard.OverallEffortLevel = getFloat64Ptr("overall_effort_level")
	reportCard.OverallImprovement = getFloat64Ptr("overall_improvement")
	reportCard.AverageMidtermScore = getFloat64Ptr("average_midterm_score")
	reportCard.AverageFinalExamScore = getFloat64Ptr("average_final_exam_score")
	reportCard.AverageQuizScore = getFloat64Ptr("average_quiz_score")
	reportCard.AverageAssignmentScore = getFloat64Ptr("average_assignment_score")
	reportCard.AveragePracticalScore = getFloat64Ptr("average_practical_score")
	reportCard.AverageOralPresentationScore = getFloat64Ptr("average_oral_presentation_score")
	reportCard.ImprovementTrend = getStringPtr("improvement_trend")
	reportCard.ConsistencyRating = getFloat64Ptr("consistency_rating")
	reportCard.PerformanceStability = getStringPtr("performance_stability")
	reportCard.AttendanceRate = getFloat64Ptr("attendance_rate")
	reportCard.EngagementLevel = getStringPtr("engagement_level")
	reportCard.ClassParticipation = getStringPtr("class_participation")
	reportCard.AcademicGoals = getStringPtr("academic_goals")
	reportCard.ShortTermObjectives = getStringPtr("short_term_objectives")
	reportCard.LongTermObjectives = getStringPtr("long_term_objectives")
	reportCard.ParentTeacherRecommendations = getStringPtr("parent_teacher_recommendations")
	reportCard.TeacherComments = getStringPtr("teacher_comments")
	reportCard.PrincipalComments = getStringPtr("principal_comments")
	reportCard.OverallRemarks = getStringPtr("overall_remarks")
	reportCard.RecommendedResources = getStringPtr("recommended_resources")
	reportCard.SuggestedActivities = getStringPtr("suggested_activities")
	reportCard.NextReviewDate = getTimePtr("next_review_date")

	// Handle subject reports array
	if subjectReportsInterface, exists := reportCardMap["subject_reports"]; exists && subjectReportsInterface != nil {
		if subjectReportsArray, ok := subjectReportsInterface.([]interface{}); ok {
			subjectReports := make([]model.SubjectReportSummary, 0, len(subjectReportsArray))
			for _, subjectInterface := range subjectReportsArray {
				if subjectMap, ok := subjectInterface.(map[string]interface{}); ok {
					summary := model.SubjectReportSummary{}
					if subject, ok := subjectMap["subject"].(string); ok {
						summary.Subject = subject
					}
					if score, ok := subjectMap["score"].(float64); ok {
						summary.Score = int(score)
					} else if score, ok := subjectMap["score"].(int); ok {
						summary.Score = score
					}
					if grade, ok := subjectMap["grade"].(string); ok && strings.TrimSpace(grade) != "" {
						trimmed := strings.TrimSpace(grade)
						summary.Grade = &trimmed
					}
					if conceptual, ok := subjectMap["conceptual_understanding"].(float64); ok {
						summary.ConceptualUnderstanding = &conceptual
					}
					if problemSolving, ok := subjectMap["problem_solving"].(float64); ok {
						summary.ProblemSolving = &problemSolving
					}
					if analytical, ok := subjectMap["analytical_thinking"].(float64); ok {
						summary.AnalyticalThinking = &analytical
					}
					if improvements, ok := subjectMap["areas_for_improvement"].(string); ok && strings.TrimSpace(improvements) != "" {
						trimmed := strings.TrimSpace(improvements)
						summary.AreasForImprovement = &trimmed
					}
					if strengths, ok := subjectMap["key_strengths"].(string); ok && strings.TrimSpace(strengths) != "" {
						trimmed := strings.TrimSpace(strengths)
						summary.KeyStrengths = &trimmed
					}
					if recommendations, ok := subjectMap["subject_specific_recommendations"].(string); ok && strings.TrimSpace(recommendations) != "" {
						trimmed := strings.TrimSpace(recommendations)
						summary.SubjectSpecificRecommendations = &trimmed
					}
					subjectReports = append(subjectReports, summary)
				}
			}
			reportCard.SubjectReports = subjectReports
		}
	}

	// Save to database
	savedReportCard, err := repository.SaveReportCard(*reportCard)
	if err != nil {
		return nil, fmt.Errorf("failed to save report card: %v", err)
	}

	// Build the response data structure - only include fields that have actual data
	reportCardResponse := map[string]interface{}{
		"student_id":    savedReportCard.StudentID,
		"student_name":  savedReportCard.StudentName,
		"academic_term": savedReportCard.AcademicTerm,
		"generated_at":  savedReportCard.GeneratedAt,
	}

	// Add optional fields only if they exist
	if savedReportCard.OverallGPA != nil {
		reportCardResponse["overall_gpa"] = *savedReportCard.OverallGPA
	}
	if savedReportCard.OverallGrade != nil {
		reportCardResponse["overall_grade"] = *savedReportCard.OverallGrade
	}
	if savedReportCard.OverallPercentage != nil {
		reportCardResponse["overall_percentage"] = *savedReportCard.OverallPercentage
	}
	if savedReportCard.ClassRank != nil {
		reportCardResponse["class_rank"] = *savedReportCard.ClassRank
	}
	if savedReportCard.TotalStudentsInClass != nil {
		reportCardResponse["total_students_in_class"] = *savedReportCard.TotalStudentsInClass
	}
	if savedReportCard.SubjectsCount != nil {
		reportCardResponse["subjects_count"] = *savedReportCard.SubjectsCount
	}
	if savedReportCard.HighestSubjectScore != nil {
		reportCardResponse["highest_subject_score"] = *savedReportCard.HighestSubjectScore
	}
	if savedReportCard.LowestSubjectScore != nil {
		reportCardResponse["lowest_subject_score"] = *savedReportCard.LowestSubjectScore
	}
	if savedReportCard.AverageSubjectScore != nil {
		reportCardResponse["average_subject_score"] = *savedReportCard.AverageSubjectScore
	}
	if savedReportCard.BestPerformingSubject != nil {
		reportCardResponse["best_performing_subject"] = *savedReportCard.BestPerformingSubject
	}
	if savedReportCard.WeakestSubject != nil {
		reportCardResponse["weakest_subject"] = *savedReportCard.WeakestSubject
	}
	if savedReportCard.AcademicStrengths != nil {
		reportCardResponse["academic_strengths"] = *savedReportCard.AcademicStrengths
	}
	if savedReportCard.AreasNeedingImprovement != nil {
		reportCardResponse["areas_needing_improvement"] = *savedReportCard.AreasNeedingImprovement
	}
	if savedReportCard.RecommendedActions != nil {
		reportCardResponse["recommended_actions"] = *savedReportCard.RecommendedActions
	}
	if savedReportCard.StudyRecommendations != nil {
		reportCardResponse["study_recommendations"] = *savedReportCard.StudyRecommendations
	}
	if savedReportCard.OverallConceptualUnderstanding != nil {
		reportCardResponse["overall_conceptual_understanding"] = *savedReportCard.OverallConceptualUnderstanding
	}
	if savedReportCard.OverallProblemSolving != nil {
		reportCardResponse["overall_problem_solving"] = *savedReportCard.OverallProblemSolving
	}
	if savedReportCard.OverallKnowledgeApplication != nil {
		reportCardResponse["overall_knowledge_application"] = *savedReportCard.OverallKnowledgeApplication
	}
	if savedReportCard.OverallAnalyticalThinking != nil {
		reportCardResponse["overall_analytical_thinking"] = *savedReportCard.OverallAnalyticalThinking
	}
	if savedReportCard.OverallCreativity != nil {
		reportCardResponse["overall_creativity"] = *savedReportCard.OverallCreativity
	}
	if savedReportCard.OverallPracticalSkills != nil {
		reportCardResponse["overall_practical_skills"] = *savedReportCard.OverallPracticalSkills
	}
	if savedReportCard.OverallParticipation != nil {
		reportCardResponse["overall_participation"] = *savedReportCard.OverallParticipation
	}
	if savedReportCard.OverallDiscipline != nil {
		reportCardResponse["overall_discipline"] = *savedReportCard.OverallDiscipline
	}
	if savedReportCard.OverallPunctuality != nil {
		reportCardResponse["overall_punctuality"] = *savedReportCard.OverallPunctuality
	}
	if savedReportCard.OverallTeamwork != nil {
		reportCardResponse["overall_teamwork"] = *savedReportCard.OverallTeamwork
	}
	if savedReportCard.OverallEffortLevel != nil {
		reportCardResponse["overall_effort_level"] = *savedReportCard.OverallEffortLevel
	}
	if savedReportCard.OverallImprovement != nil {
		reportCardResponse["overall_improvement"] = *savedReportCard.OverallImprovement
	}
	if savedReportCard.ImprovementTrend != nil {
		reportCardResponse["improvement_trend"] = *savedReportCard.ImprovementTrend
	}
	if savedReportCard.ConsistencyRating != nil {
		reportCardResponse["consistency_rating"] = *savedReportCard.ConsistencyRating
	}
	if savedReportCard.PerformanceStability != nil {
		reportCardResponse["performance_stability"] = *savedReportCard.PerformanceStability
	}
	if savedReportCard.EngagementLevel != nil {
		reportCardResponse["engagement_level"] = *savedReportCard.EngagementLevel
	}
	if savedReportCard.AcademicGoals != nil {
		reportCardResponse["academic_goals"] = *savedReportCard.AcademicGoals
	}
	if savedReportCard.TeacherComments != nil {
		reportCardResponse["teacher_comments"] = *savedReportCard.TeacherComments
	}
	if savedReportCard.OverallRemarks != nil {
		reportCardResponse["overall_remarks"] = *savedReportCard.OverallRemarks
	}
	if len(savedReportCard.SubjectReports) > 0 {
		reportCardResponse["subject_reports"] = savedReportCard.SubjectReports
	}

	return map[string]interface{}{
		"report_card_data": reportCardResponse,
	}, nil
}
