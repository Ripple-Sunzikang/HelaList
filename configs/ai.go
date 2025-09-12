package configs

import (
	"os"
)

// AI相关配置
type AIConfig struct {
	DeepSeekAPIKey string
	DeepSeekModel  string
	DeepSeekAPIURL string
}

var AI = AIConfig{
	DeepSeekAPIKey: getEnvOrDefault("DEEPSEEK_API_KEY", "sk-fcccf8df73c54f9f9f657fb2abdcd202"),
	DeepSeekModel:  getEnvOrDefault("DEEPSEEK_MODEL", "deepseek-chat"),
	DeepSeekAPIURL: getEnvOrDefault("DEEPSEEK_API_URL", "https://api.deepseek.com/chat/completions"),
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
