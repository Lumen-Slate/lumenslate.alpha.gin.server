// controller/classroom_controller.go
package controller

import (
	"lumenslate/internal/common"
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
	"lumenslate/internal/serializer"
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
type ClassroomCreateRequest struct {
	Subject       string             `json:"subject" binding:"required"`
	TeacherIDs    []string           `json:"teacherIds"`
	Teachers      []model.Teacher    `json:"teachers"`
	AssignmentIDs []string           `json:"assignmentIds"`
	Assignments   []model.Assignment `json:"assignments"`
	Credits       int                `json:"credits" binding:"required,min=0"`
	Tags          []string           `json:"tags"`
}

func CreateClassroom(c *gin.Context) {
	var req ClassroomCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	classroom := *model.NewClassroom()
	classroom.ID = uuid.New().String()
	classroom.Subject = req.Subject
	classroom.Credits = req.Credits
	classroom.Tags = req.Tags

	// Validate and process teachers
	for _, t := range req.Teachers {
		if t.ID == "" {
			t.ID = uuid.New().String()
		}
		if err := common.Validate.Struct(t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid teacher: " + err.Error()})
			return
		}
		if err := repo.SaveTeacher(t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save teacher: " + err.Error()})
			return
		}
		classroom.TeacherIDs = append(classroom.TeacherIDs, t.ID)
	}
	classroom.TeacherIDs = append(classroom.TeacherIDs, req.TeacherIDs...)

	// Validate and process assignments
	for _, a := range req.Assignments {
		if a.ID == "" {
			a.ID = uuid.New().String()
		}
		if err := common.Validate.Struct(a); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment: " + err.Error()})
			return
		}
		if err := repo.SaveAssignment(a); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save assignment: " + err.Error()})
			return
		}
		classroom.AssignmentIDs = append(classroom.AssignmentIDs, a.ID)
	}
	classroom.AssignmentIDs = append(classroom.AssignmentIDs, req.AssignmentIDs...)

	// Debugging log (optional)
	// log.Printf("TeacherIDs: %+v", classroom.TeacherIDs)
	// log.Printf("AssignmentIDs: %+v", classroom.AssignmentIDs)

	// Validate classroom
	if err := common.Validate.Struct(classroom); err != nil {
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
// @Param subject query string false "Filter by subject"
// @Param teacherId query string false "Filter by teacher ID"
// @Param tags query string false "Filter by tags"
// @Param q query string false "Search in subject (partial match)"
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
	var classroom model.Classroom
	if err := c.ShouldBindJSON(&classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	classroom.ID = id
	classroom.UpdatedAt = time.Now()

	if err := common.Validate.Struct(classroom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveClassroom(classroom); err != nil {
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

	updates["updatedAt"] = time.Now()

	updated, err := repo.PatchClassroom(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Patch failed"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
