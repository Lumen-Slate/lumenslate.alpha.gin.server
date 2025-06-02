// controller/assignment_controller.go
package controller

import (
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"net/http"

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
	a.ID = uuid.New().String() // Auto-generate ID
	if err := service.CreateAssignment(a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
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
	a, err := service.GetAssignment(id)
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
	err := service.DeleteAssignment(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete assignment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assignment deleted"})
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
	var updated model.Assignment
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated.ID = id
	if err := service.UpdateAssignment(id, updated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, updated)
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
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")
	points := c.Query("points")
	due := c.Query("dueDate")

	assignments, err := service.FilterAssignments(limit, offset, points, due)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}
	c.JSON(http.StatusOK, assignments)
}
