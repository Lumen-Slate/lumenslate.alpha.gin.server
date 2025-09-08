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
	quest "lumenslate/internal/repository/questions"

	"github.com/google/uuid"
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
	AssessmentResult   interface{}       `json:"assessment_result"`
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

// toCamelCase converts snake_case string to camelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// convertKeysToCamelCase recursively converts all keys in a map/slice from snake_case to camelCase
func convertKeysToCamelCase(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			camelKey := toCamelCase(key)
			result[camelKey] = convertKeysToCamelCase(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertKeysToCamelCase(item)
		}
		return result
	default:
		return data
	}
}

func LumenAgent(file, fileType, teacherId, role, message, createdAt, updatedAt string) (map[string]interface{}, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		log.Printf("ERROR: Failed to establish gRPC connection: %v", err)
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	res, err := client.LumenAgent(ctx, req)
	if err != nil {
		log.Printf("ERROR: gRPC Agent call failed: %v", err)
		errorResponse := createErrorResponse(teacherId, err.Error(), res)
		return errorResponse, nil
	}

	// Get the raw agent response
	rawAgentResponse := res.GetAgentResponse()

	// Try to parse the agent response as JSON to see if it contains database operations
	var agentResponse AgentResponse
	agentName := "root_agent"
	responseMessage := "Agent response processed successfully"
	var responseData interface{}

	if err := json.Unmarshal([]byte(rawAgentResponse), &agentResponse); err == nil {
		// Handle report card field name mismatch: microservice returns "report_card" but we expect "report_card_data"
		if agentResponse.ReportCardData == nil {
			var rawMap map[string]interface{}
			if err := json.Unmarshal([]byte(rawAgentResponse), &rawMap); err == nil {
				if reportCardData, exists := rawMap["report_card"]; exists {
					agentResponse.ReportCardData = reportCardData
				}
			}
		}

		// Check for assignment result FIRST (prioritize assessor agent)
		// Handle both "assignment_result" and "assessment_result" field names
		var assignmentResultData interface{}
		if agentResponse.AssignmentResult != nil {
			assignmentResultData = agentResponse.AssignmentResult
		} else if agentResponse.AssessmentResult != nil {
			assignmentResultData = agentResponse.AssessmentResult
		}

		if assignmentResultData != nil {
			// Handle assignment result saving (from assessor agent)
			if assignmentResult, err := handleAssignmentResultSaving(assignmentResultData); err == nil {
				responseData = assignmentResult
				agentName = "assessor_agent"
				responseMessage = "Assignment assessment completed successfully"
			} else {
				log.Printf("ERROR: Failed to handle assignment result saving: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else if len(agentResponse.QuestionsRequested) > 0 {
			// Handle question generation (both AGG and AGT)
			if questionsData, err := handleQuestionGeneration(agentResponse.QuestionsRequested, rawAgentResponse); err == nil {
				responseData = questionsData
				agentName = "assignment_generator_general"
				responseMessage = "Assignment generated successfully"
			} else {
				log.Printf("ERROR: Failed to handle question generation: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else if agentResponse.AssessmentData != nil {
			// Handle legacy assessment data saving (for backward compatibility)
			if assessmentData, err := handleAssessmentSaving(agentResponse.AssessmentData, teacherId); err == nil {
				responseData = assessmentData
				agentName = "assessor_agent"
				responseMessage = "Subject assessment report processed successfully"
			} else {
				log.Printf("ERROR: Failed to handle assessment saving: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else if agentResponse.ReportCardData != nil {
			// Handle report card generation and saving
			if reportCardData, err := handleReportCardGeneration(agentResponse.ReportCardData, teacherId); err == nil {
				responseData = reportCardData
				agentName = "report_card_generator"
				responseMessage = "Report card generated and saved successfully"
			} else {
				log.Printf("ERROR: Failed to handle report card generation: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else {
			// Regular agent response without database operations
			responseData = map[string]interface{}{
				"agentResponse": rawAgentResponse,
			}
		}
	} else {
		// Couldn't parse as JSON, this is likely from general_chat_agent
		responseMessage = rawAgentResponse
		responseData = map[string]interface{}{}
		agentName = "general_chat_agent"
	}

	// Prepare final response
	finalResponse := map[string]interface{}{
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
	}

	// Convert the entire response to camelCase (including nested structs)
	jsonBytes, err := json.Marshal(finalResponse)
	if err != nil {
		log.Printf("ERROR: Failed to marshal response data to JSON: %v", err)
		return finalResponse, nil
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		log.Printf("ERROR: Failed to unmarshal response data from JSON: %v", err)
		return finalResponse, nil
	}

	// Apply camelCase conversion to the properly converted data
	camelCaseResponseData := convertKeysToCamelCase(jsonMap)

	// Type assert back to map[string]interface{} for return
	finalResponseData, ok := camelCaseResponseData.(map[string]interface{})
	if !ok {
		log.Printf("ERROR: Failed to convert response data to map[string]interface{}")
		return finalResponse, nil
	}

	return finalResponseData, nil
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

func handleQuestionGeneration(questionsRequested []QuestionRequest, rawAgentResponse string) (map[string]interface{}, error) {
	// Parse the raw agent response to extract title and body
	var agentResponseData map[string]interface{}
	var assignmentTitle, assignmentBody string

	if err := json.Unmarshal([]byte(rawAgentResponse), &agentResponseData); err == nil {
		if title, ok := agentResponseData["title"].(string); ok {
			assignmentTitle = title
		}
		if body, ok := agentResponseData["body"].(string); ok {
			assignmentBody = body
		}
	}

	// Default values if not provided by agent
	if assignmentTitle == "" {
		assignmentTitle = "Generated Assignment"
	}
	if assignmentBody == "" {
		assignmentBody = "Assignment generated from agent request"
	}

	// Test database connectivity
	if err := repository.DebugTestDatabaseConnection(); err != nil {
		return nil, fmt.Errorf("database connectivity failed: %v", err)
	}

	// Group requests by subject to handle multiple difficulty levels for the same subject
	subjectRequests := make(map[string][]QuestionRequest)

	for _, request := range questionsRequested {
		if request.Type == "assignment_generator_general" {
			subject := strings.ToLower(strings.TrimSpace(request.Subject))
			subjectRequests[subject] = append(subjectRequests[subject], request)
		}
	}

	// Collect all selected questions and their IDs for assignment creation
	var allSelectedQuestions []model.Questions
	mcqCount := 0
	msqCount := 0
	natCount := 0
	subjectiveCount := 0

	// Track question IDs by type for assignment
	var mcqIds []string
	var msqIds []string
	var natIds []string
	var subjectiveIds []string

	for subjectKey, requests := range subjectRequests {
		// Convert subject string to Subject enum
		subject, validSubject := model.GetSubjectFromString(subjectKey)

		if !validSubject {
			continue
		}

		// Process all requests for this subject
		for _, req := range requests {
			numQuestions := req.NumberOfQuestions
			difficulty := strings.ToLower(strings.TrimSpace(req.Difficulty))

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
					availableQuestions, err = repository.GetQuestionsBySubject(subject)
				}

				if difficultyEnum != "" {
					availableQuestions, err = repository.GetQuestionsBySubjectAndDifficulty(subject, difficultyEnum)
				}
			} else {
				// No difficulty specified, get all questions for this subject
				availableQuestions, err = repository.GetQuestionsBySubject(subject)
			}

			if err != nil {
				continue
			}

			if len(availableQuestions) >= numQuestions {
				// Randomly sample the requested number of questions
				rng := rand.New(rand.NewSource(time.Now().UnixNano()))
				selectedQuestions := make([]model.Questions, numQuestions)
				perm := rng.Perm(len(availableQuestions))
				for i := 0; i < numQuestions; i++ {
					selectedQuestions[i] = availableQuestions[perm[i]]
				}
				allSelectedQuestions = append(allSelectedQuestions, selectedQuestions...)

				// Count questions by type and collect IDs
				for _, q := range selectedQuestions {
					questionID := q.ID

					// Check if this is an MCQ
					if mcq, err := quest.GetMCQByID(questionID); err == nil && mcq != nil {
						mcqCount++
						mcqIds = append(mcqIds, questionID)
						continue
					}

					// Check if this is an MSQ
					if msq, err := quest.GetMSQByID(questionID); err == nil && msq != nil {
						msqCount++
						msqIds = append(msqIds, questionID)
						continue
					}

					// Check if this is a NAT
					if nat, err := quest.GetNATByID(questionID); err == nil && nat != nil {
						natCount++
						natIds = append(natIds, questionID)
						continue
					}

					// Check if this is a Subjective
					if subjective, err := quest.GetSubjectiveByID(questionID); err == nil && subjective != nil {
						subjectiveCount++
						subjectiveIds = append(subjectiveIds, questionID)
						continue
					}
				}
			}
		}
	}

	// Create assignment if we have selected questions
	if len(allSelectedQuestions) == 0 {
		return nil, fmt.Errorf("no questions could be selected for the assignment")
	}

	// Create assignment using existing assignment repository
	assignmentID := uuid.New().String()
	dueDate := time.Now().AddDate(0, 0, 7) // Default due date 7 days from now
	assignment := model.Assignment{
		ID:            assignmentID,
		Title:         assignmentTitle,
		Body:          assignmentBody,
		DueDate:       dueDate,
		Points:        len(allSelectedQuestions), // 1 point per question
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsActive:      true,
		CommentIds:    []string{},
		MCQIds:        mcqIds,
		MSQIds:        msqIds,
		NATIds:        natIds,
		SubjectiveIds: subjectiveIds,
	}

	// Save assignment to database
	if err := repository.SaveAssignment(assignment); err != nil {
		return nil, fmt.Errorf("failed to save assignment: %v", err)
	}

	// Prepare response
	responseData := map[string]interface{}{
		"assignmentId":    assignment.ID,
		"title":           assignment.Title,
		"body":            assignment.Body,
		"mcqCount":        mcqCount,
		"msqCount":        msqCount,
		"natCount":        natCount,
		"subjectiveCount": subjectiveCount,
	}

	// Convert the entire response to camelCase (including nested structs)
	camelCaseResponseData := convertKeysToCamelCase(responseData)

	// Type assert back to map[string]interface{} for return
	finalResponseData, ok := camelCaseResponseData.(map[string]interface{})
	if !ok {
		log.Printf("ERROR: Failed to convert response data to map[string]interface{}")
		return responseData, nil // Fallback to original data
	}

	return finalResponseData, nil
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
	subjectString := strings.ToLower(strings.TrimSpace(assessmentData.Subject))

	subject, validSubject := model.GetSubjectFromString(subjectString)
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
	fieldCount := 0
	if ptr := toStringPtr(assessmentData.GradeLetter); ptr != nil {
		subjectReport.GradeLetter = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.ClassName); ptr != nil {
		subjectReport.ClassName = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.InstructorName); ptr != nil {
		subjectReport.InstructorName = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.Term); ptr != nil {
		subjectReport.Term = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.Remarks); ptr != nil {
		subjectReport.Remarks = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.LearningObjectivesMastered); ptr != nil {
		subjectReport.LearningObjectivesMastered = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.AreasForImprovement); ptr != nil {
		subjectReport.AreasForImprovement = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.RecommendedResources); ptr != nil {
		subjectReport.RecommendedResources = ptr
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.TargetGoals); ptr != nil {
		subjectReport.TargetGoals = ptr
		fieldCount++
	}

	// Set optional integer fields
	if ptr := toIntPtr(assessmentData.MidtermScore); ptr != nil {
		subjectReport.MidtermScore = ptr
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.FinalExamScore); ptr != nil {
		subjectReport.FinalExamScore = ptr
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.QuizScore); ptr != nil {
		subjectReport.QuizScore = ptr
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.AssignmentScore); ptr != nil {
		subjectReport.AssignmentScore = ptr
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.PracticalScore); ptr != nil {
		subjectReport.PracticalScore = ptr
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.OralPresentationScore); ptr != nil {
		subjectReport.OralPresentationScore = ptr
		fieldCount++
	}

	// Set optional float fields
	if ptr := toFloat64Ptr(assessmentData.ConceptualUnderstanding); ptr != nil {
		subjectReport.ConceptualUnderstanding = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.ProblemSolving); ptr != nil {
		subjectReport.ProblemSolving = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.KnowledgeApplication); ptr != nil {
		subjectReport.KnowledgeApplication = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.AnalyticalThinking); ptr != nil {
		subjectReport.AnalyticalThinking = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Creativity); ptr != nil {
		subjectReport.Creativity = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.PracticalSkills); ptr != nil {
		subjectReport.PracticalSkills = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Participation); ptr != nil {
		subjectReport.Participation = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Discipline); ptr != nil {
		subjectReport.Discipline = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Punctuality); ptr != nil {
		subjectReport.Punctuality = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Teamwork); ptr != nil {
		subjectReport.Teamwork = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.EffortLevel); ptr != nil {
		subjectReport.EffortLevel = ptr
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Improvement); ptr != nil {
		subjectReport.Improvement = ptr
		fieldCount++
	}

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

	responseFieldCount := 4 // Base fields
	// Add optional fields only if they exist
	if savedReport.GradeLetter != nil {
		assessmentDataResponse["gradeLetter"] = *savedReport.GradeLetter
		responseFieldCount++
	}
	if savedReport.ClassName != nil {
		assessmentDataResponse["className"] = *savedReport.ClassName
		responseFieldCount++
	}
	if savedReport.InstructorName != nil {
		assessmentDataResponse["instructorName"] = *savedReport.InstructorName
		responseFieldCount++
	}
	if savedReport.Term != nil {
		assessmentDataResponse["term"] = *savedReport.Term
		responseFieldCount++
	}
	if savedReport.Remarks != nil {
		assessmentDataResponse["remarks"] = *savedReport.Remarks
		responseFieldCount++
	}
	if savedReport.MidtermScore != nil {
		assessmentDataResponse["midtermScore"] = *savedReport.MidtermScore
		responseFieldCount++
	}
	if savedReport.FinalExamScore != nil {
		assessmentDataResponse["finalExamScore"] = *savedReport.FinalExamScore
		responseFieldCount++
	}
	if savedReport.QuizScore != nil {
		assessmentDataResponse["quizScore"] = *savedReport.QuizScore
		responseFieldCount++
	}
	if savedReport.AssignmentScore != nil {
		assessmentDataResponse["assignmentScore"] = *savedReport.AssignmentScore
		responseFieldCount++
	}
	if savedReport.PracticalScore != nil {
		assessmentDataResponse["practicalScore"] = *savedReport.PracticalScore
		responseFieldCount++
	}
	if savedReport.OralPresentationScore != nil {
		assessmentDataResponse["oralPresentationScore"] = *savedReport.OralPresentationScore
		responseFieldCount++
	}
	if savedReport.ConceptualUnderstanding != nil {
		assessmentDataResponse["conceptualUnderstanding"] = *savedReport.ConceptualUnderstanding
		responseFieldCount++
	}
	if savedReport.ProblemSolving != nil {
		assessmentDataResponse["problemSolving"] = *savedReport.ProblemSolving
		responseFieldCount++
	}
	if savedReport.KnowledgeApplication != nil {
		assessmentDataResponse["knowledgeApplication"] = *savedReport.KnowledgeApplication
		responseFieldCount++
	}
	if savedReport.AnalyticalThinking != nil {
		assessmentDataResponse["analyticalThinking"] = *savedReport.AnalyticalThinking
		responseFieldCount++
	}
	if savedReport.Creativity != nil {
		assessmentDataResponse["creativity"] = *savedReport.Creativity
		responseFieldCount++
	}
	if savedReport.PracticalSkills != nil {
		assessmentDataResponse["practicalSkills"] = *savedReport.PracticalSkills
		responseFieldCount++
	}
	if savedReport.Participation != nil {
		assessmentDataResponse["participation"] = *savedReport.Participation
		responseFieldCount++
	}
	if savedReport.Discipline != nil {
		assessmentDataResponse["discipline"] = *savedReport.Discipline
		responseFieldCount++
	}
	if savedReport.Punctuality != nil {
		assessmentDataResponse["punctuality"] = *savedReport.Punctuality
		responseFieldCount++
	}
	if savedReport.Teamwork != nil {
		assessmentDataResponse["teamwork"] = *savedReport.Teamwork
		responseFieldCount++
	}
	if savedReport.EffortLevel != nil {
		assessmentDataResponse["effortLevel"] = *savedReport.EffortLevel
		responseFieldCount++
	}
	if savedReport.Improvement != nil {
		assessmentDataResponse["improvement"] = *savedReport.Improvement
		responseFieldCount++
	}
	if savedReport.LearningObjectivesMastered != nil {
		assessmentDataResponse["learningObjectivesMastered"] = *savedReport.LearningObjectivesMastered
		responseFieldCount++
	}
	if savedReport.AreasForImprovement != nil {
		assessmentDataResponse["areasForImprovement"] = *savedReport.AreasForImprovement
		responseFieldCount++
	}
	if savedReport.RecommendedResources != nil {
		assessmentDataResponse["recommendedResources"] = *savedReport.RecommendedResources
		responseFieldCount++
	}
	if savedReport.TargetGoals != nil {
		assessmentDataResponse["targetGoals"] = *savedReport.TargetGoals
		responseFieldCount++
	}

	finalResponse := map[string]interface{}{
		"assessmentData": assessmentDataResponse,
	}

	// Convert the entire response to camelCase (including nested structs)
	camelCaseResponseData := convertKeysToCamelCase(finalResponse)

	// Type assert back to map[string]interface{} for return
	finalResponseData, ok := camelCaseResponseData.(map[string]interface{})
	if !ok {
		return finalResponse, nil // Fallback to original data
	}

	return finalResponseData, nil
}

func handleReportCardGeneration(reportCardDataInterface interface{}, teacherId string) (map[string]interface{}, error) {
	// The report card data is now pre-extracted from the agent response
	// We can directly use it to save to database and pass through to the frontend
	reportCardData := reportCardDataInterface

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
		// Continue and return the data even if saving fails
		savedReportCard = nil
	}

	// Prepare response data
	// Convert all keys in the report card data to camelCase
	camelCaseReportCard := convertKeysToCamelCase(reportCardData)

	responseData := map[string]interface{}{
		"reportCard": camelCaseReportCard,
	}

	// Add database metadata if save was successful
	if savedReportCard != nil {
		responseData["databaseId"] = savedReportCard.ID.Hex()
		responseData["savedAt"] = savedReportCard.CreatedAt.Format("2006-01-02T15:04:05.000Z")
	}

	return responseData, nil
}

// handleAssignmentResultSaving processes assignment result data from assessor agent
func handleAssignmentResultSaving(assignmentResultData interface{}) (map[string]interface{}, error) {
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
		mcqResultsBytes, err := json.Marshal(mcqResults)
		if err == nil {
			var mcqResultsList []model.MCQResult
			if err := json.Unmarshal(mcqResultsBytes, &mcqResultsList); err == nil {
				assignmentResult.MCQResults = mcqResultsList
			}
		}
	}

	if msqResults, ok := resultMap["msq_results"]; ok {
		msqResultsBytes, err := json.Marshal(msqResults)
		if err == nil {
			var msqResultsList []model.MSQResult
			if err := json.Unmarshal(msqResultsBytes, &msqResultsList); err == nil {
				assignmentResult.MSQResults = msqResultsList
			}
		}
	}

	if natResults, ok := resultMap["nat_results"]; ok {
		natResultsBytes, err := json.Marshal(natResults)
		if err == nil {
			var natResultsList []model.NATResult
			if err := json.Unmarshal(natResultsBytes, &natResultsList); err == nil {
				assignmentResult.NATResults = natResultsList
			}
		}
	}

	if subjectiveResults, ok := resultMap["subjective_results"]; ok {
		subjectiveResultsBytes, err := json.Marshal(subjectiveResults)
		if err == nil {
			var subjectiveResultsList []model.SubjectiveResult
			if err := json.Unmarshal(subjectiveResultsBytes, &subjectiveResultsList); err == nil {
				assignmentResult.SubjectiveResults = subjectiveResultsList
			}
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

	// Return the saved assignment result data directly (flattened)
	responseData := map[string]interface{}{
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
	}

	// Convert the entire response to camelCase (including nested structs)
	camelCaseResponseData := convertKeysToCamelCase(responseData)

	// Type assert back to map[string]interface{} for return
	finalResponseData, ok := camelCaseResponseData.(map[string]interface{})
	if !ok {
		return responseData, nil // Fallback to original data
	}

	return finalResponseData, nil
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

func RAGAgentClient(corpusName string, message string) (*pb.RAGAgentResponse, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		log.Printf("ERROR: Failed to establish gRPC connection: %v", err)
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.RAGAgentRequest{
		CorpusName: corpusName,
		Message:    message,
	}

	resp, err := client.RAGAgent(ctx, req)
	if err != nil {
		log.Printf("ERROR: RAG agent call failed: %v", err)
		return nil, fmt.Errorf("failed to call RAG agent: %v", err)
	}

	return resp, nil
}
