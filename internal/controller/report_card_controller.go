package controller

import (
	"log"
	"net/http"

	"lumenslate/internal/repository"

	"github.com/gin-gonic/gin"
)

// ReportCardRequest represents the request body for creating report cards
type ReportCardRequest struct {
	UserID       string `json:"userId" validate:"required"`
	StudentID    int    `json:"studentId" validate:"required"`
	StudentName  string `json:"studentName" validate:"required"`
	AcademicTerm string `json:"academicTerm" validate:"required"`
	// Add other fields as needed
}

// GetAllReportCardsHandler godoc
// @Summary      Get all report cards
// @Description  Retrieves all report cards with optional filtering
// @Tags         report-cards
// @Accept       json
// @Produce      json
// @Param        userId       query     string  false  "Filter by user ID"
// @Param        studentId    query     string  false  "Filter by student ID"
// @Param        academicTerm query     string  false  "Filter by academic term"
// @Param        limit        query     string  false  "Limit number of results (default 10)"
// @Param        offset       query     string  false  "Offset for pagination (default 0)"
// @Success      200          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Router       /api/report-cards [get]
func GetAllReportCardsHandler(c *gin.Context) {
	log.Println("[ReportCard] /api/report-cards GET called")

	// Build filters from query parameters
	filters := make(map[string]string)
	if userId := c.Query("userId"); userId != "" {
		filters["userId"] = userId
	}
	if studentId := c.Query("studentId"); studentId != "" {
		filters["studentId"] = studentId
	}
	if academicTerm := c.Query("academicTerm"); academicTerm != "" {
		filters["academicTerm"] = academicTerm
	}
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	log.Printf("[ReportCard] Filters: %+v", filters)

	reportCards, err := repository.GetAllReportCards(filters)
	if err != nil {
		log.Printf("[ReportCard] Error getting report cards: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[ReportCard] Successfully retrieved %d report cards", len(reportCards))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reportCards,
		"count":   len(reportCards),
	})
}

// GetReportCardByIDHandler godoc
// @Summary      Get report card by ID
// @Description  Retrieves a specific report card by its ID
// @Tags         report-cards
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Report Card ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/report-cards/{id} [get]
func GetReportCardByIDHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[ReportCard] /api/report-cards/%s GET called", id)

	reportCard, err := repository.GetReportCardByID(id)
	if err != nil {
		log.Printf("[ReportCard] Error getting report card by ID: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Report card not found"})
		return
	}

	log.Printf("[ReportCard] Successfully retrieved report card")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reportCard,
	})
}

// DeleteReportCardHandler godoc
// @Summary      Delete report card
// @Description  Deletes a specific report card by its ID
// @Tags         report-cards
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Report Card ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/report-cards/{id} [delete]
func DeleteReportCardHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[ReportCard] /api/report-cards/%s DELETE called", id)

	err := repository.DeleteReportCard(id)
	if err != nil {
		log.Printf("[ReportCard] Error deleting report card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[ReportCard] Successfully deleted report card")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Report card deleted successfully",
	})
}

// UpdateReportCardHandler godoc
// @Summary      Update report card
// @Description  Updates a specific report card by its ID
// @Tags         report-cards
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "Report Card ID"
// @Param        body body      map[string]interface{} true  "Update data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/report-cards/{id} [put]
func UpdateReportCardHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[ReportCard] /api/report-cards/%s PUT called", id)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		log.Printf("[ReportCard] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[ReportCard] Update request: %+v", updates)

	updatedReportCard, err := repository.UpdateReportCard(id, updates)
	if err != nil {
		log.Printf("[ReportCard] Error updating report card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[ReportCard] Successfully updated report card")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedReportCard,
		"message": "Report card updated successfully",
	})
}
