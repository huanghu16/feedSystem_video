package video

import (
	"feedSystem_video/internal/db"

	"gorm.io/gorm"
)

// Repo 视频数据访问
type Repo struct{}

func NewRepo() *Repo {
	return &Repo{}
}

// Create 创建视频记录
func (r *Repo) Create(video *Video) error {
	return db.DB.Create(video).Error
}

// ListByAuthorID 按作者 ID 查询视频列表（按时间倒序）
func (r *Repo) ListByAuthorID(authorID uint) ([]Video, error) {
	var videos []Video
	err := db.DB.Where("author_id = ?", authorID).
		Order("created_at DESC").
		Find(&videos).Error
	return videos, err
}

// GetByID 根据 ID 查询视频详情
func (r *Repo) GetByID(id uint) (*Video, error) {
	var video Video
	result := db.DB.First(&video, id)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			return nil, nil
		}
		return nil, result.Error
	}
	return &video, nil
}

// ==================== 点赞 ====================

// CreateLike 创建点赞记录
func (r *Repo) CreateLike(videoID, accountID uint) error {
	like := Like{VideoID: videoID, AccountID: accountID} // 点赞记录
	return db.DB.Create(&like).Error
}

// DeleteLike 删除点赞记录（取消点赞）
func (r *Repo) DeleteLike(videoID, accountID uint) error {

	return db.DB.Where("video_id = ? AND account_id = ?", videoID, accountID).
		Delete(&Like{}).Error
}

// IsLiked 查询是否已赞
func (r *Repo) IsLiked(videoID, accountID uint) (bool, error) {
	var count int64
	err := db.DB.Model(&Like{}).
		Where("video_id = ? AND account_id = ?", videoID, accountID).
		Count(&count).Error
	return count > 0, err
}

// IncrementLikesCount 增加视频点赞数
func (r *Repo) IncrementLikesCount(videoID uint) error {
	return db.DB.Model(&Video{}).
		Where("id = ?", videoID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + ?", 1)).Error
}

// DecrementLikesCount 减少视频点赞数
func (r *Repo) DecrementLikesCount(videoID uint) error {
	return db.DB.Model(&Video{}).
		Where("id = ?", videoID).
		UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - ?, 0)", 1)).Error
}

// ==================== 播放量 ====================

// IncrementPlayCount 增加视频播放量
func (r *Repo) IncrementPlayCount(videoID uint) error {
	return db.DB.Model(&Video{}).
		Where("id = ?", videoID).
		UpdateColumn("play_count", gorm.Expr("play_count + ?", 1)).Error
}

// ==================== 评论 ====================

// CreateComment 创建评论
func (r *Repo) CreateComment(comment *Comment) error {
	return db.DB.Create(comment).Error
}

// ListCommentsByVideoID 查询视频的评论列表（按时间正序）
func (r *Repo) ListCommentsByVideoID(videoID uint) ([]Comment, error) {
	var comments []Comment
	err := db.DB.Where("video_id = ?", videoID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// ==================== 热门列表 ====================

// ListHotVideos 获取热门视频列表（按播放量倒序，限制数量）
func (r *Repo) ListHotVideos(limit int) ([]Video, error) {
	var videos []Video
	err := db.DB.Order("play_count DESC").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}
