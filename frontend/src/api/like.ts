import { postJson } from './client'

/**
 * 点赞视频
 */
export function like(videoId: number) {
  return postJson<{ message: string }>('/like/like', { video_id: videoId })
}

/**
 * 取消点赞视频
 */
export function unlike(videoId: number) {
  return postJson<{ message: string }>('/like/unlike', { video_id: videoId })
}

/**
 * 检查视频是否已点赞
 */
export function isLiked(videoId: number) {
  return postJson<{ is_liked: boolean }>('/like/isLiked', { video_id: videoId })
}
