// controller/questions/nat_controller.go
package questions

import (
	"lumenslate/internal/model/questions"
	service "lumenslate/internal/service/questions"
	"net/http"

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
	var n questions.NAT
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n.ID = uuid.New().String()
	if err := service.CreateNAT(n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create NAT"})
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
	n, err := service.GetNAT(id)
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
	if err := service.DeleteNAT(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete NAT"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "NAT deleted"})
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
	filters := map[string]string{
		"bankId": c.Query("bankId"),
		"limit":  c.DefaultQuery("limit", "10"),
		"offset": c.DefaultQuery("offset", "0"),
	}
	nats, err := service.GetAllNATs(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch NATs"})
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
	var n questions.NAT
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n.ID = id
	if err := service.UpdateNAT(id, n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update NAT"})
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
	if err := service.PatchNAT(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch NAT"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "NAT updated"})
}

// @Summary Bulk Create NATs
// @Tags NATs
// @Accept json
// @Produce json
// @Param nats body []questions.NAT true "List of NATs"
// @Success 201 {array} questions.NAT
// @Router /nats/bulk [post]
func CreateBulkNATs(c *gin.Context) {
	var nats []questions.NAT
	if err := c.ShouldBindJSON(&nats); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i := range nats {
		nats[i].ID = uuid.New().String()
	}

	if err := service.CreateBulkNATs(nats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create NATs"})
		return
	}

	c.JSON(http.StatusCreated, nats)
}
