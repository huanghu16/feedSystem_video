package feed

import (
	"feedSystem_video/internal/db"
	"feedSystem_video/internal/video"
)

// Repo Feed 数据访问
type Repo struct{}

func NewRepo() *Repo {
	return &Repo{}
}

// VideoCounts 视频计数快照
// 用于在缓存命中后，用实时计数覆盖缓存里的过期计数
type VideoCounts struct {
	ID            uint `gorm:"column:id"`
	LikesCount    int  `gorm:"column:likes_count"`
	CommentsCount int  `gorm:"column:comments_count"`
	PlayCount     int  `gorm:"column:play_count"`
}

// CountsByIDs 批量查询视频的实时计数
// 走主键 IN 查询，一次拿回整页视频的计数，成本可忽略
func (r *Repo) CountsByIDs(ids []uint) (map[uint]VideoCounts, error) {
	result := make(map[uint]VideoCounts, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	var rows []VideoCounts
	if err := db.DB.Model(&video.Video{}).
		Select("id, likes_count, comments_count, play_count").
		Where("id IN ?", ids).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.ID] = row
	}
	return result, nil
}

// ListLatest 分页查询最新视频列表
// page 从 1 开始，size 为每页条数
// 返回视频列表和总数
func (r *Repo) ListLatest(page, size int) ([]video.Video, int64, error) {
	var videos []video.Video
	var total int64

	// 先查总数
	if err := db.DB.Model(&video.Video{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	if err := db.DB.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}
