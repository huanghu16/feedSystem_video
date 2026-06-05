package main

import (
	"feedSystem_video/backend/internal/apierror"
	"feedSystem_video/backend/internal/config"
	"feedSystem_video/backend/internal/db"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	if err := config.Load("configs/config.yaml"); err != nil {
		panic(fmt.Sprintf("配置加载失败: %v", err))
	}
	// 2. 初始化数据库
	db.Init()
	// 3. 创建 Gin 引擎
	r := gin.Default()
	// 4. 健康检查路由
	r.GET("/healthz", func(c *gin.Context) {
		apierror.OK(c, gin.H{"status": "ok"})
	})
	// 5. 启动 HTTP 服务
	r.Run(fmt.Sprintf(":%d", config.C.Server.Port))
}
