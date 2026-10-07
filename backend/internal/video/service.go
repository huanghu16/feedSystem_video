package video

import (
	"context"
	"errors"
	"feedSystem_video/internal/account"
	"feedSystem_video/internal/middleware/feedcache"
	"feedSystem_video/internal/middleware/rabbitmq"
	"feedSystem_video/internal/middleware/redis"
	"fmt"
	"log"
)

// Service 视频业务逻辑
type Service struct {
	repo        *Repo         // 关联数据库
	accountRepo *account.Repo // 关联账号
}

// NewService 创建 Service 实例
func NewService(repo *Repo, accountRepo *account.Repo) *Service {
	return &Service{
		repo:        repo,
		accountRepo: accountRepo,
	}
}

var (
	ErrVideoNotFound = errors.New("视频不存在")
)

// Publish 发布视频
func (s *Service) Publish(req *PublishRequest, authorID uint) (*Video, error) {
	author, err := s.accountRepo.FindByID(authorID)
	if err != nil {
		return nil, err
	}
	if author == nil {
		return nil, errors.New("用户不存在")
	}

	video := &Video{
		AuthorID:    authorID,
		Username:    author.Username,
		Title:       req.Title,
		Description: req.Description,
		PublishDate: req.PublishDate,
		PlayURL:     req.PlayURL,
		CoverURL:    req.CoverURL,
	}

	if err := s.repo.Create(video); err != nil {
		return nil, fmt.Errorf("创建视频失败: %w", err)
	}

	// 失效 Feed 缓存（通过 feedcache 包统一管理 key，避免硬编码不同步）
	feedcache.InvalidateLatestCache()

	return video, nil
}

// ListByAuthor 按作者查询视频列表（带分页）
func (s *Service) ListByAuthor(req *ListByAuthorRequest) (*ListByAuthorResponse, error) {
	page, size := normalizePaging(req.Page, req.Size)

	videos, total, err := s.repo.ListByAuthorID(req.AuthorID, page, size)
	if err != nil {
		return nil, err
	}

	items := make([]VideoItem, 0, len(videos))
	for _, v := range videos {
		items = append(items, VideoToItem(&v))
	}

	hasMore := int64(page*size) < total

	return &ListByAuthorResponse{
		List:    items,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: hasMore,
	}, nil
}

// ==================== 点赞 ====================

// Like 点赞（同步事务写库，保证数据立即持久化）
func (s *Service) Like(videoID, accountID uint) error {
	// 检查是否已赞
	isLiked, err := s.repo.IsLiked(videoID, accountID)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.New("已经点赞过了")
	}

	// 同步事务写库：创建点赞记录 + 增加计数，保证原子性和即时一致性
	if err := s.repo.CreateLikeTx(videoID, accountID); err != nil {
		return fmt.Errorf("点赞失败: %w", err)
	}

	// 发 MQ 事件，通知消费者异步处理（不影响主流程）
	_ = rabbitmq.Publish(rabbitmq.ExchangeLike, rabbitmq.RoutingKeyLike, rabbitmq.LikeEvent{
		VideoID: videoID, AccountID: accountID,
	})

	return nil
}

// Unlike 取消点赞（同步事务写库，保证数据立即持久化）
func (s *Service) Unlike(videoID, accountID uint) error {
	isLiked, err := s.repo.IsLiked(videoID, accountID)
	if err != nil {
		return err
	}
	if !isLiked {
		return errors.New("还没有点赞")
	}

	// 同步事务写库：删除点赞记录 + 减少计数，保证原子性和即时一致性
	if err := s.repo.DeleteLikeTx(videoID, accountID); err != nil {
		return fmt.Errorf("取消点赞失败: %w", err)
	}

	// 发 MQ 事件，通知消费者异步处理
	_ = rabbitmq.Publish(rabbitmq.ExchangeLike, rabbitmq.RoutingKeyUnlike, rabbitmq.LikeEvent{
		VideoID: videoID, AccountID: accountID,
	})

	return nil
}

// IsLiked 查询是否已赞
func (s *Service) IsLiked(videoID, accountID uint) (bool, error) {
	return s.repo.IsLiked(videoID, accountID)
}

// ==================== 评论 ====================

