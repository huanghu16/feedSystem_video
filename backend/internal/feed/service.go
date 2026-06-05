package feed

// Service Feed 业务逻辑
type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// ListLatest 获取最新视频列表
func (s *Service) ListLatest() ([]FeedVideoItem, error) {
	videos, err := s.repo.ListLatest(20) // 默认返回 20 条
	if err != nil {
		return nil, err
	}

	items := make([]FeedVideoItem, len(videos))
	for i, v := range videos {
		items[i] = FeedVideoItem{
			ID:         v.ID,         // 视频 ID
			AuthorID:   v.AuthorID,   // 作者 ID
			Username:   v.Username,   // 作者名
			Title:      v.Title,      // 标题
			PlayURL:    v.PlayURL,    // 播放地址
			CoverURL:   v.CoverURL,   // 封面
			LikesCount: v.LikesCount, // 点赞数
		}
	}
	return items, nil
}
