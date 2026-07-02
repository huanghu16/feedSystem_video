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
