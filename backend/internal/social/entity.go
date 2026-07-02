package social

import (
	"time"

	"gorm.io/gorm"
)

// Social 关注关系表
// follower_id 关注 vlogger_id
// 复合唯一索引 idx_follower_vlogger 防止重复关注
// 注意：GORM 软删除 + 唯一索引存在已知冲突，若需支持取消后重新关注，
// 可将 DeletedAt 纳入复合索引或使用业务层校验（当前采用业务层预检）
type Social struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FollowerID uint           `gorm:"uniqueIndex:idx_follower_vlogger;not null" json:"follower_id"` // 粉丝
	VloggerID  uint           `gorm:"uniqueIndex:idx_follower_vlogger;not null" json:"vlogger_id"`  // 被关注者
	CreatedAt  time.Time      `json:"created_at"`                                                   // 关注时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                                               // 软删除
}

func (Social) TableName() string {
	return "socials"
}

// --- DTO ---

// FollowRequest 关注请求
type FollowRequest struct {
	VloggerID uint `json:"vlogger_id" binding:"required"`
}

// UnfollowRequest 取消关注请求
type UnfollowRequest struct {
	VloggerID uint `json:"vlogger_id" binding:"required"`
}

// GetFollowersRequest 查询粉丝请求
type GetFollowersRequest struct {
	// 从 JWT 获取当前用户 ID
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// GetVloggersRequest 查询关注列表请求
type GetVloggersRequest struct {
	// 从 JWT 获取当前用户 ID
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// FollowerItem 粉丝列表项
type FollowerItem struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// VloggerItem 关注列表项（含用户资料统计）
type VloggerItem struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	AvatarURL      string `json:"avatar_url"`
	Bio            string `json:"bio"`
	VideoCount     int64  `json:"video_count"`
	FansCount      int64  `json:"fans_count"`
	FollowingCount int64  `json:"following_count"`
}

// IsFollowingRequest 检查是否关注请求
type IsFollowingRequest struct {
	VloggerID uint `json:"vlogger_id" binding:"required"`
}

// IsFollowingResponse 检查是否关注响应
type IsFollowingResponse struct {
	IsFollowing bool `json:"is_following"`
}

// FollowCountsResponse 粉丝/关注数响应
type FollowCountsResponse struct {
	FollowersCount int64 `json:"followers_count"` // 粉丝数
	VloggersCount  int64 `json:"vloggers_count"`  // 关注数
}
