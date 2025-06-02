// controller/post_controller.go
package controller

import (
	"lumenslate/internal/model"
	"lumenslate/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create Post
// @Tags Posts
// @Accept json
// @Produce json
// @Param post body model.Post true "Post JSON"
// @Success 201 {object} model.Post
// @Router /posts [post]
func CreatePost(c *gin.Context) {
	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post.ID = uuid.New().String() // Auto-generate ID
	if err := service.CreatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}
	c.JSON(http.StatusCreated, post)
}

// @Summary Get Post by ID
// @Tags Posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} model.Post
// @Router /posts/{id} [get]
func GetPost(c *gin.Context) {
	id := c.Param("id")
	post, err := service.GetPost(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// @Summary Delete Post
// @Tags Posts
// @Param id path string true "Post ID"
// @Success 200 {object} map[string]string
// @Router /posts/{id} [delete]
func DeletePost(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}

// @Summary Get All Posts
// @Tags Posts
// @Produce json
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Param userId query string false "Filter by user ID"
// @Success 200 {array} model.Post
// @Router /posts [get]
func GetAllPosts(c *gin.Context) {
	filters := map[string]string{
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
		"userId": c.Query("userId"),
	}
	posts, err := service.GetAllPosts(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// @Summary Update Post
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param post body model.Post true "Updated Post"
// @Success 200 {object} model.Post
// @Router /posts/{id} [put]
func UpdatePost(c *gin.Context) {
	id := c.Param("id")
	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post.ID = id
	if err := service.UpdatePost(id, post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// @Summary Patch Post
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]string
// @Router /posts/{id} [patch]
func PatchPost(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.PatchPost(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Patch failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post updated"})
}
