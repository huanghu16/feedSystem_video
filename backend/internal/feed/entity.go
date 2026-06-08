package feed

// FeedVideoItem Feed 视频项
type FeedVideoItem struct {
	ID         uint   `json:"id"`          // 视频 ID
	AuthorID   uint   `json:"author_id"`   // 作者 ID
	Username   string `json:"username"`    // 作者用户名
	Title      string `json:"title"`       // 标题
	PlayURL    string `json:"play_url"`    // 播放地址
	CoverURL   string `json:"cover_url"`   // 封面地址
	LikesCount int    `json:"likes_count"` // 点赞数
	PlayCount  int    `json:"play_count"`  // 播放量
}

// ListLatestRequest 最新视频请求
type ListLatestRequest struct {
	// 空请求，查询全局最新
}
