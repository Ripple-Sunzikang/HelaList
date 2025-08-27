package server

import (
	"HelaList/internal/server/handler"
	"HelaList/internal/server/webdav"

	"github.com/gin-gonic/gin"
)

// 总路由
func Init() *gin.Engine {
	r := gin.Default()
	registerUserRoutes(r)
	registerStorageRoutes(r)
	registerMetaRoutes(r)
	registerWebdavRoutes(r)
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

func registerWebdavRoutes(r *gin.Engine) {
	// 创建 WebDAV Handler 实例
	webdavHandler := &webdav.Handler{
		Prefix:     "/webdav",
		LockSystem: webdav.NewMemLS(), // 使用内存锁系统
		// Logger: 可选的日志函数，用于记录错误
	}

	// 使用 gin.WrapH 将 http.Handler 包装为 Gin 中间件
	// 支持 WebDAV 方法：OPTIONS, DELETE, MKCOL
	r.Any("/webdav/*path", gin.WrapH(webdavHandler))
}
