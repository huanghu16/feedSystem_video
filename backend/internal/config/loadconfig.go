package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// AppConfig 是全局配置
// 所有子配置都内嵌在这里，通过 YAML 文件 + 环境变量加载
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	MySQL    MySQLConfig    `yaml:"mysql"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Storage  StorageConfig  `yaml:"storage"`
}

// 服务器配置
type ServerConfig struct {
	Port            int `yaml:"port"`
	ReadTimeoutSec  int `yaml:"read_timeout_sec"`  // HTTP 读超时（秒）
	WriteTimeoutSec int `yaml:"write_timeout_sec"` // HTTP 写超时（秒）
}

// MySQL 配置
type MySQLConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Database       string `yaml:"database"`
	MaxOpenConns   int    `yaml:"max_open_conns"`    // 最大连接数
	MaxIdleConns   int    `yaml:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifeSec int    `yaml:"conn_max_life_sec"` // 连接最大存活时间（秒）
	ConnMaxIdleSec int    `yaml:"conn_max_idle_sec"` // 空闲连接最大存活时间（秒）
	LogLevel       string `yaml:"log_level"`         // 日志级别: silent/error/warn/info
}

// Redis 配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// JWT 配置
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	AccessTTL  int    `yaml:"access_ttl"`
	RefreshTTL int    `yaml:"refresh_ttl"`
}

// RabbitMQ 配置
type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// StorageConfig 文件存储与工具配置
type StorageConfig struct {
	UploadDir  string `yaml:"upload_dir"`  // 上传文件存储目录
	StaticPath string `yaml:"static_path"` // 静态文件 URL 前缀
	FFmpegPath string `yaml:"ffmpeg_path"` // FFmpeg 可执行文件路径，留空则使用 PATH 中的 ffmpeg
}

// 全局配置实例，其他包通过 config.C 访问配置
var C AppConfig

// Load 加载配置文件，优先级：环境变量 > YAML 文件 > 默认值
func Load(configPath string) error {
	//尝试加载 .env 文件
	_ = godotenv.Load()
	//1. 读取文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		setDefaults()
		applyEnvOverrides()
		validate()
		return nil
	}
	//2.解析yaml文件
	if err := yaml.Unmarshal(data, &C); err != nil {
		return err
	}
	//3.覆盖环境变量
	applyEnvOverrides()
	//4.验证配置
	validate()

	return nil

}

// setDefaults 设置默认值（当配置文件不存在时使用）
// 敏感字段（密码、Secret）默认留空，强制通过配置文件或环境变量提供
func setDefaults() {
	C.Server.Port = 8080
	C.Server.ReadTimeoutSec = 30
	C.Server.WriteTimeoutSec = 30

	C.MySQL.Host = "127.0.0.1"
	C.MySQL.Port = 3306
	C.MySQL.User = "root"
	C.MySQL.Password = ""
	C.MySQL.Database = "feed_system_video"
	C.MySQL.MaxOpenConns = 50
	C.MySQL.MaxIdleConns = 10
	C.MySQL.ConnMaxLifeSec = 3600
	C.MySQL.ConnMaxIdleSec = 300
	C.MySQL.LogLevel = "warn"

	C.Redis.Host = "127.0.0.1"
	C.Redis.Port = 6379
	C.Redis.Password = ""
	C.Redis.DB = 0

	C.JWT.Secret = ""
	C.JWT.AccessTTL = 900
	C.JWT.RefreshTTL = 604800

	C.RabbitMQ.Host = "127.0.0.1"
	C.RabbitMQ.Port = 5672
	C.RabbitMQ.User = "guest"
	C.RabbitMQ.Password = "guest"

	C.Storage.UploadDir = "uploads"
	C.Storage.StaticPath = "/static"
	C.Storage.FFmpegPath = "ffmpeg"
}

