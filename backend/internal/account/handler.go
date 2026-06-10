package account

import (
	"errors"
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/middleware/jwt"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// GetProfile 处理 POST /account/getProfile
func (h *Handler) GetProfile(c *gin.Context) {
	var req GetProfileRequest // 从 JWT 获取当前用户 ID
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

	// 检查文件类型
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[strings.ToLower(ext)] {
		apierror.FailParam(c, "只支持 JPG、PNG、GIF、WebP 格式的图片")
		return
	}

	// 检查文件大小（限制10MB）
	const maxSize = 10 * 1024 * 1024 // 10MB
	if header.Size > maxSize {
		apierror.FailParam(c, "图片大小不能超过10MB")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), "avatar", ext)
	savePath := filepath.Join("uploads", filename)

	// 确保 uploads 目录存在
	os.MkdirAll("uploads", os.ModePerm)

	// 创建目标文件
	out, err := os.Create(savePath)
	if err != nil {
		apierror.FailServer(c, "保存文件失败")
		return
	}
	defer out.Close()

	// 复制文件内容
	if _, err := io.Copy(out, file); err != nil {
		apierror.FailServer(c, "写入文件失败")
		return
	}

	// 获取当前用户ID
	accountID, exists := c.Get(jwt.AccountIDKey)
	if !exists {
		apierror.FailServer(c, "未登录")
		return
	}
	userID := accountID.(uint)

	// 更新数据库中的头像URL
	avatarURL := fmt.Sprintf("/static/%s", filename)
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

	// 获取当前用户ID
	accountID, exists := c.Get(jwt.AccountIDKey)
	if !exists {
		apierror.FailServer(c, "未登录")
		return
	}
	userID := accountID.(uint)

	if err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "原密码错误" {
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

	// 获取当前用户ID
	accountID, exists := c.Get(jwt.AccountIDKey)
	if !exists {
		apierror.FailServer(c, "未登录")
		return
	}
	userID := accountID.(uint)

	if err := h.service.UpdateBio(userID, req.Bio); err != nil {
		apierror.FailServer(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"message": "简介更新成功"})
}
