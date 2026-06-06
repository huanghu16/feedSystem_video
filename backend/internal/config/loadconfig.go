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
}

// 服务器配置
type ServerConfig struct {
	Port int `yaml:"port"`
}

// MySQL 配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
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

// setDefaults 设置硬编码的默认值（当配置文件不存在时使用）
func setDefaults() {
	C.Server.Port = 8080

	C.MySQL.Host = "127.0.0.1"
	C.MySQL.Port = 3308
	C.MySQL.User = "root"
	C.MySQL.Password = "123456"
	C.MySQL.Database = "feed_system_video"

	C.Redis.Host = "127.0.0.1"
	C.Redis.Port = 6379
	C.Redis.Password = ""
	C.Redis.DB = 0

	C.JWT.Secret = "feedSystem-dev-secret-key"
	C.JWT.AccessTTL = 900
	C.JWT.RefreshTTL = 604800

	C.RabbitMQ.Host = "127.0.0.1"
	C.RabbitMQ.Port = 5672
	C.RabbitMQ.User = "guest"
	C.RabbitMQ.Password = "guest"
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
	// 覆盖配置
	overrideInt("SERVER_PORT", &C.Server.Port)
	overrideStr("MYSQL_HOST", &C.MySQL.Host)
	overrideInt("MYSQL_PORT", &C.MySQL.Port)
	overrideStr("MYSQL_USER", &C.MySQL.User)
	overrideStr("MYSQL_PASSWORD", &C.MySQL.Password)
	overrideStr("MYSQL_DATABASE", &C.MySQL.Database)
	overrideStr("REDIS_HOST", &C.Redis.Host)
	overrideInt("REDIS_PORT", &C.Redis.Port)
	overrideStr("JWT_SECRET", &C.JWT.Secret)
	overrideStr("RABBITMQ_HOST", &C.RabbitMQ.Host)
	overrideInt("RABBITMQ_PORT", &C.RabbitMQ.Port)
	overrideStr("RABBITMQ_USER", &C.RabbitMQ.User)
	overrideStr("RABBITMQ_PASS", &C.RabbitMQ.Password)
}

// validate 检查必要配置是否已设置
func validate() {
	// JWT Secret 不能为空
	if C.JWT.Secret == "" {
		panic("JWT_SECRET is required")
	}
}
