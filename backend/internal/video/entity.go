package video

import (
	"time"

	"gorm.io/gorm"
)

// Video 视频表模型
type Video struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	AuthorID   uint           `gorm:"index;not null" json:"author_id"`
	Username   string         `gorm:"type:varchar(64);not null" json:"username"`
	Title      string         `gorm:"type:varchar(256);not null" json:"title"`
	PlayURL    string         `gorm:"type:varchar(512);not null" json:"play_url"`
	CoverURL   string         `gorm:"type:varchar(512);default:''" json:"cover_url"`
	LikesCount int            `gorm:"default:0" json:"likes_count"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Video) TableName() string {
	return "videos"
}

// --- DTO ---

// PublishRequest 发布视频请求
type PublishRequest struct {
	Title    string `json:"title" binding:"required,max=256"`
	PlayURL  string `json:"play_url" binding:"required"`
	CoverURL string `json:"cover_url"`
}

// ListByAuthorRequest 按作者查询请求
type ListByAuthorRequest struct {
	AuthorID uint `json:"author_id" binding:"required"`
}

// VideoItem 视频列表项
type VideoItem struct {
	ID         uint      `json:"id"`
	AuthorID   uint      `json:"author_id"`
	Username   string    `json:"username"`
	Title      string    `json:"title"`
	PlayURL    string    `json:"play_url"`
	CoverURL   string    `json:"cover_url"`
	LikesCount int       `json:"likes_count"`
	CreatedAt  time.Time `json:"created_at"`
}
