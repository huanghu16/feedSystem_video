// 和后端 DTO 对应的 TypeScript 类型

// 登录返回的令牌
export interface TokenResponse {
    access_token: string    // 访问令牌
    refresh_token: string   // 刷新令牌
    expires_in: number  // 令牌过期时间
}

// 账户信息
export interface Account {
    id: number
    username: string
    avatar_url: string   // 头像
    bio: string  // 简介
    created_at: string  // 创建时间
}

// 视频信息
export interface VideoItem {
    id: number
    author_id: number  // 作者id
    username: string
    title: string
    play_url: string   // 播放地址
    cover_url: string  // 封面
    likes_count: number  // 点赞数
    play_count: number  // 播放量
    comments_count: number  // 评论数
    created_at: string  // 创建时间
}

// 订阅视频信息
export interface FeedVideoItem {
    id: number
    author_id: number
    username: string
    title: string
    play_url: string
    cover_url: string
    likes_count: number
    play_count: number  // 播放量
    comments_count: number  // 评论数
}

// 评论信息
export interface CommentItem {
    id: number
    username: string
    content: string
    created_at: string
}

// 粉丝信息
export interface FollowerItem {
    id: number
    username: string
}

// 视频作者信息
export interface VloggerItem {
    id: number
    username: string
}

// 粉丝数和视频作者数
export interface FollowCounts {
    followers_count: number  // 粉丝数
    vloggers_count: number  // 视频作者数
}