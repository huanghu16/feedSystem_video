import { postJson } from './client'
import type { FeedVideoItem } from './types'

/**
 * 获取最新视频列表
 * @return FeedVideoItem[] 最新视频列表
 */
export function listLatest() {   
  return postJson<FeedVideoItem[]>('/feed/listLatest')
} 