package controller

import (
	"context"
	"lumenslate/internal/model"
	"lumenslate/internal/repository"
	"lumenslate/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type UserController struct {
	Repo *repository.UserRepository
}

func NewUserController(repo *repository.UserRepository) *UserController {
	return &UserController{Repo: repo}
}

func (uc *UserController) CreateUser(c *gin.Context) {
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
	if err := uc.Repo.CreateUser(context.Background(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.Repo.GetUserByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uc *UserController) UpdateUser(c *gin.Context) {
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
	// Optionally set updatedAt if you add that field
	if err := uc.Repo.UpdateUser(context.Background(), id, bson.M{
		"name":         u.Name,
		"email":        u.Email,
		"role":         u.Role,
		"phone_number": u.PhoneNumber,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := uc.Repo.DeleteUser(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func (uc *UserController) ListUsers(c *gin.Context) {
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
	users, err := uc.Repo.ListUsers(context.Background(), query, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (uc *UserController) PatchUser(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates["updatedAt"] = time.Now()
	if err := uc.Repo.UpdateUser(context.Background(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch user"})
		return
	}
	user, _ := uc.Repo.GetUserByID(context.Background(), id)
	c.JSON(http.StatusOK, user)
}
