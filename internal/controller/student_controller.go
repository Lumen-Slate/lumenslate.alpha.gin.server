// controller/student_controller.go
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

// @Summary Create Student
// @Tags Students
// @Accept json
// @Produce json
// @Param student body model.Student true "Student JSON"
// @Success 201 {object} model.Student
// @Router /students [post]
func CreateStudent(c *gin.Context) {
	var student model.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initialize with default values
	student = *model.NewStudent()
	student.ID = uuid.New().String()

	// Validate the struct
	if err := common.Validate.Struct(student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.CreateStudent(student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student"})
		return
	}
	c.JSON(http.StatusCreated, student)
}

// @Summary Get Student by ID
// @Tags Students
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} model.Student
// @Router /students/{id} [get]
func GetStudent(c *gin.Context) {
	id := c.Param("id")
	student, err := service.GetStudent(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}
	c.JSON(http.StatusOK, student)
}

// @Summary Delete Student
// @Tags Students
// @Param id path string true "Student ID"
// @Success 200 {object} map[string]string
// @Router /students/{id} [delete]
func DeleteStudent(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteStudent(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete student"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Student deleted successfully"})
}

// @Summary Get All Students
// @Tags Students
// @Produce json
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Param email query string false "Filter by email"
// @Param rollNo query string false "Filter by roll number"
// @Success 200 {array} model.Student
// @Router /students [get]
func GetAllStudents(c *gin.Context) {
	filters := map[string]string{
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
		"email":  c.Query("email"),
		"rollNo": c.Query("rollNo"),
	}
	students, err := service.GetAllStudents(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
		return
	}
	c.JSON(http.StatusOK, students)
}

// @Summary Update Student
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param student body model.Student true "Updated Student"
// @Success 200 {object} model.Student
// @Router /students/{id} [put]
func UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	var student model.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student.ID = id
	student.UpdatedAt = time.Now()

	// Validate the struct
	if err := common.Validate.Struct(student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateStudent(id, student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, student)
}

// @Summary Patch a student
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} model.Student
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students/{id} [patch]
func PatchStudent(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// Get the updated student
	updated, err := service.PatchStudent(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch student"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
