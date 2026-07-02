import { postJson } from './client'
import type { FeedVideoItem } from './types'

/**
 * 最新视频列表分页响应
 */
export interface ListLatestResponse {
  list: FeedVideoItem[]
  total: number
  page: number
  size: number
  has_more: boolean
}

/**
 * 获取最新视频列表（带分页）
 * @param page 页码，默认 1
 * @param size 每页条数，默认 10
 * @return ListLatestResponse 分页响应
 */
export function listLatest(page: number = 1, size: number = 20) {
  return postJson<ListLatestResponse>('/feed/listLatest', { page, size })
}
