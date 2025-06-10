// controller/assignment_controller.go
package controller

import (
	"lumenslate/internal/common"
	"lumenslate/internal/model"
	"lumenslate/internal/model/questions"
	repo "lumenslate/internal/repository"
	quest "lumenslate/internal/repository/questions"
	"lumenslate/internal/serializer"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create Assignment
// @Tags Assignments
// @Accept json
// @Produce json
// @Param assignment body model.Assignment true "Assignment JSON"
// @Success 201 {object} model.Assignment
// @Router /assignments [post]
func CreateAssignment(c *gin.Context) {
	var req struct {
		Title         string                 `json:"title" binding:"required"`
		Body          string                 `json:"body" binding:"required"`
		DueDate       time.Time              `json:"dueDate" binding:"required"`
		Points        int                    `json:"points" binding:"required,min=0"`
		MCQs          []questions.MCQ        `json:"mcqs"`
		MSQs          []questions.MSQ        `json:"msqs"`
		NATs          []questions.NAT        `json:"nats"`
		Subjectives   []questions.Subjective `json:"subjectives"`
		Comments      []model.Comment        `json:"comments"`
		MCQIds        []string               `json:"mcqIds"`
		MSQIds        []string               `json:"msqIds"`
		NATIds        []string               `json:"natIds"`
		SubjectiveIds []string               `json:"subjectiveIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	assignment := model.Assignment{
		ID:            uuid.New().String(),
		Title:         req.Title,
		Body:          req.Body,
		DueDate:       req.DueDate,
		Points:        req.Points,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsActive:      true,
		CommentIds:    []string{},
		MCQIds:        []string{},
		MSQIds:        []string{},
		NATIds:        []string{},
		SubjectiveIds: []string{},
	}

	// Save MCQs
	for _, q := range req.MCQs {
		q.ID = uuid.New().String()
		if err := common.Validate.Struct(q); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MCQ: " + err.Error()})
			return
		}
		if err := quest.SaveMCQ(q); err == nil {
			assignment.MCQIds = append(assignment.MCQIds, q.ID)
		}
	}

	// Save MSQs
	for _, q := range req.MSQs {
		q.ID = uuid.New().String()
		if err := common.Validate.Struct(q); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MSQ: " + err.Error()})
			return
		}
		if err := quest.SaveMSQ(q); err == nil {
			assignment.MSQIds = append(assignment.MSQIds, q.ID)
		}
	}

	// Save NATs
	for _, q := range req.NATs {
		q.ID = uuid.New().String()
		if err := common.Validate.Struct(q); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid NAT: " + err.Error()})
			return
		}
		if err := quest.SaveNAT(q); err == nil {
			assignment.NATIds = append(assignment.NATIds, q.ID)
		}
	}

	// Save Subjectives
	for _, q := range req.Subjectives {
		q.ID = uuid.New().String()
		if err := common.Validate.Struct(q); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Subjective: " + err.Error()})
			return
		}
		if err := quest.SaveSubjective(q); err == nil {
			assignment.SubjectiveIds = append(assignment.SubjectiveIds, q.ID)
		}
	}

	// Save Comments
	for _, cm := range req.Comments {
		cm.ID = uuid.New().String()
		if err := repo.SaveComment(cm); err == nil {
			assignment.CommentIds = append(assignment.CommentIds, cm.ID)
		}
	}

	// Append question IDs from request
	assignment.MCQIds = append(assignment.MCQIds, req.MCQIds...)
	assignment.MSQIds = append(assignment.MSQIds, req.MSQIds...)
	assignment.NATIds = append(assignment.NATIds, req.NATIds...)
	assignment.SubjectiveIds = append(assignment.SubjectiveIds, req.SubjectiveIds...)

	// Save assignment
	if err := repo.SaveAssignment(assignment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

// @Summary Get Assignment by ID
// @Tags Assignments
// @Produce json
// @Param id path string true "Assignment ID"
// @Param extended query string false "Extended view with populated relations"
// @Success 200 {object} model.Assignment
// @Success 200 {object} serializer.AssignmentExtended
// @Router /assignments/{id} [get]
func GetAssignment(c *gin.Context) {
	id := c.Param("id")

	assignment, err := repo.GetAssignmentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	// Check if extended query param is true
	extended := c.DefaultQuery("extended", "false") == "true"

	if extended {
		ext := serializer.NewAssignmentExtended(assignment)
		c.JSON(http.StatusOK, ext)
		return
	}

	// Default plain assignment
	c.JSON(http.StatusOK, assignment)
}

// @Summary Delete Assignment
// @Tags Assignments
// @Param id path string true "Assignment ID"
// @Success 200 {object} map[string]string
// @Router /assignments/{id} [delete]
func DeleteAssignment(c *gin.Context) {
	id := c.Param("id")
	if err := repo.DeleteAssignment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assignment deleted successfully"})
}

// @Summary Update Assignment
// @Tags Assignments
// @Accept json
// @Produce json
// @Param id path string true "Assignment ID"
// @Param assignment body model.Assignment true "Updated Assignment"
// @Success 200 {object} model.Assignment
// @Router /assignments/{id} [put]
func UpdateAssignment(c *gin.Context) {
	id := c.Param("id")
	var a model.Assignment
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a.ID = id
	a.UpdatedAt = time.Now()

	// Validate the struct
	if err := common.Validate.Struct(a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveAssignment(a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, a)
}

// @Summary Get All Assignments
// @Tags Assignments
// @Produce json
// @Param points query string false "Filter by points"
// @Param dueDate query string false "Filter by due date"
// @Param limit query string false "Pagination limit"
// @Param offset query string false "Pagination offset"
// @Success 200 {array} model.Assignment
// @Router /assignments [get]
func GetAllAssignments(c *gin.Context) {
	filters := make(map[string]string)
	if points := c.Query("points"); points != "" {
		filters["points"] = points
	}
	if due := c.Query("dueDate"); due != "" {
		filters["dueDate"] = due
	}
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	assignments, err := repo.FilterAssignments(
		filters["limit"],
		filters["offset"],
		filters["points"],
		filters["dueDate"],
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if extended query param is true
	extended := c.DefaultQuery("extended", "false") == "true"

	if extended {
		extendedList := make([]*serializer.AssignmentExtended, 0, len(assignments))
		for i := range assignments {
			ext := serializer.NewAssignmentExtended(&assignments[i])
			extendedList = append(extendedList, ext)
		}
		c.JSON(http.StatusOK, extendedList)
		return
	}

	// Default plain assignment list
	c.JSON(http.StatusOK, assignments)
}

// @Summary Patch Assignment
// @Tags Assignments
// @Accept json
// @Produce json
// @Param id path string true "Assignment ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} model.Assignment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assignments/{id} [patch]
func PatchAssignment(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// Get the updated assignment
	updated, err := repo.PatchAssignment(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}
