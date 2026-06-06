package worker

import (
	"encoding/json"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/video"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartLikeWorker 启动点赞事件消费者
func StartLikeWorker() {

	//创建channel
	ch, err := rabbitmq.Conn.Channel() // 创建 Channel
	if err != nil {
		log.Printf("[LikeWorker] 创建 Channel 失败: %v", err)
		return
	}
	defer ch.Close() // 关闭 Channel

	// 声明 Queue
	q, err := ch.QueueDeclare(
		"like.events.queue", // queue name，队列名称
		true,                // durable ，持久化
		false,               // auto-delete ，队列无消费者时是否自动删除
		false,               // exclusive ，是否排他
		false,               // no-wait ，是否等待响应
		nil,
	)
	if err != nil {
		log.Printf("[LikeWorker] 声明 Queue 失败: %v", err)
		return
	}

	// 绑定 Queue 到 Exchange
	ch.QueueBind(
		q.Name,                // queue name
		"like.*",              // routing key 模式：匹配 like.created 和 like.deleted
		rabbitmq.ExchangeLike, // exchange
		false,                 // no-wait , 是否等待响应
		nil,                   // args ，可选参数
	)

	// 开始消费
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag
		false,  // auto-ack：设为 false，手动确认
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	if err != nil {
		log.Printf("[LikeWorker] 开始消费失败: %v", err)
		return
	}

	log.Println("[LikeWorker] 开始监听点赞事件")

	// 逐条处理消息
	for msg := range msgs {
		processLikeMessage(msg) // 处理点赞消息
		msg.Ack(false)          // 手动确认消息已处理
	}
}

// processLikeMessage 处理单条点赞消息
func processLikeMessage(msg amqp.Delivery) {
	routingKey := msg.RoutingKey // 路由键

	switch routingKey {
	case rabbitmq.RoutingKeyLike:
		// 点赞事件
		var event rabbitmq.LikeEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil { // 解析消息
			log.Printf("[LikeWorker] 解析消息失败: %v", err)
			return
		}
		// 写入数据库
		repo := video.NewRepo()
		if err := repo.CreateLike(event.VideoID, event.AccountID); err != nil { // 写入数据库
			log.Printf("[LikeWorker] 写入点赞失败: %v", err)
			return
		}
		// 更新点赞计数
		_ = repo.IncrementLikesCount(event.VideoID)
		log.Printf("[LikeWorker] 处理点赞: video=%d, account=%d", event.VideoID, event.AccountID)

	case rabbitmq.RoutingKeyUnlike:
		// 取消点赞事件
		var event rabbitmq.LikeEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("[LikeWorker] 解析消息失败: %v", err)
			return
		}
		repo := video.NewRepo()
		_ = repo.DeleteLike(event.VideoID, event.AccountID)
		_ = repo.DecrementLikesCount(event.VideoID)
		log.Printf("[LikeWorker] 处理取消点赞: video=%d, account=%d", event.VideoID, event.AccountID)
	}
}
