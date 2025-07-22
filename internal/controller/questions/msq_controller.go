// controller/questions/msq_controller.go
package questions

import (
	"lumenslate/internal/model/questions"
	repo "lumenslate/internal/repository/questions"
	"lumenslate/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create MSQ
// @Tags MSQs
// @Accept json
// @Produce json
// @Param msq body questions.MSQ true "MSQ JSON"
// @Success 201 {object} questions.MSQ
// @Router /msqs [post]
func CreateMSQ(c *gin.Context) {
	m := questions.NewMSQ()
	if err := c.ShouldBindJSON(m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID before validation
	m.ID = uuid.New().String()

	// Validate the struct
	if err := utils.Validate.Struct(m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveMSQ(*m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MSQ"})
		return
	}
	c.JSON(http.StatusCreated, m)
}

// @Summary Get MSQ by ID
// @Tags MSQs
// @Produce json
// @Param id path string true "MSQ ID"
// @Success 200 {object} questions.MSQ
// @Router /msqs/{id} [get]
func GetMSQ(c *gin.Context) {
	id := c.Param("id")
	m, err := repo.GetMSQByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "MSQ not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// @Summary Delete MSQ
// @Tags MSQs
// @Param id path string true "MSQ ID"
// @Success 200 {object} map[string]string
// @Router /msqs/{id} [delete]
func DeleteMSQ(c *gin.Context) {
	id := c.Param("id")
	if err := repo.DeleteMSQ(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete MSQ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MSQ deleted"})
}

// @Summary Get All MSQs
// @Tags MSQs
// @Produce json
// @Param bankId query string false "Filter by bank ID"
// @Param limit query string false "Pagination limit"
// @Param offset query string false "Pagination offset"
// @Success 200 {array} questions.MSQ
// @Router /msqs [get]
func GetAllMSQs(c *gin.Context) {
	filters := map[string]string{
		"bankId": c.Query("bankId"),
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
	}
	msqs, err := repo.GetAllMSQs(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MSQs"})
		return
	}
	c.JSON(http.StatusOK, msqs)
}

// @Summary Update MSQ
// @Tags MSQs
// @Accept json
// @Produce json
// @Param id path string true "MSQ ID"
// @Param msq body questions.MSQ true "Updated MSQ"
// @Success 200 {object} questions.MSQ
// @Router /msqs/{id} [put]
func UpdateMSQ(c *gin.Context) {
	id := c.Param("id")
	var m questions.MSQ
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the struct
	if err := utils.Validate.Struct(m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m.ID = id
	m.UpdatedAt = time.Now()
	if err := repo.SaveMSQ(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update MSQ"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// @Summary Patch MSQ
// @Tags MSQs
// @Accept json
// @Produce json
// @Param id path string true "MSQ ID"
// @Param updates body map[string]interface{} true "Updates"
// @Success 200 {object} questions.MSQ
// @Router /msqs/{id} [patch]
func PatchMSQ(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	updated, err := repo.PatchMSQ(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch MSQ"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// @Summary Bulk Create MSQs
// @Tags MSQs
// @Accept json
// @Produce json
// @Param msqs body []questions.MSQ true "List of MSQs"
// @Success 201 {array} questions.MSQ
// @Router /msqs/bulk [post]
func CreateBulkMSQs(c *gin.Context) {
	var msqs []questions.MSQ
	if err := c.ShouldBindJSON(&msqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	for i := range msqs {
		msqs[i].ID = uuid.New().String()
		msqs[i].CreatedAt = now
		msqs[i].UpdatedAt = now
		msqs[i].IsActive = true

		// Validate each MSQ
		if err := utils.Validate.Struct(msqs[i]); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if err := repo.SaveBulkMSQs(msqs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MSQs"})
		return
	}

	c.JSON(http.StatusCreated, msqs)
}
