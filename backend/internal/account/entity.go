package account

import (
	"time"

	"gorm.io/gorm"
)

// Account 用户表模型
// 对应数据库中的 accounts 表
type Account struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Password     string         `gorm:"type:varchar(256);not null" json:"-"` // json:"-" 表示不返回给前端
	AvatarURL    string         `gorm:"type:varchar(512);default:''" json:"avatar_url"`
	Bio          string         `gorm:"type:varchar(256);default:''" json:"bio"`
	Token        string         `gorm:"type:varchar(512);default:''" json:"-"`
	RefreshToken string         `gorm:"type:varchar(128);default:''" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// TableName 指定表名（GORM 默认用结构体名的复数形式）
func (Account) TableName() string {
	return "accounts"
}

// --- 请求/响应 DTO ---

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// --- 登录相关 DTO ---

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // access_token 有效期（秒）
}

// TokenResponse 刷新 token 响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// ProfileResponse 用户资料响应
type ProfileResponse struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	AvatarURL      string `json:"avatar_url"`
	Bio            string `json:"bio"`             // 简介
	FansCount      int64  `json:"fans_count"`      // 粉丝数
	FollowingCount int64  `json:"following_count"` // 关注数
	VideoCount     int64  `json:"video_count"`
	LikesCount     int64  `json:"likes_count"`
}

// GetProfileRequest 获取用户资料请求
type GetProfileRequest struct {
	UserID uint `json:"user_id" form:"user_id" binding:"required"`
}

// UpdateAvatarRequest 更新头像请求
type UpdateAvatarRequest struct {
	AvatarURL string `json:"avatar_url" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=64"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

// UpdateBioRequest 更新简介请求
type UpdateBioRequest struct {
	Bio string `json:"bio" binding:"max=256"`
}
