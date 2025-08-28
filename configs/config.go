package configs

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
	SiteURL        string      `json:"site_url" env:"SITE_URL"`
	Cdn            string      `json:"cdn" env:"CDN"`
	JwtSecret      string      `json:"jwt_secret" env:"JWT_SECRET"`
	TokenExpiresIn int         `json:"token_expires_in" env:"TOKEN_EXPIRES_IN"`
	Database       Database    `json:"database" envPrefix:"DB_"`
	Tasks          TasksConfig `json:"tasks" envPrefix:"TASKS_"`
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
			DSN:      "host=localhost user=suzuki password=suzuki dbname=hela port=5432 sslmode=disable TimeZone=Asia/Shanghai",
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

func init() {
	// 实际项目中，这里通常会从文件或环境变量加载配置
	// 此处我们先用默认配置来初始化
	Conf = DefaultConfig("")
}
