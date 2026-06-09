package video

import (
	"time"

	"gorm.io/gorm"
)

// Video 视频表模型
type Video struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	AuthorID      uint           `gorm:"index;not null" json:"author_id"`
	Username      string         `gorm:"type:varchar(64);not null" json:"username"`
	Title         string         `gorm:"type:varchar(256);not null" json:"title"`
	PlayURL       string         `gorm:"type:varchar(512);not null" json:"play_url"`
	CoverURL      string         `gorm:"type:varchar(512);default:''" json:"cover_url"`
	LikesCount    int            `gorm:"default:0" json:"likes_count"`
	PlayCount     int            `gorm:"default:0" json:"play_count"`
	CommentsCount int            `gorm:"default:0" json:"comments_count"`  // 评论数 (新增)
	HotScore      float64        `gorm:"-:all" json:"hot_score,omitempty"` // 热度分数（新增）（不存储在数据库，仅用于热榜查询）
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"` // 更新时间 (新增)
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
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
	ID            uint      `json:"id"`
	AuthorID      uint      `json:"author_id"`
	Username      string    `json:"username"`
	Title         string    `json:"title"`
	PlayURL       string    `json:"play_url"`
	CoverURL      string    `json:"cover_url"`
	LikesCount    int       `json:"likes_count"`
	PlayCount     int       `json:"play_count"`
	CommentsCount int       `json:"comments_count"` // 新增
	CreatedAt     time.Time `json:"created_at"`
}

// --- 点赞相关 ---

// Like 点赞记录表
type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VideoID   uint      `gorm:"index;not null" json:"video_id"`
	AccountID uint      `gorm:"index;not null" json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Like) TableName() string {
	return "likes"
}

// LikeRequest 点赞/取消点赞请求
type LikeRequest struct {
	VideoID uint `json:"video_id" binding:"required"`
}

// UnlikeRequest 取消点赞请求
type UnlikeRequest struct {
	VideoID uint `json:"video_id" binding:"required"`
}

// IsLikedRequest 查询是否已赞请求
type IsLikedRequest struct {
	VideoID uint `json:"video_id" binding:"required"`
}

// IsLikedResponse 是否已赞响应
type IsLikedResponse struct {
	IsLiked bool `json:"is_liked"`
}

// ListMyLikedVideosRequest 查询我赞过的视频
type ListMyLikedVideosRequest struct {
	// 空请求，从 JWT 获取用户 ID
}

// --- 评论相关 ---

// Comment 评论表
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VideoID   uint      `gorm:"index;not null" json:"video_id"`
	AccountID uint      `gorm:"index;not null" json:"account_id"`
	Username  string    `gorm:"type:varchar(64);not null" json:"username"`
	Content   string    `gorm:"type:varchar(512);not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (Comment) TableName() string {
	return "comments"
}

// PublishCommentRequest 发布评论请求
type PublishCommentRequest struct {
	VideoID uint   `json:"video_id" binding:"required"`
	Content string `json:"content" binding:"required,max=512"`
}

// CommentItem 评论列表项
type CommentItem struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ListCommentsRequest 评论列表请求
type ListCommentsRequest struct {
	VideoID uint `json:"video_id" binding:"required"`
}
