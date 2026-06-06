package redis

import (
	"context"
	"feedSystem_video/internal/config"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client 是全局 Redis 客户端
var Client *redis.Client

// Init 初始化 Redis
func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.C.Redis.Host, config.C.Redis.Port), // Redis 服务器地址和端口
		Password: config.C.Redis.Password,                                        // Redis 密码
		DB:       config.C.Redis.DB,                                              // Redis 数据库索引
	})

	// 测试连通性（3 秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 创建一个带有超时的上下文
	defer cancel()                                                          // 取消上下文,释放资源

	_, err := Client.Ping(ctx).Result() // Ping Redis 服务器
	if err != nil {
		log.Printf("[Redis] 连接失败: %v（将降级运行，不影响核心功能）", err)
		return
	}

	log.Println("[Redis] connected successfully")
}
