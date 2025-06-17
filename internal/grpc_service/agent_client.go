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
		return nil, err
	}

	// Get the raw agent response
	rawAgentResponse := res.GetAgentResponse()

	// Try to parse the agent response as JSON to see if it contains database operations
	var agentResponse AgentResponse
	finalResponse := rawAgentResponse
	agentName := res.GetAgentName()
	responseMessage := res.GetMessage()

	if err := json.Unmarshal([]byte(rawAgentResponse), &agentResponse); err == nil {
		// Successfully parsed JSON, check for database operations
		if len(agentResponse.QuestionsRequested) > 0 {
			// Handle question generation
			if questionsResult, err := handleQuestionGeneration(agentResponse.QuestionsRequested, userId); err == nil {
				finalResponse = questionsResult
				agentName = "assignment_generator_general"
				responseMessage = "Assignment generated successfully"
			} else {
				log.Printf("Error handling question generation: %v", err)
				// Keep the original response on error
			}
		} else if agentResponse.AssessmentData != nil {
			// Handle assessment data saving
			if assessmentResult, err := handleAssessmentSaving(agentResponse.AssessmentData, userId); err == nil {
				finalResponse = assessmentResult
				agentName = "assessment_agent"
				responseMessage = "Subject Assessment report processed"
			} else {
				log.Printf("Error handling assessment saving: %v", err)
				// Keep the original response on error
			}
		}
	}

	// Return the processed response
	return map[string]interface{}{
		"message":        responseMessage,
		"user_id":        res.GetUserId(),
		"agent_name":     agentName,
		"agent_response": finalResponse,
		"session_id":     res.GetSessionId(),
		"createdAt":      res.GetCreatedAt(),
		"updatedAt":      res.GetUpdatedAt(),
		"response_time":  res.GetResponseTime(),
		"role":           res.GetRole(),
		"feedback":       res.GetFeedback(),
	}, nil
}

