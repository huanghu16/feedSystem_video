/**
 * 全局类型定义
 * VideoItem 和 FeedVideoItem 字段完全相同，合并为 VideoItem 一个类型
 * FeedVideoItem 保留为类型别名，保持向后兼容
 */

export interface TokenResponse {
  access_token: string
  refresh_token: string
}

export interface Account {
  id: number
  username: string
  avatar_url: string
  bio: string
  created_at: string
}

export interface VideoItem {
  id: number
  author_id: number
  username: string
  title: string
  description: string
  publish_date: string
  play_url: string
  cover_url: string
  likes_count: number
  play_count: number
  comments_count: number
  created_at: string
}

// FeedVideoItem 与 VideoItem 字段完全相同，使用类型别名消除冗余
export type FeedVideoItem = VideoItem

export interface CommentItem {
  id: number
  username: string
  avatar_url: string
  content: string
  created_at: string
}

export interface FollowerItem {
  id: number
  username: string
}

export interface VloggerItem {
  id: number
  username: string
  avatar_url: string
  bio: string
  video_count: number
  fans_count: number
  following_count: number
}

export interface FollowCounts {
  fans_count: number
  following_count: number
}
