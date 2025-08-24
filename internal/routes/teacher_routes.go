package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterTeacherRoutes(r *gin.RouterGroup) {
	teachers := r.Group("/teachers")
	{
		teachers.GET("", controller.GetAllTeachers)
		teachers.GET(":id", controller.GetTeacher)
		teachers.POST("", controller.CreateTeacher)
		teachers.PUT(":id", controller.UpdateTeacher)
		teachers.PATCH(":id", controller.PatchTeacher)
		teachers.DELETE(":id", controller.DeleteTeacher)
	}
}
