package video

import (
	"errors"
	"feedSystem_video/internal/account"
	"fmt"
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
		PlayURL:  req.PlayURL,
		CoverURL: req.CoverURL,
	}

	if err := s.repo.Create(video); err != nil {
		return nil, fmt.Errorf("创建视频失败: %w", err)
	}

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
			ID:         v.ID,
			AuthorID:   v.AuthorID, // 作者ID
			Username:   v.Username,
			Title:      v.Title,
			PlayURL:    v.PlayURL,    // 播放地址
			CoverURL:   v.CoverURL,   // 封面地址
			LikesCount: v.LikesCount, // 点赞数
			CreatedAt:  v.CreatedAt,  // 创建时间
		}
	}
	return items, nil
}

// ==================== 点赞 ====================

// Like 点赞
func (s *Service) Like(videoID, accountID uint) error {
	// 检查是否已赞
	isLiked, err := s.repo.IsLiked(videoID, accountID)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.New("已经点赞过了")
	}

	// 创建点赞记录
	if err := s.repo.CreateLike(videoID, accountID); err != nil {
		return err
	}

	// 增加视频点赞数
	return s.repo.IncrementLikesCount(videoID)
}

// Unlike 取消点赞
func (s *Service) Unlike(videoID, accountID uint) error {
	// 检查是否已赞
	isLiked, err := s.repo.IsLiked(videoID, accountID)
	if err != nil {
		return err
	}
	if !isLiked {
		return errors.New("还没有点赞")
	}

	// 删除点赞记录
	if err := s.repo.DeleteLike(videoID, accountID); err != nil {
		return err
	}

	// 减少视频点赞数
	return s.repo.DecrementLikesCount(videoID)
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
