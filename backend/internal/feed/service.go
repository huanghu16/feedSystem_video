package feed

import (
	"context"
	"feedSystem_video/internal/middleware/redis"
	"fmt"
	"time"
)

// Service Feed 业务逻辑
type Service struct {
	repo *Repo
}

// NewService 创建 Service
func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// 缓存 key 前缀（带版本号，方便以后批量失效）
const feedCachePrefix = "v1:feed:latest"

// ListLatest 获取最新视频列表（带 Redis 缓存）
func (s *Service) ListLatest() ([]FeedVideoItem, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("%s:all", feedCachePrefix)

	// 第一步：先查 Redis 缓存
	var cached []FeedVideoItem                        // 缓存数据
	hit, err := redis.GetJSON(ctx, cacheKey, &cached) // 从 Redis 获取缓存数据
	if err != nil {
		// 缓存读取异常，降级查数据库
		fmt.Println("[Feed] 缓存读取异常，降级查数据库")
	} else if hit {
		// 缓存命中，直接返回
		fmt.Println("[Feed] 缓存命中")
		return cached, nil
	}

	// 第二步：缓存未命中，查数据库
	fmt.Println("[Feed] 缓存未命中，查数据库")
	videos, err := s.repo.ListLatest(20) // 默认返回 20 条
	if err != nil {
		return nil, err
	}

	// 转换成 FeedVideoItem
	items := make([]FeedVideoItem, len(videos))
	for i, v := range videos {
		items[i] = FeedVideoItem{
			ID:         v.ID,         // 视频 ID
			AuthorID:   v.AuthorID,   // 作者 ID
			Username:   v.Username,   // 作者名
			Title:      v.Title,      // 标题
			PlayURL:    v.PlayURL,    // 播放地址
			CoverURL:   v.CoverURL,   // 封面
			LikesCount: v.LikesCount, // 点赞数
		}
	}

	// 第三步：回填缓存（5 分钟过期）
	_ = redis.SetJSON(ctx, cacheKey, items, 5*time.Minute)

	return items, nil
}
