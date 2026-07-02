package rabbitmq

import (
	"context"
	"encoding/json"
	"feedSystem_video/internal/config"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Conn 是全局 RabbitMQ 连接
var Conn *amqp.Connection

// connected 标记连接是否可用（原子操作，供 Publish 快速判断）
var connected atomic.Bool

var (
	mu sync.Mutex // 保护重连过程
)

// Init 初始化 RabbitMQ 连接，失败返回 error（调用方可决定降级或重试）
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		config.C.RabbitMQ.User,
		config.C.RabbitMQ.Password,
		config.C.RabbitMQ.Host,
		config.C.RabbitMQ.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[RabbitMQ] 连接失败: %v（将降级运行）", err)
		return err
	}

	Conn = conn
	connected.Store(true)
	log.Println("[RabbitMQ] 已成功连接")

	// 启动后台 goroutine 监听连接关闭，自动重连
	go watchConnection()

	return nil
}

// watchConnection 监听连接关闭事件，自动重连
func watchConnection() {
	closeCh := Conn.NotifyClose(make(chan *amqp.Error, 1))
	for range closeCh {
		connected.Store(false)
		log.Println("[RabbitMQ] 连接已断开，开始自动重连...")
		reconnectLoop()
		// 重连成功后重新监听新连接的关闭事件
		closeCh = Conn.NotifyClose(make(chan *amqp.Error, 1))
	}
}

// reconnectLoop 持续重连直到成功
func reconnectLoop() {
	const maxInterval = 30 * time.Second
	interval := 2 * time.Second

	for {
		mu.Lock()
		url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
			config.C.RabbitMQ.User,
			config.C.RabbitMQ.Password,
			config.C.RabbitMQ.Host,
			config.C.RabbitMQ.Port,
		)
		conn, err := amqp.Dial(url)
		mu.Unlock()

		if err != nil {
			log.Printf("[RabbitMQ] 重连失败: %v，%v 后重试", err, interval)
			time.Sleep(interval)
			// 指数退避，上限 30 秒
			interval = time.Duration(float64(interval) * 1.5)
			if interval > maxInterval {
				interval = maxInterval
			}
			continue
		}

		Conn = conn
		connected.Store(true)
		log.Println("[RabbitMQ] 重连成功")
		return
	}
}

// IsConnected 返回当前连接是否可用
func IsConnected() bool {
	return connected.Load()
}

// Publish 发布消息到指定 exchange
// message 会被自动 json.Marshal，调用方可直接传入 struct
func Publish(exchange, routingKey string, message interface{}) error {
	if !connected.Load() || Conn == nil {
		return fmt.Errorf("RabbitMQ 未连接")
	}

	// 序列化消息为 JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %w", err)
	}

	// 每次发布创建新 Channel（Channel 是轻量级的，用完即关）
	ch, err := Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

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

	// 发布消息（设置 DeliveryMode=2 持久化，防止 RabbitMQ 重启丢消息）
	return ch.PublishWithContext(
		context.Background(),
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 消息持久化
		},
	)
}
