// controller/classroom_controller.go
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

// @Summary Create Classroom
// @Tags Classrooms
// @Accept json
// @Produce json
// @Param classroom body model.Classroom true "Classroom JSON"
// @Success 201 {object} model.Classroom
// @Router /classrooms [post]
func CreateClassroom(c *gin.Context) {
	var classroom model.Classroom
	if err := c.ShouldBindJSON(&classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initialize with default values
	classroom = *model.NewClassroom()
	classroom.ID = uuid.New().String()

	// Validate the struct
	if err := common.Validate.Struct(classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.CreateClassroom(classroom); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create classroom"})
		return
	}
	c.JSON(http.StatusCreated, classroom)
}

// @Summary Get Classroom by ID
// @Tags Classrooms
// @Produce json
// @Param id path string true "Classroom ID"
// @Success 200 {object} model.Classroom
// @Router /classrooms/{id} [get]
func GetClassroom(c *gin.Context) {
	id := c.Param("id")
	classroom, err := service.GetClassroom(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}
	c.JSON(http.StatusOK, classroom)
}

// @Summary Delete Classroom
// @Tags Classrooms
// @Param id path string true "Classroom ID"
// @Success 200 {object} map[string]string
// @Router /classrooms/{id} [delete]
func DeleteClassroom(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteClassroom(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete classroom"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Classroom deleted successfully"})
}

// @Summary Get All Classrooms
// @Tags Classrooms
// @Produce json
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {array} model.Classroom
// @Router /classrooms [get]
func GetAllClassrooms(c *gin.Context) {
	filters := map[string]string{
		"limit":     c.DefaultQuery("limit", "10"),
		"offset":    c.DefaultQuery("offset", "0"),
		"subject":   c.Query("subject"),
		"teacherId": c.Query("teacherId"),
		"tags":      c.Query("tags"),
	}
	classrooms, err := service.GetAllClassrooms(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classrooms"})
		return
	}
	c.JSON(http.StatusOK, classrooms)
}

// @Summary Update Classroom
// @Tags Classrooms
// @Accept json
// @Produce json
// @Param id path string true "Classroom ID"
// @Param classroom body model.Classroom true "Updated Classroom"
// @Success 200 {object} model.Classroom
// @Router /classrooms/{id} [put]
func UpdateClassroom(c *gin.Context) {
	id := c.Param("id")
	var classroom model.Classroom
	if err := c.ShouldBindJSON(&classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	classroom.ID = id
	classroom.UpdatedAt = time.Now()

	// Validate the struct
	if err := common.Validate.Struct(classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateClassroom(id, classroom); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, classroom)
}

// @Summary Patch Classroom
// @Tags Classrooms
// @Accept json
// @Produce json
// @Param id path string true "Classroom ID"
// @Param updates body map[string]interface{} true "Partial updates"
// @Success 200 {object} model.Classroom
// @Router /classrooms/{id} [patch]
func PatchClassroom(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// Get the updated classroom
	updated, err := service.PatchClassroom(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Patch failed"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
