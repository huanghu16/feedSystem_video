package main

import (
	"feedSystem_video/internal/config"
	"feedSystem_video/internal/db"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/worker"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 加载配置
	if err := config.Load("configs/config.yaml"); err != nil {
		panic(fmt.Sprintf("配置加载失败: %v", err))
	}

	db.Init() //初始化数据库

	// 连接 RabbitMQ（失败则重试，Worker 没有 MQ 就没意义）
	for {
		if err := rabbitmq.Init(); err == nil {
			break
		}
		log.Println("[Worker] RabbitMQ 未连接，5 秒后重试...")
		time.Sleep(5 * time.Second)
	}

	log.Println("[Worker] 启动成功，开始消费消息")

	// 启动 LikeWorker（CommentWorker 和 SocialWorker 已移除：service 层为同步写库，无需异步消费）
	go worker.StartLikeWorker()

	// 等待中断信号（Ctrl+C 或 SIGTERM），优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[Worker] 收到信号 %v，开始优雅关闭...", sig)

	// 关闭 RabbitMQ 连接（Channel 的 defer Close 会在 Conn 关闭后触发）
	if rabbitmq.Conn != nil {
		if err := rabbitmq.Conn.Close(); err != nil {
			log.Printf("[Worker] 关闭 RabbitMQ 连接失败: %v", err)
		}
	}

	// 关闭数据库连接
	if sqlDB, err := db.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}

	log.Println("[Worker] 已优雅关闭")
}
