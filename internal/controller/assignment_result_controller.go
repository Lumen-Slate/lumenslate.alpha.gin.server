package controller

import (
	"log"
	"net/http"

	"lumenslate/internal/model"
	"lumenslate/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetAllAssignmentResultsHandler godoc
// @Summary      Get all assignment results
// @Description  Retrieves all assignment results with optional filtering
// @Tags         assignment-results
// @Accept       json
// @Produce      json
// @Param        studentId      query     string  false  "Filter by student ID"
// @Param        assignmentId   query     string  false  "Filter by assignment ID"
// @Param        limit          query     string  false  "Limit number of results (default 10)"
// @Param        offset         query     string  false  "Offset for pagination (default 0)"
// @Success      200            {object}  map[string]interface{}
// @Failure      500            {object}  map[string]interface{}
// @Router       /api/assignment-results [get]
func GetAllAssignmentResultsHandler(c *gin.Context) {
	log.Println("[AssignmentResult] /api/assignment-results GET called")

	// Build filters from query parameters
	filters := make(map[string]string)
	if studentId := c.Query("studentId"); studentId != "" {
		filters["studentId"] = studentId
	}
	if assignmentId := c.Query("assignmentId"); assignmentId != "" {
		filters["assignmentId"] = assignmentId
	}
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	log.Printf("[AssignmentResult] Filters: %+v", filters)

	results, err := repository.GetAllAssignmentResults(filters)
	if err != nil {
		log.Printf("[AssignmentResult] Error getting assignment results: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AssignmentResult] Successfully retrieved %d assignment results", len(results))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

// GetAssignmentResultByIDHandler godoc
// @Summary      Get assignment result by ID
// @Description  Retrieves a specific assignment result by its ID
// @Tags         assignment-results
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Assignment Result ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/assignment-results/{id} [get]
func GetAssignmentResultByIDHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[AssignmentResult] /api/assignment-results/%s GET called", id)

	result, err := repository.GetAssignmentResultByID(id)
	if err != nil {
		log.Printf("[AssignmentResult] Error getting assignment result by ID: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment result not found"})
		return
	}

	log.Printf("[AssignmentResult] Successfully retrieved assignment result")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// CreateAssignmentResultHandler godoc
// @Summary      Create assignment result
// @Description  Creates a new assignment result
// @Tags         assignment-results
// @Accept       json
// @Produce      json
// @Param        body body      model.AssignmentResult  true  "Assignment result data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/assignment-results [post]
func CreateAssignmentResultHandler(c *gin.Context) {
	log.Println("[AssignmentResult] /api/assignment-results POST called")

	var result model.AssignmentResult
	if err := c.ShouldBindJSON(&result); err != nil {
		log.Printf("[AssignmentResult] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AssignmentResult] Create request: %+v", result)

	createdResult, err := repository.CreateAssignmentResult(result)
	if err != nil {
		log.Printf("[AssignmentResult] Error creating assignment result: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AssignmentResult] Successfully created assignment result")
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    createdResult,
		"message": "Assignment result created successfully",
	})
}

// UpdateAssignmentResultHandler godoc
// @Summary      Update assignment result
// @Description  Updates a specific assignment result by its ID
// @Tags         assignment-results
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "Assignment Result ID"
// @Param        body body      map[string]interface{} true  "Update data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/assignment-results/{id} [put]
func UpdateAssignmentResultHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[AssignmentResult] /api/assignment-results/%s PUT called", id)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		log.Printf("[AssignmentResult] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AssignmentResult] Update request: %+v", updates)

	updatedResult, err := repository.UpdateAssignmentResult(id, updates)
	if err != nil {
		log.Printf("[AssignmentResult] Error updating assignment result: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AssignmentResult] Successfully updated assignment result")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedResult,
		"message": "Assignment result updated successfully",
	})
}

// DeleteAssignmentResultHandler godoc
// @Summary      Delete assignment result
// @Description  Deletes a specific assignment result by its ID
// @Tags         assignment-results
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Assignment Result ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/assignment-results/{id} [delete]
func DeleteAssignmentResultHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[AssignmentResult] /api/assignment-results/%s DELETE called", id)

	err := repository.DeleteAssignmentResult(id)
	if err != nil {
		log.Printf("[AssignmentResult] Error deleting assignment result: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[AssignmentResult] Successfully deleted assignment result")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Assignment result deleted successfully",
	})
}
