// controller/post_controller.go
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

// @Summary Create Thread
// @Tags Threads
// @Accept json
// @Produce json
// @Param thread body model.Thread true "Thread JSON"
// @Success 201 {object} model.Thread
// @Router /threads [post]
func CreateThread(c *gin.Context) {
	// Create new Thread with default values
	thread := *model.NewThread()

	// Bind JSON to the struct
	if err := c.ShouldBindJSON(&thread); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID
	thread.ID = uuid.New().String()

	// Validate the struct
	if err := common.Validate.Struct(thread); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.CreateThread(thread); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}
	c.JSON(http.StatusCreated, thread)
}

// @Summary Get Thread by ID
// @Tags Threads
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} model.Thread
// @Router /threads/{id} [get]
func GetThread(c *gin.Context) {
	id := c.Param("id")
	thread, err := service.GetThread(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}
	c.JSON(http.StatusOK, thread)
}

// @Summary Delete Thread
// @Tags Threads
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]string
// @Router /threads/{id} [delete]
func DeleteThread(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteThread(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Thread deleted successfully"})
}

// @Summary Get All Threads
// @Tags Threads
// @Produce json
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Param userId query string false "Filter by user ID"
// @Success 200 {array} model.Thread
// @Router /threads [get]
func GetAllThreads(c *gin.Context) {
	filters := map[string]string{
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
		"userId": c.Query("userId"),
	}
	threads, err := service.GetAllThreads(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
		return
	}
	c.JSON(http.StatusOK, threads)
}

// @Summary Update Thread
// @Tags Threads
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Param thread body model.Thread true "Updated Thread"
// @Success 200 {object} model.Thread
// @Router /threads/{id} [put]
func UpdateThread(c *gin.Context) {
	id := c.Param("id")
	var thread model.Thread
	if err := c.ShouldBindJSON(&thread); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	thread.ID = id
	thread.UpdatedAt = time.Now()

	// Validate the struct
	if err := common.Validate.Struct(thread); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateThread(id, thread); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, thread)
}

// @Summary Patch Thread
// @Tags Threads
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} model.Thread
// @Router /threads/{id} [patch]
func PatchThread(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	// Get the updated thread
	updated, err := service.PatchThread(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Patch failed"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
