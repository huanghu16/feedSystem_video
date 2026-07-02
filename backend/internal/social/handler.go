package social

import (
	"errors"
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/middleware/jwt"

	"github.com/gin-gonic/gin"
)

// Handler 关注 HTTP 处理
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// getAccountID 从 Gin Context 安全获取当前用户 ID（带类型断言防护）
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

// normalizePaging 规范化分页参数
func normalizePaging(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 50 {
		size = 50
	}
	return page, size
}

// Follow 处理 POST /social/follow
func (h *Handler) Follow(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	followerID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.Follow(followerID, req.VloggerID); err != nil {
		if errors.Is(err, ErrAlreadyFollowing) || errors.Is(err, ErrCannotFollowSelf) {
			apierror.FailParam(c, err.Error())
			return
		}
		apierror.FailServer(c, "关注失败")
		return
	}

	apierror.OK(c, gin.H{"message": "关注成功"})
}

// Unfollow 处理 POST /social/unfollow
func (h *Handler) Unfollow(c *gin.Context) {
	var req UnfollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	followerID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.Unfollow(followerID, req.VloggerID); err != nil {
		if errors.Is(err, ErrNotFollowing) {
			apierror.FailParam(c, err.Error())
			return
		}
		apierror.FailServer(c, "取消关注失败")
		return
	}

	apierror.OK(c, gin.H{"message": "取消关注成功"})
}

// GetFollowers 处理 POST /social/getAllFollowers
func (h *Handler) GetFollowers(c *gin.Context) {
	var req GetFollowersRequest
	_ = c.ShouldBindJSON(&req) // 分页参数可选，绑定失败用默认值

	userID, ok := getAccountID(c)
	if !ok {
		return
	}

	page, size := normalizePaging(req.Page, req.PageSize)
	items, total, err := h.service.GetFollowers(userID, page, size)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, gin.H{
		"list":     items,
		"total":    total,
		"page":     page,
		"size":     size,
		"has_more": int64(page*size) < total,
	})
}

// GetVloggers 处理 POST /social/getAllVloggers
func (h *Handler) GetVloggers(c *gin.Context) {
	var req GetVloggersRequest
	_ = c.ShouldBindJSON(&req) // 分页参数可选，绑定失败用默认值

	userID, ok := getAccountID(c)
	if !ok {
		return
	}

	page, size := normalizePaging(req.Page, req.PageSize)
	items, total, err := h.service.GetVloggers(userID, page, size)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, gin.H{
		"list":     items,
		"total":    total,
		"page":     page,
		"size":     size,
		"has_more": int64(page*size) < total,
	})
}

// GetCounts 处理 POST /social/getCounts
func (h *Handler) GetCounts(c *gin.Context) {
	userID, ok := getAccountID(c)
	if !ok {
		return
	}

	counts, err := h.service.GetCounts(userID)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, counts)
}

// IsFollowing 处理 POST /social/isFollowing
func (h *Handler) IsFollowing(c *gin.Context) {
	var req IsFollowingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	followerID, ok := getAccountID(c)
	if !ok {
		return
	}

	resp, err := h.service.CheckIsFollowing(followerID, req.VloggerID)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, resp)
}
