package notification

import "time"

// Notification 通知表模型
type Notification struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RecipientID uint      `gorm:"index;not null" json:"recipient_id"`    // 收件人（视频作者/被关注者）
	ActorID     uint      `gorm:"not null" json:"actor_id"`              // 操作者（点赞/评论/关注的人）
	ActorName   string    `gorm:"type:varchar(64)" json:"actor_name"`    // 操作者用户名
	Type        string    `gorm:"type:varchar(20);not null" json:"type"` // like/comment/follow
	VideoID     uint      `gorm:"default:0" json:"video_id"`             // 相关视频（关注时为0）
	Content     string    `gorm:"type:varchar(256)" json:"content"`      // 通知文案
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// 通知类型常量
const (
	TypeLike    = "like"
	TypeComment = "comment"
	TypeFollow  = "follow"
)

// --- 响应 DTO ---

// NotificationItem 通知列表项
type NotificationItem struct {
	ID        uint      `json:"id"`
	ActorName string    `json:"actor_name"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// ListResponse 通知列表响应
type ListResponse struct {
	List  []NotificationItem `json:"list"`
	Total int64              `json:"total"`
}

// UnreadCountResponse 未读数响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
