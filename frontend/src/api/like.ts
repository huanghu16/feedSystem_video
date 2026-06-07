import { postJson } from './client'

/**
 * 点赞视频
 * @param videoId 
 * @returns 
 */
export function like(videoId: number) {
  return postJson('/like/like', { video_id: videoId })
}

/**
 * 取消点赞视频
 * @param videoId 
 * @returns 
 */
export function unlike(videoId: number) {
  return postJson('/like/unlike', { video_id: videoId })
}

/**
 * 检查视频是否已点赞
 * @param videoId 
 * @returns 
 */
export function isLiked(videoId: number) {
  return postJson<{ is_liked: boolean }>('/like/isLiked', { video_id: videoId })
}