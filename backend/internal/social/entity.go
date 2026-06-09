package social

import (
	"time"

	"gorm.io/gorm"
)

// Social 关注关系表
// follower_id 关注 vlogger_id
type Social struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FollowerID uint           `gorm:"index;not null" json:"follower_id"` // 粉丝
	VloggerID  uint           `gorm:"index;not null" json:"vlogger_id"`  // 被关注者
	CreatedAt  time.Time      `json:"created_at"`                        // 关注时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                    // 软删除
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
}

// GetVloggersRequest 查询关注列表请求
type GetVloggersRequest struct {
	// 从 JWT 获取当前用户 ID
}

// FollowerItem 粉丝列表项
type FollowerItem struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// VloggerItem 关注列表项
type VloggerItem struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
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
