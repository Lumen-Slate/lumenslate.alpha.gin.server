package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterClassroomRoutes(r *gin.RouterGroup) {
	cls := r.Group("/classrooms")
	{
		cls.GET("", controller.GetAllClassrooms)
		cls.GET(":id", controller.GetClassroom)
		cls.POST("", controller.CreateClassroom)
		cls.PUT(":id", controller.UpdateClassroom)
		cls.PATCH(":id", controller.PatchClassroom)
		cls.DELETE(":id", controller.DeleteClassroom)
	}
}