// PublishComment 发布评论
func (s *Service) PublishComment(req *PublishCommentRequest, accountID uint, username string) (*Comment, error) {
	// 检查视频是否存在
	video, err := s.repo.GetByID(req.VideoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}

	comment := &Comment{
		VideoID:   req.VideoID,
		AccountID: accountID,
		Username:  username,
		Content:   req.Content,
	}

	// 使用事务创建评论并增加计数（保证原子性）
	if err := s.repo.CreateCommentTx(comment); err != nil {
		return nil, err
	}

	// 查询评论者头像，填充到返回结果中
	if author, err := s.accountRepo.FindByID(accountID); err == nil && author != nil {
		comment.AvatarURL = author.AvatarURL
	}

	// 发 MQ 事件，通知消费者异步处理
	_ = rabbitmq.Publish(rabbitmq.ExchangeComment, rabbitmq.RoutingKeyComment, rabbitmq.CommentEvent{
		VideoID: req.VideoID, AccountID: accountID, Username: username, Content: req.Content,
	})

	return comment, nil
}

// ListComments 查询评论列表（带分页）
func (s *Service) ListComments(req *ListCommentsRequest) (*ListCommentsResponse, error) {
	page, size := normalizePaging(req.Page, req.Size)

	comments, total, err := s.repo.ListCommentsByVideoID(req.VideoID, page, size)
	if err != nil {
		return nil, err
	}

	items := make([]CommentItem, 0, len(comments))
	for _, c := range comments {
		items = append(items, CommentToItem(&c))
	}

	hasMore := int64(page*size) < total

	return &ListCommentsResponse{
		List:    items,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: hasMore,
	}, nil
}

// RecordPlay 记录视频播放
func (s *Service) RecordPlay(videoID uint) error {
	return s.repo.IncrementPlayCount(videoID)
}

// ListHotVideos 获取热门视频列表（带 Redis 缓存）
// 热榜需要全表算分 + 排序，成本高；而它是"趋势"数据，允许秒级滞后，
// 因此用短 TTL 缓存扛住高频访问
func (s *Service) ListHotVideos(limit int) ([]VideoItem, error) {
	// limit 不同结果不同，必须进 key
	cacheKey := fmt.Sprintf("%s:%d", CacheKeyHot, limit)

	var cached []Video
	if ok, _ := redis.GetJSON(context.Background(), cacheKey, &cached); ok {
		log.Printf("[ListHotVideos] 缓存命中，返回 %d 条记录", len(cached))
		return videosToItems(cached), nil
	}

	videos, err := s.repo.ListHotVideos(limit)
	if err != nil {
		log.Printf("[ListHotVideos] 查询失败: %v", err)
		return nil, err
	}

	if err := redis.SetJSON(context.Background(), cacheKey, videos, CacheTTLHot); err != nil {
		log.Printf("[ListHotVideos] 缓存写入失败（不影响响应）: %v", err)
	}

	log.Printf("[ListHotVideos] 查询成功，返回 %d 条记录", len(videos))
	return videosToItems(videos), nil
}

// videosToItems 批量转换 Video 列表为 VideoItem 列表
func videosToItems(videos []Video) []VideoItem {
	items := make([]VideoItem, 0, len(videos))
	for i := range videos {
		items = append(items, VideoToItem(&videos[i]))
	}
	return items
}

// SearchVideos 搜索视频
func (s *Service) SearchVideos(keyword string, page, size int) (*SearchVideosResponse, error) {
	if keyword == "" {
		return nil, errors.New("搜索关键词不能为空")
	}

	page, size = normalizePaging(page, size)

	videos, total, err := s.repo.SearchVideos(keyword, page, size)
	if err != nil {
		log.Printf("[SearchVideos] 搜索失败: %v", err)
		return nil, err
	}

	items := make([]VideoItem, 0, len(videos))
	for _, v := range videos {
		items = append(items, VideoToItem(&v))
	}

	hasMore := int64(page*size) < total

	return &SearchVideosResponse{
		List:    items,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: hasMore,
	}, nil
}

// DeleteVideo 删除视频
func (s *Service) DeleteVideo(videoID uint, authorID uint) error {
	video, err := s.repo.GetByID(videoID)
	if err != nil {
		return err
	}
	if video == nil {
		return ErrVideoNotFound
	}
	if video.AuthorID != authorID {
		return errors.New("无权限删除此视频")
	}

	if err := s.repo.DeleteVideo(videoID, authorID); err != nil {
		return fmt.Errorf("删除视频失败: %w", err)
	}

	// 失效 Feed 缓存
	feedcache.InvalidateLatestCache()

	return nil
}

// DeleteVideosBatch 批量删除视频
func (s *Service) DeleteVideosBatch(videoIDs []uint, authorID uint) error {
	if len(videoIDs) == 0 {
		return errors.New("请选择要删除的视频")
	}

	if err := s.repo.DeleteVideosBatch(videoIDs, authorID); err != nil {
		return fmt.Errorf("批量删除视频失败: %w", err)
	}

	// 失效 Feed 缓存
	feedcache.InvalidateLatestCache()

	return nil
}

// normalizePaging 规范化分页参数
// page 从 1 开始，size 默认 10，最大 50
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
