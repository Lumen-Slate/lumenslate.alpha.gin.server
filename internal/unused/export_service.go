// package service

// import (
// 	"fmt"
// 	qrepo "lumenslate/internal/repository/questions"
// )

// type QuestionForPDF struct {
// 	ID              string
// 	Type            string
// 	Question        string
// 	Options         []string
// 	AnswerIndex     *int
// 	AnswerIndices   []int
// 	Answer          *float64
// 	IdealAnswer     *string
// 	GradingCriteria []string
// }

// type GroupedQuestions struct {
// 	Type      string
// 	Questions []QuestionForPDF
// }

// func GetQuestionsGroupedForPDF(ids []string) ([]GroupedQuestions, error) {
// 	var all []QuestionForPDF

// 	for _, id := range ids {
// 		if q, err := qrepo.GetMCQByID(id); err == nil {
// 			all = append(all, QuestionForPDF{
// 				ID:          q.ID,
// 				Type:        "MCQ",
// 				Question:    q.Question,
// 				Options:     q.Options,
// 				AnswerIndex: &q.AnswerIndex,
// 			})
// 			continue
// 		}
// 		if q, err := qrepo.GetMSQByID(id); err == nil {
// 			all = append(all, QuestionForPDF{
// 				ID:            q.ID,
// 				Type:          "MSQ",
// 				Question:      q.Question,
// 				Options:       q.Options,
// 				AnswerIndices: q.AnswerIndices,
// 			})
// 			continue
// 		}
// 		if q, err := qrepo.GetNATByID(id); err == nil {
// 			all = append(all, QuestionForPDF{
// 				ID:       q.ID,
// 				Type:     "NAT",
// 				Question: q.Question,
// 				Answer:   &q.Answer,
// 			})
// 			continue
// 		}
// 		if q, err := qrepo.GetSubjectiveByID(id); err == nil {
// 			all = append(all, QuestionForPDF{
// 				ID:              q.ID,
// 				Type:            "Subjective",
// 				Question:        q.Question,
// 				IdealAnswer:     q.IdealAnswer,
// 				GradingCriteria: q.GradingCriteria,
// 			})
// 		}
// 	}

// 	if len(all) == 0 {
// 		return nil, fmt.Errorf("no valid questions found")
// 	}

// 	groupMap := make(map[string][]QuestionForPDF)
// 	for _, q := range all {
// 		groupMap[q.Type] = append(groupMap[q.Type], q)
// 	}

// 	var grouped []GroupedQuestions
// 	for _, t := range []string{"MCQ", "MSQ", "NAT", "Subjective"} {
// 		if qlist, ok := groupMap[t]; ok {
// 			grouped = append(grouped, GroupedQuestions{
// 				Type:      t,
// 				Questions: qlist,
// 			})
// 		}
// 	}

// 	return grouped, nil
// }
