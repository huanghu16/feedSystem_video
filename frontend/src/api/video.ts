import { postJson, postForm, getJson } from './client'
import type { VideoItem } from './types'

/**
 * 上传视频
 * @param file
 * @return play_url 视频播放地址, cover_url 自动生成的封面地址
 */
export function uploadVideo(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return postForm<{ play_url: string; cover_url?: string }>('/video/uploadVideo', formData)
}

/**
 * 上传封面图片
 * @param file 封面图片文件
 * @return url 封面地址
 */
export function uploadCover(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return postForm<{ url: string }>('/video/uploadCover', formData)
}

/**
 * 发布视频
 */
export function publishVideo(title: string, playUrl: string, coverUrl: string, description?: string, publishDate?: string) {
  return postJson<VideoItem>('/video/publish', {
    title,
    play_url: playUrl,
    cover_url: coverUrl,
    description: description || '',
    publish_date: publishDate || ''
  })
}

/**
 * 获取作者的视频列表（带分页）
 */
export interface ListByAuthorResponse {
  list: VideoItem[]
  total: number
  page: number
  size: number
  has_more: boolean
}

export function listByAuthor(authorId: number, page: number = 1, size: number = 12) {
  return getJson<ListByAuthorResponse>('/video/listByAuthorID', { author_id: authorId, page, size })
}

/**
 * 获取视频详情
 */
export function getDetail(videoId: number) {
  return getJson<VideoItem>('/video/getDetail', { id: videoId })
}

/**
 * 记录视频播放
 */
export function recordPlay(videoId: number) {
  return postJson<void>('/video/recordPlay', { video_id: videoId })
}

/**
 * 获取热门视频列表（按播放量排序）
 */
export function listHotVideos(limit: number = 10) {
  return getJson<VideoItem[]>('/video/listHot', { limit })
}

/**
 * 搜索视频（带分页）
 */
export interface SearchVideosResponse {
  list: VideoItem[]
  total: number
  page: number
  size: number
  has_more: boolean
}

export function searchVideos(keyword: string, page: number = 1, size: number = 10) {
  return getJson<SearchVideosResponse>('/video/search', { keyword, page, size })
}

/**
 * 删除视频
 */
export function deleteVideo(videoId: number) {
  return postJson<void>('/video/delete', { video_id: videoId })
}

/**
 * 批量删除视频
 */
export function deleteVideosBatch(videoIds: number[]) {
  return postJson<void>('/video/deleteBatch', { video_ids: videoIds })
}
