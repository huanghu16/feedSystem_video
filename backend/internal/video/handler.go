package video

import (
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/middleware/jwt"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// 自动生成封面
	coverURL := ""
	if isVideoFile(filename) {
		fmt.Printf("检测到视频文件，开始生成封面...\n")
		coverFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".jpg"
		coverPath := filepath.Join("uploads", coverFilename)

		// 尝试生成封面
		if generateCover(savePath, coverPath) {
			coverURL = fmt.Sprintf("/static/%s", coverFilename)
			fmt.Printf("✅ 封面生成成功: %s\n", coverFilename)
		} else {
			fmt.Printf("❌ 封面生成失败，将使用空封面\n")
		}
	} else {
		fmt.Printf("非视频文件，跳过封面生成\n")
	}

	fmt.Printf("返回数据 - play_url: %s, cover_url: %s\n", playURL, coverURL)

	apierror.OK(c, gin.H{
		"play_url":  playURL,
		"cover_url": coverURL,
	})
}

// generateCover 使用 FFmpeg 从视频生成封面
func generateCover(videoPath, coverPath string) bool {
	// FFmpeg 绝对路径
	ffmpegPath := `E:\ffmpeg\ffmpeg-8.1.1-essentials_build\bin\ffmpeg.exe`

	// 检查 FFmpeg 是否存在
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		fmt.Printf("[FFmpeg] FFmpeg 不存在: %s\n", ffmpegPath)
		return false
	}

	// 提取视频第一帧作为封面
	fmt.Printf("[FFmpeg] 执行命令: %s -i %s -vframes 1 -q:v 2 %s\n", ffmpegPath, videoPath, coverPath)
	cmd := exec.Command(ffmpegPath, "-i", videoPath, "-vframes", "1", "-q:v", "2", coverPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[FFmpeg] 执行失败: %v\n", err)
		fmt.Printf("[FFmpeg] 错误输出: %s\n", string(output))
		return false
	}

	// 检查封面文件是否生成
	if _, err := os.Stat(coverPath); os.IsNotExist(err) {
		fmt.Printf("[FFmpeg] 封面文件不存在: %s\n", coverPath)
		return false
	}

	// 获取文件大小
	info, err := os.Stat(coverPath)
	if err == nil {
		fmt.Printf("[FFmpeg] 封面文件大小: %.2fKB\n", float64(info.Size())/1024)
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

	accountID, _ := c.Get(jwt.AccountIDKey)
	accountIDUint := accountID.(uint)

	if err := h.service.DeleteVideo(req.VideoID, accountIDUint); err != nil {
		if err == ErrVideoNotFound {
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

	accountID, _ := c.Get(jwt.AccountIDKey)
	accountIDUint := accountID.(uint)

	if err := h.service.DeleteVideosBatch(req.VideoIDs, accountIDUint); err != nil {
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

	// 检查文件类型
	if !isImageFile(header.Filename) {
		apierror.FailParam(c, "请上传图片文件")
		return
	}

	// 检查文件大小（限制10MB）
	const maxSize = 10 * 1024 * 1024 // 10MB
	if header.Size > maxSize {
		apierror.FailParam(c, fmt.Sprintf("图片大小不能超过10MB，当前文件大小: %.2fMB", float64(header.Size)/(1024*1024)))
		return
	}

	// 生成唯一文件名
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
	_, err = io.Copy(out, file)
	if err != nil {
		apierror.FailServer(c, fmt.Sprintf("写入文件失败: %v", err))
		return
	}

	// 返回可访问的 URL
	coverURL := fmt.Sprintf("/static/%s", filename)
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
