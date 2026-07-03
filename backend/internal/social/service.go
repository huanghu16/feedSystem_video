package social

import (
	"errors"

	"feedSystem_video/internal/account"
	"feedSystem_video/internal/middleware/rabbitmq"
)

// Service 关注业务逻辑
type Service struct {
	repo        *Repo         // 关联数据库
	accountRepo *account.Repo // 关联账号
}

func NewService(repo *Repo, accountRepo *account.Repo) *Service {
	return &Service{
		repo:        repo,
		accountRepo: accountRepo,
	}
}

var (
	ErrAlreadyFollowing = errors.New("已经关注了")
	ErrNotFollowing     = errors.New("还没有关注")
	ErrCannotFollowSelf = errors.New("不能关注自己")
)

// Follow 关注用户
func (s *Service) Follow(followerID, vloggerID uint) error {
	if followerID == vloggerID {
		return ErrCannotFollowSelf
	}

	// 检查是否已关注
	isFollowing, err := s.repo.IsFollowing(followerID, vloggerID)
	if err != nil {
		return err
	}
	if isFollowing {
		return ErrAlreadyFollowing
	}

	if err := s.repo.CreateFollow(followerID, vloggerID); err != nil {
		return err
	}

	// 发 MQ 事件，通知消费者异步处理
	_ = rabbitmq.Publish(rabbitmq.ExchangeSocial, rabbitmq.RoutingKeyFollow, rabbitmq.SocialEvent{
		FollowerID: followerID, VloggerID: vloggerID,
	})

	return nil
}

// Unfollow 取消关注
func (s *Service) Unfollow(followerID, vloggerID uint) error {
	isFollowing, err := s.repo.IsFollowing(followerID, vloggerID) //
	if err != nil {
		return err
	}
	if !isFollowing {
		return ErrNotFollowing
	}

	return s.repo.DeleteFollow(followerID, vloggerID)
}

// GetFollowers 查询粉丝列表（JOIN 用户表，消除 N+1 查询，带分页）
func (s *Service) GetFollowers(vloggerID uint, page, size int) ([]FollowerItem, int64, error) {
	return s.repo.GetFollowersWithUser(vloggerID, page, size)
}

// GetVloggers 查询关注列表（JOIN 用户表，消除 N+1 查询，带分页）
func (s *Service) GetVloggers(followerID uint, page, size int) ([]VloggerItem, int64, error) {
	return s.repo.GetVloggersWithUser(followerID, page, size)
}

// GetCounts 查询粉丝数和关注数
func (s *Service) GetCounts(userID uint) (*FollowCountsResponse, error) {
	followers, err := s.repo.GetFollowersCount(userID)
	if err != nil {
		return nil, err
	}

	vloggers, err := s.repo.GetVloggersCount(userID)
	if err != nil {
		return nil, err
	}

	return &FollowCountsResponse{
		FollowersCount: followers,
		VloggersCount:  vloggers,
	}, nil
}

// CheckIsFollowing 检查是否已关注
func (s *Service) CheckIsFollowing(followerID, vloggerID uint) (*IsFollowingResponse, error) {
	isFollowing, err := s.repo.IsFollowing(followerID, vloggerID)
	if err != nil {
		return nil, err
	}
	return &IsFollowingResponse{
		IsFollowing: isFollowing,
	}, nil
}
