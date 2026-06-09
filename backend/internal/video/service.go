package video

import (
	"context"
	"encoding/json"
	"errors"
	"feedSystem_video/internal/account"
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
	// 查作者信息（获取 username）
	author, err := s.accountRepo.FindByID(authorID)
	if err != nil {
		return nil, err
	}
	if author == nil {
		return nil, errors.New("用户不存在")
	}

	video := &Video{
		AuthorID: authorID,
		Username: author.Username,
		Title:    req.Title,
		PlayURL:  req.PlayURL,  // 播放地址
		CoverURL: req.CoverURL, // 封面地址
	}

	// 保存视频
	if err := s.repo.Create(video); err != nil {
		return nil, fmt.Errorf("创建视频失败: %w", err)
	}

	// 清除 Feed 缓存，让下次查询时重新从数据库加载
	_ = redis.Del(context.Background(), "v1:feed:latest:all")

	return video, nil
}

// ListByAuthor 按作者查询视频列表
func (s *Service) ListByAuthor(authorID uint) ([]VideoItem, error) {
	videos, err := s.repo.ListByAuthorID(authorID)
	if err != nil {
		return nil, err
	}

	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		items[i] = VideoItem{
			ID:            v.ID,
			AuthorID:      v.AuthorID, // 作者ID
			Username:      v.Username,
			Title:         v.Title,
			PlayURL:       v.PlayURL,       // 播放地址
			CoverURL:      v.CoverURL,      // 封面地址
			LikesCount:    v.LikesCount,    // 点赞数
			PlayCount:     v.PlayCount,     // 播放数
			CommentsCount: v.CommentsCount, // 评论数
			CreatedAt:     v.CreatedAt,     // 创建时间
		}
	}
	return items, nil
}

// ==================== 点赞 ====================

// Like 点赞（发 MQ 异步处理）
func (s *Service) Like(videoID, accountID uint) error {
	// 检查是否已赞
	isLiked, err := s.repo.IsLiked(videoID, accountID)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.New("已经点赞过了")
	}

	// 发 MQ 消息（异步）
	event := rabbitmq.LikeEvent{VideoID: videoID, AccountID: accountID}
	eventJSON, _ := json.Marshal(event)

	err = rabbitmq.Publish(rabbitmq.ExchangeLike, rabbitmq.RoutingKeyLike, string(eventJSON))
	if err != nil {
		// MQ 发送失败，降级为同步写库
		log.Printf("[Like] MQ 发送失败，降级同步写库: %v", err)
		return s.repo.CreateLike(videoID, accountID)
	}

	return nil
}

// Unlike 取消点赞（发 MQ 异步处理）
func (s *Service) Unlike(videoID, accountID uint) error {
	isLiked, err := s.repo.IsLiked(videoID, accountID)
	if err != nil {
		return err
	}
	if !isLiked {
		return errors.New("还没有点赞")
	}

	event := rabbitmq.LikeEvent{VideoID: videoID, AccountID: accountID}
	eventJSON, _ := json.Marshal(event)

	err = rabbitmq.Publish(rabbitmq.ExchangeLike, rabbitmq.RoutingKeyUnlike, string(eventJSON))
	if err != nil {
		log.Printf("[Unlike] MQ 发送失败，降级同步写库: %v", err)
		return s.repo.DeleteLike(videoID, accountID)
	}

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
		VideoID:   req.VideoID, // 视频ID
		AccountID: accountID,   // 用户ID
		Username:  username,    // 用户名
		Content:   req.Content, // 评论内容
	}

	if err := s.repo.CreateComment(comment); err != nil {
		return nil, err
	}

	// 增加视频评论数
	_ = s.repo.IncrementCommentsCount(req.VideoID)

	return comment, nil
}

// ListComments 查询评论列表
func (s *Service) ListComments(videoID uint) ([]CommentItem, error) {
	comments, err := s.repo.ListCommentsByVideoID(videoID)
	if err != nil {
		return nil, err
	}

	items := make([]CommentItem, len(comments))
	for i, c := range comments {
		items[i] = CommentItem{
			ID:        c.ID,
			Username:  c.Username,
			Content:   c.Content,   // 评论内容
			CreatedAt: c.CreatedAt, // 创建时间
		}
	}
	return items, nil
}

// RecordPlay 记录视频播放
func (s *Service) RecordPlay(videoID uint) error {
	return s.repo.IncrementPlayCount(videoID)
}

// ListHotVideos 获取热门视频列表
func (s *Service) ListHotVideos(limit int) ([]VideoItem, error) {
	videos, err := s.repo.ListHotVideos(limit)
	if err != nil {
		log.Printf("[ListHotVideos] 查询失败: %v", err)
		return nil, err
	}

	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		items[i] = VideoItem{
			ID:            v.ID,
			AuthorID:      v.AuthorID,
			Username:      v.Username,
			Title:         v.Title,
			PlayURL:       v.PlayURL,
			CoverURL:      v.CoverURL,
			LikesCount:    v.LikesCount,
			PlayCount:     v.PlayCount,
			CommentsCount: v.CommentsCount,
			CreatedAt:     v.CreatedAt,
		}
	}
	log.Printf("[ListHotVideos] 查询成功，返回 %d 条记录", len(items))
	return items, nil
}

// SearchVideos 搜索视频
func (s *Service) SearchVideos(keyword string, page, size int) (*SearchVideosResponse, error) {
	if keyword == "" {
		return nil, errors.New("搜索关键词不能为空")
	}

	videos, total, err := s.repo.SearchVideos(keyword, page, size) // 调用 repo, 返回视频列表和总数
	if err != nil {
		log.Printf("[SearchVideos] 搜索失败: %v", err)
		return nil, err
	}

	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		items[i] = VideoItem{
			ID:            v.ID,
			AuthorID:      v.AuthorID,
			Username:      v.Username,
			Title:         v.Title,
			PlayURL:       v.PlayURL,
			CoverURL:      v.CoverURL,
			LikesCount:    v.LikesCount,
			PlayCount:     v.PlayCount,
			CommentsCount: v.CommentsCount,
			CreatedAt:     v.CreatedAt,
		}
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
