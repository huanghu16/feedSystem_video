import { postJson } from './client'
import type { CommentItem } from './types'

/**
 * 评论列表分页响应
 */
export interface ListCommentsResponse {
  list: CommentItem[]
  total: number
  page: number
  size: number
  has_more: boolean
}

/**
 * 获取视频的评论列表（带分页）
 * @param videoId 视频id
 * @param page 页码，默认 1
 * @param size 每页条数，默认 10
 */
export function listComments(videoId: number, page: number = 1, size: number = 20) {
  return postJson<ListCommentsResponse>('/comment/listAll', { video_id: videoId, page, size })
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
