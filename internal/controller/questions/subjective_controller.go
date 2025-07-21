package questions

import (
	model "lumenslate/internal/model/questions"
	repo "lumenslate/internal/repository/questions"
	"lumenslate/internal/utils"
	"net/http"
	"time"

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
	// Create new Subjective with default values
	s := *model.NewSubjective()

	// Bind JSON to the struct
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID
	s.ID = uuid.New().String()

	// Validate the struct
	if err := utils.Validate.Struct(s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveSubjective(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	s, err := repo.GetSubjectiveByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subjective not found"})
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
	if err := repo.DeleteSubjective(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Subjective deleted successfully"})
}

// @Summary Get all Subjectives
// @Tags Subjectives
// @Param bankId query string false "Bank ID"
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {array} questions.Subjective
// @Router /subjectives [get]
func GetAllSubjectives(c *gin.Context) {
	filters := make(map[string]string)
	if bankID := c.Query("bankId"); bankID != "" {
		filters["bankId"] = bankID
	}
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	subjectives, err := repo.GetAllSubjectives(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subjectives)
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
	var s model.Subjective
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.ID = id
	s.UpdatedAt = time.Now()

	// Validate the struct
	if err := utils.Validate.Struct(s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveSubjective(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	updated, err := repo.PatchSubjective(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary Bulk Create Subjectives
// @Tags Subjectives
// @Accept json
// @Produce json
// @Param data body []questions.Subjective true "List of Subjective Questions"
// @Success 201 {array} questions.Subjective
// @Router /subjectives/bulk [post]
func CreateBulkSubjectives(c *gin.Context) {
	var subjectives []model.Subjective
	if err := c.ShouldBindJSON(&subjectives); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate each Subjective and set timestamps
	for i := range subjectives {
		subjectives[i].ID = uuid.New().String()
		subjectives[i].CreatedAt = time.Now()
		subjectives[i].UpdatedAt = time.Now()
		subjectives[i].IsActive = true

		if err := utils.Validate.Struct(subjectives[i]); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if err := repo.SaveBulkSubjectives(subjectives); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subjectives)
}
