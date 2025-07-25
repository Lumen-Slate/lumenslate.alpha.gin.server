package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

func Agent(file, fileType, teacherId, role, message, createdAt, updatedAt string) (map[string]interface{}, error) {
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

	res, err := client.Agent(ctx, req)
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
			if assignmentResult, err := handleAssignmentResultSaving(assignmentResultData, teacherId); err == nil {
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
			if questionsData, err := handleQuestionGeneration(agentResponse.QuestionsRequested, teacherId, rawAgentResponse); err == nil {
				responseData = questionsData
				agentName = "assignment_generator_general"
				responseMessage = "Assignment generated successfully"
			} else {
				log.Printf("ERROR: Failed to handle question generation: %v", err)
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

func handleQuestionGeneration(questionsRequested []QuestionRequest, teacherId string, rawAgentResponse string) (map[string]interface{}, error) {
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

func RAGAgentClient(teacherId, message, file string) (*pb.RAGAgentResponse, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		log.Printf("ERROR: Failed to establish gRPC connection: %v", err)
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
		log.Printf("ERROR: RAG agent call failed: %v", err)
		return nil, fmt.Errorf("failed to call RAG agent: %v", err)
	}

	return resp, nil
}
