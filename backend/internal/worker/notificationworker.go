package worker

import (
	"encoding/json"
	"fmt"
	"log"

	"feedSystem_video/internal/account"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/notification"
	"feedSystem_video/internal/video"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartNotificationWorker 启动通知事件消费者
// 监听点赞、评论、关注三类事件，生成通知记录
// ACK 策略与 LikeWorker 一致：解析失败 Ack 丢弃，写库失败 Nack 不重投
func StartNotificationWorker() {
	ch, err := rabbitmq.Conn.Channel()
	if err != nil {
		log.Printf("[NotificationWorker] 创建 Channel 失败: %v", err)
		return
	}
	defer ch.Close()

	// 声明通知队列
	q, err := ch.QueueDeclare(
		"notification.events.queue", // queue name
		true,                        // durable
		false,                       // auto-delete
		false,                       // exclusive
		false,                       // no-wait
		nil,
	)
	if err != nil {
		log.Printf("[NotificationWorker] 声明 Queue 失败: %v", err)
		return
	}

	// 绑定到三个 Exchange，接收所有事件
	for _, bind := range []struct {
		exchange   string
		routingKey string
	}{
		{rabbitmq.ExchangeLike, "like.*"},
		{rabbitmq.ExchangeComment, "comment.*"},
		{rabbitmq.ExchangeSocial, "social.*"},
	} {
		if err := ch.QueueBind(q.Name, bind.routingKey, bind.exchange, false, nil); err != nil {
			log.Printf("[NotificationWorker] 绑定 %s 失败: %v", bind.exchange, err)
			return
		}
	}

	// 每次只推送 1 条未确认消息
	if err := ch.Qos(1, 0, false); err != nil {
		log.Printf("[NotificationWorker] 设置 QoS 失败: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[NotificationWorker] 开始消费失败: %v", err)
		return
	}

	log.Println("[NotificationWorker] 开始监听通知事件")

	for msg := range msgs {
		ok := processNotificationMessage(msg)
		if ok {
			_ = msg.Ack(false)
		} else {
			_ = msg.Nack(false, false)
		}
	}

	log.Println("[NotificationWorker] 消费通道已关闭")
}

// processNotificationMessage 处理单条事件消息，生成通知记录
func processNotificationMessage(msg amqp.Delivery) bool {
	notifRepo := notification.NewRepo()

	switch msg.RoutingKey {

	// ---- 点赞事件 ----
	case rabbitmq.RoutingKeyLike:
		var event rabbitmq.LikeEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("[NotificationWorker] 解析点赞消息失败（丢弃）: %v", err)
			return true
		}

		// 查视频作者（通知收件人）
		videoRepo := video.NewRepo()
		v, err := videoRepo.GetByID(event.VideoID)
		if err != nil || v == nil {
			log.Printf("[NotificationWorker] 查询视频失败: %v", err)
			return true // 视频不存在则丢弃，不阻塞队列
		}

		// 不给自己发通知
		if v.AuthorID == event.AccountID {
			return true
		}

		// 查操作者用户名
		accountRepo := account.NewRepo()
		actor, _ := accountRepo.FindByID(event.AccountID)
		actorName := "未知用户"
		if actor != nil {
			actorName = actor.Username
		}

		n := &notification.Notification{
			RecipientID: v.AuthorID,
			ActorID:     event.AccountID,
			ActorName:   actorName,
			Type:        notification.TypeLike,
			VideoID:     event.VideoID,
			Content:     fmt.Sprintf("%s 赞了你的视频", actorName),
		}
		if err := notifRepo.Create(n); err != nil {
			log.Printf("[NotificationWorker] 写入点赞通知失败: %v", err)
			return false
		}
		log.Printf("[NotificationWorker] 点赞通知: %s → user %d", actorName, v.AuthorID)

	// ---- 评论事件 ----
	case rabbitmq.RoutingKeyComment:
		var event rabbitmq.CommentEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("[NotificationWorker] 解析评论消息失败（丢弃）: %v", err)
			return true
		}

		videoRepo := video.NewRepo()
		v, err := videoRepo.GetByID(event.VideoID)
		if err != nil || v == nil {
			log.Printf("[NotificationWorker] 查询视频失败: %v", err)
			return true
		}

		if v.AuthorID == event.AccountID {
			return true
		}

		// 截取评论内容（最多 50 字）
		content := event.Content
		if len(content) > 50 {
			content = content[:50] + "..."
		}

		n := &notification.Notification{
			RecipientID: v.AuthorID,
			ActorID:     event.AccountID,
			ActorName:   event.Username,
			Type:        notification.TypeComment,
			VideoID:     event.VideoID,
			Content:     fmt.Sprintf("%s 评论了你的视频: %s", event.Username, content),
		}
		if err := notifRepo.Create(n); err != nil {
			log.Printf("[NotificationWorker] 写入评论通知失败: %v", err)
			return false
		}
		log.Printf("[NotificationWorker] 评论通知: %s → user %d", event.Username, v.AuthorID)

	// ---- 关注事件 ----
	case rabbitmq.RoutingKeyFollow:
		var event rabbitmq.SocialEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("[NotificationWorker] 解析关注消息失败（丢弃）: %v", err)
			return true
		}

		accountRepo := account.NewRepo()
		actor, _ := accountRepo.FindByID(event.FollowerID)
		actorName := "未知用户"
		if actor != nil {
			actorName = actor.Username
		}

		n := &notification.Notification{
			RecipientID: event.VloggerID,
			ActorID:     event.FollowerID,
			ActorName:   actorName,
			Type:        notification.TypeFollow,
			Content:     fmt.Sprintf("%s 关注了你", actorName),
		}
		if err := notifRepo.Create(n); err != nil {
			log.Printf("[NotificationWorker] 写入关注通知失败: %v", err)
			return false
		}
		log.Printf("[NotificationWorker] 关注通知: %s → user %d", actorName, event.VloggerID)

	default:
		log.Printf("[NotificationWorker] 未知路由键（丢弃）: %s", msg.RoutingKey)
		return true
	}

	return true
}
