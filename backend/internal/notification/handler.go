package notification

import (
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/middleware/jwt"

	"github.com/gin-gonic/gin"
)

// Handler 通知 HTTP 处理层
type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

// getAccountID 从 Gin Context 获取当前用户 ID
func getAccountID(c *gin.Context) (uint, bool) {
	val, exists := c.Get(jwt.AccountIDKey)
	if !exists {
		apierror.FailAuth(c, "未登录")
		return 0, false
	}
	uid, ok := val.(uint)
	if !ok {
		apierror.FailAuth(c, "用户信息异常")
		return 0, false
	}
	return uid, true
}

// UnreadCount 获取未读通知数
// POST /notification/unreadCount
func (h *Handler) UnreadCount(c *gin.Context) {
	recipientID, ok := getAccountID(c)
	if !ok {
		return
	}
	count, err := h.repo.CountUnread(recipientID)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}
	apierror.OK(c, UnreadCountResponse{Count: count})
}

// List 获取通知列表（读取后自动标记已读）
// POST /notification/list
func (h *Handler) List(c *gin.Context) {
	recipientID, ok := getAccountID(c)
	if !ok {
		return
	}

	list, total, err := h.repo.ListByRecipient(recipientID, 1, 20)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	items := make([]NotificationItem, 0, len(list))
	for _, n := range list {
		items = append(items, NotificationItem{
			ID:        n.ID,
			ActorName: n.ActorName,
			Type:      n.Type,
			Content:   n.Content,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		})
	}

	// 读取后标记为已读
	_ = h.repo.MarkAllRead(recipientID)

	apierror.OK(c, ListResponse{List: items, Total: total})
}
