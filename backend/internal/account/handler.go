package account

import (
	"errors"
	"feedSystem_video/internal/apierror"

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

// Login 处理 POST /account/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	resp, err := h.service.Login(&req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || err.Error() == "密码错误" {
			apierror.FailParam(c, "用户名或密码错误") // 不告诉用户是用户名错还是密码错（安全）
			return
		}
		apierror.FailServer(c, "登录失败")
		return
	}

	apierror.OK(c, resp)
}
