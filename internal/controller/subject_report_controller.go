package controller

import (
	"log"
	"net/http"
	"strconv"

	"lumenslate/internal/repository"

	"github.com/gin-gonic/gin"
)

// SubjectReportRequest represents the request body for creating subject reports
type SubjectReportRequest struct {
	UserID      string `json:"userId" validate:"required"`
	StudentID   int    `json:"studentId" validate:"required"`
	StudentName string `json:"studentName" validate:"required"`
	Subject     string `json:"subject" validate:"required"`
	Score       int    `json:"score" validate:"required"`
	// Add other fields as needed
}

// GetAllSubjectReportsHandler godoc
// @Summary      Get all subject reports
// @Description  Retrieves all subject reports with optional filtering
// @Tags         subject-reports
// @Accept       json
// @Produce      json
// @Param        userId    query     string  false  "Filter by user ID"
// @Param        studentId query     string  false  "Filter by student ID"
// @Param        subject   query     string  false  "Filter by subject"
// @Param        limit     query     string  false  "Limit number of results (default 10)"
// @Param        offset    query     string  false  "Offset for pagination (default 0)"
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /api/subject-reports [get]
func GetAllSubjectReportsHandler(c *gin.Context) {
	// Build filters from query parameters
	filters := make(map[string]string)
	if userId := c.Query("userId"); userId != "" {
		filters["userId"] = userId
	}
	if studentId := c.Query("studentId"); studentId != "" {
		filters["studentId"] = studentId
	}
	if subject := c.Query("subject"); subject != "" {
		filters["subject"] = subject
	}
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	reports, err := repository.GetAllSubjectReports(filters)
	if err != nil {
		log.Printf("[SubjectReport] Error getting subject reports: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reports,
		"count":   len(reports),
	})
}

// GetSubjectReportByIDHandler godoc
// @Summary      Get subject report by ID
// @Description  Retrieves a specific subject report by its ID
// @Tags         subject-reports
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subject Report ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/subject-reports/{id} [get]
func GetSubjectReportByIDHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[SubjectReport] /api/subject-reports/%s GET called", id)

	report, err := repository.GetSubjectReportByID(id)
	if err != nil {
		log.Printf("[SubjectReport] Error getting subject report by ID: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Subject report not found"})
		return
	}

	log.Printf("[SubjectReport] Successfully retrieved subject report")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// DeleteSubjectReportHandler godoc
// @Summary      Delete subject report
// @Description  Deletes a specific subject report by its ID
// @Tags         subject-reports
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subject Report ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/subject-reports/{id} [delete]
func DeleteSubjectReportHandler(c *gin.Context) {
	id := c.Param("id")

	err := repository.DeleteSubjectReport(id)
	if err != nil {
		log.Printf("[SubjectReport] Error deleting subject report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subject report deleted successfully",
	})
}

// UpdateSubjectReportHandler godoc
// @Summary      Update subject report
// @Description  Updates a specific subject report by its ID
// @Tags         subject-reports
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "Subject Report ID"
// @Param        body body      map[string]interface{} true  "Update data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/subject-reports/{id} [put]
func UpdateSubjectReportHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[SubjectReport] /api/subject-reports/%s PUT called", id)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		log.Printf("[SubjectReport] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[SubjectReport] Update request: %+v", updates)

	updatedReport, err := repository.UpdateSubjectReport(id, updates)
	if err != nil {
		log.Printf("[SubjectReport] Error updating subject report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[SubjectReport] Successfully updated subject report")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedReport,
		"message": "Subject report updated successfully",
	})
}

// GetSubjectReportsByStudentIDHandler godoc
// @Summary      Get subject reports by student ID
// @Description  Retrieves all subject reports for a specific student
// @Tags         subject-reports
// @Accept       json
// @Produce      json
// @Param        studentId path     string  true  "Student ID"
// @Success      200       {object} map[string]interface{}
// @Failure      400       {object} map[string]interface{}
// @Failure      500       {object} map[string]interface{}
// @Router       /api/students/{studentId}/subject-reports [get]
func GetSubjectReportsByStudentIDHandler(c *gin.Context) {
	studentIdStr := c.Param("studentId")
	log.Printf("[SubjectReport] /api/students/%s/subject-reports GET called", studentIdStr)

	studentId, err := strconv.Atoi(studentIdStr)
	if err != nil {
		log.Printf("[SubjectReport] Invalid student ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	reports, err := repository.GetSubjectReportsByStudentID(studentId)
	if err != nil {
		log.Printf("[SubjectReport] Error getting subject reports by student ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[SubjectReport] Successfully retrieved %d subject reports for student %d", len(reports), studentId)
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       reports,
		"count":      len(reports),
		"student_id": studentId,
	})
}
