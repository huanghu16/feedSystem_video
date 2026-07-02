package account

import (
	"errors"
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/middleware/jwt"
	"feedSystem_video/internal/middleware/storage"

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

func (h *Handler) Register(c *gin.Context) {
	// 解析请求参数
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}
	// 调用 Service，执行业务逻辑
	resp, err := h.service.Register(&req)
	if err != nil {
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
		// 用户不存在或密码错误，统一返回"用户名或密码错误"（安全考量）
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrPasswordWrong) {
			apierror.FailParam(c, "用户名或密码错误")
			return
		}
		apierror.FailServer(c, "登录失败")
		return
	}

	apierror.OK(c, resp)
}

// GetProfile 处理 POST /account/getProfile
func (h *Handler) GetProfile(c *gin.Context) {
	var req GetProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	profile, err := h.service.GetProfile(req.UserID)
	if err != nil {
		apierror.FailServer(c, "获取用户资料失败")
		return
	}

	apierror.OK(c, profile)
}

// UploadAvatar 处理 POST /account/uploadAvatar
func (h *Handler) UploadAvatar(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		apierror.FailParam(c, "请上传文件")
		return
	}
	defer file.Close()

	// 使用公共上传工具保存文件
	const maxSize = 10 * 1024 * 1024 // 10MB
	_, avatarURL, err := storage.SaveUploadFile(file, header.Filename, storage.ImageExts, maxSize)
	if err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	// 获取当前用户ID（带类型断言防护）
	userID, ok := getAccountID(c)
	if !ok {
		return
	}

	// 更新数据库中的头像URL
	if err := h.service.UpdateAvatar(userID, avatarURL); err != nil {
		apierror.FailServer(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"avatar_url": avatarURL})
}

// ChangePassword 处理 POST /account/changePassword
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	userID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrOldPasswordWrong) {
			apierror.FailParam(c, "原密码错误")
			return
		}
		apierror.FailServer(c, "修改密码失败")
		return
	}

	apierror.OK(c, gin.H{"message": "密码修改成功"})
}

// UpdateBio 处理 POST /account/updateBio
func (h *Handler) UpdateBio(c *gin.Context) {
	var req UpdateBioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	userID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.UpdateBio(userID, req.Bio); err != nil {
		if errors.Is(err, ErrBioTooLong) {
			apierror.FailParam(c, err.Error())
			return
		}
		apierror.FailServer(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"message": "简介更新成功"})
}
