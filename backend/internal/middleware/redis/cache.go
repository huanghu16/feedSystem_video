package redis

import (
	"context"
	"encoding/json"
	"time"
)

// SetJSON 将对象序列化为 JSON 存入 Redis
func SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if Client == nil {
		return nil // Redis 不可用时静默跳过
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, data, ttl).Err()
}

// GetJSON 从 Redis 读取 JSON 并反序列化到目标对象
// 返回 true 表示命中缓存，false 表示未命中
func GetJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	if Client == nil {
		return false, nil // Redis 不可用时视为未命中
	}

	data, err := Client.Get(ctx, key).Bytes() // 从 Redis 获取缓存数据
	if err != nil {
		// redis.Nil 表示 key 不存在，不是真正的错误
		return false, nil
	}

	if err := json.Unmarshal(data, dest); err != nil { // 反序列化数据
		return false, err
	}

	return true, nil
}

// Del 删除缓存 key
func Del(ctx context.Context, keys ...string) error {
	if Client == nil {
		return nil
	}
	return Client.Del(ctx, keys...).Err()
}
