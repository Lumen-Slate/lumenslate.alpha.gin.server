// controller/comment_controller.go
package controller

import (
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create Comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param comment body model.Comment true "Comment JSON"
// @Success 201 {object} model.Comment
// @Router /comments [post]
func CreateComment(c *gin.Context) {
	var comment model.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment.ID = uuid.New().String() // Auto-generate ID
	if err := service.CreateComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// @Summary Get Comment by ID
// @Tags Comments
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} model.Comment
// @Router /comments/{id} [get]
func GetComment(c *gin.Context) {
	id := c.Param("id")
	comment, err := service.GetComment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	c.JSON(http.StatusOK, comment)
}

// @Summary Delete Comment
// @Tags Comments
// @Param id path string true "Comment ID"
// @Success 200 {object} map[string]string
// @Router /comments/{id} [delete]
func DeleteComment(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteComment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

// @Summary Get All Comments
// @Tags Comments
// @Produce json
// @Success 200 {array} model.Comment
// @Router /comments [get]
func GetAllComments(c *gin.Context) {
	comments, err := service.GetAllComments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}
	c.JSON(http.StatusOK, comments)
}

// @Summary Update Comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param comment body model.Comment true "Updated Comment"
// @Success 200 {object} model.Comment
// @Router /comments/{id} [put]
func UpdateComment(c *gin.Context) {
	id := c.Param("id")
	var comment model.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment.ID = id
	if err := service.UpdateComment(id, comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, comment)
}

// @Summary Patch a comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /comments/{id} [patch]
func PatchComment(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.PatchComment(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment updated"})
}
