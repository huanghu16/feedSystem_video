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
	items, err := h.service.ListLatest()
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, items)
}
