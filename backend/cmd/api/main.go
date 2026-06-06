package main

import (
	"feedSystem_video/internal/account"
	"feedSystem_video/internal/config"
	"feedSystem_video/internal/db"
	httpHandler "feedSystem_video/internal/http"
	"feedSystem_video/internal/middleware/redis" //新增
	"feedSystem_video/internal/social"
	"feedSystem_video/internal/video"

	"fmt"
)

func main() {
	// 1. 加载配置
	if err := config.Load("configs/config.yaml"); err != nil {
		panic(fmt.Sprintf("配置加载失败: %v", err))
	}
	//初始化数据库
	db.Init()
	//初始化 Redis
	redis.Init()

	//自动建表
	db.DB.AutoMigrate(
		&account.Account{},
		&video.Video{},
		&video.Like{},
		&video.Comment{},
		&social.Social{}, // 新增
	)

	r := httpHandler.SetupRouter()
	//启动 HTTP 服务
	r.Run(fmt.Sprintf(":%d", config.C.Server.Port))
}
