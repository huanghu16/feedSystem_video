package social

import (
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

// Follow 处理 POST /social/follow
func (h *Handler) Follow(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	followerID, _ := c.Get(jwt.AccountIDKey)
	followerIDUint := followerID.(uint)

	if err := h.service.Follow(followerIDUint, req.VloggerID); err != nil {
		if err == ErrAlreadyFollowing || err == ErrCannotFollowSelf {
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

	followerID, _ := c.Get(jwt.AccountIDKey)
	followerIDUint := followerID.(uint)

	if err := h.service.Unfollow(followerIDUint, req.VloggerID); err != nil {
		if err == ErrNotFollowing {
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
	userID, _ := c.Get(jwt.AccountIDKey)
	userIDUint := userID.(uint)

	items, err := h.service.GetFollowers(userIDUint)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, items)
}

// GetVloggers 处理 POST /social/getAllVloggers
func (h *Handler) GetVloggers(c *gin.Context) {
	userID, _ := c.Get(jwt.AccountIDKey)
	userIDUint := userID.(uint)

	items, err := h.service.GetVloggers(userIDUint)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, items)
}

// GetCounts 处理 POST /social/getCounts
func (h *Handler) GetCounts(c *gin.Context) {
	userID, _ := c.Get(jwt.AccountIDKey)
	userIDUint := userID.(uint)

	counts, err := h.service.GetCounts(userIDUint)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, counts)
}
