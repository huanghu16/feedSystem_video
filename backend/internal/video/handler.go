package video

import (
	"feedSystem_video/backend/internal/apierror"
	"feedSystem_video/backend/internal/middleware/jwt"
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
	if _, err := io.Copy(out, file); err != nil {
		apierror.FailServer(c, "写入文件失败")
		return
	}

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
