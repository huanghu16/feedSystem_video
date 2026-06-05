package social

import (
	"feedSystem_video/internal/db"
)

// Repo 关注数据访问
type Repo struct{}

func NewRepo() *Repo {
	return &Repo{}
}

// CreateFollow 创建关注关系
func (r *Repo) CreateFollow(followerID, vloggerID uint) error {
	social := Social{FollowerID: followerID, VloggerID: vloggerID}
	return db.DB.Create(&social).Error
}

// DeleteFollow 删除关注关系（取消关注）
func (r *Repo) DeleteFollow(followerID, vloggerID uint) error {
	return db.DB.Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Delete(&Social{}).Error
}

// IsFollowing 查询是否已关注
func (r *Repo) IsFollowing(followerID, vloggerID uint) (bool, error) {
	var count int64
	err := db.DB.Model(&Social{}).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowers 查询粉丝列表
func (r *Repo) GetFollowers(vloggerID uint) ([]Social, error) {
	var list []Social
	err := db.DB.Where("vlogger_id = ?", vloggerID).Find(&list).Error
	return list, err
}

// GetVloggers 查询关注列表
func (r *Repo) GetVloggers(followerID uint) ([]Social, error) {
	var list []Social
	err := db.DB.Where("follower_id = ?", followerID).Find(&list).Error
	return list, err
}

// GetFollowersCount 查询粉丝数
func (r *Repo) GetFollowersCount(vloggerID uint) (int64, error) {
	var count int64
	err := db.DB.Model(&Social{}).Where("vlogger_id = ?", vloggerID).Count(&count).Error
	return count, err
}

// GetVloggersCount 查询关注数
func (r *Repo) GetVloggersCount(followerID uint) (int64, error) {
	var count int64
	err := db.DB.Model(&Social{}).Where("follower_id = ?", followerID).Count(&count).Error
	return count, err
}
