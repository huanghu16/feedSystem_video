package db

import (
	"feedSystem_video/internal/config"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局数据库连接实例
var DB *gorm.DB

// Init 初始化数据库连接
func Init() {
	//构造DSN,用来创建数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		config.C.MySQL.User,
		config.C.MySQL.Password,
		config.C.MySQL.Host,
		config.C.MySQL.Port,
		config.C.MySQL.Database,
	)

	// 根据配置解析日志级别
	logLevel := parseLogLevel(config.C.MySQL.LogLevel)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("[DB] 连接失败: %v", err)
	}

	// 获取底层 *sql.DB 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("[DB] 获取底层连接失败: %v", err)
	}

	// 配置连接池参数
	sqlDB.SetMaxOpenConns(config.C.MySQL.MaxOpenConns)                                   // 最大连接数
	sqlDB.SetMaxIdleConns(config.C.MySQL.MaxIdleConns)                                   // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Duration(config.C.MySQL.ConnMaxLifeSec) * time.Second) // 连接最大存活时间
	sqlDB.SetConnMaxIdleTime(time.Duration(config.C.MySQL.ConnMaxIdleSec) * time.Second) // 空闲连接最大存活时间

	// 测试连通性
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("[DB] Ping 失败: %v", err)
	}

	log.Printf("[DB] MySQL connected successfully (pool: maxOpen=%d, maxIdle=%d, maxLife=%ds, maxIdleTime=%ds)",
		config.C.MySQL.MaxOpenConns, config.C.MySQL.MaxIdleConns,
		config.C.MySQL.ConnMaxLifeSec, config.C.MySQL.ConnMaxIdleSec)
}

// parseLogLevel 将字符串日志级别转换为 gorm logger.Level
func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	case "warn", "":
		return logger.Warn
	default:
		return logger.Warn
	}
}
