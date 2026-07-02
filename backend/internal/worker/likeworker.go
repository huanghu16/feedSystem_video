package worker

import (
	"encoding/json"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/video"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartLikeWorker 启动点赞事件消费者
// ACK 策略：
//   - 消息解析失败（格式错误）：Ack 丢弃无效消息
//   - 写库失败（DB 故障）：Nack 不重投（requeue=false），避免无限循环；后续可接入死信队列
//   - 写库成功：Ack 确认
func StartLikeWorker() {
	ch, err := rabbitmq.Conn.Channel()
	if err != nil {
		log.Printf("[LikeWorker] 创建 Channel 失败: %v", err)
		return
	}
	defer ch.Close()

	// 声明 Queue
	q, err := ch.QueueDeclare(
		"like.events.queue", // queue name
		true,                // durable
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,
	)
	if err != nil {
		log.Printf("[LikeWorker] 声明 Queue 失败: %v", err)
		return
	}

	// 绑定 Queue 到 Exchange
	if err := ch.QueueBind(
		q.Name,                // queue name
		"like.*",              // routing key 模式
		rabbitmq.ExchangeLike, // exchange
		false,                 // no-wait
		nil,                   // args
	); err != nil {
		log.Printf("[LikeWorker] 绑定 Queue 失败: %v", err)
		return
	}

	// 设置 QoS：每次只推送 1 条未确认消息，避免堆积
	if err := ch.Qos(1, 0, false); err != nil {
		log.Printf("[LikeWorker] 设置 QoS 失败: %v", err)
	}

	// 开始消费
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag
		false,  // auto-ack：手动确认
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
		ok := processLikeMessage(msg)
		if ok {
			// 处理成功，确认消息
			if err := msg.Ack(false); err != nil {
				log.Printf("[LikeWorker] Ack 失败: %v", err)
			}
		} else {
			// 处理失败，Nack 不重投（避免无限循环，后续可接入死信队列）
			if err := msg.Nack(false, false); err != nil {
				log.Printf("[LikeWorker] Nack 失败: %v", err)
			}
		}
	}

	log.Println("[LikeWorker] 消费通道已关闭，停止消费")
}

// processLikeMessage 处理单条点赞消息，返回 true 表示成功（应 Ack），false 表示失败（应 Nack）
func processLikeMessage(msg amqp.Delivery) bool {
	routingKey := msg.RoutingKey

	switch routingKey {
	case rabbitmq.RoutingKeyLike:
		// 点赞事件
		var event rabbitmq.LikeEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			// 解析失败：消息格式错误，丢弃（Ack）
			log.Printf("[LikeWorker] 解析点赞消息失败（丢弃）: %v", err)
			return true
		}
		// 使用事务写库：创建点赞记录 + 增加计数，保证原子性
		repo := video.NewRepo()
		if err := repo.CreateLikeTx(event.VideoID, event.AccountID); err != nil {
			log.Printf("[LikeWorker] 写入点赞失败: %v", err)
			return false
		}
		log.Printf("[LikeWorker] 处理点赞: video=%d, account=%d", event.VideoID, event.AccountID)

	case rabbitmq.RoutingKeyUnlike:
		// 取消点赞事件
		var event rabbitmq.LikeEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("[LikeWorker] 解析取消点赞消息失败（丢弃）: %v", err)
			return true
		}
		// 使用事务写库：删除点赞记录 + 减少计数，保证原子性
		repo := video.NewRepo()
		if err := repo.DeleteLikeTx(event.VideoID, event.AccountID); err != nil {
			log.Printf("[LikeWorker] 删除点赞失败: %v", err)
			return false
		}
		log.Printf("[LikeWorker] 处理取消点赞: video=%d, account=%d", event.VideoID, event.AccountID)

	default:
		// 未知路由键，丢弃
		log.Printf("[LikeWorker] 未知路由键（丢弃）: %s", routingKey)
		return true
	}

	return true
}
