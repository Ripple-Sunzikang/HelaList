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
	Database Database `json:"database" envPrefix:"DB_"`
}

func DefaultConfig(dataDir string) *Config {
	return &Config{
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

func init() {
	// 实际项目中，这里通常会从文件或环境变量加载配置
	// 此处我们先用默认配置来初始化
	Conf = DefaultConfig("")
}
