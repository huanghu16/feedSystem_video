package notification

import "feedSystem_video/internal/db"

// Repo 通知数据仓库
type Repo struct{}

func NewRepo() *Repo {
	return &Repo{}
}

// Create 创建通知记录
func (r *Repo) Create(n *Notification) error {
	return db.DB.Create(n).Error
}

// ListByRecipient 查询用户的通知列表（最新在前，带分页）
func (r *Repo) ListByRecipient(recipientID uint, page, size int) ([]Notification, int64, error) {
	var list []Notification
	var total int64

	db.DB.Model(&Notification{}).Where("recipient_id = ?", recipientID).Count(&total)

	offset := (page - 1) * size
	err := db.DB.Where("recipient_id = ?", recipientID).
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&list).Error
	return list, total, err
}

// CountUnread 统计未读通知数
func (r *Repo) CountUnread(recipientID uint) (int64, error) {
	var count int64
	err := db.DB.Model(&Notification{}).
		Where("recipient_id = ? AND is_read = ?", recipientID, false).
		Count(&count).Error
	return count, err
}

// MarkAllRead 将用户所有通知标记为已读
func (r *Repo) MarkAllRead(recipientID uint) error {
	return db.DB.Model(&Notification{}).
		Where("recipient_id = ? AND is_read = ?", recipientID, false).
		Update("is_read", true).Error
}
