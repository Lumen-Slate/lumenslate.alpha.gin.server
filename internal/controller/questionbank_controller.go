// controller/question_bank_controller.go
package controller

import (
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create QuestionBank
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Param questionBank body model.QuestionBank true "QuestionBank JSON"
// @Success 201 {object} model.QuestionBank
// @Router /question-banks [post]
func CreateQuestionBank(c *gin.Context) {
	var q model.QuestionBank
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q.ID = uuid.New().String() // Auto-generate ID
	if err := service.CreateQuestionBank(q); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question bank"})
		return
	}
	c.JSON(http.StatusCreated, q)
}

// @Summary Get QuestionBank by ID
// @Tags QuestionBanks
// @Produce json
// @Param id path string true "QuestionBank ID"
// @Success 200 {object} model.QuestionBank
// @Router /question-banks/{id} [get]
func GetQuestionBank(c *gin.Context) {
	id := c.Param("id")
	q, err := service.GetQuestionBank(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, q)
}

// @Summary Delete QuestionBank
// @Tags QuestionBanks
// @Param id path string true "QuestionBank ID"
// @Success 200 {object} map[string]string
// @Router /question-banks/{id} [delete]
func DeleteQuestionBank(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteQuestionBank(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// @Summary Get All QuestionBanks
// @Tags QuestionBanks
// @Produce json
// @Param topic query string false "Filter by topic"
// @Param name query string false "Filter by name"
// @Param teacherId query string false "Filter by teacher ID"
// @Param tags query string false "Tags (comma-separated)"
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {array} model.QuestionBank
// @Router /question-banks [get]
func GetAllQuestionBanks(c *gin.Context) {
	filters := map[string]string{
		"topic":     c.Query("topic"),
		"name":      c.Query("name"),
		"teacherId": c.Query("teacherId"),
		"tags":      c.Query("tags"), // Added tags filter
		"limit":     c.DefaultQuery("limit", "10"),
		"offset":    c.DefaultQuery("offset", "0"),
	}
	items, err := service.GetAllQuestionBanks(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// @Summary Update QuestionBank
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Param id path string true "QuestionBank ID"
// @Param questionBank body model.QuestionBank true "Updated QuestionBank"
// @Success 200 {object} model.QuestionBank
// @Router /question-banks/{id} [put]
func UpdateQuestionBank(c *gin.Context) {
	id := c.Param("id")
	var q model.QuestionBank
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q.ID = id
	if err := service.UpdateQuestionBank(id, q); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, q)
}

// @Summary Patch QuestionBank
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Param id path string true "QuestionBank ID"
// @Param updates body map[string]interface{} true "Updates"
// @Success 200 {object} map[string]string
// @Router /question-banks/{id} [patch]
func PatchQuestionBank(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.PatchQuestionBank(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Patch failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}
