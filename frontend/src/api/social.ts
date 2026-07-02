import { postJson } from './client'
import type { VloggerItem } from './types'

/**
 * 关注列表分页响应
 */
export interface GetVloggersResponse {
  list: VloggerItem[]
  total: number
  page: number
  size: number
  has_more: boolean
}

/**
 * 获取关注列表（当前用户关注的人，含资料统计）
 * @param page 页码，默认 1
 * @param size 每页条数，默认 20
 */
export async function getVloggers(page: number = 1, size: number = 50): Promise<GetVloggersResponse> {
    return await postJson('/social/getAllVloggers', { page, page_size: size })
}

/**
 * 检查是否已关注
 * @param vloggerId 用户id
 * @returns 是否已关注
 */
export async function isFollowing(vloggerId: number): Promise<{ is_following: boolean }> {
    return await postJson('/social/isFollowing', { vlogger_id: vloggerId })
}

/**
 * 关注用户
 * @param vloggerId 要关注的用户id
 * @returns 关注结果
 */
export async function follow(vloggerId: number): Promise<{ message: string }> {
    return await postJson('/social/follow', { vlogger_id: vloggerId })
}

/**
 * 取消关注
 * @param vloggerId 要取消关注的用户id
 * @returns 取消关注结果
 */
export async function unfollow(vloggerId: number): Promise<{ message: string }> {
    return await postJson('/social/unfollow', { vlogger_id: vloggerId })
}
