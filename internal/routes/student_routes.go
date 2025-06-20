package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterStudentRoutes(r *gin.Engine) {
	students := r.Group("/students")
	{
		students.GET("", controller.GetAllStudents)
		students.GET(":id", controller.GetStudent)
		students.POST("", controller.CreateStudent)
		students.PUT(":id", controller.UpdateStudent)
		students.PATCH(":id", controller.PatchStudent)
		students.DELETE(":id", controller.DeleteStudent)
	}
}
