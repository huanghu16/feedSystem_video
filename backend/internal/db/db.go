package db

import (
	"feedSystem_video/internal/config"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局数据库连接实例
var DB *gorm.DB

// Init 初始化数据库连接
func Init() {
	//构造DSN,用来创建数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.C.MySQL.User,
		config.C.MySQL.Password,
		config.C.MySQL.Host,
		config.C.MySQL.Port,
		config.C.MySQL.Database,
	)
	// 连接数据库（开发环境打印 SQL）
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("[DB] 连接失败: %v", err)
	}

	// 测试连通性
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("[DB] 获取底层连接失败: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("[DB] Ping 失败: %v", err)
	}

	log.Println("[DB] MySQL connected successfully")
}
