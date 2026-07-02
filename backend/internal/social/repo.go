package social

import (
	"feedSystem_video/internal/db"

	"gorm.io/gorm"
)

// Repo 关注数据访问
type Repo struct {
	db *gorm.DB
}

func NewRepo() *Repo {
	return &Repo{db: db.DB}
}

// CreateFollow 创建关注关系
func (r *Repo) CreateFollow(followerID, vloggerID uint) error {
	social := Social{FollowerID: followerID, VloggerID: vloggerID}
	return r.db.Create(&social).Error
}

// DeleteFollow 删除关注关系（取消关注）
func (r *Repo) DeleteFollow(followerID, vloggerID uint) error {
	return r.db.Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Delete(&Social{}).Error
}

// IsFollowing 检查是否已关注
func (r *Repo) IsFollowing(followerID, vloggerID uint) (bool, error) {
	var count int64
	err := r.db.Model(&Social{}).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetFollowers 查询粉丝列表（仅关注关系记录）
func (r *Repo) GetFollowers(vloggerID uint) ([]Social, error) {
	var list []Social
	err := r.db.Where("vlogger_id = ?", vloggerID).Find(&list).Error
	return list, err
}

// GetVloggers 查询关注列表（仅关注关系记录）
func (r *Repo) GetVloggers(followerID uint) ([]Social, error) {
	var list []Social
	err := r.db.Where("follower_id = ?", followerID).Find(&list).Error
	return list, err
}

// GetFollowersWithUser 查询粉丝列表并 JOIN 用户表获取用户名（消除 N+1 查询）
// followerID 即粉丝的 ID，socials.vlogger_id = vloggerID
// 支持分页，page 从 1 开始
func (r *Repo) GetFollowersWithUser(vloggerID uint, page, size int) ([]FollowerItem, int64, error) {
	var items []FollowerItem
	var total int64

	// 先查总数
	if err := r.db.Table("socials").
		Where("vlogger_id = ? AND deleted_at IS NULL", vloggerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err := r.db.Table("socials").
		Select("accounts.id, accounts.username").
		Joins("INNER JOIN accounts ON accounts.id = socials.follower_id").
		Where("socials.vlogger_id = ? AND socials.deleted_at IS NULL", vloggerID).
		Offset(offset).
		Limit(size).
		Scan(&items).Error
	return items, total, err
}

// GetVloggersWithUser 查询关注列表并 JOIN 用户表获取用户名和统计信息（消除 N+1 查询）
// socials.follower_id = followerID，即当前用户关注的人
// 支持分页，page 从 1 开始
func (r *Repo) GetVloggersWithUser(followerID uint, page, size int) ([]VloggerItem, int64, error) {
	var items []VloggerItem
	var total int64

	// 先查总数
	if err := r.db.Table("socials").
		Where("follower_id = ? AND deleted_at IS NULL", followerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，使用子查询获取每个用户的统计数据
	offset := (page - 1) * size
	err := r.db.Table("socials").
		Select(`accounts.id, 
			accounts.username, 
			accounts.avatar_url,
			accounts.bio,
			(SELECT COUNT(*) FROM videos WHERE videos.author_id = accounts.id AND videos.deleted_at IS NULL) AS video_count,
			(SELECT COUNT(*) FROM socials s2 WHERE s2.vlogger_id = accounts.id AND s2.deleted_at IS NULL) AS fans_count,
			(SELECT COUNT(*) FROM socials s3 WHERE s3.follower_id = accounts.id AND s3.deleted_at IS NULL) AS following_count`).
		Joins("INNER JOIN accounts ON accounts.id = socials.vlogger_id").
		Where("socials.follower_id = ? AND socials.deleted_at IS NULL", followerID).
		Order("socials.created_at DESC").
		Offset(offset).
		Limit(size).
		Scan(&items).Error
	return items, total, err
}

// GetFollowersCount 查询粉丝数
func (r *Repo) GetFollowersCount(vloggerID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Social{}).Where("vlogger_id = ?", vloggerID).Count(&count).Error
	return count, err
}

// GetVloggersCount 查询关注数
func (r *Repo) GetVloggersCount(followerID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Social{}).Where("follower_id = ?", followerID).Count(&count).Error
	return count, err
}
