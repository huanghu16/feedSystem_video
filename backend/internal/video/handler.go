package video

import (
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/middleware/jwt"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler 视频 HTTP 处理
type Handler struct {
	service *Service
}

// NewHandler 创建 Handler 实例
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service, // 把参数(service)传进来的 service 变量，赋值给 Handler 结构体的 service 字段
	}
}

// Publish 处理 POST /video/publish（需要 JWT）
func (h *Handler) Publish(c *gin.Context) {
	var req PublishRequest // 请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	// 从 JWT Context 获取当前用户 ID
	authorID, _ := c.Get(jwt.AccountIDKey)
	authorIDUint := authorID.(uint)

	video, err := h.service.Publish(&req, authorIDUint)
	if err != nil {
		apierror.FailServer(c, err.Error())
		return
	}

	apierror.OK(c, video)
}

// UploadVideo 处理 POST /video/uploadVideo（需要 JWT）
// 接收 multipart 文件上传，保存到本地
func (h *Handler) UploadVideo(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		apierror.FailParam(c, "请上传文件")
		return
	}
	defer file.Close()

	// 检查文件大小（限制100MB）
	const maxSize = 100 * 1024 * 1024 // 100MB
	if header.Size > maxSize {
		apierror.FailParam(c, fmt.Sprintf("文件大小不能超过100MB，当前文件大小: %.2fMB", float64(header.Size)/(1024*1024)))
		return
	}

	// 生成唯一文件名：时间戳_原始文件名
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
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
	written, err := io.Copy(out, file)
	if err != nil {
		apierror.FailServer(c, fmt.Sprintf("写入文件失败: %v", err))
		return
	}

	// 记录日志
	fmt.Printf("视频上传成功: %s (大小: %.2fMB)\n", filename, float64(written)/(1024*1024))

	// 返回可访问的 URL
	playURL := fmt.Sprintf("/static/%s", filename)
	apierror.OK(c, gin.H{"play_url": playURL})
}

// ListByAuthor 处理 POST /video/listByAuthorID
func (h *Handler) ListByAuthor(c *gin.Context) {
	var req ListByAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	items, err := h.service.ListByAuthor(req.AuthorID)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, items)
}

// ==================== 点赞 Handler ====================

// Like 处理 POST /like/like
func (h *Handler) Like(c *gin.Context) {
	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil { // 解析请求参数
		apierror.FailParam(c, err.Error())
		return
	}

	accountID, _ := c.Get(jwt.AccountIDKey)
	accountIDUint := accountID.(uint)

	if err := h.service.Like(req.VideoID, accountIDUint); err != nil { // 调用 Service，执行业务逻辑
		apierror.FailServer(c, err.Error()) // 返回错误响应
		return
	}

	apierror.OK(c, gin.H{"message": "点赞成功"})
}

// Unlike 处理 POST /like/unlike
func (h *Handler) Unlike(c *gin.Context) {
	var req UnlikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	accountID, _ := c.Get(jwt.AccountIDKey)
	accountIDUint := accountID.(uint)

	if err := h.service.Unlike(req.VideoID, accountIDUint); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"message": "取消点赞成功"})
}

// IsLiked 处理 POST /like/isLiked
func (h *Handler) IsLiked(c *gin.Context) {
	var req IsLikedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	accountID, exists := c.Get(jwt.AccountIDKey)
	if !exists {
		apierror.OK(c, IsLikedResponse{IsLiked: false})
		return
	}
	accountIDUint := accountID.(uint)

	isLiked, err := h.service.IsLiked(req.VideoID, accountIDUint)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, IsLikedResponse{IsLiked: isLiked})
}

// ==================== 评论 Handler ====================

// PublishComment 处理 POST /comment/publish
func (h *Handler) PublishComment(c *gin.Context) {
	var req PublishCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	accountID, _ := c.Get(jwt.AccountIDKey)
	accountIDUint := accountID.(uint)
	username, _ := c.Get("username")
	usernameStr := username.(string)

	comment, err := h.service.PublishComment(&req, accountIDUint, usernameStr)
	if err != nil {
		if err == ErrVideoNotFound {
			apierror.FailParam(c, "视频不存在")
			return
		}
		apierror.FailServer(c, "发布失败")
		return
	}

	apierror.OK(c, comment)
}

// ListComments 处理 POST /comment/listAll
func (h *Handler) ListComments(c *gin.Context) {
	var req ListCommentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	items, err := h.service.ListComments(req.VideoID)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, items)
}

// GetDetail 处理 POST /video/getDetail
func (h *Handler) GetDetail(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	video, err := h.service.repo.GetByID(req.ID)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}
	if video == nil {
		apierror.FailParam(c, "视频不存在")
		return
	}

	apierror.OK(c, video)
}

// RecordPlay 处理 POST /video/recordPlay
func (h *Handler) RecordPlay(c *gin.Context) {
	var req struct {
		VideoID uint `json:"video_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	if err := h.service.RecordPlay(req.VideoID); err != nil {
		apierror.FailServer(c, "记录播放失败")
		return
	}

	apierror.OK(c, gin.H{"message": "播放记录成功"})
}

// ListHotVideos 处理 POST /video/listHot
func (h *Handler) ListHotVideos(c *gin.Context) {
	var req struct {
		Limit int `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Limit = 10 // 默认返回10个
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}

	items, err := h.service.ListHotVideos(req.Limit)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, items)
}
