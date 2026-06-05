package video

import (
	"errors"
	"feedSystem_video/backend/internal/account"
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
