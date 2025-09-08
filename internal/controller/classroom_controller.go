// controller/classroom_controller.go
package controller

import (
	"log"
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
	"lumenslate/internal/serializer"
	"lumenslate/internal/utils"
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

	classroom.ID = uuid.New().String()
	classroom.CreatedAt = time.Now()
	classroom.UpdatedAt = time.Now()
	if classroom.ClassroomCode == "" {
		classroom.ClassroomCode = utils.GenerateRandomCode(12)
	}
	if classroom.IsActive == false {
		classroom.IsActive = true
	}

	// Validate classroom
	if err := utils.Validate.Struct(classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveClassroom(classroom); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create classroom: " + err.Error()})
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
	classroom, err := repo.GetClassroomByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	extended := c.DefaultQuery("extended", "false") == "true"
	if extended {
		ext := serializer.NewClassroomExtended(classroom)
		c.JSON(http.StatusOK, ext)
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
	if err := repo.DeleteClassroom(id); err != nil {
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
// @Param name query string false "Filter by name"
// @Param teacherId query string false "Filter by teacher ID"
// @Param tags query string false "Filter by tags"
// @Param q query string false "Search in name (partial match)"
// @Success 200 {array} model.Classroom
// @Router /classrooms [get]
func GetAllClassrooms(c *gin.Context) {
	filters := map[string]string{
		"limit":     c.DefaultQuery("limit", "10"),
		"offset":    c.DefaultQuery("offset", "0"),
		"name":      c.Query("name"),
		"teacherId": c.Query("teacherId"),
		"tags":      c.Query("tags"),
	}
	if q := c.Query("q"); q != "" {
		filters["q"] = q
	}
	classrooms, err := repo.GetAllClassrooms(filters)
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
	log.Printf("[UpdateClassroom] Received request for ID: %s", id)
	var classroom model.Classroom
	if err := c.ShouldBindJSON(&classroom); err != nil {
		log.Printf("[UpdateClassroom] JSON binding error for ID %s: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[UpdateClassroom] Request body: %+v", classroom)
	classroom.ID = id
	classroom.UpdatedAt = time.Now()

	if err := utils.Validate.Struct(classroom); err != nil {
		log.Printf("[UpdateClassroom] Validation error for ID %s: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[UpdateClassroom] Validation passed for ID: %s", id)
	if err := repo.SaveClassroom(classroom); err != nil {
		log.Printf("[UpdateClassroom] DB save error for ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	log.Printf("[UpdateClassroom] Classroom updated successfully: %s", id)
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

	updates["updatedAt"] = time.Now()

	updated, err := repo.PatchClassroom(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Patch failed"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
