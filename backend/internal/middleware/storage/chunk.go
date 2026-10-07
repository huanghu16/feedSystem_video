package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"feedSystem_video/internal/config"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// chunkRoot 分片临时根目录
func chunkRoot() string {
	return config.C.Storage.UploadDir + "_tmp"
}

// ValidUploadID 校验 uploadID：必须是 64 位小写十六进制（sha256）
func ValidUploadID(id string) bool {
	if len(id) != 64 {
		return false
	}
	for _, ch := range id {
		if !(ch >= '0' && ch <= '9') && !(ch >= 'a' && ch <= 'f') {
			return false
		}
	}
	return true
}

// ChunkDir 某个上传任务的分片目录
func ChunkDir(uploadID string) string {
	return filepath.Join(chunkRoot(), uploadID)
}

func chunkFilePath(uploadID string, index int) string {
	return filepath.Join(ChunkDir(uploadID), fmt.Sprintf("%d.part", index))
}

// SaveChunk 保存单个分片
func SaveChunk(uploadID string, index int, r io.Reader, maxSize int64) error {
	if !ValidUploadID(uploadID) {
		return fmt.Errorf("非法的 uploadID")
	}
	if index < 0 {
		return fmt.Errorf("非法的分片序号")
	}

	if err := os.MkdirAll(ChunkDir(uploadID), os.ModePerm); err != nil {
		return fmt.Errorf("创建分片目录失败: %w", err)
	}

	finalPath := chunkFilePath(uploadID, index)
	tmpPath := finalPath + ".uploading"

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("创建分片文件失败: %w", err)
	}

	written, copyErr := io.Copy(out, io.LimitReader(r, maxSize+1))
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("写入分片失败: %w", copyErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("关闭分片文件失败: %w", closeErr)
	}
	if written > maxSize {
		os.Remove(tmpPath)
		return fmt.Errorf("分片超过大小限制（最大 %d 字节）", maxSize)
	}

	// 原子落位：只有完整写完的分片才以 .part 出现
	// 这样 ListUploadedChunks 看到 .part 就一定是完整分片，不会把半截数据当已完成
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("保存分片失败: %w", err)
	}
	return nil
}

// ListUploadedChunks 返回已完成的分片序号（断点续传的依据）
func ListUploadedChunks(uploadID string) []int {
	if !ValidUploadID(uploadID) {
		return nil
	}
	entries, err := os.ReadDir(ChunkDir(uploadID))
	if err != nil {
		return nil
	}
	indexes := make([]int, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".part") {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSuffix(e.Name(), ".part"))
		if err != nil {
			continue
		}
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	return indexes
}

// MergeChunks 按序拼接分片并落盘，返回写入字节数
func MergeChunks(uploadID string, totalChunks int, dstPath string) (int64, error) {
	if !ValidUploadID(uploadID) {
		return 0, fmt.Errorf("非法的 uploadID")
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return 0, fmt.Errorf("创建目录失败: %w", err)
	}

	tmpPath := dstPath + ".merging"
	out, err := os.Create(tmpPath)
	if err != nil {
		return 0, fmt.Errorf("创建目标文件失败: %w", err)
	}

	var total int64
	for i := 0; i < totalChunks; i++ {
		in, err := os.Open(chunkFilePath(uploadID, i))
		if err != nil {
			out.Close()
			os.Remove(tmpPath)
			return 0, fmt.Errorf("分片 %d 缺失: %w", i, err)
		}
		n, err := io.Copy(out, in)
		in.Close()
		if err != nil {
			out.Close()
			os.Remove(tmpPath)
			return 0, fmt.Errorf("写入分片 %d 失败: %w", i, err)
		}
		total += n
	}

	if err := out.Close(); err != nil {
		os.Remove(tmpPath)
		return 0, fmt.Errorf("关闭目标文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		return 0, fmt.Errorf("落盘失败: %w", err)
	}
	return total, nil
}

// StoredPath hash 对应文件的磁盘路径
func StoredPath(uploadID, ext string) string {
	return filepath.Join(config.C.Storage.UploadDir, uploadID+ext)
}

// StoredFileURL 文件已存在则返回其 URL —— 秒传就靠这个判断
func StoredFileURL(uploadID, ext string) (string, bool) {
	if !ValidUploadID(uploadID) {
		return "", false
	}
	if _, err := os.Stat(StoredPath(uploadID, ext)); err != nil {
		return "", false
	}
	return fmt.Sprintf("%s/%s", config.C.Storage.StaticPath, uploadID+ext), true
}

// FileSHA256 计算文件内容 sha256（十六进制小写）
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CleanChunkDir 删除某个上传任务的全部分片
func CleanChunkDir(uploadID string) {
	if !ValidUploadID(uploadID) {
		return
	}
	os.RemoveAll(ChunkDir(uploadID))
}

// CleanStaleChunkDirs 清理超过 maxAge 未更新的分片目录
func CleanStaleChunkDirs(maxAge time.Duration) int {
	entries, err := os.ReadDir(chunkRoot())
	if err != nil {
		return 0
	}
	removed := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil || time.Since(info.ModTime()) < maxAge {
			continue
		}
		if err := os.RemoveAll(filepath.Join(chunkRoot(), e.Name())); err == nil {
			removed++
		}
	}
	return removed
}
