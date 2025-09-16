package configs

import (
	"HelaList/internal/redis"
	"os"
)

var Conf *Config

type Database struct {
	Type     string `json:"type" env:"TYPE"`
	Host     string `json:"host" env:"HOST"`
	Port     int    `json:"port" env:"PORT"`
	User     string `json:"json" env:"USER"`
	Password string `json:"password" env:"PASSWORD"`
	Name     string `json:"name" env:"NAME"`
	DSN      string `json:"dsn" env:"DSN"`
}

type Config struct {
	SiteURL        string       `json:"site_url" env:"SITE_URL"`
	Cdn            string       `json:"cdn" env:"CDN"`
	JwtSecret      string       `json:"jwt_secret" env:"JWT_SECRET"`
	TokenExpiresIn int          `json:"token_expires_in" env:"TOKEN_EXPIRES_IN"`
	Database       Database     `json:"database" envPrefix:"DB_"`
	Redis          redis.Config `json:"redis" envPrefix:"REDIS_"`
	Tasks          TasksConfig  `json:"tasks" envPrefix:"TASKS_"`
	RAG            RAGConfig    `json:"rag" envPrefix:"RAG_"`
}

func DefaultConfig(dataDir string) *Config {
	return &Config{
		TokenExpiresIn: 24,
		Database: Database{
			Type:     "postgresql",
			Host:     "localhost",
			Port:     5432,
			User:     "suzuki",
			Password: "suzuki",
			Name:     "hela",
			DSN:      "host=localhost user=suzuki password=suzuki dbname=hela port=5432 sslmode=disable TimeZone=Asia/Shanghai client_encoding=UTF8",
		},
		Redis: *redis.DefaultConfig(),
		RAG: RAGConfig{
			Enabled:           true,
			EmbeddingProvider: "qwen",
			EmbeddingModel:    "text-embedding-v2",
			EmbeddingAPIKey:   os.Getenv("QWEN_API_KEY"),
			EmbeddingBaseURL:  "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding",
			ChunkSize:         500,
			ChunkOverlap:      50,
			TopK:              5,
			MinScore:          0.7,
		},
	}
}

type TaskConfig struct {
	Workers        int  `json:"workers" env:"WORKERS"`
	MaxRetry       int  `json:"max_retry" env:"MAX_RETRY"`
	TaskPersistant bool `json:"task_persistant" env:"TASK_PERSISTANT"`
}

type TasksConfig struct {
	Download           TaskConfig `json:"download" envPrefix:"DOWNLOAD_"`
	Transfer           TaskConfig `json:"transfer" envPrefix:"TRANSFER_"`
	Upload             TaskConfig `json:"upload" envPrefix:"UPLOAD_"`
	Copy               TaskConfig `json:"copy" envPrefix:"COPY_"`
	Move               TaskConfig `json:"move" envPrefix:"MOVE_"`
	Decompress         TaskConfig `json:"decompress" envPrefix:"DECOMPRESS_"`
	DecompressUpload   TaskConfig `json:"decompress_upload" envPrefix:"DECOMPRESS_UPLOAD_"`
	AllowRetryCanceled bool       `json:"allow_retry_canceled" env:"ALLOW_RETRY_CANCELED"`
}

type RAGConfig struct {
	Enabled           bool    `json:"enabled" env:"ENABLED" envDefault:"true"`
	EmbeddingProvider string  `json:"embedding_provider" env:"EMBEDDING_PROVIDER" envDefault:"qwen"`
	EmbeddingModel    string  `json:"embedding_model" env:"EMBEDDING_MODEL" envDefault:"text-embedding-v2"`
	EmbeddingAPIKey   string  `json:"embedding_api_key" env:"EMBEDDING_API_KEY"`
	EmbeddingBaseURL  string  `json:"embedding_base_url" env:"EMBEDDING_BASE_URL" envDefault:"https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"`
	ChunkSize         int     `json:"chunk_size" env:"CHUNK_SIZE" envDefault:"500"`
	ChunkOverlap      int     `json:"chunk_overlap" env:"CHUNK_OVERLAP" envDefault:"50"`
	TopK              int     `json:"top_k" env:"TOP_K" envDefault:"5"`
	MinScore          float64 `json:"min_score" env:"MIN_SCORE" envDefault:"0.7"`
}

func init() {
	// 实际项目中，这里通常会从文件或环境变量加载配置
	// 此处我们先用默认配置来初始化
	Conf = DefaultConfig("")
}
