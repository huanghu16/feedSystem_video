package video

import (
	"errors"
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/config"
	"feedSystem_video/internal/middleware/jwt"
	"feedSystem_video/internal/middleware/storage"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler 视频 HTTP 处理
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

// getUsername 从 Gin Context 安全获取当前用户名（带类型断言防护）
func getUsername(c *gin.Context) (string, bool) {
	val, exists := c.Get("username")
	if !exists {
		apierror.FailAuth(c, "未登录")
		return "", false
	}
	name, ok := val.(string)
	if !ok {
		apierror.FailAuth(c, "用户信息异常")
		return "", false
	}
	return name, true
}

// Publish 处理 POST /video/publish（需要 JWT）
func (h *Handler) Publish(c *gin.Context) {
	var req PublishRequest // 请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	// 从 JWT Context 安全获取当前用户 ID
	authorID, ok := getAccountID(c)
	if !ok {
		return
	}

	video, err := h.service.Publish(&req, authorID)
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

	// 使用公共上传工具保存文件（限制100MB）
	const maxSize = 100 * 1024 * 1024 // 100MB
	savePath, playURL, err := storage.SaveUploadFile(file, header.Filename, storage.VideoExts, maxSize)
	if err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	log.Printf("视频上传成功: %s", playURL)

	// 自动生成封面
	coverURL := ""
	if isVideoFile(header.Filename) {
		log.Printf("检测到视频文件，开始生成封面...\n")
		coverFilename := strings.TrimSuffix(filepath.Base(savePath), filepath.Ext(savePath)) + ".jpg"
		coverPath := filepath.Join(config.C.Storage.UploadDir, coverFilename)

		if generateCover(savePath, coverPath) {
			coverURL = fmt.Sprintf("%s/%s", config.C.Storage.StaticPath, coverFilename)
			log.Printf("封面生成成功: %s\n", coverFilename)
		} else {
			log.Printf("封面生成失败，将使用空封面\n")
		}
	}

	log.Printf("返回数据 - play_url: %s, cover_url: %s\n", playURL, coverURL)

	apierror.OK(c, gin.H{
		"play_url":  playURL,
		"cover_url": coverURL,
	})
}

// generateCover 使用 FFmpeg 从视频生成封面
// FFmpeg 路径从 config.C.Storage.FFmpegPath 读取，支持绝对路径或 PATH 中的 "ffmpeg"
func generateCover(videoPath, coverPath string) bool {
	ffmpegPath := config.C.Storage.FFmpegPath

	// 如果是相对路径（如 "ffmpeg"），依赖系统 PATH 查找；否则检查文件是否存在
	if filepath.IsAbs(ffmpegPath) {
		if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
			log.Printf("[FFmpeg] 可执行文件不存在: %s\n", ffmpegPath)
			return false
		}
	}

	// 提取视频第一帧作为封面
	log.Printf("[FFmpeg] 执行命令: %s -i %s -vframes 1 -q:v 2 %s\n", ffmpegPath, videoPath, coverPath)
	cmd := exec.Command(ffmpegPath, "-i", videoPath, "-vframes", "1", "-q:v", "2", coverPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[FFmpeg] 执行失败: %v\n", err)
		log.Printf("[FFmpeg] 错误输出: %s\n", string(output))
		return false
	}

	// 检查封面文件是否生成
	if _, err := os.Stat(coverPath); os.IsNotExist(err) {
		log.Printf("[FFmpeg] 封面文件不存在: %s\n", coverPath)
		return false
	}

	// 获取文件大小
	info, err := os.Stat(coverPath)
	if err == nil {
		log.Printf("[FFmpeg] 封面文件大小: %.2fKB\n", float64(info.Size())/1024)
	}

	return true
}

// isVideoFile 检查是否为视频文件
func isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
	for _, vext := range videoExts {
		if ext == vext {
			return true
		}
	}
	return false
}

// ListByAuthor 处理 POST /video/listByAuthorID
func (h *Handler) ListByAuthor(c *gin.Context) {
	var req ListByAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	resp, err := h.service.ListByAuthor(&req)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, resp)
}

// ==================== 点赞 Handler ====================

// Like 处理 POST /like/like
func (h *Handler) Like(c *gin.Context) {
	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil { // 解析请求参数
		apierror.FailParam(c, err.Error())
		return
	}

	accountID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.Like(req.VideoID, accountID); err != nil { // 调用 Service，执行业务逻辑
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

	accountID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.Unlike(req.VideoID, accountID); err != nil {
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

	accountID, ok := getAccountID(c)
	if !ok {
		apierror.OK(c, IsLikedResponse{IsLiked: false})
		return
	}

	isLiked, err := h.service.IsLiked(req.VideoID, accountID)
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

	accountID, ok := getAccountID(c)
	if !ok {
		return
	}
	usernameStr, ok := getUsername(c)
	if !ok {
		return
	}

	comment, err := h.service.PublishComment(&req, accountID, usernameStr)
	if err != nil {
		if errors.Is(err, ErrVideoNotFound) {
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

	resp, err := h.service.ListComments(&req)
	if err != nil {
		apierror.FailServer(c, "查询失败")
		return
	}

	apierror.OK(c, resp)
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

// SearchVideos 处理 POST /video/search
func (h *Handler) SearchVideos(c *gin.Context) {
	var req SearchVideosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 50 {
		req.Size = 10
	}

	resp, err := h.service.SearchVideos(req.Keyword, req.Page, req.Size)
	if err != nil {
		apierror.FailServer(c, "搜索失败")
		return
	}

	apierror.OK(c, resp)
}

// DeleteVideo 处理 POST /video/delete（需要 JWT）
func (h *Handler) DeleteVideo(c *gin.Context) {
	var req struct {
		VideoID uint `json:"video_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	accountID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteVideo(req.VideoID, accountID); err != nil {
		if errors.Is(err, ErrVideoNotFound) {
			apierror.FailParam(c, "视频不存在")
			return
		}
		apierror.FailServer(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"message": "删除成功"})
}

// DeleteVideosBatch 处理 POST /video/deleteBatch（需要 JWT）
func (h *Handler) DeleteVideosBatch(c *gin.Context) {
	var req struct {
		VideoIDs []uint `json:"video_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	if len(req.VideoIDs) == 0 {
		apierror.FailParam(c, "请选择要删除的视频")
		return
	}

	accountID, ok := getAccountID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteVideosBatch(req.VideoIDs, accountID); err != nil {
		apierror.FailServer(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"message": fmt.Sprintf("成功删除 %d 个视频", len(req.VideoIDs))})
}

// UploadCover 处理 POST /video/uploadCover（需要 JWT）
// 接收封面图片上传
func (h *Handler) UploadCover(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		apierror.FailParam(c, "请上传文件")
		return
	}
	defer file.Close()

	// 使用公共上传工具保存文件（限制10MB）
	const maxSize = 10 * 1024 * 1024 // 10MB
	_, coverURL, err := storage.SaveUploadFile(file, header.Filename, storage.ImageExts, maxSize)
	if err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"url": coverURL})
}

// isImageFile 检查是否为图片文件
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	for _, iext := range imageExts {
		if ext == iext {
			return true
		}
	}
	return false
}
