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

// ListLatest 查询最新视频（按时间倒序）
func (r *Repo) ListLatest(limit int) ([]video.Video, error) {
	var videos []video.Video
	err := db.DB.Order("created_at DESC").
		Limit(limit).
		Find(&videos).Error
	return videos, err
}
