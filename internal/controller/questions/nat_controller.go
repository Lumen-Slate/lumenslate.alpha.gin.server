// controller/questions/nat_controller.go
package questions

import (
	"lumenslate/internal/common"
	model "lumenslate/internal/model/questions"
	repo "lumenslate/internal/repository/questions"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create NAT
// @Tags NATs
// @Accept json
// @Produce json
// @Param nat body questions.NAT true "NAT JSON"
// @Success 201 {object} questions.NAT
// @Router /nats [post]
func CreateNAT(c *gin.Context) {
	var n model.NAT
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create new NAT with default values
	n = *model.NewNAT()
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID
	n.ID = uuid.New().String()

	// Validate the struct
	if err := common.Validate.Struct(n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveNAT(n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, n)
}

// @Summary Get NAT by ID
// @Tags NATs
// @Produce json
// @Param id path string true "NAT ID"
// @Success 200 {object} questions.NAT
// @Router /nats/{id} [get]
func GetNAT(c *gin.Context) {
	id := c.Param("id")
	n, err := repo.GetNATByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NAT not found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

// @Summary Delete NAT
// @Tags NATs
// @Param id path string true "NAT ID"
// @Success 200 {object} map[string]string
// @Router /nats/{id} [delete]
func DeleteNAT(c *gin.Context) {
	id := c.Param("id")
	if err := repo.DeleteNAT(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "NAT deleted successfully"})
}

// @Summary Get All NATs
// @Tags NATs
// @Produce json
// @Param bankId query string false "Filter by bank ID"
// @Param limit query string false "Pagination limit"
// @Param offset query string false "Pagination offset"
// @Success 200 {array} questions.NAT
// @Router /nats [get]
func GetAllNATs(c *gin.Context) {
	filters := make(map[string]string)
	if bankID := c.Query("bankId"); bankID != "" {
		filters["bankId"] = bankID
	}
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	nats, err := repo.GetAllNATs(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nats)
}

// @Summary Update NAT
// @Tags NATs
// @Accept json
// @Produce json
// @Param id path string true "NAT ID"
// @Param nat body questions.NAT true "Updated NAT"
// @Success 200 {object} questions.NAT
// @Router /nats/{id} [put]
func UpdateNAT(c *gin.Context) {
	id := c.Param("id")
	var n model.NAT
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	n.ID = id
	n.UpdatedAt = time.Now()

	// Validate the struct
	if err := common.Validate.Struct(n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repo.SaveNAT(n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, n)
}

// @Summary Patch NAT
// @Tags NATs
// @Accept json
// @Produce json
// @Param id path string true "NAT ID"
// @Param updates body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]string
// @Router /nats/{id} [patch]
func PatchNAT(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	updated, err := repo.PatchNAT(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary Bulk Create NATs
// @Tags NATs
// @Accept json
// @Produce json
// @Param nats body []questions.NAT true "List of NATs"
// @Success 201 {array} questions.NAT
// @Router /nats/bulk [post]
func CreateBulkNATs(c *gin.Context) {
	var nats []model.NAT
	if err := c.ShouldBindJSON(&nats); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate each NAT and set timestamps
	for i := range nats {
		nats[i].ID = uuid.New().String()
		nats[i].CreatedAt = time.Now()
		nats[i].UpdatedAt = time.Now()
		nats[i].IsActive = true

		if err := common.Validate.Struct(nats[i]); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if err := repo.SaveBulkNATs(nats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, nats)
}