// applyEnvOverrides 用环境变量覆盖配置（Docker 部署时使用）
func applyEnvOverrides() {
	// 辅助函数：只在环境变量非空时覆盖 string 字段
	overrideStr := func(envVar string, target *string) {
		if value := os.Getenv(envVar); value != "" { //os.Getenv() 读取环境变量
			*target = value
		}
	}

	// 辅助函数：只在环境变量非空时覆盖 int 字段
	overrideInt := func(envVar string, target *int) {
		if value := os.Getenv(envVar); value != "" {
			if num, err := strconv.Atoi(value); err == nil { //strconv.Atoi() 将字符串转换为整数
				*target = num
			}
		}
	}

	// --- Server ---
	overrideInt("SERVER_PORT", &C.Server.Port)
	overrideInt("SERVER_READ_TIMEOUT_SEC", &C.Server.ReadTimeoutSec)
	overrideInt("SERVER_WRITE_TIMEOUT_SEC", &C.Server.WriteTimeoutSec)

	// --- MySQL ---
	overrideStr("MYSQL_HOST", &C.MySQL.Host)
	overrideInt("MYSQL_PORT", &C.MySQL.Port)
	overrideStr("MYSQL_USER", &C.MySQL.User)
	overrideStr("MYSQL_PASSWORD", &C.MySQL.Password)
	overrideStr("MYSQL_DATABASE", &C.MySQL.Database)
	overrideInt("MYSQL_MAX_OPEN_CONNS", &C.MySQL.MaxOpenConns)
	overrideInt("MYSQL_MAX_IDLE_CONNS", &C.MySQL.MaxIdleConns)
	overrideInt("MYSQL_CONN_MAX_LIFE_SEC", &C.MySQL.ConnMaxLifeSec)
	overrideInt("MYSQL_CONN_MAX_IDLE_SEC", &C.MySQL.ConnMaxIdleSec)
	overrideStr("MYSQL_LOG_LEVEL", &C.MySQL.LogLevel)

	// --- Redis ---
	overrideStr("REDIS_HOST", &C.Redis.Host)
	overrideInt("REDIS_PORT", &C.Redis.Port)
	overrideStr("REDIS_PASSWORD", &C.Redis.Password)
	overrideInt("REDIS_DB", &C.Redis.DB)

	// --- JWT ---
	overrideStr("JWT_SECRET", &C.JWT.Secret)
	overrideInt("JWT_ACCESS_TTL", &C.JWT.AccessTTL)
	overrideInt("JWT_REFRESH_TTL", &C.JWT.RefreshTTL)

	// --- RabbitMQ ---
	overrideStr("RABBITMQ_HOST", &C.RabbitMQ.Host)
	overrideInt("RABBITMQ_PORT", &C.RabbitMQ.Port)
	overrideStr("RABBITMQ_USER", &C.RabbitMQ.User)
	overrideStr("RABBITMQ_PASS", &C.RabbitMQ.Password)

	// --- Storage ---
	overrideStr("STORAGE_UPLOAD_DIR", &C.Storage.UploadDir)
	overrideStr("STORAGE_STATIC_PATH", &C.Storage.StaticPath)
	overrideStr("FFMPEG_PATH", &C.Storage.FFmpegPath)
}

// validate 检查必要配置是否已设置
func validate() {
	// JWT Secret 不能为空
	if C.JWT.Secret == "" {
		panic("JWT_SECRET is required (set in config.yaml or env JWT_SECRET)")
	}
	// 确保关键默认值不为零值（防止 yaml 未覆盖字段时出现非法值）
	if C.Server.Port == 0 {
		C.Server.Port = 8080
	}
	if C.MySQL.MaxOpenConns == 0 {
		C.MySQL.MaxOpenConns = 50
	}
	if C.MySQL.MaxIdleConns == 0 {
		C.MySQL.MaxIdleConns = 10
	}
	if C.Storage.UploadDir == "" {
		C.Storage.UploadDir = "uploads"
	}
	if C.Storage.StaticPath == "" {
		C.Storage.StaticPath = "/static"
	}
	if C.Storage.FFmpegPath == "" {
		C.Storage.FFmpegPath = "ffmpeg"
	}
}
