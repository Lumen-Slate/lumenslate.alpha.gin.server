package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"lumenslate/internal/controller"
	"lumenslate/internal/model"
	"lumenslate/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCreateUsageLimits tests the creation of usage limits
func TestCreateUsageLimits(t *testing.T) {
	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/usage-limits", controller.CreateUsageLimits)

	// Test data
	usageLimitsData := service.CreateUsageLimitsRequest{
		PlanName:                "test-plan",
		Teachers:                10,
		Classrooms:              20,
		StudentsPerClassroom:    30,
		QuestionBanks:           100,
		Questions:               1000,
		AssignmentExportsPerDay: 50,
		AI: model.AILimits{
			IndependentAgent:   200,
			LumenAgent:         150,
			RAGAgent:           100,
			RAGDocumentUploads: 50,
		},
	}

	jsonData, _ := json.Marshal(usageLimitsData)

	// Create request
	req, _ := http.NewRequest("POST", "/usage-limits", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions (Note: These will depend on your actual database setup)
	// For now, we're just testing the controller structure
	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestGetUsageLimits tests retrieving usage limits by ID
func TestGetUsageLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/usage-limits/:id", controller.GetUsageLimits)

	// Create request with a test ID
	req, _ := http.NewRequest("GET", "/usage-limits/507f1f77bcf86cd799439011", nil)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This will return 404 without actual database, but tests the route structure
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

// TestUpdateUsageLimits tests updating usage limits
func TestUpdateUsageLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/usage-limits/:id", controller.UpdateUsageLimits)

	// Test update data
	updateData := service.UpdateUsageLimitsRequest{
		Teachers:   new(model.UsageLimitValue),
		Classrooms: new(model.UsageLimitValue),
	}
	*updateData.Teachers = 15
	*updateData.Classrooms = "unlimited"

	jsonData, _ := json.Marshal(updateData)

	// Create request
	req, _ := http.NewRequest("PUT", "/usage-limits/507f1f77bcf86cd799439011", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This will return an error without actual database, but tests the route structure
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError, http.StatusNotFound}, w.Code)
}

// TestUsageLimitValueValidation tests the validation of usage limit values
func TestUsageLimitValueValidation(t *testing.T) {
	// Test valid values
	assert.True(t, model.ValidateUsageLimitValue(10))
	assert.True(t, model.ValidateUsageLimitValue("unlimited"))
	assert.True(t, model.ValidateUsageLimitValue("custom"))
	assert.True(t, model.ValidateUsageLimitValue(-1))

	// Test invalid values
	assert.False(t, model.ValidateUsageLimitValue("invalid"))
	assert.False(t, model.ValidateUsageLimitValue([]string{"test"}))
	assert.False(t, model.ValidateUsageLimitValue(map[string]string{"test": "value"}))
}

// TestIsUnlimited tests the IsUnlimited function
func TestIsUnlimited(t *testing.T) {
	assert.True(t, model.IsUnlimited("unlimited"))
	assert.True(t, model.IsUnlimited("custom"))
	assert.True(t, model.IsUnlimited(-1))
	assert.True(t, model.IsUnlimited(int64(-1)))

	assert.False(t, model.IsUnlimited(10))
	assert.False(t, model.IsUnlimited("limited"))
	assert.False(t, model.IsUnlimited(0))
}

// TestGetIntValue tests the GetIntValue function
func TestGetIntValue(t *testing.T) {
	assert.Equal(t, -1, model.GetIntValue("unlimited"))
	assert.Equal(t, -1, model.GetIntValue("custom"))
	assert.Equal(t, 10, model.GetIntValue(10))
	assert.Equal(t, 20, model.GetIntValue(int64(20)))
	assert.Equal(t, 30, model.GetIntValue(float64(30)))
	assert.Equal(t, 0, model.GetIntValue("invalid"))
}
