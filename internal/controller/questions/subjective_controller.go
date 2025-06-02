package questions

import (
	questionsModel "lumenslate/internal/model/questions"
	questionsService "lumenslate/internal/service/questions"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create Subjective
// @Tags Subjectives
// @Accept json
// @Produce json
// @Param data body questions.Subjective true "Subjective Question"
// @Success 201 {object} questions.Subjective
// @Router /subjectives [post]
func CreateSubjective(c *gin.Context) {
	var s questionsModel.Subjective
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.ID = uuid.New().String()

	if err := questionsService.CreateSubjective(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Subjective"})
		return
	}

	c.JSON(http.StatusCreated, s)
}

// @Summary Get Subjective by ID
// @Tags Subjectives
// @Produce json
// @Param id path string true "Subjective ID"
// @Success 200 {object} questions.Subjective
// @Router /subjectives/{id} [get]
func GetSubjective(c *gin.Context) {
	id := c.Param("id")
	s, err := questionsService.GetSubjective(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// @Summary Delete Subjective
// @Tags Subjectives
// @Param id path string true "Subjective ID"
// @Success 200 {object} map[string]string
// @Router /subjectives/{id} [delete]
func DeleteSubjective(c *gin.Context) {
	id := c.Param("id")
	if err := questionsService.DeleteSubjective(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// @Summary Get all Subjectives
// @Tags Subjectives
// @Param bankId query string false "Bank ID"
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {array} questions.Subjective
// @Router /subjectives [get]
func GetAllSubjectives(c *gin.Context) {
	filters := map[string]string{
		"bankId": c.Query("bankId"),
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
	}
	items, err := questionsService.GetAllSubjectives(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// @Summary Update Subjective
// @Tags Subjectives
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param data body questions.Subjective true "Subjective"
// @Success 200 {object} questions.Subjective
// @Router /subjectives/{id} [put]
func UpdateSubjective(c *gin.Context) {
	id := c.Param("id")
	var s questionsModel.Subjective
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := questionsService.UpdateSubjective(id, s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// @Summary Patch Subjective
// @Tags Subjectives
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param updates body map[string]interface{} true "Updates"
// @Success 200 {object} map[string]string
// @Router /subjectives/{id} [patch]
func PatchSubjective(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := questionsService.PatchSubjective(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Patch failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

// @Summary Bulk Create Subjectives
// @Tags Subjectives
// @Accept json
// @Produce json
// @Param data body []questions.Subjective true "List of Subjective Questions"
// @Success 201 {array} questions.Subjective
// @Router /subjectives/bulk [post]
func CreateBulkSubjectives(c *gin.Context) {
	var subjectives []questionsModel.Subjective
	if err := c.ShouldBindJSON(&subjectives); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i := range subjectives {
		subjectives[i].ID = uuid.New().String()
	}

	if err := questionsService.CreateBulkSubjectives(subjectives); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subjectives"})
		return
	}

	c.JSON(http.StatusCreated, subjectives)
}
