package feed

import (
	"context"
	"feedSystem_video/internal/middleware/feedcache"
	"feedSystem_video/internal/middleware/redis"
	"feedSystem_video/internal/video"
	"log"
)

// Service Feed 业务逻辑
type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// ListLatest 获取最新视频列表（带分页 + Redis 缓存）
// 缓存策略：仅缓存第一页（page=1）的默认条数，其他页直接查库
func (s *Service) ListLatest(req *ListLatestRequest) (*ListLatestResponse, error) {
	page, size := normalizePaging(req.Page, req.Size)

	// 仅缓存第一页
	if page == 1 {
		// 先尝试读缓存
		var cached []video.Video
		if ok, _ := redis.GetJSON(context.Background(), feedcache.CacheKeyLatest, &cached); ok {
			log.Printf("[Feed] 缓存命中，返回 %d 条视频", len(cached))
			total := int64(len(cached))
			items := videosToItems(cached)
			return &ListLatestResponse{
				List:    items,
				Total:   total,
				Page:    page,
				Size:    size,
				HasMore: total > int64(size),
			}, nil
		}
	}

	// 查库
	videos, total, err := s.repo.ListLatest(page, size)
	if err != nil {
		log.Printf("[Feed] 查询最新视频失败: %v", err)
		return nil, err
	}

	items := videosToItems(videos)

	// 仅缓存第一页
	if page == 1 {
		if err := redis.SetJSON(context.Background(), feedcache.CacheKeyLatest, videos, feedcache.CacheTTL); err != nil {
			log.Printf("[Feed] 缓存写入失败（不影响响应）: %v", err)
		}
	}

	hasMore := int64(page*size) < total

	return &ListLatestResponse{
		List:    items,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: hasMore,
	}, nil
}

// videosToItems 批量转换 Video 列表为 FeedVideoItem 列表
func videosToItems(videos []video.Video) []FeedVideoItem {
	items := make([]FeedVideoItem, 0, len(videos))
	for i := range videos {
		items = append(items, video.VideoToItem(&videos[i]))
	}
	return items
}

// normalizePaging 规范化分页参数
func normalizePaging(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 50 {
		size = 50
	}
	return page, size
}
