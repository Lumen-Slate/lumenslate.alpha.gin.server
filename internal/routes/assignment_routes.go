package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterAssignmentRoutes(r *gin.RouterGroup) {
	a := r.Group("/assignments")
	{
		a.POST("/", controller.CreateAssignment)
		a.GET("/", controller.GetAllAssignments)
		a.GET("/:id", controller.GetAssignment)
		a.PUT("/:id", controller.UpdateAssignment)
		a.PATCH("/:id", controller.PatchAssignment)
		a.DELETE("/:id", controller.DeleteAssignment)
	}
}
