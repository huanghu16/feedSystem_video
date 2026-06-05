package account

import (
	"errors"
	"feedSystem_video/backend/internal/apierror"

	"github.com/gin-gonic/gin"
)

// Handler 封装 Account 的 HTTP 处理
type Handler struct {
	service *Service
}

// NewHandler 创建 Handler 实例
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(c *gin.Context) {
	// 解析请求参数
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 校验失败（比如用户名为空、密码太短）
		apierror.FailParam(c, err.Error())
		return
	}
	// 调用 Service，执行业务逻辑
	resp, err := h.service.Register(&req)
	if err != nil {
		//3.根据错误类型返回不同的 HTTP 响应
		if errors.Is(err, ErrUserAlreadyExists) {
			apierror.FailParam(c, "用户名已存在")
			return
		}
		// 其他未预期的错误，返回 500
		apierror.FailServer(c, "注册失败")
		return
	}

	// 成功，则返回统一格式的响应
	apierror.OK(c, resp)
}
