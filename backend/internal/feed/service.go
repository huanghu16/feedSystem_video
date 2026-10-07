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

// latestCachePayload 首页缓存载荷
// 只缓存"列表结构 + 总数"，计数在返回前用实时值覆盖（见 applyRealtimeCounts）
type latestCachePayload struct {
	Videos []video.Video `json:"videos"`
	Total  int64         `json:"total"`
	Size   int           `json:"size"` // 生成该缓存时的每页条数
}

// ListLatest 获取最新视频列表（带分页 + Redis 缓存）
// 缓存策略：
//  1. 只缓存第一页
//  2. 缓存的是"列表结构"，计数每次用实时值覆盖
//     —— 计数变化频率极高，不能跟着列表缓存一起等到过期
func (s *Service) ListLatest(req *ListLatestRequest) (*ListLatestResponse, error) {
	page, size := normalizePaging(req.Page, req.Size)

	if page == 1 {
		var payload latestCachePayload
		// payload.Size >= size 才可用：缓存可能按更小的 size 生成，条数不够
		if ok, _ := redis.GetJSON(context.Background(), feedcache.CacheKeyLatest, &payload); ok && payload.Size >= size {
			videos := payload.Videos
			if len(videos) > size {
				videos = videos[:size]
			}
			items := videosToItems(videos)

			// 关键：用实时计数覆盖缓存里的计数快照
			s.applyRealtimeCounts(items)

			log.Printf("[Feed] 缓存命中，返回 %d 条视频", len(items))
			return &ListLatestResponse{
				List:    items,
				Total:   payload.Total,
				Page:    page,
				Size:    size,
				HasMore: payload.Total > int64(page*size),
			}, nil
		}
	}

	// 查库（此时计数本身就是实时值，无需覆盖）
	videos, total, err := s.repo.ListLatest(page, size)
	if err != nil {
		log.Printf("[Feed] 查询最新视频失败: %v", err)
		return nil, err
	}

	items := videosToItems(videos)

	// 仅缓存第一页
	if page == 1 {
		payload := latestCachePayload{Videos: videos, Total: total, Size: size}
		if err := redis.SetJSON(context.Background(), feedcache.CacheKeyLatest, payload, feedcache.CacheTTL); err != nil {
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

// applyRealtimeCounts 用数据库里的实时计数覆盖列表上的计数快照
// 列表结构可以缓存，但计数变化频率极高，必须每次取实时值
func (s *Service) applyRealtimeCounts(items []FeedVideoItem) {
	if len(items) == 0 {
		return
	}

	ids := make([]uint, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}

	counts, err := s.repo.CountsByIDs(ids)
	if err != nil {
		// 降级：查询失败时保留缓存里的计数快照，不影响接口可用性
		log.Printf("[Feed] 查询实时计数失败，降级使用缓存快照: %v", err)
		return
	}

	for i := range items {
		if c, ok := counts[items[i].ID]; ok {
			items[i].LikesCount = c.LikesCount
			items[i].CommentsCount = c.CommentsCount
			items[i].PlayCount = c.PlayCount
		}
	}
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
