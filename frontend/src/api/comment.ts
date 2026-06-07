import { postJson } from './client'
import type { CommentItem } from './types'

/**
 * 获取视频的评论列表
 * @param videoId 视频id
 * @return CommentItem[] 评论列表
 */
export function listComments(videoId: number) {
  return postJson<CommentItem[]>('/comment/listAll', { video_id: videoId })
}

/**
 * 发表评论
 * @param videoId 视频id
 * @param content 评论内容
 * @return CommentItem 评论项
 */
export function publishComment(videoId: number, content: string) {
  return postJson<CommentItem>('/comment/publish', { video_id: videoId, content })
}