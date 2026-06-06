package rabbitmq

import (
	"context"
	"feedSystem_video/internal/config"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Conn 是全局 RabbitMQ 连接
var Conn *amqp.Connection

// Init 初始化 RabbitMQ 连接
func Init() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", // 用户名、密码、主机、端口
		config.C.RabbitMQ.User,
		config.C.RabbitMQ.Password,
		config.C.RabbitMQ.Host,
		config.C.RabbitMQ.Port,
	)

	var err error
	Conn, err = amqp.Dial(url) // 创建 RabbitMQ 连接
	if err != nil {
		log.Printf("[RabbitMQ] 连接失败: %v（将降级运行）", err)
		return err
	}
	log.Println("[RabbitMQ] 已成功连接")
	return nil
}

// Publish 发布消息到指定 exchange
func Publish(exchange, routingKey string, message interface{}) error {
	if Conn == nil {
		return fmt.Errorf("RabbitMQ 未连接")
	}

	// 每次发布创建新 Channel（Channel 是轻量级的，用完即关）
	ch, err := Conn.Channel() // 创建 Channel
	if err != nil {
		return err
	}
	defer ch.Close() // 关闭 Channel

	// 声明 Exchange（确保存在）
	err = ch.ExchangeDeclare(
		exchange, // 交换机名称
		"topic",  // type：topic 模式支持路由键匹配
		true,     // durable：持久化（RabbitMQ 重启后还在）
		false,    // auto-deleted 是否自动删除
		false,    // internal 是否排他
		false,    // no-wait 是否等待响应
		nil,      // args 额外的参数
	)
	if err != nil {
		return err
	}

	// 发布消息
	return ch.PublishWithContext(
		context.Background(), // context，发送消息的超时时间
		exchange,             // exchange，消息要发送到的交换机
		routingKey,           // routing key ，路由键，用于匹配绑定的队列
		false,                // mandatory，是否强制路由
		false,                // immediate，是否立即投递
		amqp.Publishing{ // 发送的消息
			ContentType: "application/json",
			Body:        []byte(fmt.Sprintf("%v", message)),
		},
	)
}
