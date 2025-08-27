package server

import (
	"HelaList/internal/server/handler"

	"github.com/gin-gonic/gin"
)

// 总路由
func Init() *gin.Engine {
	r := gin.Default()
	registerUserRoutes(r)
	return r
}

func registerUserRoutes(r *gin.Engine) {
	api := r.Group("/api")
	user := api.Group("/user")
	{
		user.GET("/get", handler.GetUser)
		user.POST("/create", handler.CreateUser)
		user.POST("/update", handler.UpdateUser)
		user.POST("/delete", handler.DeleteUser)
	}
}
