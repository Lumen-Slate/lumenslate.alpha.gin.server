// controller/assignment_controller.go
package controller

import (
	"lumenslate/internal/common"
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
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
	var a model.Assignment
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create new Assignment with default values
	a = *model.NewAssignment()
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID
	a.ID = uuid.New().String()

	// Validate the struct
	if err := common.Validate.Struct(a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveAssignment(a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, a)
}

// @Summary Get Assignment by ID
// @Tags Assignments
// @Produce json
// @Param id path string true "Assignment ID"
// @Success 200 {object} model.Assignment
// @Router /assignments/{id} [get]
func GetAssignment(c *gin.Context) {
	id := c.Param("id")
	a, err := repo.GetAssignmentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}
	c.JSON(http.StatusOK, a)
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
