package configs

import (
	"os"
)

// AI相关配置
type AIConfig struct {
	QwenAPIKey string
	QwenModel  string
	QwenAPIURL string
}

var AI = AIConfig{
	QwenAPIKey: os.Getenv("QWEN_API_KEY"),
	QwenModel:  getEnvOrDefault("QWEN_MODEL", "qwen-vl-plus"),
	QwenAPIURL: getEnvOrDefault("QWEN_API_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"),
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
