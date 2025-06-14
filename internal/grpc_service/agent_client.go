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

func Agent(file, fileType, teacherId, role, message, createdAt, updatedAt string) (map[string]interface{}, error) {
	log.Printf("=== AGENT FUNCTION START ===")
	log.Printf("Input parameters - file: '%s', fileType: '%s', teacherId: '%s', role: '%s', message length: %d",
		file, fileType, teacherId, role, len(message))
	log.Printf("Timestamps - createdAt: '%s', updatedAt: '%s'", createdAt, updatedAt)

	client, conn, err := DialGRPC()
	if err != nil {
		log.Printf("ERROR: Failed to establish gRPC connection: %v", err)
		return nil, err
	}
	defer conn.Close()
	log.Printf("Successfully established gRPC connection")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Printf("Created context with 60 second timeout")

	req := &pb.AgentRequest{
		File:      file,
		FileType:  fileType,
		TeacherId: teacherId,
		Role:      role,
		Message:   message,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	log.Printf("Sending gRPC request to Agent service...")

	res, err := client.Agent(ctx, req)
	if err != nil {
		log.Printf("ERROR: gRPC Agent call failed: %v", err)
		errorResponse := createErrorResponse(teacherId, err.Error(), res)
		log.Printf("Returning error response: %+v", errorResponse)
		return errorResponse, nil
	}
	log.Printf("Successfully received gRPC response")

	// Get the raw agent response
	rawAgentResponse := res.GetAgentResponse()
	log.Printf("=== RAW AGENT RESPONSE ===")
	log.Printf("Response length: %d characters", len(rawAgentResponse))
	log.Printf("Raw response preview (first 500 chars): %.500s", rawAgentResponse)

	// Try to parse the agent response as JSON to see if it contains database operations
	var agentResponse AgentResponse
	agentName := "root_agent"
	responseMessage := "Agent response processed successfully"
	var responseData interface{}

	// Add debug logging to see the actual response structure
	log.Printf("=== ATTEMPTING JSON PARSING ===")
	log.Printf("Trying to parse response as structured JSON...")

	if err := json.Unmarshal([]byte(rawAgentResponse), &agentResponse); err == nil {
		log.Printf("✓ Successfully parsed JSON response")

		// Handle report card field name mismatch: microservice returns "report_card" but we expect "report_card_data"
		if agentResponse.ReportCardData == nil {
			var rawMap map[string]interface{}
			if err := json.Unmarshal([]byte(rawAgentResponse), &rawMap); err == nil {
				if reportCardData, exists := rawMap["report_card"]; exists {
					agentResponse.ReportCardData = reportCardData
					log.Printf("✓ Found 'report_card' field and mapped it to ReportCardData")
				}
			}
		}

		// Successfully parsed JSON, log what we found
		log.Printf("=== PARSED AGENT RESPONSE STRUCTURE ===")
		log.Printf("QuestionsRequested count: %d", len(agentResponse.QuestionsRequested))
		log.Printf("AssignmentResult present: %v", agentResponse.AssignmentResult != nil)
		log.Printf("AssessmentResult present: %v", agentResponse.AssessmentResult != nil)
		log.Printf("AssessmentData present: %v", agentResponse.AssessmentData != nil)
		log.Printf("ReportCardData present: %v", agentResponse.ReportCardData != nil)

		// Check for assignment result FIRST (prioritize assessor agent)
		// Handle both "assignment_result" and "assessment_result" field names
		var assignmentResultData interface{}
		if agentResponse.AssignmentResult != nil {
			assignmentResultData = agentResponse.AssignmentResult
			log.Printf("=== DETECTED: Assignment result from 'assignment_result' field ===")
		} else if agentResponse.AssessmentResult != nil {
			assignmentResultData = agentResponse.AssessmentResult
			log.Printf("=== DETECTED: Assignment result from 'assessment_result' field ===")
		}

		if assignmentResultData != nil {
			log.Printf("Processing assignment result (assessor agent)...")
			// Handle assignment result saving (from assessor agent)
			if assignmentResult, err := handleAssignmentResultSaving(assignmentResultData, teacherId); err == nil {
				responseData = assignmentResult
				agentName = "assessor_agent"
				responseMessage = "Assignment assessment completed successfully"
				log.Printf("✓ Assignment result processed successfully")
			} else {
				log.Printf("ERROR: Failed to handle assignment result saving: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else if len(agentResponse.QuestionsRequested) > 0 {
			log.Printf("=== DETECTED: Question generation request ===")
			log.Printf("Number of question requests: %d", len(agentResponse.QuestionsRequested))
			// Handle question generation (both AGG and AGT)
			if questionsData, err := handleQuestionGeneration(agentResponse.QuestionsRequested, teacherId, rawAgentResponse); err == nil {
				responseData = questionsData
				// Determine agent name based on the request - both AGG and AGT use same handler
				agentName = "assignment_generator_general" // Both AGG and AGT return this structure
				responseMessage = "Assignment generated successfully"
				log.Printf("✓ Question generation processed successfully")
			} else {
				log.Printf("ERROR: Failed to handle question generation: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else if agentResponse.AssessmentData != nil {
			log.Printf("=== DETECTED: Legacy assessment data ===")
			// Handle legacy assessment data saving (for backward compatibility)
			if assessmentData, err := handleAssessmentSaving(agentResponse.AssessmentData, teacherId); err == nil {
				responseData = assessmentData
				agentName = "assessor_agent"
				responseMessage = "Subject assessment report processed successfully"
				log.Printf("✓ Legacy assessment data processed successfully")
			} else {
				log.Printf("ERROR: Failed to handle assessment saving: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else if agentResponse.ReportCardData != nil {
			log.Printf("=== DETECTED: Report card generation request ===")
			// Handle report card generation and saving
			if reportCardData, err := handleReportCardGeneration(agentResponse.ReportCardData, teacherId); err == nil {
				responseData = reportCardData
				agentName = "report_card_generator"
				responseMessage = "Report card generated and saved successfully"
				log.Printf("✓ Report card generation processed successfully")
			} else {
				log.Printf("ERROR: Failed to handle report card generation: %v", err)
				errorResponse := createErrorResponse(teacherId, err.Error(), res)
				return errorResponse, nil
			}
		} else {
			log.Printf("=== DETECTED: Regular agent response (no database operations) ===")
			// Regular agent response without database operations
			responseData = map[string]interface{}{
				"agentResponse": rawAgentResponse,
			}
		}
	} else {
		// Couldn't parse as JSON, this is likely from general_chat_agent
		// Put the response in the message field and leave data empty
		log.Printf("✗ Failed to parse JSON: %v", err)
		log.Printf("=== DETECTED: General chat response (non-JSON) ===")
		log.Printf("Treating as general chat response from general_chat_agent")
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

	log.Printf("=== FINAL RESPONSE SUMMARY ===")
	log.Printf("Agent Name: %s", agentName)
	log.Printf("Response Message: %s", responseMessage)
	log.Printf("Data Type: %T", responseData)
	log.Printf("Session ID: %s", res.GetSessionId())
	log.Printf("Response Time: %s", res.GetResponseTime())
	log.Printf("=== AGENT FUNCTION END ===")

	// Return the standardized response format
	return finalResponse, nil
}

func createErrorResponse(teacherId, errorMessage string, res *pb.AgentResponse) map[string]interface{} {
	log.Printf("=== CREATING ERROR RESPONSE ===")
	log.Printf("Teacher ID: %s", teacherId)
	log.Printf("Error Message: %s", errorMessage)
	log.Printf("Response object present: %v", res != nil)

	var sessionId, createdAt, updatedAt, responseTime, feedback string
	if res != nil {
		sessionId = res.GetSessionId()
		createdAt = res.GetCreatedAt()
		updatedAt = res.GetUpdatedAt()
		responseTime = res.GetResponseTime()
		feedback = res.GetFeedback()
		log.Printf("Extracted from response - SessionId: %s, CreatedAt: %s, UpdatedAt: %s", sessionId, createdAt, updatedAt)
	} else {
		log.Printf("No response object available, using empty values")
	}

	errorResponse := map[string]interface{}{
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

	log.Printf("Error response created: %+v", errorResponse)
	return errorResponse
}

func handleQuestionGeneration(questionsRequested []QuestionRequest, teacherId string, rawAgentResponse string) (map[string]interface{}, error) {
	log.Printf("=== HANDLE QUESTION GENERATION START ===")
	log.Printf("Teacher ID: %s", teacherId)
	log.Printf("Total question requests: %d", len(questionsRequested))
	log.Printf("Raw agent response length: %d characters", len(rawAgentResponse))

	// Parse the raw agent response to extract title and body
	var agentResponseData map[string]interface{}
	var assignmentTitle, assignmentBody string

	log.Printf("Attempting to extract assignment title and body from raw response...")
	if err := json.Unmarshal([]byte(rawAgentResponse), &agentResponseData); err == nil {
		log.Printf("✓ Successfully parsed raw response for title/body extraction")
		if title, ok := agentResponseData["title"].(string); ok {
			assignmentTitle = title
			log.Printf("Found assignment title: '%s'", title)
		} else {
			log.Printf("No title field found in response")
		}
		if body, ok := agentResponseData["body"].(string); ok {
			assignmentBody = body
			log.Printf("Found assignment body (length: %d characters)", len(body))
		} else {
			log.Printf("No body field found in response")
		}
	} else {
		log.Printf("✗ Failed to parse raw response for title/body: %v", err)
	}

	// Default values if not provided by agent
	if assignmentTitle == "" {
		assignmentTitle = "Generated Assignment"
		log.Printf("Using default assignment title: '%s'", assignmentTitle)
	}
	if assignmentBody == "" {
		assignmentBody = "Assignment generated from agent request"
		log.Printf("Using default assignment body: '%s'", assignmentBody)
	}

	// Log all incoming requests
	log.Printf("=== ANALYZING QUESTION REQUESTS ===")
	for i, req := range questionsRequested {
		log.Printf("Request %d:", i+1)
		log.Printf("  Type: '%s'", req.Type)
		log.Printf("  Subject: '%s'", req.Subject)
		log.Printf("  Number of Questions: %d", req.NumberOfQuestions)
		log.Printf("  Difficulty: '%s'", req.Difficulty)
	}

	// Test database connectivity
	log.Printf("=== DATABASE CONNECTIVITY CHECK ===")
	if err := repository.DebugTestDatabaseConnection(); err != nil {
		log.Printf("ERROR: Database connectivity test failed: %v", err)
		return nil, fmt.Errorf("database connectivity failed: %v", err)
	}
	log.Printf("✓ Database connectivity confirmed")

	// Debug database contents
	log.Printf("=== DATABASE CONTENTS DEBUG ===")
	repository.DebugGetAllSubjects()
	repository.DebugGetSampleQuestions(5)
	repository.DebugFindMathQuestions()
	repository.DebugListAllCollections()

	// Group requests by subject to handle multiple difficulty levels for the same subject
	subjectRequests := make(map[string][]QuestionRequest)
	log.Printf("=== GROUPING REQUESTS BY SUBJECT ===")

	for i, request := range questionsRequested {
		if request.Type == "assignment_generator_general" {
			subject := strings.ToLower(strings.TrimSpace(request.Subject))
			log.Printf("Request %d: '%s' -> normalized to: '%s'", i+1, request.Subject, subject)
			subjectRequests[subject] = append(subjectRequests[subject], request)
		} else {
			log.Printf("Request %d: Skipping non-assignment_generator_general type: '%s'", i+1, request.Type)
		}
	}

	log.Printf("Grouped into %d unique subjects:", len(subjectRequests))
	for subject, requests := range subjectRequests {
		log.Printf("  Subject '%s': %d requests", subject, len(requests))
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

	log.Printf("=== PROCESSING EACH SUBJECT ===")
	for subjectKey, requests := range subjectRequests {
		log.Printf("--- Processing subject: '%s' ---", subjectKey)

		// Convert subject string to Subject enum
		subject, validSubject := model.GetSubjectFromString(subjectKey)
		log.Printf("Subject conversion: '%s' -> enum '%s', valid: %v", subjectKey, subject, validSubject)

		if !validSubject {
			log.Printf("WARNING: Skipping invalid subject: %s", subjectKey)
			continue
		}

		// Process all requests for this subject
		for reqIndex, req := range requests {
			log.Printf("  Request %d for subject '%s':", reqIndex+1, subject)
			numQuestions := req.NumberOfQuestions
			difficulty := strings.ToLower(strings.TrimSpace(req.Difficulty))
			log.Printf("    Difficulty: '%s', Count: %d", difficulty, numQuestions)

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
					log.Printf("    WARNING: Invalid difficulty '%s', getting all questions for subject '%s'", difficulty, subject)
					availableQuestions, err = repository.GetQuestionsBySubject(subject)
				}

				if difficultyEnum != "" {
					log.Printf("    Querying database: subject='%s', difficulty='%s'", subject, difficultyEnum)
					availableQuestions, err = repository.GetQuestionsBySubjectAndDifficulty(subject, difficultyEnum)
				}
			} else {
				// No difficulty specified, get all questions for this subject
				log.Printf("    No difficulty specified, getting all questions for subject '%s'", subject)
				availableQuestions, err = repository.GetQuestionsBySubject(subject)
			}

			if err != nil {
				log.Printf("    ERROR: Failed to get questions for subject %s: %v", subject, err)
				continue
			}

			log.Printf("    Database query result: %d available questions for %d requested", len(availableQuestions), numQuestions)

			if len(availableQuestions) >= numQuestions {
				// Randomly sample the requested number of questions
				log.Printf("    ✓ Sufficient questions available, randomly selecting %d questions", numQuestions)
				rng := rand.New(rand.NewSource(time.Now().UnixNano()))
				selectedQuestions := make([]model.Questions, numQuestions)
				perm := rng.Perm(len(availableQuestions))
				for i := 0; i < numQuestions; i++ {
					selectedQuestions[i] = availableQuestions[perm[i]]
					log.Printf("      Selected question %d: ID=%s", i+1, selectedQuestions[i].ID)
				}
				allSelectedQuestions = append(allSelectedQuestions, selectedQuestions...)

				// Count questions by type and collect IDs
				log.Printf("    Categorizing selected questions by type...")
				for qIndex, q := range selectedQuestions {
					questionID := q.ID
					log.Printf("      Processing question %d (ID: %s)", qIndex+1, questionID)

					// Check if this is an MCQ
					if mcq, err := quest.GetMCQByID(questionID); err == nil && mcq != nil {
						mcqCount++
						mcqIds = append(mcqIds, questionID)
						log.Printf("        → Categorized as MCQ (total MCQ: %d)", mcqCount)
						continue
					}

					// Check if this is an MSQ
					if msq, err := quest.GetMSQByID(questionID); err == nil && msq != nil {
						msqCount++
						msqIds = append(msqIds, questionID)
						log.Printf("        → Categorized as MSQ (total MSQ: %d)", msqCount)
						continue
					}

					// Check if this is a NAT
					if nat, err := quest.GetNATByID(questionID); err == nil && nat != nil {
						natCount++
						natIds = append(natIds, questionID)
						log.Printf("        → Categorized as NAT (total NAT: %d)", natCount)
						continue
					}

					// Check if this is a Subjective
					if subjective, err := quest.GetSubjectiveByID(questionID); err == nil && subjective != nil {
						subjectiveCount++
						subjectiveIds = append(subjectiveIds, questionID)
						log.Printf("        → Categorized as Subjective (total Subjective: %d)", subjectiveCount)
						continue
					}

					log.Printf("        WARNING: Question ID %s could not be found in any specific collection", questionID)
				}

				log.Printf("    ✓ Successfully processed %d questions for subject '%s'", numQuestions, subject)
			} else {
				log.Printf("    ✗ Insufficient questions - requested: %d, available: %d", numQuestions, len(availableQuestions))
			}
		}
	}

	// Create assignment if we have selected questions
	log.Printf("=== ASSIGNMENT CREATION ===")
	log.Printf("Total selected questions: %d", len(allSelectedQuestions))
	if len(allSelectedQuestions) == 0 {
		log.Printf("ERROR: No questions could be selected for the assignment")
		return nil, fmt.Errorf("no questions could be selected for the assignment")
	}

	log.Printf("Question distribution:")
	log.Printf("  MCQ: %d questions", mcqCount)
	log.Printf("  MSQ: %d questions", msqCount)
	log.Printf("  NAT: %d questions", natCount)
	log.Printf("  Subjective: %d questions", subjectiveCount)

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

	log.Printf("Assignment details:")
	log.Printf("  ID: %s", assignment.ID)
	log.Printf("  Title: %s", assignment.Title)
	log.Printf("  Body length: %d characters", len(assignment.Body))
	log.Printf("  Due Date: %s", assignment.DueDate.Format("2006-01-02 15:04:05"))
	log.Printf("  Total Points: %d", assignment.Points)
	log.Printf("  Question IDs - MCQ: %d, MSQ: %d, NAT: %d, Subjective: %d",
		len(assignment.MCQIds), len(assignment.MSQIds), len(assignment.NATIds), len(assignment.SubjectiveIds))

	// Save assignment to database
	log.Printf("Saving assignment to database...")
	if err := repository.SaveAssignment(assignment); err != nil {
		log.Printf("ERROR: Failed to save assignment to database: %v", err)
		return nil, fmt.Errorf("failed to save assignment: %v", err)
	}

	log.Printf("✓ Assignment saved successfully to database")

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

	log.Printf("=== QUESTION GENERATION RESPONSE ===")
	log.Printf("Response data: %+v", responseData)
	log.Printf("=== HANDLE QUESTION GENERATION END ===")

	// Return the new simplified response structure
	return responseData, nil
}

func handleAssessmentSaving(assessmentDataInterface interface{}, teacherId string) (map[string]interface{}, error) {
	log.Printf("=== HANDLE ASSESSMENT SAVING START ===")
	log.Printf("Teacher ID: %s", teacherId)
	log.Printf("Assessment data type: %T", assessmentDataInterface)

	// Parse the assessment data
	log.Printf("Marshaling assessment data interface...")
	assessmentBytes, err := json.Marshal(assessmentDataInterface)
	if err != nil {
		log.Printf("ERROR: Failed to marshal assessment data: %v", err)
		return nil, fmt.Errorf("failed to marshal assessment data: %v", err)
	}
	log.Printf("✓ Successfully marshaled assessment data (%d bytes)", len(assessmentBytes))

	var assessmentData AssessmentData
	log.Printf("Unmarshaling to AssessmentData struct...")
	if err := json.Unmarshal(assessmentBytes, &assessmentData); err != nil {
		log.Printf("ERROR: Failed to unmarshal assessment data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal assessment data: %v", err)
	}
	log.Printf("✓ Successfully unmarshaled to AssessmentData struct")

	// Check for mandatory fields
	log.Printf("=== VALIDATING MANDATORY FIELDS ===")
	var missingFields []string

	if assessmentData.StudentID == nil {
		missingFields = append(missingFields, "student_id")
		log.Printf("✗ Missing: student_id")
	} else {
		log.Printf("✓ Found: student_id = %v (type: %T)", assessmentData.StudentID, assessmentData.StudentID)
	}

	if assessmentData.StudentName == "" {
		missingFields = append(missingFields, "student_name")
		log.Printf("✗ Missing: student_name")
	} else {
		log.Printf("✓ Found: student_name = '%s'", assessmentData.StudentName)
	}

	if assessmentData.Subject == "" {
		missingFields = append(missingFields, "subject")
		log.Printf("✗ Missing: subject")
	} else {
		log.Printf("✓ Found: subject = '%s'", assessmentData.Subject)
	}

	if assessmentData.Score == nil {
		missingFields = append(missingFields, "score")
		log.Printf("✗ Missing: score")
	} else {
		log.Printf("✓ Found: score = %v (type: %T)", assessmentData.Score, assessmentData.Score)
	}

	if len(missingFields) > 0 {
		log.Printf("ERROR: Missing mandatory fields: %s", strings.Join(missingFields, ", "))
		return nil, fmt.Errorf("missing mandatory fields: %s", strings.Join(missingFields, ", "))
	}
	log.Printf("✓ All mandatory fields validated successfully")

	// Validate and convert subject to enum
	log.Printf("=== SUBJECT VALIDATION ===")
	subjectString := strings.ToLower(strings.TrimSpace(assessmentData.Subject))
	log.Printf("Subject normalization: '%s' -> '%s'", assessmentData.Subject, subjectString)

	subject, validSubject := model.GetSubjectFromString(subjectString)
	if !validSubject {
		log.Printf("ERROR: Invalid subject: %s", assessmentData.Subject)
		return nil, fmt.Errorf("invalid subject: %s. Available subjects are: math, science, english, history, geography", assessmentData.Subject)
	}
	log.Printf("✓ Subject validated: '%s' -> enum '%s'", assessmentData.Subject, subject)

	// Convert studentID to int
	log.Printf("=== STUDENT ID CONVERSION ===")
	var studentID int
	switch v := assessmentData.StudentID.(type) {
	case float64:
		studentID = int(v)
		log.Printf("Converted float64 %f to int %d", v, studentID)
	case int:
		studentID = v
		log.Printf("Used int value: %d", studentID)
	case string:
		if id, err := strconv.Atoi(v); err == nil {
			studentID = id
			log.Printf("Converted string '%s' to int %d", v, studentID)
		} else {
			log.Printf("ERROR: Invalid student_id string format: %v", v)
			return nil, fmt.Errorf("invalid student_id format: %v", v)
		}
	default:
		log.Printf("ERROR: Invalid student_id type: %T", v)
		return nil, fmt.Errorf("invalid student_id type: %T", v)
	}

	// Convert score to int
	log.Printf("=== SCORE CONVERSION ===")
	var score int
	switch v := assessmentData.Score.(type) {
	case float64:
		score = int(v)
		log.Printf("Converted float64 %f to int %d", v, score)
	case int:
		score = v
		log.Printf("Used int value: %d", score)
	case string:
		if s, err := strconv.Atoi(v); err == nil {
			score = s
			log.Printf("Converted string '%s' to int %d", v, score)
		} else {
			log.Printf("ERROR: Invalid score string format: %v", v)
			return nil, fmt.Errorf("invalid score format: %v", v)
		}
	default:
		log.Printf("ERROR: Invalid score type: %T", v)
		return nil, fmt.Errorf("invalid score type: %T", v)
	}

	// Create SubjectReport object
	now := time.Now()
	log.Printf("=== CREATING SUBJECT REPORT ===")
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
	log.Printf("Base SubjectReport created:")
	log.Printf("  UserID: %s", subjectReport.UserID)
	log.Printf("  StudentID: %d", subjectReport.StudentID)
	log.Printf("  StudentName: %s", subjectReport.StudentName)
	log.Printf("  Subject: %s", subjectReport.Subject)
	log.Printf("  Score: %d", subjectReport.Score)

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

	log.Printf("=== SETTING OPTIONAL FIELDS ===")
	// Set optional string fields
	fieldCount := 0
	if ptr := toStringPtr(assessmentData.GradeLetter); ptr != nil {
		subjectReport.GradeLetter = ptr
		log.Printf("Set GradeLetter: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.ClassName); ptr != nil {
		subjectReport.ClassName = ptr
		log.Printf("Set ClassName: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.InstructorName); ptr != nil {
		subjectReport.InstructorName = ptr
		log.Printf("Set InstructorName: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.Term); ptr != nil {
		subjectReport.Term = ptr
		log.Printf("Set Term: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.Remarks); ptr != nil {
		subjectReport.Remarks = ptr
		log.Printf("Set Remarks: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.LearningObjectivesMastered); ptr != nil {
		subjectReport.LearningObjectivesMastered = ptr
		log.Printf("Set LearningObjectivesMastered: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.AreasForImprovement); ptr != nil {
		subjectReport.AreasForImprovement = ptr
		log.Printf("Set AreasForImprovement: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.RecommendedResources); ptr != nil {
		subjectReport.RecommendedResources = ptr
		log.Printf("Set RecommendedResources: %s", *ptr)
		fieldCount++
	}
	if ptr := toStringPtr(assessmentData.TargetGoals); ptr != nil {
		subjectReport.TargetGoals = ptr
		log.Printf("Set TargetGoals: %s", *ptr)
		fieldCount++
	}

	// Set optional integer fields
	if ptr := toIntPtr(assessmentData.MidtermScore); ptr != nil {
		subjectReport.MidtermScore = ptr
		log.Printf("Set MidtermScore: %d", *ptr)
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.FinalExamScore); ptr != nil {
		subjectReport.FinalExamScore = ptr
		log.Printf("Set FinalExamScore: %d", *ptr)
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.QuizScore); ptr != nil {
		subjectReport.QuizScore = ptr
		log.Printf("Set QuizScore: %d", *ptr)
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.AssignmentScore); ptr != nil {
		subjectReport.AssignmentScore = ptr
		log.Printf("Set AssignmentScore: %d", *ptr)
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.PracticalScore); ptr != nil {
		subjectReport.PracticalScore = ptr
		log.Printf("Set PracticalScore: %d", *ptr)
		fieldCount++
	}
	if ptr := toIntPtr(assessmentData.OralPresentationScore); ptr != nil {
		subjectReport.OralPresentationScore = ptr
		log.Printf("Set OralPresentationScore: %d", *ptr)
		fieldCount++
	}

	// Set optional float fields
	if ptr := toFloat64Ptr(assessmentData.ConceptualUnderstanding); ptr != nil {
		subjectReport.ConceptualUnderstanding = ptr
		log.Printf("Set ConceptualUnderstanding: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.ProblemSolving); ptr != nil {
		subjectReport.ProblemSolving = ptr
		log.Printf("Set ProblemSolving: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.KnowledgeApplication); ptr != nil {
		subjectReport.KnowledgeApplication = ptr
		log.Printf("Set KnowledgeApplication: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.AnalyticalThinking); ptr != nil {
		subjectReport.AnalyticalThinking = ptr
		log.Printf("Set AnalyticalThinking: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Creativity); ptr != nil {
		subjectReport.Creativity = ptr
		log.Printf("Set Creativity: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.PracticalSkills); ptr != nil {
		subjectReport.PracticalSkills = ptr
		log.Printf("Set PracticalSkills: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Participation); ptr != nil {
		subjectReport.Participation = ptr
		log.Printf("Set Participation: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Discipline); ptr != nil {
		subjectReport.Discipline = ptr
		log.Printf("Set Discipline: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Punctuality); ptr != nil {
		subjectReport.Punctuality = ptr
		log.Printf("Set Punctuality: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Teamwork); ptr != nil {
		subjectReport.Teamwork = ptr
		log.Printf("Set Teamwork: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.EffortLevel); ptr != nil {
		subjectReport.EffortLevel = ptr
		log.Printf("Set EffortLevel: %f", *ptr)
		fieldCount++
	}
	if ptr := toFloat64Ptr(assessmentData.Improvement); ptr != nil {
		subjectReport.Improvement = ptr
		log.Printf("Set Improvement: %f", *ptr)
		fieldCount++
	}

	log.Printf("Set %d optional fields", fieldCount)

	// Save to database
	log.Printf("=== SAVING TO DATABASE ===")
	log.Printf("Calling repository.SaveSubjectReport...")
	savedReport, err := repository.SaveSubjectReport(subjectReport)
	if err != nil {
		log.Printf("ERROR: Failed to save subject report to database: %v", err)
		return nil, fmt.Errorf("failed to save subject report: %v", err)
	}
	log.Printf("✓ Successfully saved subject report to database")
	log.Printf("Saved report: %+v", savedReport)

	// Build the response data structure
	log.Printf("=== BUILDING RESPONSE DATA ===")
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

	log.Printf("Built response with %d fields in assessmentData", responseFieldCount)
	log.Printf("Final response keys: %v", func() []string {
		keys := make([]string, 0, len(assessmentDataResponse))
		for k := range assessmentDataResponse {
			keys = append(keys, k)
		}
		return keys
	}())
	log.Printf("=== HANDLE ASSESSMENT SAVING END ===")

	return finalResponse, nil
}

func handleReportCardGeneration(reportCardDataInterface interface{}, teacherId string) (map[string]interface{}, error) {
	log.Printf("=== HANDLE REPORT CARD GENERATION START ===")
	log.Printf("Teacher ID: %s", teacherId)
	log.Printf("Report card data type: %T", reportCardDataInterface)

	// The report card data is now pre-extracted from the agent response
	// We can directly use it to save to database and pass through to the frontend
	reportCardData := reportCardDataInterface
	log.Printf("✓ Using pre-extracted report card data (type: %T)", reportCardData)

	// Convert the report card data to our model structure
	log.Printf("Converting report card data to model structure...")
	reportCardBytes, err := json.Marshal(reportCardData)
	if err != nil {
		log.Printf("ERROR: Failed to marshal report card data: %v", err)
		return nil, fmt.Errorf("failed to marshal report card data: %v", err)
	}
	log.Printf("✓ Successfully marshaled report card data (%d bytes)", len(reportCardBytes))

	var agentReportCardData model.AgentReportCardData
	if err := json.Unmarshal(reportCardBytes, &agentReportCardData); err != nil {
		log.Printf("ERROR: Failed to unmarshal report card data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal report card data: %v", err)
	}
	log.Printf("✓ Successfully unmarshaled to AgentReportCardData model")

	// Create the AgentReportCard with metadata
	log.Printf("Creating AgentReportCard with metadata...")
	agentReportCard := model.AgentReportCard{
		UserID:     teacherId,
		ReportCard: agentReportCardData,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	log.Printf("AgentReportCard created:")
	log.Printf("  UserID: %s", agentReportCard.UserID)
	log.Printf("  CreatedAt: %s", agentReportCard.CreatedAt.Format("2006-01-02 15:04:05"))
	log.Printf("  UpdatedAt: %s", agentReportCard.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Save to database
	log.Printf("Saving AgentReportCard to database...")
	savedReportCard, err := repository.CreateAgentReportCard(agentReportCard)
	if err != nil {
		log.Printf("WARNING: Failed to save agent report card to database: %v", err)
		log.Printf("Continuing with response generation despite database save failure...")
		// Continue and return the data even if saving fails
	} else {
		log.Printf("✓ Successfully saved AgentReportCard to database")
		log.Printf("Saved report card: %+v", savedReportCard)
	}

	// Prepare response data
	log.Printf("Preparing response data...")

	// Convert all keys in the report card data to camelCase
	log.Printf("Converting report card keys from snake_case to camelCase...")
	camelCaseReportCard := convertKeysToCamelCase(reportCardData)
	log.Printf("✓ Successfully converted all keys to camelCase")

	responseData := map[string]interface{}{
		"reportCard": camelCaseReportCard,
	}

	// Add database metadata if save was successful
	if savedReportCard != nil {
		responseData["databaseId"] = savedReportCard.ID.Hex()
		responseData["savedAt"] = savedReportCard.CreatedAt.Format("2006-01-02T15:04:05.000Z")
		log.Printf("Added database metadata to response:")
		log.Printf("  Database ID: %s", savedReportCard.ID.Hex())
		log.Printf("  Saved At: %s", savedReportCard.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	} else {
		log.Printf("No database metadata available (save failed)")
	}

	log.Printf("=== REPORT CARD GENERATION RESPONSE ===")
	log.Printf("Response data keys: %v", func() []string {
		keys := make([]string, 0, len(responseData))
		for k := range responseData {
			keys = append(keys, k)
		}
		return keys
	}())
	log.Printf("Response structure:")
	log.Printf("  - reportCard: Contains complete report card data (all keys converted to camelCase)")
	if _, exists := responseData["databaseId"]; exists {
		log.Printf("  - databaseId: MongoDB ObjectId as hex string")
		log.Printf("  - savedAt: ISO timestamp of database save")
	}
	log.Printf("=== HANDLE REPORT CARD GENERATION END ===")

	return responseData, nil
}

// handleAssignmentResultSaving processes assignment result data from assessor agent
func handleAssignmentResultSaving(assignmentResultData interface{}, teacherId string) (map[string]interface{}, error) {
	log.Printf("=== HANDLE ASSIGNMENT RESULT SAVING START ===")
	log.Printf("Teacher ID: %s", teacherId)
	log.Printf("Assignment result data type: %T", assignmentResultData)
	log.Printf("Assignment result data: %+v", assignmentResultData)

	// Convert interface{} to map[string]interface{}
	log.Printf("Converting assignment result data to map...")
	resultMap, ok := assignmentResultData.(map[string]interface{})
	if !ok {
		log.Printf("ERROR: Invalid assignment result data format - expected map[string]interface{}, got %T", assignmentResultData)
		return nil, fmt.Errorf("invalid assignment result data format")
	}
	log.Printf("✓ Successfully converted to map[string]interface{}")

	log.Printf("=== ASSIGNMENT RESULT MAP CONTENTS ===")
	for key, value := range resultMap {
		log.Printf("Key: '%s', Value: %+v (Type: %T)", key, value, value)
	}

	// Convert to AssignmentResult model
	log.Printf("=== CONVERTING TO ASSIGNMENT RESULT MODEL ===")
	var assignmentResult model.AssignmentResult

	// Map the fields from the response
	log.Printf("Mapping assignment_id field...")
	if assignmentId, ok := resultMap["assignment_id"].(string); ok {
		assignmentResult.AssignmentID = assignmentId
		log.Printf("✓ Found assignment_id: %s", assignmentId)
	} else {
		log.Printf("ERROR: assignment_id is required but not found or not a string")
		log.Printf("Available keys: %v", func() []string {
			keys := make([]string, 0, len(resultMap))
			for k := range resultMap {
				keys = append(keys, k)
			}
			return keys
		}())
		return nil, fmt.Errorf("assignment_id is required")
	}

	log.Printf("Mapping student_id field...")
	if studentId, ok := resultMap["student_id"].(string); ok {
		assignmentResult.StudentID = studentId
		log.Printf("✓ Found student_id: %s", studentId)
	} else {
		log.Printf("ERROR: student_id is required but not found or not a string")
		return nil, fmt.Errorf("student_id is required")
	}

	log.Printf("Mapping total_points_awarded field...")
	if totalPoints, ok := resultMap["total_points_awarded"].(float64); ok {
		assignmentResult.TotalPointsAwarded = int(totalPoints)
		log.Printf("✓ Found total_points_awarded: %f -> %d", totalPoints, int(totalPoints))
	} else {
		log.Printf("WARNING: total_points_awarded not found or not a float64, using default value 0")
	}

	log.Printf("Mapping total_max_points field...")
	if maxPoints, ok := resultMap["total_max_points"].(float64); ok {
		assignmentResult.TotalMaxPoints = int(maxPoints)
		log.Printf("✓ Found total_max_points: %f -> %d", maxPoints, int(maxPoints))
	} else {
		log.Printf("WARNING: total_max_points not found or not a float64, using default value 0")
	}

	log.Printf("Mapping percentage_score field...")
	if percentage, ok := resultMap["percentage_score"].(float64); ok {
		assignmentResult.PercentageScore = percentage
		log.Printf("✓ Found percentage_score: %f", percentage)
	} else {
		log.Printf("WARNING: percentage_score not found or not a float64, using default value 0.0")
	}

	// Convert the results arrays into the proper structs
	log.Printf("=== PROCESSING QUESTION RESULTS ===")

	if mcqResults, ok := resultMap["mcq_results"]; ok {
		log.Printf("Processing mcq_results...")
		log.Printf("MCQ results type: %T", mcqResults)
		log.Printf("MCQ results value: %+v", mcqResults)
		mcqResultsBytes, err := json.Marshal(mcqResults)
		if err != nil {
			log.Printf("ERROR: Failed to marshal MCQ results: %v", err)
		} else {
			log.Printf("MCQ results JSON (%d bytes): %s", len(mcqResultsBytes), string(mcqResultsBytes))
			var mcqResultsList []model.MCQResult
			if err := json.Unmarshal(mcqResultsBytes, &mcqResultsList); err == nil {
				assignmentResult.MCQResults = mcqResultsList
				log.Printf("✓ Successfully processed %d MCQ results", len(mcqResultsList))
				for i, result := range mcqResultsList {
					log.Printf("  MCQ %d: QuestionID=%s, StudentAnswer=%d, CorrectAnswer=%d, IsCorrect=%v, PointsAwarded=%d",
						i+1, result.QuestionID, result.StudentAnswer, result.CorrectAnswer, result.IsCorrect, result.PointsAwarded)
				}
			} else {
				log.Printf("ERROR: Failed to unmarshal MCQ results: %v", err)
			}
		}
	} else {
		log.Printf("No mcq_results found in response")
	}

	if msqResults, ok := resultMap["msq_results"]; ok {
		log.Printf("Processing msq_results...")
		log.Printf("MSQ results type: %T", msqResults)
		log.Printf("MSQ results value: %+v", msqResults)
		msqResultsBytes, err := json.Marshal(msqResults)
		if err != nil {
			log.Printf("ERROR: Failed to marshal MSQ results: %v", err)
		} else {
			log.Printf("MSQ results JSON (%d bytes): %s", len(msqResultsBytes), string(msqResultsBytes))
			var msqResultsList []model.MSQResult
			if err := json.Unmarshal(msqResultsBytes, &msqResultsList); err == nil {
				assignmentResult.MSQResults = msqResultsList
				log.Printf("✓ Successfully processed %d MSQ results", len(msqResultsList))
				for i, result := range msqResultsList {
					log.Printf("  MSQ %d: QuestionID=%s, StudentAnswers=%v, CorrectAnswers=%v, IsCorrect=%v, PointsAwarded=%d",
						i+1, result.QuestionID, result.StudentAnswers, result.CorrectAnswers, result.IsCorrect, result.PointsAwarded)
				}
			} else {
				log.Printf("ERROR: Failed to unmarshal MSQ results: %v", err)
			}
		}
	} else {
		log.Printf("No msq_results found in response")
	}

	if natResults, ok := resultMap["nat_results"]; ok {
		log.Printf("Processing nat_results...")
		log.Printf("NAT results type: %T", natResults)
		log.Printf("NAT results value: %+v", natResults)
		natResultsBytes, err := json.Marshal(natResults)
		if err != nil {
			log.Printf("ERROR: Failed to marshal NAT results: %v", err)
		} else {
			log.Printf("NAT results JSON (%d bytes): %s", len(natResultsBytes), string(natResultsBytes))
			var natResultsList []model.NATResult
			if err := json.Unmarshal(natResultsBytes, &natResultsList); err == nil {
				assignmentResult.NATResults = natResultsList
				log.Printf("✓ Successfully processed %d NAT results", len(natResultsList))
				for i, result := range natResultsList {
					log.Printf("  NAT %d: QuestionID=%s, StudentAnswer=%v, CorrectAnswer=%v, IsCorrect=%v, PointsAwarded=%d",
						i+1, result.QuestionID, result.StudentAnswer, result.CorrectAnswer, result.IsCorrect, result.PointsAwarded)
				}
			} else {
				log.Printf("ERROR: Failed to unmarshal NAT results: %v", err)
			}
		}
	} else {
		log.Printf("No nat_results found in response")
	}

	if subjectiveResults, ok := resultMap["subjective_results"]; ok {
		log.Printf("Processing subjective_results...")
		log.Printf("Subjective results type: %T", subjectiveResults)
		log.Printf("Subjective results value: %+v", subjectiveResults)
		subjectiveResultsBytes, err := json.Marshal(subjectiveResults)
		if err != nil {
			log.Printf("ERROR: Failed to marshal Subjective results: %v", err)
		} else {
			log.Printf("Subjective results JSON (%d bytes): %s", len(subjectiveResultsBytes), string(subjectiveResultsBytes))
			var subjectiveResultsList []model.SubjectiveResult
			if err := json.Unmarshal(subjectiveResultsBytes, &subjectiveResultsList); err == nil {
				assignmentResult.SubjectiveResults = subjectiveResultsList
				log.Printf("✓ Successfully processed %d Subjective results", len(subjectiveResultsList))
				for i, result := range subjectiveResultsList {
					log.Printf("  Subjective %d: QuestionID=%s, StudentAnswer='%s', PointsAwarded=%d, MaxPoints=%d, AssessmentFeedback='%s'",
						i+1, result.QuestionID, result.StudentAnswer, result.PointsAwarded, result.MaxPoints, result.AssessmentFeedback)
				}
			} else {
				log.Printf("ERROR: Failed to unmarshal Subjective results: %v", err)
			}
		}
	} else {
		log.Printf("No subjective_results found in response")
	}

	// Set metadata
	log.Printf("=== SETTING METADATA ===")
	assignmentResult.ID = primitive.NewObjectID()
	assignmentResult.CreatedAt = time.Now()
	assignmentResult.UpdatedAt = time.Now()
	log.Printf("Generated ObjectID: %s", assignmentResult.ID.Hex())
	log.Printf("CreatedAt: %s", assignmentResult.CreatedAt.Format("2006-01-02 15:04:05"))
	log.Printf("UpdatedAt: %s", assignmentResult.UpdatedAt.Format("2006-01-02 15:04:05"))

	log.Printf("=== ASSIGNMENT RESULT SUMMARY BEFORE SAVE ===")
	log.Printf("AssignmentID: %s", assignmentResult.AssignmentID)
	log.Printf("StudentID: %s", assignmentResult.StudentID)
	log.Printf("TotalPointsAwarded: %d", assignmentResult.TotalPointsAwarded)
	log.Printf("TotalMaxPoints: %d", assignmentResult.TotalMaxPoints)
	log.Printf("PercentageScore: %f", assignmentResult.PercentageScore)
	log.Printf("MCQ Results: %d", len(assignmentResult.MCQResults))
	log.Printf("MSQ Results: %d", len(assignmentResult.MSQResults))
	log.Printf("NAT Results: %d", len(assignmentResult.NATResults))
	log.Printf("Subjective Results: %d", len(assignmentResult.SubjectiveResults))

	// Save to database using repository
	log.Printf("=== SAVING TO DATABASE ===")
	log.Printf("Calling repository.CreateAssignmentResult...")
	savedResult, err := repository.CreateAssignmentResult(assignmentResult)
	if err != nil {
		log.Printf("ERROR: Failed to save assignment result to database: %v", err)
		return nil, fmt.Errorf("failed to save assignment result: %v", err)
	}
	assignmentResult = *savedResult
	log.Printf("✓ Successfully saved assignment result to database")
	log.Printf("Saved assignment result ID: %s", assignmentResult.ID.Hex())

	// Return the saved assignment result data directly (flattened)
	log.Printf("=== BUILDING RESPONSE ===")

	// Convert result arrays to camelCase for consistent frontend consumption
	log.Printf("Converting result arrays to camelCase...")
	camelCaseMcqResults := convertKeysToCamelCase(assignmentResult.MCQResults)
	camelCaseMsqResults := convertKeysToCamelCase(assignmentResult.MSQResults)
	camelCaseNatResults := convertKeysToCamelCase(assignmentResult.NATResults)
	camelCaseSubjectiveResults := convertKeysToCamelCase(assignmentResult.SubjectiveResults)
	log.Printf("✓ Successfully converted all result arrays to camelCase")

	responseData := map[string]interface{}{
		"id":                 assignmentResult.ID.Hex(),
		"assignmentId":       assignmentResult.AssignmentID,
		"studentId":          assignmentResult.StudentID,
		"totalPointsAwarded": assignmentResult.TotalPointsAwarded,
		"totalMaxPoints":     assignmentResult.TotalMaxPoints,
		"percentageScore":    assignmentResult.PercentageScore,
		"mcqResults":         camelCaseMcqResults,
		"msqResults":         camelCaseMsqResults,
		"natResults":         camelCaseNatResults,
		"subjectiveResults":  camelCaseSubjectiveResults,
		"createdAt":          assignmentResult.CreatedAt,
		"updatedAt":          assignmentResult.UpdatedAt,
	}

	log.Printf("Response data summary:")
	log.Printf("  Database ID: %s", responseData["id"])
	log.Printf("  Assignment ID: %s", responseData["assignmentId"])
	log.Printf("  Student ID: %s", responseData["studentId"])
	log.Printf("  Total Points: %d/%d (%.2f%%)", responseData["totalPointsAwarded"], responseData["totalMaxPoints"], responseData["percentageScore"])
	log.Printf("  Results count - MCQ: %d, MSQ: %d, NAT: %d, Subjective: %d (all in camelCase)",
		len(assignmentResult.MCQResults), len(assignmentResult.MSQResults),
		len(assignmentResult.NATResults), len(assignmentResult.SubjectiveResults))
	log.Printf("=== HANDLE ASSIGNMENT RESULT SAVING END ===")

	return responseData, nil
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
	log.Printf("=== RAG AGENT CLIENT START ===")
	log.Printf("Teacher ID: %s", teacherId)
	log.Printf("Message length: %d characters", len(message))
	log.Printf("File parameter: %s", file)

	log.Printf("Establishing gRPC connection...")
	client, conn, err := DialGRPC()
	if err != nil {
		log.Printf("ERROR: Failed to establish gRPC connection: %v", err)
		return nil, err
	}
	defer conn.Close()
	log.Printf("✓ Successfully established gRPC connection")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	log.Printf("Created context with 30 second timeout")

	req := &pb.RAGAgentRequest{
		TeacherId: teacherId,
		Message:   message,
		File:      file,
	}
	log.Printf("Created RAG agent request:")
	log.Printf("  TeacherId: %s", req.TeacherId)
	log.Printf("  Message: %.100s%s", req.Message, func() string {
		if len(req.Message) > 100 {
			return "..."
		}
		return ""
	}())
	log.Printf("  File: %s", req.File)

	log.Printf("Calling RAG agent service...")
	resp, err := client.RAGAgent(ctx, req)
	if err != nil {
		log.Printf("ERROR: RAG agent call failed: %v", err)
		return nil, fmt.Errorf("failed to call RAG agent: %v", err)
	}
	log.Printf("✓ Successfully received RAG agent response")

	log.Printf("=== RAG AGENT RESPONSE SUMMARY ===")
	if resp != nil {
		log.Printf("Response received successfully")
		log.Printf("Message field length: %d", len(resp.GetMessage()))
		log.Printf("Agent name: %s", resp.GetAgentName())
		log.Printf("Agent response length: %d", len(resp.GetAgentResponse()))
		log.Printf("Session ID: %s", resp.GetSessionId())
		log.Printf("Response time: %s", resp.GetResponseTime())
		log.Printf("Role: %s", resp.GetRole())
		log.Printf("Feedback: %s", resp.GetFeedback())
	} else {
		log.Printf("WARNING: Received nil response")
	}
	log.Printf("=== RAG AGENT CLIENT END ===")

	return resp, nil
}
