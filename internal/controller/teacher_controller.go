// controller/teacher_controller.go
package controller

import (
	"lumenslate/internal/model"
	"net/http"
	"server/internal/service"

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
	var teacher model.Teacher
	if err := c.ShouldBindJSON(&teacher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	teacher.ID = uuid.New().String() // Auto-generate ID
	if err := service.CreateTeacher(teacher); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create teacher"})
		return
	}
	c.JSON(http.StatusCreated, teacher)
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
// @Success 200 {array} model.Teacher
// @Router /teachers [get]
func GetAllTeachers(c *gin.Context) {
	filters := map[string]string{} // Define filters as needed
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
	var teacher model.Teacher
	if err := c.ShouldBindJSON(&teacher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	teacher.ID = id
	if err := service.UpdateTeacher(id, teacher); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, teacher)
}

// @Summary Patch a teacher
// @Tags Teachers
// @Accept json
// @Produce json
// @Param id path string true "Teacher ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /teachers/{id} [patch]
func PatchTeacher(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.PatchTeacher(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch teacher"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Teacher updated"})
}
