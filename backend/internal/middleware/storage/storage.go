package storage

import (
	"feedSystem_video/internal/config"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveUploadFile 保存上传文件到指定目录，返回可访问的 URL
// 参数：
//   - file: 文件内容 reader
//   - originalName: 原始文件名（用于提取扩展名）
//   - allowedExts: 允许的扩展名集合（小写，含点号，如 ".jpg"）
//   - maxSize: 最大文件大小（字节）
//
// 返回：
//   - savePath: 文件在磁盘上的完整路径
//   - url: 可访问的 URL（如 /static/xxx.jpg）
//   - error: 错误信息
func SaveUploadFile(file io.Reader, originalName string, allowedExts map[string]bool, maxSize int64) (savePath, url string, err error) {
	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(originalName))
	if !allowedExts[ext] {
		return "", "", fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), originalName)
	savePath = filepath.Join(config.C.Storage.UploadDir, filename)

	// 确保上传目录存在
	if err := os.MkdirAll(config.C.Storage.UploadDir, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建目标文件
	out, err := os.Create(savePath)
	if err != nil {
		return "", "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	// 复制文件内容
	written, err := io.Copy(out, file)
	if err != nil {
		return "", "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 检查文件大小
	if maxSize > 0 && written > maxSize {
		os.Remove(savePath) // 删除超大的文件
		return "", "", fmt.Errorf("文件大小超过限制（最大 %d 字节）", maxSize)
	}

	// 生成可访问的 URL
	url = fmt.Sprintf("%s/%s", config.C.Storage.StaticPath, filename)

	return savePath, url, nil
}

// ImageExts 常用图片扩展名集合
var ImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// VideoExts 常用视频扩展名集合
var VideoExts = map[string]bool{
	".mp4": true, ".avi": true, ".mov": true, ".mkv": true, ".flv": true, ".wmv": true,
}
