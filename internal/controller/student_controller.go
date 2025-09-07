// controller/student_controller.go
package controller

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
	"lumenslate/internal/utils"
	"net/http"
	"strconv"
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
	// Create new Student with default values
	student := *model.NewStudent()

	// Bind JSON to the struct
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID
	student.ID = uuid.New().String()

	// Validate the struct
	if err := utils.Validate.Struct(student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveStudent(student); err != nil {
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
	student, err := repo.GetStudentByID(id)
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
	if err := repo.DeleteStudent(id); err != nil {
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
// @Param classIds query string false "Filter by class IDs (comma-separated)"
// @Param q query string false "Search in name or email (partial match, name gets priority)"
// @Success 200 {array} model.Student
// @Router /students [get]
func GetAllStudents(c *gin.Context) {
	filters := map[string]string{
		"limit":    c.DefaultQuery("limit", "10"),
		"offset":   c.DefaultQuery("offset", "0"),
		"email":    c.Query("email"),
		"rollNo":   c.Query("rollNo"),
		"classIds": c.Query("classIds"),
	}
	if q := c.Query("q"); q != "" {
		filters["q"] = q
	}
	students, err := repo.GetAllStudents(filters)
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
	if err := utils.Validate.Struct(student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveStudent(student); err != nil {
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
	updated, err := repo.PatchStudent(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch student"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// GetStudentClassrooms fetches student by ID and then all classrooms for their classIds
func GetStudentClassrooms(c *gin.Context) {
	studentID := c.Param("id")
	logger := utils.NewLogger("student_controller")
	ctx := c.Request.Context()
	logger.Info(ctx, "GetStudentClassrooms request for studentID="+studentID)

	student, err := repo.GetStudentByID(studentID)
	if err != nil || student == nil {
		logger.Error(ctx, "Student not found", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}
	logger.Info(ctx, "Found student: "+student.ID+", classIDs="+strconv.Itoa(len(student.ClassIDs)))

	// Pagination params
	limit := 10
	offset := 0
	if l := c.DefaultQuery("limit", "10"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		} else {
			logger.Error(ctx, "Invalid limit param: "+l, err)
		}
	}
	if o := c.DefaultQuery("offset", "0"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		} else {
			logger.Error(ctx, "Invalid offset param: "+o, err)
		}
	}
	logger.Info(ctx, "Pagination limit="+strconv.Itoa(limit)+" offset="+strconv.Itoa(offset))

	total := len(student.ClassIDs)
	pagedIDs := student.ClassIDs
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		pagedIDs = student.ClassIDs[offset:end]
	} else {
		pagedIDs = []string{}
	}
	logger.Info(ctx, "Fetching classrooms for pagedIDs count="+strconv.Itoa(len(pagedIDs)))

	var classrooms []model.Classroom
	for _, classID := range pagedIDs {
		classroom, err := repo.GetClassroomByID(classID)
		if err != nil {
			logger.Error(ctx, "Error fetching classroomID="+classID, err)
			continue
		}
		if classroom != nil {
			classrooms = append(classrooms, *classroom)
			logger.Info(ctx, "Added classroomID="+classID)
		} else {
			logger.Error(ctx, "Classroom not found for ID="+classID, nil)
		}
	}

	if len(classrooms) == 0 {
		logger.Error(ctx, "No classrooms found for studentID="+studentID, nil)
		c.JSON(http.StatusNotFound, gin.H{"error": "No classrooms found for this student"})
		return
	}
	logger.Info(ctx, "Returning "+strconv.Itoa(len(classrooms))+" classrooms for studentID="+studentID)
	c.JSON(http.StatusOK, gin.H{
		"classrooms": classrooms,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"returned": len(classrooms),
		},
	})
}

// JoinClassroomByCode allows a student to join a classroom using classroom code
func JoinClassroomByCode(c *gin.Context) {
	var req struct {
		StudentID     string `json:"studentId" binding:"required"`
		ClassroomCode string `json:"classroomCode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student, err := repo.GetStudentByID(req.StudentID)
	if err != nil || student == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Find classroom by code
	classroom, err := repo.GetClassroomByCode(req.ClassroomCode)
	if err != nil || classroom == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	// Add classroom ID to student's ClassIDs if not already present
	alreadyJoined := false
	for _, cid := range student.ClassIDs {
		if cid == classroom.ID {
			alreadyJoined = true
			break
		}
	}
	if !alreadyJoined {
		student.ClassIDs = append(student.ClassIDs, classroom.ID)
		student.UpdatedAt = time.Now()
		if err := repo.SaveStudent(*student); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join classroom"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined classroom successfully", "student": student, "classroom": classroom})
}
