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
	// Group requests by subject to handle multiple difficulty levels for the same subject
	subjectRequests := make(map[string][]QuestionRequest)

	for _, request := range questionsRequested {
		if request.Type == "assignment_generator_general" {
			subject := strings.ToLower(strings.TrimSpace(request.Subject))
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
				log.Printf("Error getting questions for subject %s: %v", subject, err)
				continue
			}

			if len(availableQuestions) == 0 {
				// No questions available for this difficulty
				continue
			} else if numQuestions > len(availableQuestions) {
				// Not enough questions for this difficulty, skip this request
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
			}
		}

		// Get total available questions for this subject
		totalAvailable, err := repository.CountQuestionsBySubject(subject)
		if err != nil {
			totalAvailable = 0
		}

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
