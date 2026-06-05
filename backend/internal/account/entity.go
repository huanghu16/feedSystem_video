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
