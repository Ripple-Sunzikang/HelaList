package server

import (
	"HelaList/configs"
	"HelaList/internal/bootstrap"
	"HelaList/internal/rag"
	"HelaList/internal/repository"
	"HelaList/internal/server/handler"
	"HelaList/internal/server/middlewares"
	"HelaList/internal/server/webdav"
	"HelaList/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

// 总路由
func Init() *gin.Engine {
	// 初始化RAG服务和聊天服务
	initRAGAndChatServices()

	r := gin.Default()
	registerUserRoutes(r)
	registerStorageRoutes(r)
	registerMetaRoutes(r)
	registerFsRoutes(r)
	registerAIRoutes(r)
	registerWebdavRoutes(r)
	return r
}

func initRAGAndChatServices() {
	// 获取数据库连接
	db, err := bootstrap.Db.DB()
	if err != nil {
		log.Printf("Failed to get database connection for RAG: %v", err)
		return
	}

	// 创建RAG服务
	ragService := rag.NewRAGService(db, &configs.Conf.RAG)

	// 创建聊天仓库和服务
	chatRepo := repository.NewChatRepository(bootstrap.Db)
	chatService := service.NewChatService(chatRepo, ragService)

	// 初始化handler中的服务
	handler.InitRAGService(ragService)
	handler.InitChatService(chatService)

	log.Println("RAG and Chat services initialized successfully")
}

func registerUserRoutes(r *gin.Engine) {
	api := r.Group("/api")
	user := api.Group("/user")
	{
		user.POST("/login", handler.Login)   // 登录接口
		user.POST("/logout", handler.Logout) // 登出接口
		user.GET("/get", middlewares.Auth(true), handler.GetUser)
		user.POST("/create", handler.CreateUser)
		user.POST("/update", middlewares.Auth(true), handler.UpdateUser)
		user.POST("/delete", middlewares.Auth(true), handler.DeleteUser)
	}
}

func registerStorageRoutes(r *gin.Engine) {
	api := r.Group("/api")
	storage := api.Group("/storage")
	{
		storage.POST("/create", handler.CreateStorageHandler)
		storage.POST("/update", handler.UpdateStorageHandler)
		storage.POST("/load", handler.LoadStorageHandler)
		storage.DELETE("/:id", handler.DeleteStorageHandler)
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

func registerAIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	ai := api.Group("/ai")
	ai.Use(middlewares.Auth(false)) // 可选认证
	{
		ai.POST("/chat", handler.AIChatHandler)
		ai.POST("/execute", handler.ExecuteFileOperationHandler)
	}

	// 对话上下文聊天接口
	chat := api.Group("/chat")
	chat.Use(middlewares.Auth(false)) // 可选认证
	{
		chat.POST("/message", handler.ChatWithContextHandler)                   // 发送消息（支持上下文和RAG）
		chat.POST("/sessions", handler.CreateChatSessionHandler)                // 创建新会话
		chat.GET("/sessions", handler.GetUserSessionsHandler)                   // 获取用户会话列表
		chat.GET("/sessions/:sessionId", handler.GetChatSessionHandler)         // 获取会话详情
		chat.PUT("/sessions/:sessionId", handler.UpdateChatSessionHandler)      // 更新会话
		chat.DELETE("/sessions/:sessionId", handler.DeleteChatSessionHandler)   // 删除会话
		chat.GET("/sessions/:sessionId/history", handler.GetChatHistoryHandler) // 获取会话消息历史
	}

	// RAG相关接口
	rag := api.Group("/rag")
	rag.Use(middlewares.Auth(false)) // 可选认证
	{
		rag.POST("/index", handler.RAGIndexHandler)
		rag.GET("/status", handler.RAGStatusHandler)
		rag.POST("/search", handler.RAGSearchHandler)
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

func registerFsRoutes(r *gin.Engine) {
	api := r.Group("/api")
	fs := api.Group("/fs")
	fs.Use(middlewares.Auth(false)) // 应用认证中间件，false表示不需要强制认证
	{
		fs.GET("/list/*path", handler.FsListHandler)
		fs.GET("/dirs/*path", handler.FsDirsHandler)
		fs.GET("/get/*path", handler.FsGetHandler)
		fs.POST("/mkdir", handler.FsMkdir)
		fs.POST("/copy", handler.FsCopyHandler)
		fs.POST("/move", handler.FsMoveHandler)
		fs.POST("/rename", handler.FsRenameHandler)
		fs.POST("/remove", handler.FsRemoveHandler)
		fs.POST("/put", handler.FsPutHandler)
		fs.POST("/link", handler.FsLinkHandler)

		// 下载、预览和流媒体相关路由
		fs.GET("/download/*path", handler.DownloadHandler) // 文件下载
		fs.GET("/preview/*path", handler.PreviewHandler)   // 文件预览
		fs.GET("/proxy/*path", handler.ProxyHandler)       // 代理访问
		fs.GET("/stream/*path", handler.StreamHandler)     // 流媒体播放
	}
}
