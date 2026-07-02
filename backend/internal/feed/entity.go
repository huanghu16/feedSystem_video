package feed

import "feedSystem_video/internal/video"

// FeedVideoItem Feed 流视频项
// 直接复用 video.VideoItem，消除冗余类型定义
type FeedVideoItem = video.VideoItem

// ListLatestRequest 最新视频列表请求
type ListLatestRequest struct {
	Page int `json:"page"` // 页码，从 1 开始
	Size int `json:"size"` // 每页条数，默认 10，最大 50
}

// ListLatestResponse 最新视频列表响应（带分页）
type ListLatestResponse struct {
	List    []FeedVideoItem `json:"list"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	Size    int             `json:"size"`
	HasMore bool            `json:"has_more"`
}
