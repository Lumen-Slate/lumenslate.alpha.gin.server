package controller

import (
	"lumenslate/internal/model"
	repo "lumenslate/internal/repository"
	"lumenslate/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateUser(c *gin.Context) {
	u := &model.User{}
	if err := c.ShouldBindJSON(u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u.ID = uuid.New().String()
	if err := utils.Validate.Struct(u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repo.SaveUser(*u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := repo.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var u model.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := utils.Validate.Struct(u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u.ID = id
	updates := map[string]interface{}{
		"name":         u.Name,
		"email":        u.Email,
		"role":         u.Role,
		"phone_number": u.PhoneNumber,
	}
	user, err := repo.PatchUser(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := repo.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func ListUsers(c *gin.Context) {
	filters := map[string]string{
		"email":  c.Query("email"),
		"phone":  c.Query("phone"),
		"name":   c.Query("name"),
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
	}
	query := bson.M{}
	if filters["email"] != "" {
		query["email"] = filters["email"]
	}
	if filters["phone"] != "" {
		query["phone_number"] = filters["phone"]
	}
	if filters["name"] != "" {
		query["name"] = filters["name"]
	}
	users, err := repo.GetAllUsers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func PatchUser(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates["updatedAt"] = time.Now()
	user, err := repo.PatchUser(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch user"})
		return
	}
	c.JSON(http.StatusOK, user)
}
