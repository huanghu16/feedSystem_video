package worker

import (
	"encoding/json"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/video"
	"log"
)

// StartCommentWorker 启动评论事件消费者
func StartCommentWorker() {
	ch, err := rabbitmq.Conn.Channel() // 创建 Channel
	if err != nil {
		log.Printf("[CommentWorker] 创建 Channel 失败: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("comment.events.queue", true, false, false, false, nil) // 声明 Queue
	if err != nil {
		log.Printf("[CommentWorker] 声明 Queue 失败: %v", err)
		return
	}

	ch.QueueBind(q.Name, "comment.*", rabbitmq.ExchangeComment, false, nil) // 绑定 Queue

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil) // 开始消费
	if err != nil {
		log.Printf("[CommentWorker] 开始消费失败: %v", err)
		return
	}

	log.Println("[CommentWorker] 开始监听评论事件")

	for msg := range msgs {
		var event rabbitmq.CommentEvent // 解析消息
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("[CommentWorker] 解析消息失败: %v", err)
			msg.Ack(false) // 确认消息
			continue
		}

		repo := video.NewRepo() // 创建 Repo
		comment := &video.Comment{
			VideoID:   event.VideoID,
			AccountID: event.AccountID,
			Username:  event.Username,
			Content:   event.Content, // 评论内容
		}
		if err := repo.CreateComment(comment); err != nil {
			log.Printf("[CommentWorker] 写入评论失败: %v", err)
			msg.Ack(false)
			continue
		}

		log.Printf("[CommentWorker] 处理评论: video=%d, user=%s", event.VideoID, event.Username)
		msg.Ack(false)
	}
}
