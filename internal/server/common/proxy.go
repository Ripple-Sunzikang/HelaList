package common

import (
	"HelaList/configs"
	"HelaList/internal/driver"
	"HelaList/internal/model"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ShouldProxy 判断是否应该使用代理
func ShouldProxy(storage driver.Driver, filename string) bool {
	cfg := storage.Config()
	storageModel := storage.GetStorage()

	return configs.ShouldProxy(filename, storageModel.WebProxy, cfg.MustProxy())
}

// CanProxy 判断是否可以使用代理
func CanProxy(storage driver.Driver, filename string) bool {
	cfg := storage.Config()
	storageModel := storage.GetStorage()

	return configs.CanProxy(filename, storageModel.WebProxy, cfg.MustProxy())
}

// Proxy 处理代理请求
func Proxy(c *gin.Context, link *model.Link, file model.Obj, proxyRange bool) error {
	defer func() {
		if link != nil {
			link.Close()
		}
	}()

	if link.MFile != nil {
		// 处理内存文件
		return serveFile(c, file, link.MFile)
	}

	if link.URL == "" {
		return errors.New("empty download url")
	}

	// 创建代理请求
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", link.URL, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// 复制必要的请求头
	copyHeaders(req, c.Request, link.Header)

	// 处理Range请求
	if proxyRange && c.GetHeader("Range") != "" {
		req.Header.Set("Range", c.GetHeader("Range"))
	}

	// 发送请求
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow all redirects
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to proxy request")
	}
	defer resp.Body.Close()

	// 复制响应头
	c.Status(resp.StatusCode)
	copyResponseHeaders(c, resp)

	// 设置内容相关头部
	setContentHeaders(c, file, resp)

	// 复制响应体
	_, err = io.Copy(c.Writer, resp.Body)
	return err
}

// 服务内存文件
func serveFile(c *gin.Context, file model.Obj, mfile io.ReadSeeker) error {
	c.Header("Content-Type", getContentType(file.GetName()))
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\" %s\"", file.GetName()))
	c.Header("Cache-Control", "public, max-age=3600")

	http.ServeContent(c.Writer, c.Request, file.GetName(), file.GetModifiedTime(), mfile)
	return nil
}

// 复制请求头
func copyHeaders(req *http.Request, original *http.Request, additional http.Header) {
	// 复制原始请求的关键头部
	for _, header := range []string{"User-Agent", "Accept", "Accept-Language", "Accept-Encoding"} {
		if value := original.Header.Get(header); value != "" {
			req.Header.Set(header, value)
		}
	}

	// 添加额外的头部
	for key, values := range additional {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

// 复制响应头
func copyResponseHeaders(c *gin.Context, resp *http.Response) {
	for key, values := range resp.Header {
		// 跳过某些不应该转发的头部
		if shouldSkipHeader(key) {
			continue
		}
		for _, value := range values {
			c.Header(key, value)
		}
	}
}

// 设置内容相关头部
func setContentHeaders(c *gin.Context, file model.Obj, resp *http.Response) {
	// 设置文件名
	if c.GetHeader("Content-Disposition") == "" {
		dispositionType := "inline"
		if c.Query("type") == "download" || !isPreviewable(file.GetName()) {
			dispositionType = "attachment"
		}
		c.Header("Content-Disposition", fmt.Sprintf("%s; filename=\" %s\"", dispositionType, file.GetName()))
	}

	// 设置内容类型
	if c.GetHeader("Content-Type") == "" {
		c.Header("Content-Type", getContentType(file.GetName()))
	}

	// 设置缓存控制
	if c.GetHeader("Cache-Control") == "" {
		if configs.IsFileTypeIn(file.GetName(), configs.TextTypes) {
			c.Header("Cache-Control", "public, max-age=300") // 5分钟缓存
		} else {
			c.Header("Cache-Control", "public, max-age=3600") // 1小时缓存
		}
	}

	// 设置内容长度
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Header("Content-Length", contentLength)
	} else if file.GetSize() > 0 {
		c.Header("Content-Length", strconv.FormatInt(file.GetSize(), 10))
	}
}

// 判断是否应该跳过某些响应头
func shouldSkipHeader(key string) bool {
	key = strings.ToLower(key)
	skipHeaders := []string{
		"content-disposition", // Let the caller handler control this
		"connection",
		"transfer-encoding",
		"upgrade",
		"proxy-authenticate",
		"proxy-authorization",
		"te",
		"trailers",
	}

	for _, skip := range skipHeaders {
		if key == skip {
			return true
		}
	}
	return false
}

// 判断文件是否可以预览（内联显示）
func isPreviewable(filename string) bool {
	return configs.IsFileTypeIn(filename, append(configs.ProxyTypes, configs.TextTypes...))
}

// 获取文件的MIME类型
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		// 图片
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".bmp":  "image/bmp",

		// 视频
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".mkv":  "video/x-matroska",

		// 音频
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".m4a":  "audio/mp4",

		// 文档
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

		// 文本
		".txt":  "text/plain; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".htm":  "text/html; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript; charset=utf-8",
		".json": "application/json; charset=utf-8",
		".xml":  "application/xml; charset=utf-8",
		".md":   "text/markdown; charset=utf-8",
		".csv":  "text/csv; charset=utf-8",
		".yaml": "text/yaml; charset=utf-8",
		".yml":  "text/yaml; charset=utf-8",

		// 代码
		".go":   "text/plain; charset=utf-8",
		".py":   "text/plain; charset=utf-8",
		".java": "text/plain; charset=utf-8",
		".cpp":  "text/plain; charset=utf-8",
		".c":    "text/plain; charset=utf-8",
		".php":  "text/plain; charset=utf-8",
		".sh":   "text/plain; charset=utf-8",
		".bat":  "text/plain; charset=utf-8",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	return "application/octet-stream"
}

// In internal/server/common/proxy.go

func ProxyDownload(c *gin.Context, link *model.Link, filename string) error {
	defer func() {
		if link != nil {
			link.Close()
		}
	}()

	if link.URL == "" {
		return errors.New("empty download url")
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", link.URL, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// Only copy authentication headers from the driver
	for key, values := range link.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to proxy request")
	}
	defer resp.Body.Close()

	// Set headers for forced download
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "application/octet-stream") // A generic content type
	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	
c.Status(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	return err
}
