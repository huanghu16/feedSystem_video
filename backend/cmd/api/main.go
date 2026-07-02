package main

import (
	"context"
	"feedSystem_video/internal/account"
	"feedSystem_video/internal/config"
	"feedSystem_video/internal/db"
	httpHandler "feedSystem_video/internal/http"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/middleware/redis"
	"feedSystem_video/internal/social"
	"feedSystem_video/internal/video"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	// 初始化RabbitMQ 连接（失败不阻塞，降级运行）
	rabbitmq.Init()

	//自动建表
	db.DB.AutoMigrate(
		&account.Account{},
		&video.Video{},
		&video.Like{},
		&video.Comment{},
		&social.Social{},
	)

	r := httpHandler.SetupRouter()

	// 使用 http.Server 支持优雅关闭和超时配置
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.C.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(config.C.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(config.C.Server.WriteTimeoutSec) * time.Second,
	}

	// 在 goroutine 中启动 HTTP 服务
	go func() {
		log.Printf("[API] 服务启动，监听端口 %d", config.C.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[API] 服务启动失败: %v", err)
		}
	}()

	// 等待中断信号（Ctrl+C 或 SIGTERM）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[API] 收到信号 %v，开始优雅关闭...", sig)

	// 给正在处理的请求 15 秒的时间完成
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[API] 优雅关闭超时或出错: %v", err)
	}

	// 关闭 RabbitMQ 连接
	if rabbitmq.Conn != nil {
		_ = rabbitmq.Conn.Close()
	}

	// 关闭数据库连接池
	if sqlDB, err := db.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}

	log.Println("[API] 已优雅关闭")
}
