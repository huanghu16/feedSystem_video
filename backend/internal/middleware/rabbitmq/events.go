package rabbitmq

// LinkEvent 点赞事件消息
type LikeEvent struct {
	VideoID   uint `json:"video_id"`   // 视频ID
	AccountID uint `json:"account_id"` // 账号ID
}

// CommentEvent 评论事件消息
type CommentEvent struct {
	VideoID   uint   `json:"video_id"`   // 视频ID
	AccountID uint   `json:"account_id"` // 账号ID
	Username  string `json:"username"`   // 用户名
	Content   string `json:"content"`    // 评论内容
}

// SocialEvent 关注事件消息
type SocialEvent struct {
	FollowerID uint `json:"follower_id"` // 关注者ID
	VloggerID  uint `json:"vlogger_id"`  // 被关注者ID
}

// Exchange 名称常量
const (
	ExchangeLike    = "like.events"    // 点赞事件
	ExchangeComment = "comment.events" // 评论事件
	ExchangeSocial  = "social.events"  // 关注事件
)

// Routing Key 常量
const (
	RoutingKeyLike    = "like.created"    // 点赞事件
	RoutingKeyUnlike  = "like.deleted"    // 取消点赞事件
	RoutingKeyComment = "comment.created" // 评论事件
	RoutingKeyFollow  = "social.followed" // 关注事件
)
