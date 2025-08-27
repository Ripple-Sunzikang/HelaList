package server

import (
	"HelaList/internal/server/handler"

	"github.com/gin-gonic/gin"
)

// 总路由
func Init() *gin.Engine {
	r := gin.Default()
	registerUserRoutes(r)
	registerStorageRoutes(r)
	registerMetaRoutes(r)
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

func registerStorageRoutes(r *gin.Engine) {
	api := r.Group("/api")
	storage := api.Group("/storage")
	{
		storage.POST("/create", handler.CreateStorageHandler)
		storage.POST("/update", handler.UpdateStorageHandler)
		storage.POST("/load", handler.LoadStorageHandler)
		storage.GET("/all", handler.GetAllStoragesHandler)
		storage.GET("/has/:mountPath", handler.HasStorageHandler)
		storage.GET("/:mountPath", handler.GetStorageByMountPathHandler)
		storage.GET("/virtual-files", handler.GetStorageVirtualFilesByPathHandler)
	}
}

func registerMetaRoutes(r *gin.Engine) {
	api := r.Group("/api")
	meta := api.Group("/meta")
	{
		meta.POST("/create", handler.CreateMetaHandler)
		meta.POST("/update", handler.UpdateMetaHandler)
		meta.DELETE("/:id", handler.DeleteMetaByIdHandler)
		meta.GET("/:id", handler.GetMetaByIdHandler)
		meta.GET("/path/:path", handler.GetMetaByPathHandler)
		meta.GET("/nearest/:path", handler.GetNearestMetaHandler)
		meta.GET("", handler.GetMetasHandler)
	}
}
