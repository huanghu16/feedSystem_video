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

// ListByAuthorID 分页查询指定作者的视频列表
// page 从 1 开始，size 为每页条数
// 返回视频列表和总数
func (r *Repo) ListByAuthorID(authorID uint, page, size int) ([]Video, int64, error) {
	var videos []Video
	var total int64

	// 先查总数
	if err := db.DB.Model(&Video{}).Where("author_id = ?", authorID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	if err := db.DB.Where("author_id = ?", authorID).
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
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

// DeleteVideo 删除视频
func (r *Repo) DeleteVideo(videoID uint, authorID uint) error {
	return db.DB.Where("id = ? AND author_id = ?", videoID, authorID).
		Delete(&Video{}).Error
}

// DeleteVideosBatch 批量删除视频
func (r *Repo) DeleteVideosBatch(videoIDs []uint, authorID uint) error {
	return db.DB.Where("id IN ? AND author_id = ?", videoIDs, authorID).
		Delete(&Video{}).Error
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

// CreateLikeTx 在事务中创建点赞记录并增加计数（保证原子性）
func (r *Repo) CreateLikeTx(videoID, accountID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		like := Like{VideoID: videoID, AccountID: accountID}
		if err := tx.Create(&like).Error; err != nil {
			return err
		}
		// 增加视频点赞数
		if err := tx.Model(&Video{}).
			Where("id = ?", videoID).
			UpdateColumn("likes_count", gorm.Expr("likes_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteLikeTx 在事务中删除点赞记录并减少计数（保证原子性）
func (r *Repo) DeleteLikeTx(videoID, accountID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("video_id = ? AND account_id = ?", videoID, accountID).
			Delete(&Like{})
		if result.Error != nil {
			return result.Error
		}
		// 仅当确实删除了记录时才减少计数
		if result.RowsAffected > 0 {
			if err := tx.Model(&Video{}).
				Where("id = ?", videoID).
				UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - ?, 0)", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
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

// CreateCommentTx 在事务中创建评论并增加评论计数（保证原子性）
func (r *Repo) CreateCommentTx(comment *Comment) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).
			Where("id = ?", comment.VideoID).
			UpdateColumn("comments_count", gorm.Expr("comments_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	})
}

// ListCommentsByVideoID 分页查询视频的评论列表（按时间正序）
// JOIN accounts 表获取评论者头像，消除 N+1 查询
// page 从 1 开始，size 为每页条数
// 返回评论列表和总数
func (r *Repo) ListCommentsByVideoID(videoID uint, page, size int) ([]Comment, int64, error) {
	var comments []Comment
	var total int64

	// 先查总数
	if err := db.DB.Model(&Comment{}).Where("video_id = ?", videoID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，JOIN accounts 获取头像
	offset := (page - 1) * size
	if err := db.DB.Table("comments").
		Select("comments.id, comments.video_id, comments.account_id, comments.username, comments.content, comments.created_at, accounts.avatar_url AS avatar_url").
		Joins("LEFT JOIN accounts ON accounts.id = comments.account_id").
		Where("comments.video_id = ?", videoID).
		Order("comments.created_at ASC").
		Offset(offset).
		Limit(size).
		Scan(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// IncrementCommentsCount 增加视频评论数
func (r *Repo) IncrementCommentsCount(videoID uint) error {
	return db.DB.Model(&Video{}).
		Where("id = ?", videoID).
		UpdateColumn("comments_count", gorm.Expr("comments_count + ?", 1)).Error
}

// DecrementCommentsCount 减少视频评论数（如果需要删除评论功能）
func (r *Repo) DecrementCommentsCount(videoID uint) error {
	return db.DB.Model(&Video{}).
		Where("id = ?", videoID).
		UpdateColumn("comments_count", gorm.Expr("GREATEST(comments_count - ?, 0)", 1)).Error
}

// ==================== 热门列表 ====================（添加算法）

// ListHotVideos 获取热门视频列表（按时间衰减算法排序，限制数量）
func (r *Repo) ListHotVideos(limit int) ([]Video, error) {
	var videos []Video
	// 使用 MySQL 的时间函数计算热度分数
	// 优化后的热度算法：热度 = play_count / ((时间差小时数 + 24) ^ 1.2)
	// 调整说明：
	// 1. 基数从 2 改为 24：给新视频一个合理的基础时间窗口（24小时）
	// 2. 指数从 1.5 改为 1.2：让时间衰减更平缓，播放量权重更大
	// 这样既保留了时间衰减特性，又确保高播放量视频能排在前面
	err := db.DB.Select(`*, 
		COALESCE(play_count, 0) * 1.0 / 
		POWER(
			TIMESTAMPDIFF(HOUR, created_at, NOW()) + 100, 
			1.1
		) as hot_score`).
		Order("hot_score DESC").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}

// SearchVideos 搜索视频（支持分页）
func (r *Repo) SearchVideos(keyword string, page, size int) ([]Video, int64, error) {
	var videos []Video
	var total int64

	// 默认分页参数
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 50 {
		size = 10
	}

	// 计算偏移量
	offset := (page - 1) * size

	// 构建查询：模糊匹配标题或用户名
	query := db.DB.Model(&Video{}).Where("title LIKE ? OR username LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&videos).Error

	return videos, total, err
}
