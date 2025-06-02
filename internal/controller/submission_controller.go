// controller/submission_controller.go
package controller

import (
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create Submission
// @Tags Submissions
// @Accept json
// @Produce json
// @Param submission body model.Submission true "Submission JSON"
// @Success 201 {object} model.Submission
// @Router /submissions [post]
func CreateSubmission(c *gin.Context) {
	var submission model.Submission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	submission.ID = uuid.New().String() // Auto-generate ID
	if err := service.CreateSubmission(submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
		return
	}
	c.JSON(http.StatusCreated, submission)
}

// @Summary Get Submission by ID
// @Tags Submissions
// @Produce json
// @Param id path string true "Submission ID"
// @Success 200 {object} model.Submission
// @Router /submissions/{id} [get]
func GetSubmission(c *gin.Context) {
	id := c.Param("id")
	submission, err := service.GetSubmission(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}
	c.JSON(http.StatusOK, submission)
}

// @Summary Delete Submission
// @Tags Submissions
// @Param id path string true "Submission ID"
// @Success 200 {object} map[string]string
// @Router /submissions/{id} [delete]
func DeleteSubmission(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteSubmission(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete submission"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Submission deleted"})
}

// @Summary Get All Submissions
// @Tags Submissions
// @Produce json
// @Success 200 {array} model.Submission
// @Router /submissions [get]
func GetAllSubmissions(c *gin.Context) {
	filters := map[string]string{} // Define appropriate filters if needed
	submissions, err := service.GetAllSubmissions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}
	c.JSON(http.StatusOK, submissions)
}

// @Summary Update Submission
// @Tags Submissions
// @Accept json
// @Produce json
// @Param id path string true "Submission ID"
// @Param submission body model.Submission true "Updated Submission"
// @Success 200 {object} model.Submission
// @Router /submissions/{id} [put]
func UpdateSubmission(c *gin.Context) {
	id := c.Param("id")
	var submission model.Submission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	submission.ID = id
	if err := service.UpdateSubmission(id, submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, submission)
}

// @Summary Patch a submission
// @Tags Submissions
// @Accept json
// @Produce json
// @Param id path string true "Submission ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /submissions/{id} [patch]
func PatchSubmission(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.PatchSubmission(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch submission"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Submission updated"})
}
