// controller/variable_controller.go
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

// @Summary Create Variable
// @Tags Variables
// @Accept json
// @Produce json
// @Param variable body model.Variable true "Variable JSON"
// @Success 201 {object} model.Variable
// @Router /variables [post]
func CreateVariable(c *gin.Context) {
	// Create new Variable with default values
	variable := *model.NewVariable()

	// Bind JSON to the struct
	if err := c.ShouldBindJSON(&variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID
	variable.ID = uuid.New().String()

	// Validate the variable
	if err := common.Validate.Struct(variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.CreateVariable(variable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create variable"})
		return
	}
	c.JSON(http.StatusCreated, variable)
}

// @Summary Get Variable by ID
// @Tags Variables
// @Produce json
// @Param id path string true "Variable ID"
// @Success 200 {object} model.Variable
// @Router /variables/{id} [get]
func GetVariable(c *gin.Context) {
	id := c.Param("id")
	variable, err := service.GetVariable(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variable not found"})
		return
	}
	c.JSON(http.StatusOK, variable)
}

// @Summary Delete Variable
// @Tags Variables
// @Param id path string true "Variable ID"
// @Success 200 {object} map[string]string
// @Router /variables/{id} [delete]
func DeleteVariable(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteVariable(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete variable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Variable deleted successfully"})
}

// @Summary Get All Variables
// @Tags Variables
// @Produce json
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {array} model.Variable
// @Router /variables [get]
func GetAllVariables(c *gin.Context) {
	filters := map[string]string{
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
	}
	variables, err := service.GetAllVariables(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch variables"})
		return
	}
	c.JSON(http.StatusOK, variables)
}

// @Summary Update Variable
// @Tags Variables
// @Accept json
// @Produce json
// @Param id path string true "Variable ID"
// @Param variable body model.Variable true "Updated Variable"
// @Success 200 {object} model.Variable
// @Router /variables/{id} [put]
func UpdateVariable(c *gin.Context) {
	id := c.Param("id")
	var variable model.Variable
	if err := c.ShouldBindJSON(&variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variable.ID = id
	variable.UpdatedAt = time.Now()

	// Validate the variable
	if err := common.Validate.Struct(variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateVariable(id, variable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, variable)
}

// @Summary Patch Variable
// @Tags Variables
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param updates body map[string]interface{} true "Updates"
// @Success 200 {object} model.Variable
// @Router /variables/{id} [patch]
func PatchVariable(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// Get the updated variable
	updated, err := service.PatchVariable(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch variable"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary Bulk Create Variables
// @Tags Variables
// @Accept json
// @Produce json
// @Param variables body []model.Variable true "List of Variables"
// @Success 201 {array} model.Variable
// @Router /variables/bulk [post]
func CreateBulkVariables(c *gin.Context) {
	var variables []model.Variable
	if err := c.ShouldBindJSON(&variables); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	for i := range variables {
		variables[i].ID = uuid.New().String()
		variables[i].CreatedAt = now
		variables[i].UpdatedAt = now
		variables[i].IsActive = true

		// Validate each variable
		if err := common.Validate.Struct(variables[i]); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if err := service.CreateBulkVariables(variables); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create variables"})
		return
	}

	c.JSON(http.StatusCreated, variables)
}
