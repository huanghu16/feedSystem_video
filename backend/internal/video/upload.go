package video

import (
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/config"
	"feedSystem_video/internal/middleware/storage"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	maxVideoSize = 100 * 1024 * 1024 // 整段视频上限 100MB
	maxChunkSize = 5 * 1024 * 1024   // 单个分片上限 5MB
)

// UploadInitRequest 初始化上传（同时用于查询断点）
type UploadInitRequest struct {
	FileHash    string `json:"file_hash" binding:"required"`    // 前端算的 sha256，作为 upload_id
	FileName    string `json:"file_name" binding:"required"`    // 原始文件名，用于取扩展名
	FileSize    int64  `json:"file_size" binding:"required"`    // 文件总大小
	ChunkSize   int64  `json:"chunk_size" binding:"required"`   // 分片大小
	TotalChunks int    `json:"total_chunks" binding:"required"` // 分片总数
}

// UploadInitResponse 初始化响应
type UploadInitResponse struct {
	UploadID       string `json:"upload_id"`
	UploadedChunks []int  `json:"uploaded_chunks"` // 已存在的分片，前端据此跳过
	Instant        bool   `json:"instant"`         // 命中秒传
	PlayURL        string `json:"play_url,omitempty"`
	CoverURL       string `json:"cover_url,omitempty"`
}

// UploadMergeRequest 合并请求
type UploadMergeRequest struct {
	UploadID    string `json:"upload_id" binding:"required"`
	FileName    string `json:"file_name" binding:"required"`
	TotalChunks int    `json:"total_chunks" binding:"required"`
}

// UploadInit 处理 POST /video/upload/init（需要 JWT）
func (h *Handler) UploadInit(c *gin.Context) {
	var req UploadInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(req.FileName))
	if !storage.VideoExts[ext] {
		apierror.FailParam(c, "不支持的文件类型: "+ext)
		return
	}

	// 总量校验
	if req.FileSize <= 0 || req.FileSize > maxVideoSize {
		apierror.FailParam(c, fmt.Sprintf("文件大小超过限制（最大 %d 字节）", maxVideoSize))
		return
	}
	if req.ChunkSize <= 0 || req.ChunkSize > maxChunkSize {
		apierror.FailParam(c, fmt.Sprintf("分片大小非法（最大 %d 字节）", maxChunkSize))
		return
	}

	// 分片数必须与"总大小/分片大小"自洽，防止客户端伪造分片数绕过校验
	expected := int((req.FileSize + req.ChunkSize - 1) / req.ChunkSize)
	if req.TotalChunks != expected {
		apierror.FailParam(c, fmt.Sprintf("分片数量与文件大小不匹配（期望 %d）", expected))
		return
	}

	// 秒传：该 hash 的文件已存在，直接复用
	if url, ok := storage.StoredFileURL(req.FileHash, ext); ok {
		coverURL := ""
		coverName := req.FileHash + ".jpg"
		if _, err := os.Stat(filepath.Join(config.C.Storage.UploadDir, coverName)); err == nil {
			coverURL = fmt.Sprintf("%s/%s", config.C.Storage.StaticPath, coverName)
		}
		log.Printf("秒传命中: %s", url)
		apierror.OK(c, UploadInitResponse{
			UploadID: req.FileHash,
			Instant:  true,
			PlayURL:  url,
			CoverURL: coverURL,
		})
		return
	}

	apierror.OK(c, UploadInitResponse{
		UploadID:       req.FileHash,
		UploadedChunks: storage.ListUploadedChunks(req.FileHash),
	})
}

// UploadChunk 处理 POST /video/upload/chunk（需要 JWT）
// 同一分片重复上传是幂等的（覆盖写），所以前端可以安全重试
func (h *Handler) UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("upload_id")

	index, err := strconv.Atoi(c.PostForm("chunk_index"))
	if err != nil || index < 0 {
		apierror.FailParam(c, "分片序号非法")
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		apierror.FailParam(c, "请上传分片数据")
		return
	}
	defer file.Close()

	if err := storage.SaveChunk(uploadID, index, file, maxChunkSize); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	apierror.OK(c, gin.H{"chunk_index": index})
}

// UploadMerge 处理 POST /video/upload/merge（需要 JWT）
func (h *Handler) UploadMerge(c *gin.Context) {
	var req UploadMergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.FailParam(c, err.Error())
		return
	}

	if !storage.ValidUploadID(req.UploadID) {
		apierror.FailParam(c, "upload_id 非法")
		return
	}

	ext := strings.ToLower(filepath.Ext(req.FileName))
	if !storage.VideoExts[ext] {
		apierror.FailParam(c, "不支持的文件类型: "+ext)
		return
	}

	// 分片没齐就别合
	uploaded := storage.ListUploadedChunks(req.UploadID)
	if len(uploaded) != req.TotalChunks {
		apierror.FailParam(c, fmt.Sprintf("分片不完整：已收到 %d/%d", len(uploaded), req.TotalChunks))
		return
	}

	// init 之后被别人传完了，同样命中秒传
	if url, ok := storage.StoredFileURL(req.UploadID, ext); ok {
		storage.CleanChunkDir(req.UploadID)
		apierror.OK(c, gin.H{"play_url": url, "cover_url": ""})
		return
	}

	savePath := storage.StoredPath(req.UploadID, ext)
	if _, err := storage.MergeChunks(req.UploadID, req.TotalChunks, savePath); err != nil {
		apierror.FailServer(c, err.Error())
		return
	}

	// 完整性校验：合并后内容 hash 必须等于文件名（即 upload_id）
	// 网络丢包/分片错序都会在这里被抓住
	actual, err := storage.FileSHA256(savePath)
	if err != nil || actual != req.UploadID {
		os.Remove(savePath)
		storage.CleanChunkDir(req.UploadID)
		log.Printf("文件校验失败: 期望 %s 实际 %s", req.UploadID, actual)
		apierror.FailServer(c, "文件校验失败，请重新上传")
		return
	}

	storage.CleanChunkDir(req.UploadID)

	playURL := fmt.Sprintf("%s/%s", config.C.Storage.StaticPath, req.UploadID+ext)

	// 复用既有的 FFmpeg 封面生成
	coverURL := ""
	coverPath := filepath.Join(config.C.Storage.UploadDir, req.UploadID+".jpg")
	if generateCover(savePath, coverPath) {
		coverURL = fmt.Sprintf("%s/%s", config.C.Storage.StaticPath, req.UploadID+".jpg")
	}

	log.Printf("分片上传完成: %s", playURL)
	apierror.OK(c, gin.H{"play_url": playURL, "cover_url": coverURL})
}
