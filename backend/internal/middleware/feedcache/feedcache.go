package feedcache

import (
	"context"
	"feedSystem_video/internal/middleware/redis"
	"log"
	"time"
)

// 缓存相关常量，统一管理缓存 key
// video 包和 feed 包都通过此包访问缓存 key，避免不同步
const (
	CacheKeyLatest = "v1:feed:latest:all"
	CacheTTL       = 5 * time.Minute
)

// InvalidateLatestCache 失效最新视频缓存
// 供 video 包在发布/删除视频时调用，避免缓存 key 硬编码不同步
func InvalidateLatestCache() {
	if err := redis.Del(context.Background(), CacheKeyLatest); err != nil {
		log.Printf("[FeedCache] 失效缓存失败: %v", err)
	}
}