func handleQuestionGeneration(questionsRequested []QuestionRequest, userId string) (string, error) {
	// Group requests by subject to handle multiple difficulty levels for the same subject
	subjectRequests := make(map[string][]QuestionRequest)

	for _, request := range questionsRequested {
		if request.Type == "assignment_generator_general" {
			subject := strings.ToLower(strings.TrimSpace(request.Subject))
			subjectRequests[subject] = append(subjectRequests[subject], request)
		}
	}

	responseData := map[string]interface{}{
		"status":                   "success",
		"message":                  "Questions retrieved successfully",
		"total_subjects":           len(subjectRequests),
		"subjects":                 []map[string]interface{}{},
		"total_questions_returned": 0,
	}

	totalQuestionsReturned := 0

	for subjectKey, requests := range subjectRequests {
		// Convert subject string to Subject enum
		subject, validSubject := model.GetSubjectFromString(subjectKey)
		displaySubject := subjectKey

		if !validSubject {
			// Invalid subject
			totalRequested := 0
			for _, req := range requests {
				totalRequested += req.NumberOfQuestions
			}
			subjectData := map[string]interface{}{
				"subject":         displaySubject,
				"requested_count": totalRequested,
				"available_count": 0,
				"returned_count":  0,
				"questions":       []map[string]interface{}{},
				"message":         fmt.Sprintf("Subject '%s' is not supported. Available subjects: Math, Science, English, History, Geography", displaySubject),
			}
			responseData["subjects"] = append(responseData["subjects"].([]map[string]interface{}), subjectData)
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

		// Check if we got all requested questions
		if len(allQuestionsForSubject) < totalRequested {
			// Get total available questions for this subject (any difficulty)
			totalAvailable, err := repository.CountQuestionsBySubject(subject)
			if err != nil {
				totalAvailable = 0
			}

			subjectData := map[string]interface{}{
				"subject":         displaySubject,
				"requested_count": totalRequested,
				"available_count": int(totalAvailable),
				"returned_count":  0,
				"questions":       []map[string]interface{}{},
				"message":         fmt.Sprintf("Could not fulfill all requests for %s. Some difficulty levels may not have enough questions available.", displaySubject),
			}
			responseData["subjects"] = append(responseData["subjects"].([]map[string]interface{}), subjectData)
		} else if len(allQuestionsForSubject) == 0 {
			// No questions available for this subject
			totalAvailable, err := repository.CountQuestionsBySubject(subject)
			if err != nil {
				totalAvailable = 0
			}

			subjectData := map[string]interface{}{
				"subject":         displaySubject,
				"requested_count": totalRequested,
				"available_count": int(totalAvailable),
				"returned_count":  0,
				"questions":       []map[string]interface{}{},
				"message":         fmt.Sprintf("No questions available for %s with the requested difficulty levels", displaySubject),
			}
			responseData["subjects"] = append(responseData["subjects"].([]map[string]interface{}), subjectData)
		} else {
			// Successfully got all requested questions
			// Format questions for response
			formattedQuestions := make([]map[string]interface{}, len(allQuestionsForSubject))
			for i, q := range allQuestionsForSubject {
				formattedQuestions[i] = map[string]interface{}{
					"question_id": q.ID,
					"question":    q.Question,
					"options":     q.Options,
					"answer":      q.Answer,
					"difficulty":  string(q.Difficulty),
				}
			}

			// Get total available questions for this subject
			totalAvailable, err := repository.CountQuestionsBySubject(subject)
			if err != nil {
				totalAvailable = 0
			}

			subjectData := map[string]interface{}{
				"subject":         displaySubject,
				"requested_count": totalRequested,
				"available_count": int(totalAvailable),
				"returned_count":  len(formattedQuestions),
				"questions":       formattedQuestions,
			}

			totalQuestionsReturned += len(formattedQuestions)
			responseData["subjects"] = append(responseData["subjects"].([]map[string]interface{}), subjectData)
		}
	}

	responseData["total_questions_returned"] = totalQuestionsReturned

	// Format as a readable response
	responseMessage := "**Assignment Questions Generated**\n\n"

	for _, subjectInterface := range responseData["subjects"].([]map[string]interface{}) {
		subject := subjectInterface
		if msg, hasMessage := subject["message"]; hasMessage {
			// Handle case where no questions were returned due to error/insufficient questions
			responseMessage += fmt.Sprintf("**%s**: %s\n", subject["subject"], msg)
		} else {
			responseMessage += fmt.Sprintf("**%s** (%v questions):\n", subject["subject"], subject["returned_count"])

			if questions, ok := subject["questions"].([]map[string]interface{}); ok {
				for i, question := range questions {
					responseMessage += fmt.Sprintf("\n%d. %s *(%s)*\n", i+1, question["question"], question["difficulty"])
					if options, ok := question["options"].([]string); ok {
						for j, option := range options {
							responseMessage += fmt.Sprintf("   %c) %s\n", 'a'+j, option)
						}
					}
					responseMessage += fmt.Sprintf("   **Answer:** %s\n", question["answer"])
				}
			}
		}
		responseMessage += "\n" + strings.Repeat("=", 50) + "\n"
	}

	responseMessage += fmt.Sprintf("\n**Total Questions Provided:** %d", totalQuestionsReturned)

	return responseMessage, nil
}

func handleAssessmentSaving(assessmentDataInterface interface{}, userId string) (string, error) {
	// Parse the assessment data
	assessmentBytes, err := json.Marshal(assessmentDataInterface)
	if err != nil {
		return "", fmt.Errorf("failed to marshal assessment data: %v", err)
	}

	var assessmentData AssessmentData
	if err := json.Unmarshal(assessmentBytes, &assessmentData); err != nil {
		return "", fmt.Errorf("failed to unmarshal assessment data: %v", err)
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
		return "", fmt.Errorf("missing mandatory fields: %s", strings.Join(missingFields, ", "))
	}

	// Validate and convert subject to enum
	subject, validSubject := model.GetSubjectFromString(strings.ToLower(strings.TrimSpace(assessmentData.Subject)))
	if !validSubject {
		return "", fmt.Errorf("invalid subject: %s. Available subjects are: math, science, english, history, geography", assessmentData.Subject)
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
			return "", fmt.Errorf("invalid student_id format: %v", v)
		}
	default:
		return "", fmt.Errorf("invalid student_id type: %T", v)
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
			return "", fmt.Errorf("invalid score format: %v", v)
		}
	default:
		return "", fmt.Errorf("invalid score type: %T", v)
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
		return "", fmt.Errorf("failed to save subject report: %v", err)
	}

	// Format success message
	responseMessage := "**Subject Report Created Successfully**\n\n"
	responseMessage += fmt.Sprintf("**Student:** %s (ID: %d)\n", savedReport.StudentName, savedReport.StudentID)
	responseMessage += fmt.Sprintf("**Subject:** %s\n", strings.Title(string(savedReport.Subject)))
	responseMessage += fmt.Sprintf("**Overall Score:** %d/100\n", savedReport.Score)

	if savedReport.GradeLetter != nil {
		responseMessage += fmt.Sprintf("**Grade:** %s\n", *savedReport.GradeLetter)
	}

	if savedReport.ClassName != nil {
		responseMessage += fmt.Sprintf("**Class:** %s\n", *savedReport.ClassName)
	}

	if savedReport.InstructorName != nil {
		responseMessage += fmt.Sprintf("**Instructor:** %s\n", *savedReport.InstructorName)
	}

	if savedReport.Term != nil {
		responseMessage += fmt.Sprintf("**Term:** %s\n", *savedReport.Term)
	}

	// Add assessment breakdown if available
	var assessmentScores []string
	if savedReport.MidtermScore != nil {
		assessmentScores = append(assessmentScores, fmt.Sprintf("Midterm: %d", *savedReport.MidtermScore))
	}
	if savedReport.FinalExamScore != nil {
		assessmentScores = append(assessmentScores, fmt.Sprintf("Final: %d", *savedReport.FinalExamScore))
	}
	if savedReport.QuizScore != nil {
		assessmentScores = append(assessmentScores, fmt.Sprintf("Quiz: %d", *savedReport.QuizScore))
	}
	if savedReport.AssignmentScore != nil {
		assessmentScores = append(assessmentScores, fmt.Sprintf("Assignments: %d", *savedReport.AssignmentScore))
	}

	if len(assessmentScores) > 0 {
		responseMessage += fmt.Sprintf("\n**Assessment Breakdown:** %s\n", strings.Join(assessmentScores, ", "))
	}

	// Add skill evaluations if available
	var skills []string
	if savedReport.ConceptualUnderstanding != nil {
		skills = append(skills, fmt.Sprintf("Conceptual Understanding: %.1f", *savedReport.ConceptualUnderstanding))
	}
	if savedReport.ProblemSolving != nil {
		skills = append(skills, fmt.Sprintf("Problem Solving: %.1f", *savedReport.ProblemSolving))
	}
	if savedReport.AnalyticalThinking != nil {
		skills = append(skills, fmt.Sprintf("Analytical Thinking: %.1f", *savedReport.AnalyticalThinking))
	}

	if len(skills) > 0 {
		responseMessage += fmt.Sprintf("\n**Key Skills:** %s\n", strings.Join(skills, ", "))
	}

	if savedReport.Remarks != nil {
		responseMessage += fmt.Sprintf("\n**Remarks:** %s\n", *savedReport.Remarks)
	}

	if savedReport.AreasForImprovement != nil {
		responseMessage += fmt.Sprintf("\n**Areas for Improvement:** %s\n", *savedReport.AreasForImprovement)
	}

	responseMessage += fmt.Sprintf("\n**Report ID:** %s", savedReport.ID)
	responseMessage += fmt.Sprintf("\n**Created:** %s", savedReport.Timestamp.Format("2006-01-02 15:04:05 UTC"))

	return responseMessage, nil
}
