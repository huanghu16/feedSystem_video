package social

import (
	"errors"

	"feedSystem_video/internal/account"
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

	return s.repo.CreateFollow(followerID, vloggerID)
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

// GetFollowers 查询粉丝列表
func (s *Service) GetFollowers(vloggerID uint) ([]FollowerItem, error) {
	list, err := s.repo.GetFollowers(vloggerID)
	if err != nil {
		return nil, err
	}

	items := make([]FollowerItem, 0, len(list))
	for _, follow := range list {
		account, err := s.accountRepo.FindByID(follow.FollowerID) // 查询用户信息
		if err != nil || account == nil {
			continue
		}
		items = append(items, FollowerItem{
			ID:       account.ID,
			Username: account.Username,
		})
	}
	return items, nil
}

// GetVloggers 查询关注列表
func (s *Service) GetVloggers(followerID uint) ([]VloggerItem, error) {
	list, err := s.repo.GetVloggers(followerID)
	if err != nil {
		return nil, err
	}

	items := make([]VloggerItem, 0, len(list))
	for _, follow := range list {
		account, err := s.accountRepo.FindByID(follow.VloggerID)
		if err != nil || account == nil {
			continue
		}
		items = append(items, VloggerItem{
			ID:       account.ID,
			Username: account.Username,
		})
	}
	return items, nil
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
