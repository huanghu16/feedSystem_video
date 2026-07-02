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
	Description   string         `gorm:"type:text;default:''" json:"description"`
	PublishDate   string         `gorm:"type:varchar(20);default:''" json:"publish_date"`
	PlayURL       string         `gorm:"type:varchar(512);not null" json:"play_url"`
	CoverURL      string         `gorm:"type:varchar(512);default:''" json:"cover_url"`
	LikesCount    int            `gorm:"default:0" json:"likes_count"`
	PlayCount     int            `gorm:"default:0" json:"play_count"`
	CommentsCount int            `gorm:"default:0" json:"comments_count"`
	HotScore      float64        `gorm:"-:all" json:"hot_score,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Video) TableName() string {
	return "videos"
}

// --- DTO ---

// PublishRequest 发布视频请求
type PublishRequest struct {
	Title       string `json:"title" binding:"required,max=256"`
	Description string `json:"description"`
	PublishDate string `json:"publish_date"`
	PlayURL     string `json:"play_url" binding:"required"`
	CoverURL    string `json:"cover_url"`
}

// ListByAuthorRequest 按作者查询请求
type ListByAuthorRequest struct {
	AuthorID uint `json:"author_id" binding:"required"`
	Page     int  `json:"page"` // 页码，从 1 开始
	Size     int  `json:"size"` // 每页条数，默认 10，最大 50
}

// ListByAuthorResponse 按作者查询响应（带分页）
type ListByAuthorResponse struct {
	List    []VideoItem `json:"list"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
	HasMore bool        `json:"has_more"`
}

// VideoItem 视频列表项
type VideoItem struct {
	ID            uint      `json:"id"`
	AuthorID      uint      `json:"author_id"`
	Username      string    `json:"username"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	PublishDate   string    `json:"publish_date"`
	PlayURL       string    `json:"play_url"`
	CoverURL      string    `json:"cover_url"`
	LikesCount    int       `json:"likes_count"`
	PlayCount     int       `json:"play_count"`
	CommentsCount int       `json:"comments_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// --- 点赞相关 ---

// Like 点赞记录表
// 复合唯一索引 idx_video_account 防止同一用户对同一视频重复点赞
type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VideoID   uint      `gorm:"uniqueIndex:idx_video_account;not null" json:"video_id"`
	AccountID uint      `gorm:"uniqueIndex:idx_video_account;not null" json:"account_id"`
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
	AvatarURL string    `gorm:"-" json:"avatar_url"` // 不映射数据库列，JOIN 查询时填充
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
	AvatarURL string    `json:"avatar_url"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ListCommentsRequest 评论列表请求
type ListCommentsRequest struct {
	VideoID uint `json:"video_id" binding:"required"`
	Page    int  `json:"page"` // 页码，从 1 开始
	Size    int  `json:"size"` // 每页条数，默认 10，最大 50
}

// ListCommentsResponse 评论列表响应（带分页）
type ListCommentsResponse struct {
	List    []CommentItem `json:"list"`
	Total   int64         `json:"total"`
	Page    int           `json:"page"`
	Size    int           `json:"size"`
	HasMore bool          `json:"has_more"`
}

// SearchVideosRequest 搜索视频请求
type SearchVideosRequest struct {
	Keyword string `json:"keyword" binding:"required"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

// SearchVideosResponse 搜索视频响应
type SearchVideosResponse struct {
	List    []VideoItem `json:"list"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
	HasMore bool        `json:"has_more"`
}

// VideoToItem 将 Video 模型转换为 VideoItem DTO
// 提取为公共函数，消除 service 中 3 处重复的映射逻辑
func VideoToItem(v *Video) VideoItem {
	return VideoItem{
		ID:            v.ID,
		AuthorID:      v.AuthorID,
		Username:      v.Username,
		Title:         v.Title,
		Description:   v.Description,
		PublishDate:   v.PublishDate,
		PlayURL:       v.PlayURL,
		CoverURL:      v.CoverURL,
		LikesCount:    v.LikesCount,
		PlayCount:     v.PlayCount,
		CommentsCount: v.CommentsCount,
		CreatedAt:     v.CreatedAt,
	}
}

// CommentToItem 将 Comment 模型转换为 CommentItem DTO
func CommentToItem(c *Comment) CommentItem {
	return CommentItem{
		ID:        c.ID,
		Username:  c.Username,
		AvatarURL: c.AvatarURL,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
	}
}
