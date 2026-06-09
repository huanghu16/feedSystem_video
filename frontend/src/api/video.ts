import { postJson, postForm } from './client'
import type { VideoItem } from './types'

/**
 * 上传视频
 * @param file
 * @return play_url 视频播放地址
 */
export function uploadVideo(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return postForm<{ play_url: string }>('/video/uploadVideo', formData)
}

/** 
 * 发布视频
 * @param title 视频标题
 * @param playUrl 视频播放地址
 * @param coverUrl 视频封面地址
 * @return VideoItem 发布成功后返回的视频信息 
*/
export function publishVideo(title: string, playUrl: string, coverUrl: string) {
  return postJson<VideoItem>('/video/publish', { title, play_url: playUrl, cover_url: coverUrl })
}

/**
 * 获取作者的视频列表
 * @param authorId 作者id
 * @return VideoItem[] 视频列表
 */
export function listByAuthor(authorId: number) {
  return postJson<VideoItem[]>('/video/listByAuthorID', { author_id: authorId })
}

/**
 * 记录视频播放
 * @param videoId 视频id
 */
export function recordPlay(videoId: number) {
  return postJson<void>('/video/recordPlay', { video_id: videoId })
}

/**
 * 获取热门视频列表（按播放量排序）
 * @param limit 返回数量，默认10
 */
export function listHotVideos(limit: number = 10) {
  return postJson<VideoItem[]>('/video/listHot', { limit })
}

/**
 * 搜索视频
 * @param keyword 搜索关键词
 * @param page 页码，默认1
 * @param size 每页数量，默认10
 */
export interface SearchVideosResponse {
  list: VideoItem[]
  total: number
  page: number
  size: number
  has_more: boolean
}

export function searchVideos(keyword: string, page: number = 1, size: number = 10) {
  return postJson<SearchVideosResponse>('/video/search', { 
    keyword, 
    page, 
    size 
  })
}
