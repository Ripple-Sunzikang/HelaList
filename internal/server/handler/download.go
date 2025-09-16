package handler

import (
	"HelaList/configs"
	"HelaList/internal/fs"
	"HelaList/internal/model"
	"HelaList/internal/op"
	"HelaList/internal/server/common"
	"errors"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DownloadHandler now calls ProxyHandler to unify logic
func DownloadHandler(c *gin.Context) {
	// All logic is now in ProxyHandler, which respects `?type=download`
	ProxyHandler(c)
}

// PreviewHandler 处理文件预览请求
func PreviewHandler(c *gin.Context) {
	// 直接调用代理处理器进行预览
	ProxyHandler(c)
}

// ProxyHandler 处理代理请求
func ProxyHandler(c *gin.Context) {
	// 获取路径参数
	rawPath := c.Param("path")
	if rawPath == "" {
		rawPath = c.Request.Context().Value("path").(string)
	}

	// 获取用户信息
	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	if user.IsGuest() && user.Disabled {
		common.ErrorResponse(c, errors.New("guest user is disabled"), 401)
		return
	}

	// 构建完整路径
	reqPath, err := user.JoinPath(rawPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	// 获取存储和驱动
	storage, _, err := op.GetStorageAndActualPath(reqPath)
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	filename := filepath.Base(rawPath)

	// 检查是否可以代理
	if !common.CanProxy(storage, filename) {
		common.ErrorResponse(c, errors.New("proxy not allowed for this file type"), 403)
		return
	}

	// 获取文件信息和下载链接
	link, file, err := fs.Link(c.Request.Context(), reqPath, model.LinkArgs{
		Header: c.Request.Header,
		Type:   c.Query("type"),
	})
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	// 使用代理处理请求
	storageModel := storage.GetStorage()
	err = common.Proxy(c, link, file, storageModel.ProxyRange)
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}
}

// StreamHandler 处理流媒体请求（视频/音频）
func StreamHandler(c *gin.Context) {
	// 获取路径参数
	rawPath := c.Param("path")
	if rawPath == "" {
		common.ErrorResponse(c, errors.New("path is required"), 400)
		return
	}

	// 获取用户信息
	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	if user.IsGuest() && user.Disabled {
		common.ErrorResponse(c, errors.New("guest user is disabled"), 401)
		return
	}

	// 构建完整路径
	reqPath, err := user.JoinPath(rawPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	// 获取存储和驱动
	_, _, err = op.GetStorageAndActualPath(reqPath)
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	filename := filepath.Base(rawPath)

	// 检查是否是支持流式传输的文件类型
	if !isStreamable(filename) {
		common.ErrorResponse(c, errors.New("file type not streamable"), 400)
		return
	}

	// 获取文件信息和下载链接
	link, file, err := fs.Link(c.Request.Context(), reqPath, model.LinkArgs{
		Header: c.Request.Header,
		Type:   "stream",
	})
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	// 设置适当的Content-Type和其他头部
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", getStreamContentType(filename))

	// 如果文件大小已知，设置Content-Length
	if file.GetSize() > 0 {
		c.Header("Content-Length", strconv.FormatInt(file.GetSize(), 10))
	}

	// 使用代理处理流式请求（支持Range请求）
	err = common.Proxy(c, link, file, true) // 强制启用Range支持
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}
}

// 判断文件是否支持流式传输
func isStreamable(filename string) bool {
	streamableTypes := []string{
		"mp4", "avi", "mkv", "mov", "wmv", "flv", "webm", "m4v",
		"mp3", "wav", "flac", "aac", "ogg", "m4a", "wma",
	}
	return configs.IsFileTypeIn(filename, streamableTypes)
}

// 获取流媒体的Content-Type
func getStreamContentType(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "application/octet-stream"
	}

	ext = ext[1:] // 去掉点号

	videoTypes := map[string]string{
		"mp4":  "video/mp4",
		"webm": "video/webm",
		"avi":  "video/x-msvideo",
		"mov":  "video/quicktime",
		"wmv":  "video/x-ms-wmv",
		"flv":  "video/x-flv",
		"mkv":  "video/x-matroska",
		"m4v":  "video/mp4",
	}

	audioTypes := map[string]string{
		"mp3":  "audio/mpeg",
		"wav":  "audio/wav",
		"flac": "audio/flac",
		"aac":  "audio/aac",
		"ogg":  "audio/ogg",
		"m4a":  "audio/mp4",
		"wma":  "audio/x-ms-wma",
	}

	if contentType, ok := videoTypes[ext]; ok {
		return contentType
	}

	if contentType, ok := audioTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}
