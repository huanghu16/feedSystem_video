package main

import (
	"feedSystem_video/internal/config"
	"feedSystem_video/internal/db"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/worker"
	"fmt"
	"log"
	"time"
)

func main() {
	// 加载配置
	if err := config.Load("configs/config.yaml"); err != nil {
		panic(fmt.Sprintf("配置加载失败: %v", err))
	}

	db.Init() //初始化数据库

	// 连接 RabbitMQ（失败则退出，Worker 没有 MQ 就没意义）
	if err := rabbitmq.Init(); err != nil {
		log.Println("[Worker] RabbitMQ 未连接，5 秒后重试...")
		for {
			if err := rabbitmq.Init(); err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}

	log.Println("[Worker] 启动成功，开始消费消息")

	// 启动各 Worker
	go worker.StartLikeWorker()
	go worker.StartCommentWorker()
	go worker.StartSocialWorker()

	// 阻塞主 goroutine
	select {}
}
