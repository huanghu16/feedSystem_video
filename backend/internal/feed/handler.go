package feed

import (
	"feedSystem_video/internal/apierror"

	"github.com/gin-gonic/gin"
)

// Handler Feed HTTP 处理
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListLatest 处理 POST /feed/listLatest
func (h *Handler) ListLatest(c *gin.Context) {
	var req ListLatestRequest
	_ = c.ShouldBindJSON(&req) // 分页参数可选，绑定失败用默认值

	resp, err := h.service.ListLatest(&req)
	if err != nil {
		apierror.FailServer(c, "获取视频列表失败")
		return
	}

	apierror.OK(c, resp)
}
