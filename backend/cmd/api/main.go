package main

import (
	"feedSystem_video/backend/internal/account"
	"feedSystem_video/backend/internal/config"
	"feedSystem_video/backend/internal/db"
	httpHandler "feedSystem_video/backend/internal/http"
	"feedSystem_video/backend/internal/video"

	"fmt"
)

func main() {
	// 1. 加载配置
	if err := config.Load("configs/config.yaml"); err != nil {
		panic(fmt.Sprintf("配置加载失败: %v", err))
	}
	//初始化数据库
	db.Init()

	//自动建表
	db.DB.AutoMigrate(&account.Account{}, &video.Video{}) // 加上 &video.Video{}

	r := httpHandler.SetupRouter()
	//启动 HTTP 服务
	r.Run(fmt.Sprintf(":%d", config.C.Server.Port))
}
