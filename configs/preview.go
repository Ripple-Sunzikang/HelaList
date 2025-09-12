package configs

import "strings"

// 支持预览的文件类型
var (
	// 支持代理预览的文件类型
	ProxyTypes = []string{
		// 图片
		"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico",
		// 视频
		"mp4", "avi", "mkv", "mov", "wmv", "flv", "webm", "m4v",
		// 音频
		"mp3", "wav", "flac", "aac", "ogg", "m4a", "wma",
		// 文档
		"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
		// 文本
		"txt", "md", "json", "xml", "yaml", "yml", "csv",
		// 代码
		"js", "css", "html", "htm", "go", "py", "java", "cpp", "c", "php",
	}

	// 纯文本文件类型（可以直接显示内容）
	TextTypes = []string{
		"txt", "md", "json", "xml", "yaml", "yml", "csv",
		"js", "css", "html", "htm", "go", "py", "java", "cpp", "c", "php",
		"sh", "bat", "ini", "conf", "cfg", "log",
	}
)

// 检查文件扩展名是否在指定的类型列表中
func IsFileTypeIn(filename string, types []string) bool {
	ext := getFileExt(filename)
	for _, t := range types {
		if strings.EqualFold(ext, t) {
			return true
		}
	}
	return false
}

// 获取文件扩展名
func getFileExt(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return strings.ToLower(filename[idx+1:])
	}
	return ""
}

// 判断是否支持代理预览
func ShouldProxy(filename string, webProxy bool, mustProxy bool) bool {
	if mustProxy || webProxy {
		return true
	}
	return IsFileTypeIn(filename, ProxyTypes)
}

// 判断是否可以代理
func CanProxy(filename string, webProxy bool, mustProxy bool) bool {
	if mustProxy || webProxy {
		return true
	}
	if IsFileTypeIn(filename, ProxyTypes) {
		return true
	}
	if IsFileTypeIn(filename, TextTypes) {
		return true
	}
	return false
}
