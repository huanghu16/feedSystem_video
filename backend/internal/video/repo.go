package video

import "feedSystem_video/backend/internal/db"

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
