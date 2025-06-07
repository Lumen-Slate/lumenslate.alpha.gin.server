// controller/teacher_controller.go
package controller

import (
	"lumenslate/internal/common"
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create Teacher
// @Tags Teachers
// @Accept json
// @Produce json
// @Param teacher body model.Teacher true "Teacher JSON"
// @Success 201 {object} model.Teacher
// @Router /teachers [post]
func CreateTeacher(c *gin.Context) {
	t := model.NewTeacher()
	if err := c.ShouldBindJSON(t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID before validation
	t.ID = uuid.New().String()

	// Validate the struct
	if err := common.Validate.Struct(t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.CreateTeacher(*t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create teacher"})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// @Summary Get Teacher by ID
// @Tags Teachers
// @Produce json
// @Param id path string true "Teacher ID"
// @Success 200 {object} model.Teacher
// @Router /teachers/{id} [get]
func GetTeacher(c *gin.Context) {
	id := c.Param("id")
	teacher, err := service.GetTeacher(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
		return
	}
	c.JSON(http.StatusOK, teacher)
}

// @Summary Delete Teacher
// @Tags Teachers
// @Param id path string true "Teacher ID"
// @Success 200 {object} map[string]string
// @Router /teachers/{id} [delete]
func DeleteTeacher(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteTeacher(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete teacher"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Teacher deleted"})
}

// @Summary Get All Teachers
// @Tags Teachers
// @Produce json
// @Param email query string false "Filter by email"
// @Param phone query string false "Filter by phone"
// @Param name query string false "Filter by name"
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {array} model.Teacher
// @Router /teachers [get]
func GetAllTeachers(c *gin.Context) {
	filters := map[string]string{
		"email":  c.Query("email"),
		"phone":  c.Query("phone"),
		"name":   c.Query("name"),
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
	}
	teachers, err := service.GetAllTeachers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teachers"})
		return
	}
	c.JSON(http.StatusOK, teachers)
}

// @Summary Update Teacher
// @Tags Teachers
// @Accept json
// @Produce json
// @Param id path string true "Teacher ID"
// @Param teacher body model.Teacher true "Updated Teacher"
// @Success 200 {object} model.Teacher
// @Router /teachers/{id} [put]
func UpdateTeacher(c *gin.Context) {
	id := c.Param("id")
	var t model.Teacher
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the struct
	if err := common.Validate.Struct(t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t.ID = id
	t.UpdatedAt = time.Now()
	if err := service.UpdateTeacher(id, t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// @Summary Patch Teacher
// @Tags Teachers
// @Accept json
// @Produce json
// @Param id path string true "Teacher ID"
// @Param updates body map[string]interface{} true "Updates"
// @Success 200 {object} model.Teacher
// @Router /teachers/{id} [patch]
func PatchTeacher(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	updated, err := service.PatchTeacher(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch teacher"})
		return
	}
	c.JSON(http.StatusOK, updated)
}
