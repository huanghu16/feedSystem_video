package worker

import (
	"encoding/json"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/social"
	"log"
)

// StartSocialWorker 启动关注事件消费者
func StartSocialWorker() {
	ch, err := rabbitmq.Conn.Channel()
	if err != nil {
		log.Printf("[SocialWorker] 创建 Channel 失败: %v", err)
		return
	}
	defer ch.Close()

	// 声明队列
	q, err := ch.QueueDeclare("social.events.queue", true, false, false, false, nil)
	if err != nil {
		log.Printf("[SocialWorker] 声明 Queue 失败: %v", err)
		return
	}

	// 绑定队列
	ch.QueueBind(q.Name, "social.*", rabbitmq.ExchangeSocial, false, nil)

	// 开始消费
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[SocialWorker] 开始消费失败: %v", err)
		return
	}

	log.Println("[SocialWorker] 开始监听关注事件") // 打印日志

	for msg := range msgs {
		var event rabbitmq.SocialEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil { // 解析消息
			log.Printf("[SocialWorker] 解析消息失败: %v", err)
			msg.Ack(false)
			continue
		}

		repo := social.NewRepo()                                                     // 创建 Repo
		if err := repo.CreateFollow(event.FollowerID, event.VloggerID); err != nil { // 写入关注
			log.Printf("[SocialWorker] 写入关注失败: %v", err)
			msg.Ack(false)
			continue
		}

		log.Printf("[SocialWorker] 处理关注: follower=%d, vlogger=%d", event.FollowerID, event.VloggerID)
		msg.Ack(false)
	}
}
