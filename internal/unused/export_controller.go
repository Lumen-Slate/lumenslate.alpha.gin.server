// package controller

// import (
// 	"bytes"
// 	"html/template"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"time"

// 	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"

// 	"lumenslate/internal/service"
// )

// type GeneratePDFRequest struct {
// 	QuestionIDs []string `json:"questionIds"`
// 	ShowAnswers bool     `json:"showAnswers"`
// 	ShowType    bool     `json:"showType"`
// }

// // @Summary Generate a PDF of mixed question types and return file path
// // @Tags Export
// // @Accept json
// // @Produce json
// // @Param data body GeneratePDFRequest true "Question IDs and flags"
// // @Success 200 {object} map[string]string
// // @Failure 400 {object} map[string]string
// // @Failure 500 {object} map[string]string
// // @Router /generate-pdf [post]
// func GenerateQuestionPDF(c *gin.Context) {
// 	var req GeneratePDFRequest
// 	if err := c.ShouldBindJSON(&req); err != nil || len(req.QuestionIDs) == 0 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing questionIds"})
// 		return
// 	}

// 	grouped, err := service.GetQuestionsGroupedForPDF(req.QuestionIDs)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	funcMap := template.FuncMap{
// 		"add1": func(i int) int { return i + 1 },
// 		"now":  func() string { return time.Now().Format("02 Jan 2006 15:04") },
// 		"dereference": func(ptr *int) int {
// 			if ptr != nil {
// 				return *ptr
// 			}
// 			return -1
// 		},
// 	}

// 	tmplPath := filepath.Join("templates", "questions.html")
// 	tmpl, err := template.New("questions.html").Funcs(funcMap).ParseFiles(tmplPath)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template parse error: " + err.Error()})
// 		return
// 	}

// 	var htmlBuffer bytes.Buffer
// 	err = tmpl.Execute(&htmlBuffer, gin.H{
// 		"Grouped":     grouped,
// 		"ShowAnswers": req.ShowAnswers,
// 		"ShowType":    req.ShowType,
// 	})
// 	if err != nil {
// 		log.Println("Template execution error:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template rendering failed: " + err.Error()})
// 		return
// 	}

// 	pdfg, err := wkhtmltopdf.NewPDFGenerator()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generator init failed"})
// 		return
// 	}

// 	pdfg.AddPage(wkhtmltopdf.NewPageReader(&htmlBuffer))
// 	if err := pdfg.Create(); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF creation failed: " + err.Error()})
// 		return
// 	}

// 	// ✅ Save the PDF to media/ folder
// 	outputDir := "media/assignments"
// 	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
// 		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create media folder"})
// 			return
// 		}
// 	}

// 	fileName := "questions_" + uuid.New().String() + ".pdf"
// 	filePath := filepath.Join(outputDir, fileName)

// 	err = os.WriteFile(filePath, pdfg.Bytes(), 0644)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PDF file: " + err.Error()})
// 		return
// 	}

// 	// ✅ Return the file path instead of streaming
// 	c.JSON(http.StatusOK, gin.H{
// 		"message":  "PDF generated successfully",
// 		"filePath": "/" + filePath, // or just `fileName` if used with /media static server
// 	})
// }
